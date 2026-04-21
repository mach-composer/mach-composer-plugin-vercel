package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Interface interface{}

func TestSetVercelConfig(t *testing.T) {
	// All of the below env variables code is used to bypass gojsonschema's
	// inability to cast this to a []map[string]interface{}.
	environmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_ENVIRONMENT_VARIABLE", Value: "testing", Environment: []string{}},
		{Key: "TEST_ENVIRONMENT_VARIABLE_2", Value: "testing", Environment: []string{"production", "preview"}},
		{Key: "TEST_ENVIRONMENT_VARIABLE_3", Value: "secret", Environment: []string{"production"}, Sensitive: true, Comment: "A secret variable", GitBranch: "main", Target: []string{"production"}, CustomEnvironmentIDs: []string{"env_123"}},
	}

	projectDomains := []ProjectDomain{
		{Domain: "test-domain.com", RedirectStatusCode: 307},
	}

	domains := make([]interface{}, len(projectDomains))
	for i, s := range projectDomains {
		domains[i] = s
	}

	variables := make([]interface{}, len(environmentVariables))
	for i, s := range environmentVariables {
		variables[i] = s
	}

	data := map[string]any{
		"team_id":   "test-team",
		"api_token": "${sops.data.output[\"api_token\"]}",
		"project_config": map[string]any{
			"name":                         "test-project",
			"framework":                    "nextjs",
			"serverless_function_region":   "iad1",
			"build_command":                "next build",
			"ignore_command":               "if [ $VERCEL_ENV == 'production' ]; then exit 1; else exit 0; fi",
			"root_directory":               "./my-project",
			"node_version":                 "24.x",
			"manual_production_deployment": true,
			"git_repository": map[string]any{
				"production_branch": "main",
				"type":              "github",
				"repo":              "mach-composer/my-project",
			},
			"environment_variables":            variables,
			"domains":                          domains,
			"protection_bypass_for_automation": true,
			"vercel_authentication": map[string]any{
				"deployment_type": "only_preview_deployments",
			},
			"password_protection": map[string]any{
				"password":        "MyPassword",
				"deployment_type": "only_preview_deployments",
			},
		},
	}

	plugin := NewVercelPlugin()

	err := plugin.SetComponentConfig("my-component", map[string]any{
		"integrations": []string{"vercel"},
	})
	require.NoError(t, err)

	err = plugin.SetSiteConfig("my-site", data)
	require.NoError(t, err)

	result, err := plugin.RenderTerraformResources("my-site")
	require.NoError(t, err)
	assert.Contains(t, result, `api_token = sops.data.output["api_token"]`)

	component, err := plugin.RenderTerraformComponent("my-site", "test-component")
	require.NoError(t, err)
	assert.Contains(t, component.Variables, "name = \"test-project\"")
	assert.Contains(t, component.Variables, "framework = \"nextjs\"")
	assert.Contains(t, component.Variables, "serverless_function_region = \"iad1\"")
	assert.Contains(t, component.Variables, "build_command = \"next build\"")
	assert.Contains(t, component.Variables, "ignore_command = \"if [ $VERCEL_ENV == 'production' ]; then exit 1; else exit 0; fi\"")
	assert.Contains(t, component.Variables, "root_directory = \"./my-project\"")
	assert.Contains(t, component.Variables, "node_version = \"24.x\"")
	assert.Contains(t, component.Variables, "vercel_team_id = \"test-team\"")
	assert.Contains(t, component.Variables, "manual_production_deployment = true")
	assert.Contains(t, component.Variables, "production_branch = \"main\"")
	assert.Contains(t, component.Variables, "type = \"github\"")
	assert.Contains(t, component.Variables, "repo = \"mach-composer/my-project\"")
	assert.Contains(t, component.Variables, "protection_bypass_for_automation = true")
	assert.Contains(t, component.Variables, "deployment_type = \"only_preview_deployments\"")
	assert.Contains(t, component.Variables, "password = \"MyPassword\"")

	// Test environment variables

	// Test default response
	assert.Contains(t, component.Variables, "environment = [\"development\", \"preview\", \"production\"]")
	// Test custom environment variables list
	assert.Contains(t, component.Variables, "environment = [\"production\", \"preview\"]")

	// Test new environment variable fields
	assert.Contains(t, component.Variables, "sensitive = true")
	assert.Contains(t, component.Variables, "comment = \"A secret variable\"")
	assert.Contains(t, component.Variables, "git_branch = \"main\"")
	assert.Contains(t, component.Variables, "target = [\"production\"]")
	assert.Contains(t, component.Variables, "custom_environment_ids = [\"env_123\"]")

	// Test domains
	assert.Contains(t, component.Variables, "domain = \"test-domain.com\"")
	assert.Contains(t, component.Variables, "redirect_status_code = 307")

}

