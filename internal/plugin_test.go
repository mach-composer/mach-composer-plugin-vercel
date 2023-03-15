package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetVercelConfig(t *testing.T) {
	data := map[string]any{
		"team_id":   "test-team",
		"api_token": "test-token",
		"project_config": map[string]any{
			"name":                         "test-project",
			"framework":                    "nextjs",
			"serverless_function_region":   "iad1",
			"build_command":                "next build",
			"root_directory":               "./my-project",
			"manual_production_deployment": true,
			"git_repository": map[string]any{
				"type": "github",
				"repo": "mach-composer/my-project",
			},
			"environment_variables": []map[string]any{
				{"key": "TEST_ENVIRONMENT_VARIABLE", "value": "testing"},
				{"key": "TEST_ENVIRONMENT_VARIABLE_2", "value": "testing", "environment": []string{"production", "preview"}},
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
	assert.Contains(t, result, "api_token = \"test-token\"")

	component, err := plugin.RenderTerraformComponent("my-site", "test-component")
	require.NoError(t, err)
	assert.Contains(t, component.Variables, "name = \"test-project\"")
	assert.Contains(t, component.Variables, "framework = \"nextjs\"")
	assert.Contains(t, component.Variables, "serverless_function_region = \"iad1\"")
	assert.Contains(t, component.Variables, "build_command = \"next build\"")
	assert.Contains(t, component.Variables, "root_directory = \"./my-project\"")
	assert.Contains(t, component.Variables, "vercel_team_id = \"test-team\"")
	assert.Contains(t, component.Variables, "manual_production_deployment = true")
	assert.Contains(t, component.Variables, "type = \"github\"")
	assert.Contains(t, component.Variables, "repo = \"mach-composer/my-project\"")

	// Test environment variables

	// Test default response
	assert.Contains(t, component.Variables, "environment = [\"development\", \"preview\", \"production\"]")
	// Test custom environment variables list
	assert.Contains(t, component.Variables, "environment = [\"production\", \"preview\"]")
}

func TestInheritEnvironmentVariables(t *testing.T) {
	globalData := map[string]any{
		"team_id":   "test-team",
		"api_token": "test-token",
		"project_config": map[string]any{
			"manual_production_deployment": true,
			"environment_variables": []map[string]any{
				{"key": "TEST_ENVIRONMENT_VARIABLE", "value": "testing"},
			},
		},
	}

	siteData := map[string]any{
		"team_id":   "test-team",
		"api_token": "test-token",
		"project_config": map[string]any{
			"manual_production_deployment": true,
			"environment_variables": []map[string]any{
				{"key": "TEST_ENVIRONMENT_VARIABLE_2", "value": "testing", "environment": []string{"production", "preview"}},
			},
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

	assert.Contains(t, component.Variables, "environment = [\"development\", \"preview\", \"production\"]")
	assert.Contains(t, component.Variables, "environment = [\"production\", \"preview\"]")

}
