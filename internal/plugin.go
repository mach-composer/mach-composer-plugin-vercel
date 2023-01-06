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
	fmt.Println(cfg)
	return cfg.extendConfig(p.globalConfig)
}

func (p *VercelPlugin) RenderTerraformResources(site string) (string, error) {
	cfg := p.getSiteConfig(site)

	fmt.Println(cfg)

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
		manual_production_deployment = {{ .ProjectConfig.ManualProductionDeployment }}
		environment_variables = [{{range .ProjectConfig.EnvironmentVariables }}
			{
				name = {{ .Name|printf "%q" }}
				value = {{ .Value|printf "%q" }}
				environment = {{ $length := len .Environment }}{{ if eq $length 0 }}{{ .DefaultEnvironments }}{{ else }}{{ .DisplayEnvironments }}{{ end }}
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