func TestInheritance(t *testing.T) {
	globalEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_ENVIRONMENT_VARIABLE", Value: "testing", Environment: []string{}},
		{Key: "TEST_EXTEND_VARIABLE", Value: "test", Environment: []string{"production"}},
	}
	globalVariables := make([]interface{}, len(globalEnvironmentVariables))
	for i, s := range globalEnvironmentVariables {
		globalVariables[i] = s
	}
	globalData := map[string]any{
		"team_id":   "test-team",
		"api_token": "test-token",
		"project_config": map[string]any{
			"manual_production_deployment":     true,
			"protection_bypass_for_automation": true,
			"vercel_authentication": map[string]any{
				"deployment_type": "standard_protection",
			},
			"password_protection": map[string]any{
				"password":        "MyPassword",
				"deployment_type": "standard_protection",
			},
			"environment_variables": globalVariables,
		},
	}

	siteEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_ENVIRONMENT_VARIABLE_2", Value: "testing", Environment: []string{"production", "preview"}},
		{Key: "TEST_EXTEND_VARIABLE", Value: "testing", Environment: []string{"production", "preview", "development"}},
	}
	siteVariables := make([]interface{}, len(siteEnvironmentVariables))
	for i, s := range siteEnvironmentVariables {
		siteVariables[i] = s
	}

	siteData := map[string]any{
		"team_id":   "test-team-override",
		"api_token": "test-token-override",
		"project_config": map[string]any{
			"vercel_authentication": map[string]any{
				"deployment_type": "standard_protection",
			},
			"password_protection": map[string]any{
				"password":        "MyPassword",
				"deployment_type": "standard_protection",
			},
			"environment_variables": siteVariables,
		},
	}

	plugin := NewVercelPlugin()

	err := plugin.SetGlobalConfig(globalData)
	require.NoError(t, err)

	err = plugin.SetSiteConfig("my-site", siteData)
	require.NoError(t, err)

	result, err := plugin.RenderTerraformResources("my-site")
	require.NoError(t, err)
	assert.Contains(t, result, "api_token = \"test-token-override\"")

	component, err := plugin.RenderTerraformComponent("my-site", "test-component")
	require.NoError(t, err)

	// Test overriding fields
	assert.Contains(t, component.Variables, "vercel_team_id = \"test-team-override\"")

	// Test whether environment variables get extended
	assert.Contains(t, component.Variables, "environment = [\"development\", \"preview\", \"production\"]")
	assert.Contains(t, component.Variables, "environment = [\"preview\", \"production\"]")
	assert.Contains(t, component.Variables, "deployment_type = \"standard_protection\"")

	assert.Contains(t, component.Variables, "environment")
}

