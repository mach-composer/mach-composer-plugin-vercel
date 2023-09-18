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

// Creates a new VercelConfig with default values
func NewVercelConfig() VercelConfig {
	return VercelConfig{
		ProjectConfig: ProjectConfig{
			PasswordProtection: PasswordProtection{
				ProtectProduction: true,
			},
			VercelAuthentication: VercelAuthentication{
				ProtectProduction: true,
			},
		},
	}
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
	Name                          string                       `mapstructure:"name"`
	Framework                     string                       `mapstructure:"framework"`
	ManualProductionDeployment    bool                         `mapstructure:"manual_production_deployment"`
	ServerlessFunctionRegion      string                       `mapstructure:"serverless_function_region"`
	EnvironmentVariables          []ProjectEnvironmentVariable `mapstructure:"environment_variables"`
	GitRepository                 GitRepository                `mapstructure:"git_repository"`
	BuildCommand                  string                       `mapstructure:"build_command"`
	RootDirectory                 string                       `mapstructure:"root_directory"`
	ProjectDomains                []ProjectDomain              `mapstructure:"domains"`
	ProtectionBypassForAutomation bool                         `mapstructure:"protection_bypass_for_automation"`
	PasswordProtection            PasswordProtection           `mapstructure:"password_protection"`
	VercelAuthentication          VercelAuthentication         `mapstructure:"vercel_authentication"`
}

func (c *ProjectConfig) extendConfig(o *ProjectConfig) *ProjectConfig {
	if o != nil && o != (&ProjectConfig{}) {
		cfg := &ProjectConfig{
			Name:                          o.Name,
			Framework:                     o.Framework,
			ServerlessFunctionRegion:      o.ServerlessFunctionRegion,
			BuildCommand:                  o.BuildCommand,
			RootDirectory:                 o.RootDirectory,
			ManualProductionDeployment:    o.ManualProductionDeployment,
			EnvironmentVariables:          o.EnvironmentVariables,
			GitRepository:                 o.GitRepository,
			ProtectionBypassForAutomation: o.ProtectionBypassForAutomation,
			PasswordProtection:            o.PasswordProtection,
			VercelAuthentication:          o.VercelAuthentication,
			ProjectDomains:                o.ProjectDomains,
		}

		if c.Name != "" {
			cfg.Name = c.Name
		}

		if c.Framework != "" {
			cfg.Framework = c.Framework
		}

		if c.ServerlessFunctionRegion != "" {
			cfg.ServerlessFunctionRegion = c.ServerlessFunctionRegion
		}

		if c.BuildCommand != "" {
			cfg.BuildCommand = c.BuildCommand
		}

		if c.RootDirectory != "" {
			cfg.RootDirectory = c.RootDirectory
		}

		if c.ManualProductionDeployment != o.ManualProductionDeployment {
			cfg.ManualProductionDeployment = c.ManualProductionDeployment
		}

		if c.GitRepository.Type != "" || c.GitRepository.Repo != "" {
			cfg.GitRepository = c.GitRepository
		}

		if c.ProtectionBypassForAutomation {
			cfg.ProtectionBypassForAutomation = c.ProtectionBypassForAutomation
		}

		if !c.VercelAuthentication.ProtectProduction {
			cfg.VercelAuthentication.ProtectProduction = c.VercelAuthentication.ProtectProduction
		}

		if c.PasswordProtection.Password != "" || !c.PasswordProtection.ProtectProduction {
			cfg.PasswordProtection = c.PasswordProtection
		}

		if !slices.EqualFunc(c.EnvironmentVariables, o.EnvironmentVariables, EqualEnvironmentVariables) {
			// Append missing environment variables
			cfg.EnvironmentVariables = append(cfg.EnvironmentVariables, c.EnvironmentVariables...)
		}

		if !slices.EqualFunc(c.ProjectDomains, o.ProjectDomains, func(c, o ProjectDomain) bool {
			return c.Domain == o.Domain && c.GitBranch == o.GitBranch && c.Redirect == o.Redirect && c.RedirectStatusCode == o.RedirectStatusCode
		}) {
			// Append missing project domains
			cfg.ProjectDomains = append(cfg.ProjectDomains, c.ProjectDomains...)
		}

		return cfg
	}

	return c
}

type GitRepository struct {
	ProductionBranch string `mapstructure:"production_branch"`
	Type             string `mapstructure:"type"`
	Repo             string `mapstructure:"repo"`
}

type PasswordProtection struct {
	Password          string `mapstructure:"password"`
	ProtectProduction bool   `mapstructure:"protect_production"`
}

type VercelAuthentication struct {
	ProtectProduction bool `mapstructure:"protect_production"`
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

type ProjectDomain struct {
	Domain             string `mapstructure:"domain"`
	GitBranch          string `mapstructure:"git_branch"`
	Redirect           string `mapstructure:"redirect"`
	RedirectStatusCode int64  `mapstructure:"redirect_status_code"`
}
