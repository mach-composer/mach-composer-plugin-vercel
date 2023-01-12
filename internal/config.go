package internal

import (
	"github.com/google/go-cmp/cmp"
	"github.com/mach-composer/mach-composer-plugin-helpers/helpers"
	"golang.org/x/exp/slices"
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
			// Update individual fields instead of updating struct
			result := c.ProjectConfig.extendConfig(&o.ProjectConfig)
			if result != nil {
				cfg.ProjectConfig = *result
			}
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

// Returns a HCL-friendly version of the list of environments which are
// encapsulated by quotes and are comma separated
func (c *ProjectEnvironmentVariable) DisplayEnvironments() string {
	return helpers.SerializeToHCL("environment", c.Environment)
}

func EqualEnvironmentVariables(c, o ProjectEnvironmentVariable) bool {
	return c.Key == o.Key && c.Value == o.Value && slices.Equal(c.Environment, o.Environment)
}

func (c *ProjectConfig) extendConfig(o *ProjectConfig) *ProjectConfig {
	if o != nil && o != (&ProjectConfig{}) {
		cfg := &ProjectConfig{
			ManualProductionDeployment: o.ManualProductionDeployment,
			EnvironmentVariables:       o.EnvironmentVariables,
		}

		if c.ManualProductionDeployment != o.ManualProductionDeployment {
			cfg.ManualProductionDeployment = c.ManualProductionDeployment
		}

		if !slices.EqualFunc(c.EnvironmentVariables, o.EnvironmentVariables, EqualEnvironmentVariables) {
			// Append missing environment variables
			cfg.EnvironmentVariables = append(cfg.EnvironmentVariables, c.EnvironmentVariables...)
			// TODO: Update environment variables that exist in both configs with values of the site config

		}
		return cfg
	}

	return c
}
