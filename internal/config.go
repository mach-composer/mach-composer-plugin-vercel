package internal

import (
	"github.com/google/go-cmp/cmp"
	"github.com/mach-composer/mach-composer-plugin-helpers/helpers"
)

type VercelConfig struct {
	TeamID        string        `mapstructure:"team_id"`
	APIToken      string        `mapstructure:"api_token"`
	ProjectConfig ProjectConfig `mapstructure:"project_config"`
}

func (c *VercelConfig) extendConfig(o *VercelConfig) *VercelConfig {
	if o != nil && o != (&VercelConfig{}) {
		cfg := &VercelConfig{
			TeamID:        o.TeamID,
			APIToken:      o.APIToken,
			ProjectConfig: o.ProjectConfig,
		}

		if c.TeamID != "" {
			cfg.TeamID = c.TeamID
		}
		if c.APIToken != "" {
			cfg.APIToken = c.APIToken
		}
		if !cmp.Equal(c.ProjectConfig, ProjectConfig{}) {
			cfg.ProjectConfig = c.ProjectConfig
		}
		return cfg
	}

	return c
}

type ProjectConfig struct {
	ManualProductionDeployment bool                         `mapstructure:"manual_production_deployment"`
	EnvironmentVariables       []ProjectEnvironmentVariable `mapstructure:"environment_variables"`
}

type ProjectEnvironmentVariable struct {
	Key         string   `mapstructure:"key"`
	Value       string   `mapstructure:"value"`
	Environment []string `mapstructure:"environment"`
}

func (c *ProjectEnvironmentVariable) fillDefaultEnvironments() *ProjectEnvironmentVariable {
	if len(c.Environment) == 0 {
		c.Environment = []string{"development", "preview", "production"}
	}

	return c
}

// Returns a HCL-friendly version of the list of environments which are encapsulated by
// quotes and are comma separated
func (c *ProjectEnvironmentVariable) DisplayEnvironments() string {
	hcl := helpers.SerializeToHCL("environment", c.Environment)

	return hcl
}
