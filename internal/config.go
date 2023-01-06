package internal

import (
	"bytes"
	"text/template"

	"github.com/google/go-cmp/cmp"
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
	Name        string   `mapstructure:"name"`
	Value       string   `mapstructure:"value"`
	Environment []string `mapstructure:"environment"`
}

// Returns a HCL-friendly version of the list of environments which are encapsulated by
// quotes and are comma separated
func (c *ProjectEnvironmentVariable) DisplayEnvironments() (string, error) {
	tpl := `[{{ range $i, $e := . }}{{if $i}}, {{end}}{{ if last $i $}}{{ end}}"{{$e}}"{{end}}]`
	t := template.Must(template.New("template").Funcs(templateFunctions).Parse(tpl))

	var content bytes.Buffer
	err := t.Execute(&content, c.Environment)

	if err != nil {
		return "", err
	}

	return content.String(), nil
}

func (c *ProjectEnvironmentVariable) DefaultEnvironments() string {
	// Ugly but getting proper templated joined strings is hard in Go :(
	return "[\"development\", \"preview\", \"production\"]"
}