func TestSiteComponentInheritance(t *testing.T) {
	siteEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_ENVIRONMENT_VARIABLE_2", Value: "testing", Environment: []string{"production", "preview"}},
		{Key: "TEST_EXTEND_VARIABLE", Value: "testing", Environment: []string{"production", "preview", "development"}},
	}
	siteVariables := make([]interface{}, len(siteEnvironmentVariables))
	for i, s := range siteEnvironmentVariables {
		siteVariables[i] = s
	}

	siteData := map[string]any{
		"team_id":   "test-team-override",
		"api_token": "test-token-override",
		"project_config": map[string]any{
			"serverless_function_region":   "iad1",
			"manual_production_deployment": false,
			"environment_variables":        siteVariables,
			"git_repository": map[string]any{
				"production_branch": "main",
				"type":              "github",
				"repo":              "mach-composer/my-project",
			},
		},
	}

	componentEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_ENVIRONMENT_VARIABLE_3", Value: "testing"},
	}

	componentVariables := make([]interface{}, len(componentEnvironmentVariables))
	for i, s := range componentEnvironmentVariables {
		componentVariables[i] = s
	}

	componentData := map[string]any{
		"project_config": map[string]any{
			"serverless_function_region":   "fra1",
			"manual_production_deployment": true,
			"environment_variables":        componentVariables,
			"git_repository": map[string]any{
				"production_branch": "production",
			},
		},
	}

	plugin := NewVercelPlugin()

	err := plugin.SetSiteConfig("my-site", siteData)
	require.NoError(t, err)

	err = plugin.SetSiteComponentConfig("my-site", "test-component", componentData)
	require.NoError(t, err)
	component, err := plugin.RenderTerraformComponent("my-site", "test-component")
	require.NoError(t, err)

	// Test whether environment variables get extended
	assert.Contains(t, component.Variables, "vercel_project_serverless_function_region = \"fra1\"")
	assert.Contains(t, component.Variables, "vercel_project_manual_production_deployment = true")
	assert.Contains(t, component.Variables, "key = \"TEST_ENVIRONMENT_VARIABLE_2\"")
	assert.Contains(t, component.Variables, "key = \"TEST_ENVIRONMENT_VARIABLE_3\"")
	assert.Contains(t, component.Variables, "production_branch = \"production\"")
}

func TestGlobalInheritance(t *testing.T) {
	globalEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_ENVIRONMENT_VARIABLE", Value: "testing", Environment: []string{}},
		{Key: "TEST_EXTEND_VARIABLE", Value: "test", Environment: []string{"production"}},
	}
	globalVariables := make([]interface{}, len(globalEnvironmentVariables))
	for i, s := range globalEnvironmentVariables {
		globalVariables[i] = s
	}
	globalData := map[string]any{
		"team_id":   "test-team",
		"api_token": "test-token",
		"project_config": map[string]any{
			"manual_production_deployment":     true,
			"protection_bypass_for_automation": true,
			"vercel_authentication": map[string]any{
				"deployment_type": "all_deployments",
			},
			"password_protection": map[string]any{
				"password":        "MyPassword",
				"deployment_type": "all_deployments",
			},
			"environment_variables": globalVariables,
		},
	}

	siteEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_ENVIRONMENT_VARIABLE_2", Value: "testing", Environment: []string{"production", "preview"}},
		{Key: "TEST_EXTEND_VARIABLE", Value: "testing", Environment: []string{"production", "preview", "development"}},
	}
	siteVariables := make([]interface{}, len(siteEnvironmentVariables))
	for i, s := range siteEnvironmentVariables {
		siteVariables[i] = s
	}

	siteData := map[string]any{
		"team_id":   "test-team-override",
		"api_token": "test-token-override",
		"project_config": map[string]any{
			"environment_variables": siteVariables,
		},
	}

	plugin := NewVercelPlugin()

	err := plugin.SetGlobalConfig(globalData)
	require.NoError(t, err)

	err = plugin.SetSiteConfig("my-site", siteData)
	require.NoError(t, err)

	result, err := plugin.RenderTerraformResources("my-site")
	require.NoError(t, err)
	assert.Contains(t, result, "api_token = \"test-token-override\"")

	component, err := plugin.RenderTerraformComponent("my-site", "test-component")
	require.NoError(t, err)

	// Test overriding fields
	assert.Contains(t, component.Variables, "vercel_team_id = \"test-team-override\"")

	// Test whether environment variables get extended
	assert.Contains(t, component.Variables, "environment = [\"development\", \"preview\", \"production\"]")
	assert.Contains(t, component.Variables, "environment = [\"preview\", \"production\"]")
	assert.Contains(t, component.Variables, "deployment_type = \"all_deployments\"")

	assert.Contains(t, component.Variables, "environment")
}

