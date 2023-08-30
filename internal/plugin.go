package internal

import (
	"fmt"

	"github.com/mach-composer/mach-composer-plugin-helpers/helpers"
	"github.com/mach-composer/mach-composer-plugin-sdk/plugin"
	"github.com/mach-composer/mach-composer-plugin-sdk/schema"
	"github.com/mitchellh/mapstructure"
)

type VercelPlugin struct {
	environment          string
	provider             string
	globalConfig         *VercelConfig
	siteConfigs          map[string]*VercelConfig
	siteComponentConfigs map[string]map[string]*VercelConfig
	enabled              bool
}

func NewVercelPlugin() schema.MachComposerPlugin {
	state := &VercelPlugin{
		provider:    "0.15.1", // Provider version of `vercel/vercel`
		siteConfigs: map[string]*VercelConfig{},
	}
	return plugin.NewPlugin(&schema.PluginSchema{
		Identifier:          "vercel",
		Configure:           state.Configure,
		IsEnabled:           func() bool { return state.enabled },
		GetValidationSchema: state.GetValidationSchema,

		SetGlobalConfig:        state.SetGlobalConfig,
		SetSiteConfig:          state.SetSiteConfig,
		SetSiteComponentConfig: state.SetSiteComponentConfig,

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

func (p *VercelPlugin) SetGlobalConfig(data map[string]any) error {
	cfg := VercelConfig{}

	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}
	p.globalConfig = &cfg
	p.enabled = true

	return nil
}

func (p *VercelPlugin) GetValidationSchema() (*schema.ValidationSchema, error) {
	result := getSchema()
	return result, nil
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

// Set config for a combination of site and component.
func (p *VercelPlugin) SetSiteComponentConfig(site string, component string, data map[string]any) error {
	cfg := VercelConfig{}
	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}
	if p.siteComponentConfigs == nil {
		p.siteComponentConfigs = make(map[string]map[string]*VercelConfig)
	}
	if p.siteComponentConfigs[site] == nil {
		p.siteComponentConfigs[site] = make(map[string]*VercelConfig)
	}

	p.siteComponentConfigs[site][component] = &cfg
	p.enabled = true
	return nil
}

func (p *VercelPlugin) RenderTerraformStateBackend(site string) (string, error) {
	return "", nil
}

func (p *VercelPlugin) RenderTerraformProviders(site string) (string, error) {
	cfg := p.getConfig(site, "")

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

func (p *VercelPlugin) getComponentConfig(site string, component string) (*VercelConfig, error) {
	cfg, ok := p.siteComponentConfigs[site][component]
	if !ok {
		return nil, fmt.Errorf("No config found for site %s and component %s", site, component)
	}
	siteCfg, err := p.getSiteConfig(site)
	if err == nil {
		cfg = cfg.extendConfig(siteCfg)
	}

	return cfg, nil
}

func (p *VercelPlugin) getSiteConfig(site string) (*VercelConfig, error) {
	cfg, ok := p.siteConfigs[site]
	if !ok {
		return nil, fmt.Errorf("No config found for site %s", site)
	}
	cfg = cfg.extendConfig(p.globalConfig)

	return cfg, nil
}

func (p *VercelPlugin) getConfig(site string, component string) *VercelConfig {

	cfg, err := p.getComponentConfig(site, component)
	if err != nil {
		cfg, err = p.getSiteConfig(site)
		if err != nil {
			cfg = p.globalConfig
		}
	}

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
	cfg := p.getConfig(site, "")

	resourceTemplate := `
		provider "vercel" {
			{{ renderProperty "api_token" .APIToken }}
			{{ renderProperty "team" .TeamID }}
		}
	`

	return helpers.RenderGoTemplate(resourceTemplate, cfg)
}

func (p *VercelPlugin) RenderTerraformComponent(site string, component string) (*schema.ComponentSchema, error) {
	cfg := p.getConfig(site, component)
	if cfg == nil {
		return nil, nil
	}

	template := `
		{{ renderProperty "vercel_team_id" .TeamID }}
		{{ renderProperty "vercel_project_name" .ProjectConfig.Name }}
		{{ renderProperty "vercel_project_framework" .ProjectConfig.Framework }}
		{{ renderProperty "vercel_project_build_command" .ProjectConfig.BuildCommand }}
		{{ renderProperty "vercel_project_root_directory" .ProjectConfig.RootDirectory }}
		{{ renderProperty "vercel_project_serverless_function_region" .ProjectConfig.ServerlessFunctionRegion }}
		{{ renderProperty "vercel_project_manual_production_deployment" .ProjectConfig.ManualProductionDeployment }}
		vercel_project_git_repository = {
			{{ renderProperty "type" .ProjectConfig.GitRepository.Type }}
			{{ renderProperty "repo" .ProjectConfig.GitRepository.Repo }}
		}
		vercel_project_environment_variables = [{{range .ProjectConfig.EnvironmentVariables }}
			{
				{{ renderProperty "key" .Key }}
				{{ renderProperty "value" .Value }}
				{{ .DisplayEnvironments }}
			},{{end}}
		]
		vercel_project_domains = [{{range .ProjectConfig.ProjectDomains }}
			{
				{{ renderProperty "domain" .Domain }}
				{{ renderProperty "redirect_status_code" .RedirectStatusCode }}
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
