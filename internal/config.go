package internal

import (
	"encoding/json"
	"sort"
	"strconv"
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

var defaultProjectEnvironmentVariableTargets = []string{"development", "preview", "production"}

type ProjectEnvironmentVariable struct {
	Key       string   `mapstructure:"key"`
	Value     string   `mapstructure:"value"`
	Target    []string `mapstructure:"target"`
	Comment   *string  `mapstructure:"comment"`
	GitBranch *string  `mapstructure:"git_branch"`
	Sensitive *bool    `mapstructure:"sensitive"`
}

func (c ProjectEnvironmentVariable) normalize() ProjectEnvironmentVariable {
	normalized := c

	normalized.Comment = normalizeOptionalString(normalized.Comment)
	normalized.GitBranch = normalizeOptionalString(normalized.GitBranch)

	if len(normalized.Target) == 0 {
		normalized.Target = append([]string(nil), defaultProjectEnvironmentVariableTargets...)
	} else {
		normalized.Target = append([]string(nil), normalized.Target...)
	}
	sort.Strings(normalized.Target)

	return normalized
}

func (c ProjectEnvironmentVariable) Validate() error {
	if c.GitBranch == nil {
		return nil
	}

	if len(c.Target) != 1 || c.Target[0] != "preview" {
		return &InvalidEnvironmentVariableError{
			Key:     c.Key,
			Message: "git_branch can only be used when target is [\"preview\"]",
		}
	}

	return nil
}

func (c ProjectEnvironmentVariable) DisplayTarget() string {
	if len(c.Target) == 0 {
		return ""
	}

	return helpers.SerializeToHCL("target", c.Target)
}

func (c ProjectEnvironmentVariable) DisplayComment() string {
	if c.Comment == nil || *c.Comment == "" {
		return ""
	}

	return helpers.SerializeToHCL("comment", *c.Comment)
}

func (c ProjectEnvironmentVariable) DisplayGitBranch() string {
	if c.GitBranch == nil || *c.GitBranch == "" {
		return ""
	}

	return helpers.SerializeToHCL("git_branch", *c.GitBranch)
}

func (c ProjectEnvironmentVariable) DisplaySensitive() string {
	if c.Sensitive == nil {
		return ""
	}

	return helpers.SerializeToHCL("sensitive", *c.Sensitive)
}

func MergeEnvironmentVariables(o []ProjectEnvironmentVariable, c []ProjectEnvironmentVariable) []ProjectEnvironmentVariable {
	merged := make(map[string]map[string]ProjectEnvironmentVariable, len(o)+len(c))

	merge := func(items []ProjectEnvironmentVariable) {
		for _, env := range items {
			normalized := env.normalize()
			if _, exists := merged[normalized.Key]; !exists {
				merged[normalized.Key] = make(map[string]ProjectEnvironmentVariable)
			}

			for _, target := range normalized.Target {
				entry := normalized
				entry.Target = nil
				merged[normalized.Key][target] = entry
			}
		}
	}

	merge(o)
	merge(c)

	result := []ProjectEnvironmentVariable{}
	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		grouped := make(map[string]ProjectEnvironmentVariable)
		targets := make([]string, 0, len(merged[key]))
		for target := range merged[key] {
			targets = append(targets, target)
		}
		sort.Strings(targets)

		for _, target := range targets {
			env := merged[key][target]
			signature := env.payloadSignature()
			group, exists := grouped[signature]
			if !exists {
				group = env
			}
			group.Target = append(group.Target, target)
			grouped[signature] = group
		}

		signatures := make([]string, 0, len(grouped))
		for signature := range grouped {
			signatures = append(signatures, signature)
		}
		sort.Strings(signatures)

		for _, signature := range signatures {
			env := grouped[signature]
			sort.Strings(env.Target)
			result = append(result, env)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Key != result[j].Key {
			return result[i].Key < result[j].Key
		}
		if result[i].Value != result[j].Value {
			return result[i].Value < result[j].Value
		}
		if diff := compareStringSlices(result[i].Target, result[j].Target); diff != 0 {
			return diff < 0
		}
		return result[i].payloadSignature() < result[j].payloadSignature()
	})

	return result
}

type InvalidEnvironmentVariableError struct {
	Key     string
	Message string
}

func (e *InvalidEnvironmentVariableError) Error() string {
	return "invalid project environment variable " + strconv.Quote(e.Key) + ": " + e.Message
}

func (c ProjectEnvironmentVariable) payloadSignature() string {
	payload, err := json.Marshal(struct {
		Key       string  `json:"key"`
		Value     string  `json:"value"`
		Comment   *string `json:"comment,omitempty"`
		GitBranch *string `json:"git_branch,omitempty"`
		Sensitive *bool   `json:"sensitive,omitempty"`
	}{
		Key:       c.Key,
		Value:     c.Value,
		Comment:   c.Comment,
		GitBranch: c.GitBranch,
		Sensitive: c.Sensitive,
	})
	if err != nil {
		panic(err)
	}

	return string(payload)
}

func compareStringSlices(a []string, b []string) int {
	return strings.Compare(strings.Join(a, ","), strings.Join(b, ","))
}

func normalizeOptionalString(value *string) *string {
	if value == nil || *value == "" {
		return nil
	}

	return value
}

type ProjectDomain struct {
	Domain             string `mapstructure:"domain"`
	GitBranch          string `mapstructure:"git_branch"`
	Redirect           string `mapstructure:"redirect"`
	RedirectStatusCode int64  `mapstructure:"redirect_status_code"`
}