func TestSitePriorityOverGlobalInheritance(t *testing.T) {
	globalEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_ENVIRONMENT_VARIABLE", Value: "testing", Environment: []string{}},
		{Key: "TEST_EXTEND_VARIABLE", Value: "test", Environment: []string{"production"}},
	}
	globalVariables := make([]interface{}, len(globalEnvironmentVariables))
	for i, s := range globalEnvironmentVariables {
		globalVariables[i] = s
	}
	globalData := map[string]any{
		"team_id":   "test-team",
		"api_token": "test-token",
		"project_config": map[string]any{
			"manual_production_deployment":     true,
			"protection_bypass_for_automation": true,
			"vercel_authentication": map[string]any{
				"deployment_type": "standard_protection",
			},
			"password_protection": map[string]any{
				"password":        "MyPassword",
				"deployment_type": "standard_protection",
			},
			"environment_variables": globalVariables,
		},
	}

	siteEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_ENVIRONMENT_VARIABLE_2", Value: "testing", Environment: []string{"production", "preview"}},
		{Key: "TEST_EXTEND_VARIABLE", Value: "testing", Environment: []string{"production", "preview", "development"}},
	}
	siteVariables := make([]interface{}, len(siteEnvironmentVariables))
	for i, s := range siteEnvironmentVariables {
		siteVariables[i] = s
	}

	siteData := map[string]any{
		"team_id":   "test-team-override",
		"api_token": "test-token-override",
		"project_config": map[string]any{
			"environment_variables": siteVariables,
			"vercel_authentication": map[string]any{
				"deployment_type": "only_preview_deployments",
			},
		},
	}

	plugin := NewVercelPlugin()

	err := plugin.SetGlobalConfig(globalData)
	require.NoError(t, err)

	err = plugin.SetSiteConfig("my-site", siteData)
	require.NoError(t, err)

	result, err := plugin.RenderTerraformResources("my-site")
	require.NoError(t, err)
	assert.Contains(t, result, "api_token = \"test-token-override\"")

	component, err := plugin.RenderTerraformComponent("my-site", "test-component")
	require.NoError(t, err)

	// Test overriding fields
	assert.Contains(t, component.Variables, "vercel_team_id = \"test-team-override\"")

	// Test whether environment variables get extended
	assert.Contains(t, component.Variables, "environment = [\"development\", \"preview\", \"production\"]")
	assert.Contains(t, component.Variables, "environment = [\"preview\", \"production\"]")
	assert.Contains(t, component.Variables, "deployment_type = \"standard_protection\"")
	assert.Contains(t, component.Variables, "deployment_type = \"only_preview_deployments\"")

	assert.Contains(t, component.Variables, "environment")
}

func TestExtendEnvironmentVariables(t *testing.T) {
	globalEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_EXTEND_VARIABLE", Value: "test", Environment: []string{"production"}},
	}
	globalVariables := make([]interface{}, len(globalEnvironmentVariables))
	for i, s := range globalEnvironmentVariables {
		globalVariables[i] = s
	}
	globalData := map[string]any{
		"team_id":   "test-team",
		"api_token": "test-token",
		"project_config": map[string]any{
			"manual_production_deployment": true,
			"environment_variables":        globalVariables,
		},
	}

	siteEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_EXTEND_VARIABLE", Value: "testing", Environment: []string{"acceptance"}},
	}
	siteVariables := make([]interface{}, len(siteEnvironmentVariables))
	for i, s := range siteEnvironmentVariables {
		siteVariables[i] = s
	}

	siteData := map[string]any{
		"team_id":   "test-team",
		"api_token": "test-token",
		"project_config": map[string]any{
			"manual_production_deployment": true,
			"environment_variables":        siteVariables,
		},
	}

	plugin := NewVercelPlugin()

	err := plugin.SetGlobalConfig(globalData)
	require.NoError(t, err)

	err = plugin.SetSiteConfig("my-site", siteData)
	require.NoError(t, err)

	// Test whether environment variables get extended
	component, err := plugin.RenderTerraformComponent("my-site", "test-component")
	require.NoError(t, err)

	// Should only contain the site extended variable content
	assert.Contains(t, component.Variables, "key = \"TEST_EXTEND_VARIABLE\"")
	assert.Contains(t, component.Variables, "value = \"test\"")
	assert.Contains(t, component.Variables, "environment = [\"production\"]")
	assert.Contains(t, component.Variables, "value = \"testing\"")
	assert.Contains(t, component.Variables, "environment = [\"acceptance\"]")
}

