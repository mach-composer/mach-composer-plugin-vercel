package internal

import (
	"github.com/google/go-cmp/cmp"
)

type VercelConfig struct {
	TeamID        string        `mapstructure:"team_id"`
	APIToken      string        `mapstructure:"api_token"`
	ProjectConfig ProjectConfig `mapstructure:"project_config"`
}

type ProjectConfig struct {
	ManualProductionDeployment bool                         `mapstructure:"manual_production_deployment"`
	EnvironmentVariables       []ProjectEnvironmentVariable `mapstructure:"environment_variables"`
}

type ProjectEnvironmentVariable struct {
	Name        string    `mapstructure:"name"`
	Value       string    `mapstructure:"value"`
	Environment [3]string `mapstructure:"environment"`
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

func (c *ProjectEnvironmentVariable) DefaultEnvironments() string {
	// Ugly but getting proper templated joined strings is hard in Go :(
	return "[\"development\", \"preview\", \"production\"]"
}
