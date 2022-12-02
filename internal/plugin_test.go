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
			"manual_production_deployment": true,
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
	assert.Contains(t, component.Variables, "vercel_team_id = \"test-team\"")
	assert.Contains(t, component.Variables, "manual_production_deployment = true")

}