func TestMergeEnvironmentVariables(t *testing.T) {
	globalEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_EXTEND_VARIABLE", Value: "test", Environment: []string{"production"}},
	}
	globalVariables := make([]interface{}, len(globalEnvironmentVariables))
	for i, s := range globalEnvironmentVariables {
		globalVariables[i] = s
	}
	globalData := map[string]any{
		"team_id":   "test-team",
		"api_token": "test-token",
		"project_config": map[string]any{
			"manual_production_deployment": true,
			"environment_variables":        globalVariables,
		},
	}

	siteEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_EXTEND_VARIABLE", Value: "test", Environment: []string{"preview", "development"}},
	}
	siteVariables := make([]interface{}, len(siteEnvironmentVariables))
	for i, s := range siteEnvironmentVariables {
		siteVariables[i] = s
	}

	siteData := map[string]any{
		"team_id":   "test-team",
		"api_token": "test-token",
		"project_config": map[string]any{
			"manual_production_deployment": true,
			"environment_variables":        siteVariables,
		},
	}

	plugin := NewVercelPlugin()

	err := plugin.SetGlobalConfig(globalData)
	require.NoError(t, err)

	err = plugin.SetSiteConfig("my-site", siteData)
	require.NoError(t, err)

	// Test whether environment variables get extended
	component, err := plugin.RenderTerraformComponent("my-site", "test-component")
	require.NoError(t, err)

	assert.Contains(t, component.Variables, "key = \"TEST_EXTEND_VARIABLE\"")
	assert.Contains(t, component.Variables, "value = \"test\"")
	assert.Contains(t, component.Variables, "environment = [\"development\", \"preview\", \"production\"]")

	assert.NotContains(t, component.Variables, "environment = [\"development\", \"preview\"]")
	// After merge, individual environment = ["production"] should not exist since they are merged
	assert.NotContains(t, component.Variables, "\nenvironment = [\"production\"]\n")
}

func TestUpsertEnvironmentVariables(t *testing.T) {
	globalEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_EXTEND_VARIABLE", Value: "test", Environment: []string{"production"}},
	}
	globalVariables := make([]interface{}, len(globalEnvironmentVariables))
	for i, s := range globalEnvironmentVariables {
		globalVariables[i] = s
	}
	globalData := map[string]any{
		"team_id":   "test-team",
		"api_token": "test-token",
		"project_config": map[string]any{
			"manual_production_deployment": true,
			"environment_variables":        globalVariables,
		},
	}

	siteEnvironmentVariables := []ProjectEnvironmentVariable{
		{Key: "TEST_EXTEND_VARIABLE", Value: "testing", Environment: []string{"production"}},
	}
	siteVariables := make([]interface{}, len(siteEnvironmentVariables))
	for i, s := range siteEnvironmentVariables {
		siteVariables[i] = s
	}

	siteData := map[string]any{
		"team_id":   "test-team",
		"api_token": "test-token",
		"project_config": map[string]any{
			"manual_production_deployment": true,
			"environment_variables":        siteVariables,
		},
	}

	plugin := NewVercelPlugin()

	err := plugin.SetGlobalConfig(globalData)
	require.NoError(t, err)

	err = plugin.SetSiteConfig("my-site", siteData)
	require.NoError(t, err)

	// Test whether environment variables get extended
	component, err := plugin.RenderTerraformComponent("my-site", "test-component")
	require.NoError(t, err)

	assert.NotContains(t, component.Variables, "value = \"testing\"")

	assert.Contains(t, component.Variables, "key = \"TEST_EXTEND_VARIABLE\"")
	assert.Contains(t, component.Variables, "value = \"test\"")
	assert.Contains(t, component.Variables, "environment = [\"production\"]")
}

