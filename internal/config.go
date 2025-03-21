package internal

import (
	"sort"

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
				DeploymentType: "standard_protection",
			},
			VercelAuthentication: VercelAuthentication{
				DeploymentType: "standard_protection",
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
	IgnoreCommand                 string                       `mapstructure:"ignore_command"`
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
			IgnoreCommand:                 o.IgnoreCommand,
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

		if c.IgnoreCommand != "" {
			cfg.IgnoreCommand = c.IgnoreCommand
		}

		if c.RootDirectory != "" {
			cfg.RootDirectory = c.RootDirectory
		}

		if c.ManualProductionDeployment != o.ManualProductionDeployment {
			cfg.ManualProductionDeployment = c.ManualProductionDeployment
		}

		if c.GitRepository.Type != "" || c.GitRepository.Repo != "" || c.GitRepository.ProductionBranch != "" {
			result := c.GitRepository.extendConfig(&o.GitRepository)
			if result != nil {
				cfg.GitRepository = *result
			} else {
				cfg.GitRepository = c.GitRepository
			}

		}

		if c.ProtectionBypassForAutomation {
			cfg.ProtectionBypassForAutomation = c.ProtectionBypassForAutomation
		}

		if c.VercelAuthentication.DeploymentType != "" {
			cfg.VercelAuthentication.DeploymentType = c.VercelAuthentication.DeploymentType
		}

		if c.PasswordProtection.Password != "" {
			cfg.PasswordProtection = c.PasswordProtection
		}

		cfg.EnvironmentVariables = MergeEnvironmentVariables(c.EnvironmentVariables, o.EnvironmentVariables)

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

func (c *GitRepository) extendConfig(o *GitRepository) *GitRepository {
	if o != nil && o != (&GitRepository{}) {
		cfg := &GitRepository{
			ProductionBranch: o.ProductionBranch,
			Type:             o.Type,
			Repo:             o.Repo,
		}

		if c.ProductionBranch != "" {
			cfg.ProductionBranch = c.ProductionBranch
		}

		if c.Type != "" {
			cfg.Type = c.Type
		}

		if c.Repo != "" {
			cfg.Repo = c.Repo
		}

		return cfg
	}

	return c
}

type PasswordProtection struct {
	Password       string `mapstructure:"password"`
	DeploymentType string `mapstructure:"deployment_type"`
}

type VercelAuthentication struct {
	DeploymentType string `mapstructure:"deployment_type"`
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

func MergeEnvironmentVariables(o []ProjectEnvironmentVariable, c []ProjectEnvironmentVariable) []ProjectEnvironmentVariable {
	merged := make(map[string]map[string]string, len(o)+len(c))

	// process parent environments
	for _, env := range o {
		// normalize environment as default behavior for Vercel is to output to all environments
		if len(env.Environment) == 0 {
			env.Environment = []string{"development", "preview", "production"}
		}
		for _, environment := range env.Environment {
			if _, exists := merged[env.Key]; !exists {
				merged[env.Key] = make(map[string]string, 3)
			}
			merged[env.Key][environment] = env.Value
		}
	}

	// process child environments
	for _, env := range c {
		// normalize environment as default behavior for Vercel is to output to all environments
		if len(env.Environment) == 0 {
			env.Environment = []string{"development", "preview", "production"}
		}
		for _, environment := range env.Environment {
			if _, exists := merged[env.Key]; !exists {
				merged[env.Key] = make(map[string]string, 3)
			}
			merged[env.Key][environment] = env.Value
		}
	}

	// Convert the map back to a slice of ProjectEnvironmentVariable
	result := []ProjectEnvironmentVariable{}
	for key, envMap := range merged {
		// Group variables by value to consolidate environments
		valueGroups := make(map[string][]string)

		for environment, value := range envMap {
			valueGroups[value] = append(valueGroups[value], environment)
		}

		// Create final environment variables with consolidated environments
		for value, environments := range valueGroups {

			// Sort environments for consistent order
			sort.Strings(environments)

			result = append(result, ProjectEnvironmentVariable{
				Key:         key,
				Value:       value,
				Environment: environments,
			})
		}
	}

	// Sort the result by key for consistent order
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})

	return result
}

type ProjectDomain struct {
	Domain             string `mapstructure:"domain"`
	GitBranch          string `mapstructure:"git_branch"`
	Redirect           string `mapstructure:"redirect"`
	RedirectStatusCode int64  `mapstructure:"redirect_status_code"`
}
