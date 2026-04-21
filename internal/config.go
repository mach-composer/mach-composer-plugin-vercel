package internal

import (
	"fmt"
	"sort"
	"strings"

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
				DeploymentType: "",
			},
			VercelAuthentication: VercelAuthentication{
				DeploymentType: "",
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
	ManualProductionDeployment    *bool                        `mapstructure:"manual_production_deployment"`
	ServerlessFunctionRegion      string                       `mapstructure:"serverless_function_region"`
	EnvironmentVariables          []ProjectEnvironmentVariable `mapstructure:"environment_variables"`
	GitRepository                 GitRepository                `mapstructure:"git_repository"`
	BuildCommand                  string                       `mapstructure:"build_command"`
	IgnoreCommand                 string                       `mapstructure:"ignore_command"`
	RootDirectory                 string                       `mapstructure:"root_directory"`
	NodeVersion                   string                       `mapstructure:"node_version"`
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
			NodeVersion:                   o.NodeVersion,
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

		if c.NodeVersion != "" {
			cfg.NodeVersion = c.NodeVersion
		}

		if c.ManualProductionDeployment != nil {
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
	Key                    string   `mapstructure:"key"`
	Value                  string   `mapstructure:"value"`
	Environment            []string `mapstructure:"environment"`
	Comment                string   `mapstructure:"comment"`
	CustomEnvironmentIDs   []string `mapstructure:"custom_environment_ids"`
	GitBranch              string   `mapstructure:"git_branch"`
	Sensitive              bool     `mapstructure:"sensitive"`
	Target                 []string `mapstructure:"target"`
}

func (c *ProjectEnvironmentVariable) validate() error {
	if c.Sensitive && slices.Contains(c.Target, "development") {
		return fmt.Errorf("environment variable %q: target cannot include \"development\" when sensitive is true", c.Key)
	}
	return nil
}

func (c *VercelConfig) validate() error {
	for _, env := range c.ProjectConfig.EnvironmentVariables {
		if err := env.validate(); err != nil {
			return err
		}
	}
	return nil
}

// Returns a HCL-friendly version of the list of environments which are
// encapsulated by quotes and are comma separated
func (c *ProjectEnvironmentVariable) DisplayEnvironments() string {
	return helpers.SerializeToHCL("environment", c.Environment)
}

func (c *ProjectEnvironmentVariable) DisplayTarget() string {
	return helpers.SerializeToHCL("target", c.Target)
}

func (c *ProjectEnvironmentVariable) DisplayCustomEnvironmentIDs() string {
	return helpers.SerializeToHCL("custom_environment_ids", c.CustomEnvironmentIDs)
}

func MergeEnvironmentVariables(o []ProjectEnvironmentVariable, c []ProjectEnvironmentVariable) []ProjectEnvironmentVariable {
	type envKey struct {
		key string
		env string
	}
	winners := map[envKey]ProjectEnvironmentVariable{}

	apply := func(src []ProjectEnvironmentVariable) {
		for _, entry := range src {
			envs := entry.Environment
			if len(envs) == 0 {
				envs = []string{"development", "preview", "production"}
			}
			for _, e := range envs {
				winners[envKey{entry.Key, e}] = entry
			}
		}
	}
	apply(o)
	apply(c)

	// Group entries whose non-Environment fields match so we can consolidate
	// the environment list for each unique (value, metadata) combination.
	type groupKey struct {
		key       string
		value     string
		comment   string
		gitBranch string
		sensitive bool
		target    string
		customIDs string
	}
	joinKey := func(parts []string) string { return strings.Join(parts, "\x00") }

	envsByGroup := map[groupKey]map[string]struct{}{}
	sample := map[groupKey]ProjectEnvironmentVariable{}

	for ek, entry := range winners {
		gk := groupKey{
			key:       entry.Key,
			value:     entry.Value,
			comment:   entry.Comment,
			gitBranch: entry.GitBranch,
			sensitive: entry.Sensitive,
			target:    joinKey(entry.Target),
			customIDs: joinKey(entry.CustomEnvironmentIDs),
		}
		if envsByGroup[gk] == nil {
			envsByGroup[gk] = map[string]struct{}{}
			sample[gk] = entry
		}
		envsByGroup[gk][ek.env] = struct{}{}
	}

	result := make([]ProjectEnvironmentVariable, 0, len(envsByGroup))
	for gk, envSet := range envsByGroup {
		envs := make([]string, 0, len(envSet))
		for e := range envSet {
			envs = append(envs, e)
		}
		sort.Strings(envs)
		s := sample[gk]
		result = append(result, ProjectEnvironmentVariable{
			Key:                  s.Key,
			Value:                s.Value,
			Environment:          envs,
			Comment:              s.Comment,
			CustomEnvironmentIDs: s.CustomEnvironmentIDs,
			GitBranch:            s.GitBranch,
			Sensitive:            s.Sensitive,
			Target:               s.Target,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Key != result[j].Key {
			return result[i].Key < result[j].Key
		}
		return result[i].Value < result[j].Value
	})

	return result
}

type ProjectDomain struct {
	Domain             string `mapstructure:"domain"`
	GitBranch          string `mapstructure:"git_branch"`
	Redirect           string `mapstructure:"redirect"`
	RedirectStatusCode int64  `mapstructure:"redirect_status_code"`
}