func TestCompleteInheritance(t *testing.T) {
	global := map[string]any{
		"team_id": "test-team",
		"project_config": map[string]any{
			"serverless_function_region": "fra1",
		},
	}

	plugin := NewVercelPlugin()

	err := plugin.SetGlobalConfig(global)
	require.NoError(t, err)

	siteConfig := map[string]any{
		"project_config": map[string]any{
			"git_repository": map[string]any{
				"production_branch": "production",
				"type":              "github",
				"repo":              "owner/test-repo",
			},
		},
	}

	err = plugin.SetSiteConfig("my-site", siteConfig)
	require.NoError(t, err)

	componentConfig := map[string]any{
		"project_config": map[string]any{
			"manual_production_deployment": true,
		},
	}

	err = plugin.SetSiteComponentConfig("my-site", "test-component", componentConfig)
	require.NoError(t, err)

	component, err := plugin.RenderTerraformComponent("my-site", "test-component")
	require.NoError(t, err)

	fmt.Println(component)

	assert.Contains(t, component.Variables, "vercel_project_serverless_function_region = \"fra1\"")
	assert.Contains(t, component.Variables, "type = \"github\"")
	assert.Contains(t, component.Variables, "production_branch = \"production\"")
	assert.Contains(t, component.Variables, "vercel_project_manual_production_deployment = true")
}

func TestManualProductionDeploymentBehavior(t *testing.T) {
	t.Run("defaults to false when not specified", func(t *testing.T) {
		plugin := NewVercelPlugin()

		data := map[string]any{
			"team_id":   "test-team",
			"api_token": "test-token",
			"project_config": map[string]any{
				"framework": "nextjs",
			},
		}

		err := plugin.SetSiteConfig("my-site", data)
		require.NoError(t, err)

		component, err := plugin.RenderTerraformComponent("my-site", "test-component")
		require.NoError(t, err)

		assert.Contains(t, component.Variables, "vercel_project_manual_production_deployment = false")
	})

	t.Run("explicit false is preserved", func(t *testing.T) {
		plugin := NewVercelPlugin()

		data := map[string]any{
			"team_id":   "test-team",
			"api_token": "test-token",
			"project_config": map[string]any{
				"framework":                    "nextjs",
				"manual_production_deployment": false,
			},
		}

		err := plugin.SetSiteConfig("my-site", data)
		require.NoError(t, err)

		component, err := plugin.RenderTerraformComponent("my-site", "test-component")
		require.NoError(t, err)

		assert.Contains(t, component.Variables, "vercel_project_manual_production_deployment = false")
	})

	t.Run("inheritance: child overrides parent", func(t *testing.T) {
		plugin := NewVercelPlugin()

		globalData := map[string]any{
			"team_id":   "test-team",
			"api_token": "test-token",
			"project_config": map[string]any{
				"manual_production_deployment": false,
			},
		}

		siteData := map[string]any{
			"project_config": map[string]any{
				"manual_production_deployment": true,
			},
		}

		err := plugin.SetGlobalConfig(globalData)
		require.NoError(t, err)

		err = plugin.SetSiteConfig("my-site", siteData)
		require.NoError(t, err)

		component, err := plugin.RenderTerraformComponent("my-site", "test-component")
		require.NoError(t, err)

		assert.Contains(t, component.Variables, "vercel_project_manual_production_deployment = true")
	})

	t.Run("inheritance: inherits false when not specified", func(t *testing.T) {
		plugin := NewVercelPlugin()

		globalData := map[string]any{
			"team_id":   "test-team",
			"api_token": "test-token",
			"project_config": map[string]any{
				"manual_production_deployment": false,
			},
		}

		siteData := map[string]any{
			"project_config": map[string]any{
				"framework": "nextjs",
			},
		}

		err := plugin.SetGlobalConfig(globalData)
		require.NoError(t, err)

		err = plugin.SetSiteConfig("my-site", siteData)
		require.NoError(t, err)

		component, err := plugin.RenderTerraformComponent("my-site", "test-component")
		require.NoError(t, err)

		assert.Contains(t, component.Variables, "vercel_project_manual_production_deployment = false")
	})

	t.Run("inheritance: inherits true when not specified", func(t *testing.T) {
		plugin := NewVercelPlugin()

		globalData := map[string]any{
			"team_id":   "test-team",
			"api_token": "test-token",
			"project_config": map[string]any{
				"manual_production_deployment": true,
			},
		}

		siteData := map[string]any{
			"project_config": map[string]any{
				"framework": "nextjs",
			},
		}

		err := plugin.SetGlobalConfig(globalData)
		require.NoError(t, err)

		err = plugin.SetSiteConfig("my-site", siteData)
		require.NoError(t, err)

		component, err := plugin.RenderTerraformComponent("my-site", "test-component")
		require.NoError(t, err)

		assert.Contains(t, component.Variables, "vercel_project_manual_production_deployment = true")
	})
}

