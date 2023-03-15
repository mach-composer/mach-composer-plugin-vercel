package internal

import (
	"fmt"

	"github.com/mach-composer/mach-composer-plugin-helpers/helpers"
	"github.com/mach-composer/mach-composer-plugin-sdk/plugin"
	"github.com/mach-composer/mach-composer-plugin-sdk/schema"
	"github.com/mitchellh/mapstructure"
)

type VercelPlugin struct {
	environment  string
	provider     string
	globalConfig *VercelConfig
	siteConfigs  map[string]*VercelConfig
	enabled      bool
}

func NewVercelPlugin() schema.MachComposerPlugin {
	state := &VercelPlugin{
		provider:    "0.6.2", // Provider version of `vercel/vercel`
		siteConfigs: map[string]*VercelConfig{},
	}
	return plugin.NewPlugin(&schema.PluginSchema{
		Identifier: "vercel",
		Configure:  state.Configure,
		IsEnabled:  state.IsEnabled,

		SetGlobalConfig: state.SetGlobalConfig,
		SetSiteConfig:   state.SetSiteConfig,

		// Renders
		RenderTerraformProviders: state.RenderTerraformProviders,
		RenderTerraformResources: state.RenderTerraformResources,
		RenderTerraformComponent: state.RenderTerraformComponent,
	})
}

func (p *VercelPlugin) Configure(environment string, provider string) error {
	p.environment = environment
	if provider != "" {
		p.provider = provider
	}
	return nil
}

func (p *VercelPlugin) IsEnabled() bool {
	return p.enabled
}

func (p *VercelPlugin) SetGlobalConfig(data map[string]any) error {
	cfg := VercelConfig{}

	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}
	p.globalConfig = &cfg
	p.enabled = true

	return nil
}

func (p *VercelPlugin) SetSiteConfig(site string, data map[string]any) error {
	cfg := VercelConfig{}
	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}
	p.siteConfigs[site] = &cfg
	p.enabled = true
	return nil
}

func (p *VercelPlugin) RenderTerraformStateBackend(site string) (string, error) {
	return "", nil
}

func (p *VercelPlugin) RenderTerraformProviders(site string) (string, error) {
	cfg := p.getSiteConfig(site)

	if cfg == nil {
		return "", nil
	}

	result := fmt.Sprintf(`
		vercel = {
			source = "vercel/vercel"
			version = "%s"
		}
	`, helpers.VersionConstraint(p.provider))

	return result, nil
}

func (p *VercelPlugin) getSiteConfig(site string) *VercelConfig {
	cfg, ok := p.siteConfigs[site]
	if !ok {
		cfg = &VercelConfig{}
	}

	cfg = cfg.extendConfig(p.globalConfig)

	// Default behavior for Vercel is to output to all environments
	// Set this as default field unless manually filled
	for i := range cfg.ProjectConfig.EnvironmentVariables {
		if len(cfg.ProjectConfig.EnvironmentVariables[i].Environment) == 0 {
			cfg.ProjectConfig.EnvironmentVariables[i].Environment = []string{"development", "preview", "production"}
		}
	}

	return cfg
}

func (p *VercelPlugin) RenderTerraformResources(site string) (string, error) {
	cfg := p.getSiteConfig(site)

	resourceTemplate := `
		provider "vercel" {
			api_token = {{ .APIToken|printf "%q" }}
		}
	`

	return helpers.RenderGoTemplate(resourceTemplate, cfg)
}

func (p *VercelPlugin) RenderTerraformComponent(site string, component string) (*schema.ComponentSchema, error) {
	cfg := p.getSiteConfig(site)
	if cfg == nil {
		return nil, nil
	}

	template := `
		vercel_team_id = {{ .TeamID|printf "%q" }}
		name = {{ .ProjectConfig.Name|printf "%q" }}
		framework = {{ .ProjectConfig.Framework|printf "%q" }}
		build_command = {{ .ProjectConfig.BuildCommand|printf "%q" }}
		root_directory = {{ .ProjectConfig.RootDirectory|printf "%q" }}
		serverless_function_region = {{ .ProjectConfig.ServerlessFunctionRegion|printf "%q" }}
		manual_production_deployment = {{ .ProjectConfig.ManualProductionDeployment }}
		environment_variables = [{{range .ProjectConfig.EnvironmentVariables }}
			{
				key = {{ .Key|printf "%q" }}
				value = {{ .Value|printf "%q" }}
				{{ .DisplayEnvironments }}
			},{{end}}
		]
	`

	vars, err := helpers.RenderGoTemplate(template, cfg)
	if err != nil {
		return nil, err
	}

	result := &schema.ComponentSchema{
		Variables: vars,
	}

	return result, nil
}