func TestMergePreservesEnvVarFields(t *testing.T) {
	globalVars := []ProjectEnvironmentVariable{
		{Key: "SECRET", Value: "g", Environment: []string{"production"}, Sensitive: true, Comment: "global", GitBranch: "main", Target: []string{"production"}, CustomEnvironmentIDs: []string{"env_g"}},
	}
	siteVars := []ProjectEnvironmentVariable{
		{Key: "OTHER", Value: "s", Environment: []string{"preview"}, Sensitive: true, Target: []string{"preview"}, Comment: "site-only"},
	}

	toIface := func(in []ProjectEnvironmentVariable) []interface{} {
		out := make([]interface{}, len(in))
		for i, v := range in {
			out[i] = v
		}
		return out
	}

	globalData := map[string]any{
		"team_id":   "t",
		"api_token": "a",
		"project_config": map[string]any{
			"environment_variables": toIface(globalVars),
		},
	}
	siteData := map[string]any{
		"project_config": map[string]any{
			"environment_variables": toIface(siteVars),
		},
	}

	plugin := NewVercelPlugin()
	require.NoError(t, plugin.SetGlobalConfig(globalData))
	require.NoError(t, plugin.SetSiteConfig("my-site", siteData))

	component, err := plugin.RenderTerraformComponent("my-site", "test-component")
	require.NoError(t, err)

	// Global-only entry keeps all its fields through the merge
	assert.Contains(t, component.Variables, "key = \"SECRET\"")
	assert.Contains(t, component.Variables, "comment = \"global\"")
	assert.Contains(t, component.Variables, "git_branch = \"main\"")
	assert.Contains(t, component.Variables, "custom_environment_ids = [\"env_g\"]")

	// Site-only entry keeps its fields too
	assert.Contains(t, component.Variables, "key = \"OTHER\"")
	assert.Contains(t, component.Variables, "comment = \"site-only\"")
	assert.Contains(t, component.Variables, "target = [\"preview\"]")
}

func TestSensitiveTargetValidationAfterMerge(t *testing.T) {
	// Sensitive is declared at global with a safe target; site widens target to
	// include "development" — the merged config must fail validation even though
	// neither level alone violates the rule.
	globalVars := []ProjectEnvironmentVariable{
		{Key: "SECRET", Value: "v", Environment: []string{"production"}, Sensitive: true, Target: []string{"production"}},
	}
	siteVars := []ProjectEnvironmentVariable{
		{Key: "SECRET", Value: "v", Environment: []string{"development"}, Sensitive: true, Target: []string{"development"}},
	}

	toIface := func(in []ProjectEnvironmentVariable) []interface{} {
		out := make([]interface{}, len(in))
		for i, v := range in {
			out[i] = v
		}
		return out
	}

	plugin := NewVercelPlugin()
	require.NoError(t, plugin.SetGlobalConfig(map[string]any{
		"team_id":   "t",
		"api_token": "a",
		"project_config": map[string]any{
			"environment_variables": toIface(globalVars),
		},
	}))
	// Site SetSiteConfig itself should fail because siteVars already violates at its own level
	err := plugin.SetSiteConfig("my-site", map[string]any{
		"project_config": map[string]any{
			"environment_variables": toIface(siteVars),
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SECRET")
}

func TestSensitiveTargetValidation(t *testing.T) {
	envVars := []ProjectEnvironmentVariable{
		{Key: "SECRET", Value: "s", Sensitive: true, Target: []string{"development", "production"}},
	}
	variables := make([]interface{}, len(envVars))
	for i, s := range envVars {
		variables[i] = s
	}

	data := map[string]any{
		"team_id":   "t",
		"api_token": "a",
		"project_config": map[string]any{
			"environment_variables": variables,
		},
	}

	plugin := NewVercelPlugin()
	err := plugin.SetSiteConfig("my-site", data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SECRET")
	assert.Contains(t, err.Error(), "development")
}

func TestNodeVersionInheritance(t *testing.T) {
	t.Run("node_version set at site level", func(t *testing.T) {
		plugin := NewVercelPlugin()

		siteData := map[string]any{
			"team_id":   "test-team",
			"api_token": "test-token",
			"project_config": map[string]any{
				"framework":    "nextjs",
				"node_version": "20.x",
			},
		}

		err := plugin.SetSiteConfig("my-site", siteData)
		require.NoError(t, err)

		component, err := plugin.RenderTerraformComponent("my-site", "test-component")
		require.NoError(t, err)

		assert.Contains(t, component.Variables, "vercel_project_node_version = \"20.x\"")
	})

	t.Run("component overrides site node_version", func(t *testing.T) {
		plugin := NewVercelPlugin()

		siteData := map[string]any{
			"team_id":   "test-team",
			"api_token": "test-token",
			"project_config": map[string]any{
				"framework":    "nextjs",
				"node_version": "18.x",
			},
		}

		componentData := map[string]any{
			"project_config": map[string]any{
				"node_version": "20.x",
			},
		}

		err := plugin.SetSiteConfig("my-site", siteData)
		require.NoError(t, err)

		err = plugin.SetSiteComponentConfig("my-site", "test-component", componentData)
		require.NoError(t, err)

		component, err := plugin.RenderTerraformComponent("my-site", "test-component")
		require.NoError(t, err)

		assert.Contains(t, component.Variables, "vercel_project_node_version = \"20.x\"")
		assert.NotContains(t, component.Variables, "vercel_project_node_version = \"18.x\"")
	})

	t.Run("global to site to component inheritance", func(t *testing.T) {
		plugin := NewVercelPlugin()

		globalData := map[string]any{
			"team_id":   "test-team",
			"api_token": "test-token",
			"project_config": map[string]any{
				"node_version": "16.x",
			},
		}

		siteData := map[string]any{
			"project_config": map[string]any{
				"framework": "nextjs",
			},
		}

		componentData := map[string]any{
			"project_config": map[string]any{
				"node_version": "20.x",
			},
		}

		err := plugin.SetGlobalConfig(globalData)
		require.NoError(t, err)

		err = plugin.SetSiteConfig("my-site", siteData)
		require.NoError(t, err)

		err = plugin.SetSiteComponentConfig("my-site", "test-component", componentData)
		require.NoError(t, err)

		component, err := plugin.RenderTerraformComponent("my-site", "test-component")
		require.NoError(t, err)

		assert.Contains(t, component.Variables, "vercel_project_node_version = \"20.x\"")
		assert.NotContains(t, component.Variables, "vercel_project_node_version = \"16.x\"")
	})

	t.Run("empty node_version is not rendered", func(t *testing.T) {
		plugin := NewVercelPlugin()

		siteData := map[string]any{
			"team_id":   "test-team",
			"api_token": "test-token",
			"project_config": map[string]any{
				"framework": "nextjs",
				// node_version is not specified
			},
		}

		err := plugin.SetSiteConfig("my-site", siteData)
		require.NoError(t, err)

		component, err := plugin.RenderTerraformComponent("my-site", "test-component")
		require.NoError(t, err)

		// Should not contain node_version variable when empty
		assert.Contains(t, component.Variables, "vercel_team_id = \"test-team\"")
		assert.NotContains(t, component.Variables, "vercel_project_node_version")
	})
}
