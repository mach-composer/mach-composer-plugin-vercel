package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeList(t *testing.T) {
	t.Run("empty lists", func(t *testing.T) {
		result := MergeEnvironmentVariables([]ProjectEnvironmentVariable{}, []ProjectEnvironmentVariable{})
		assert.Empty(t, result)
	})

	t.Run("only parent variables", func(t *testing.T) {
		parent := []ProjectEnvironmentVariable{
			{Key: "API_URL", Value: "https://api.example.com", Environment: []string{"production", "preview"}},
		}
		result := MergeEnvironmentVariables(parent, []ProjectEnvironmentVariable{})

		assert.Len(t, result, 1)
		assert.Equal(t, "API_URL", result[0].Key)
		assert.Equal(t, "https://api.example.com", result[0].Value)
		assert.ElementsMatch(t, []string{"preview", "production"}, result[0].Environment)
	})

	t.Run("only child variables", func(t *testing.T) {
		child := []ProjectEnvironmentVariable{
			{Key: "DEBUG", Value: "true", Environment: []string{"development"}},
		}
		result := MergeEnvironmentVariables([]ProjectEnvironmentVariable{}, child)

		assert.Len(t, result, 1)
		assert.Equal(t, "DEBUG", result[0].Key)
		assert.Equal(t, "true", result[0].Value)
		assert.ElementsMatch(t, []string{"development"}, result[0].Environment)
	})

	t.Run("child overrides parent for specific environment", func(t *testing.T) {
		parent := []ProjectEnvironmentVariable{
			{Key: "API_URL", Value: "https://api.example.com", Environment: []string{"production", "preview"}},
		}
		child := []ProjectEnvironmentVariable{
			{Key: "API_URL", Value: "https://api-test.example.com", Environment: []string{"preview"}},
		}

		result := MergeEnvironmentVariables(parent, child)

		assert.Len(t, result, 2)

		// Find the production environment entry
		var prodEntry, previewEntry ProjectEnvironmentVariable
		for _, entry := range result {
			if entry.Key == "API_URL" {
				if entry.Value == "https://api.example.com" {
					prodEntry = entry
				} else if entry.Value == "https://api-test.example.com" {
					previewEntry = entry
				}
			}
		}

		assert.ElementsMatch(t, []string{"production"}, prodEntry.Environment)
		assert.ElementsMatch(t, []string{"preview"}, previewEntry.Environment)
	})

	t.Run("child adds new environments to existing key", func(t *testing.T) {
		parent := []ProjectEnvironmentVariable{
			{Key: "FEATURE_FLAG", Value: "true", Environment: []string{"production"}},
		}
		child := []ProjectEnvironmentVariable{
			{Key: "FEATURE_FLAG", Value: "true", Environment: []string{"preview"}},
		}

		result := MergeEnvironmentVariables(parent, child)

		assert.Len(t, result, 1)
		assert.Equal(t, "FEATURE_FLAG", result[0].Key)
		assert.Equal(t, "true", result[0].Value)
		assert.ElementsMatch(t, []string{"production", "preview"}, result[0].Environment)
	})

	t.Run("complex case with multiple variables and environments", func(t *testing.T) {
		parent := []ProjectEnvironmentVariable{
			{Key: "API_URL", Value: "https://api.example.com", Environment: []string{"production", "preview"}},
			{Key: "DEBUG", Value: "false", Environment: []string{"production"}},
			{Key: "DEBUG", Value: "true", Environment: []string{"development"}},
		}

		child := []ProjectEnvironmentVariable{
			{Key: "API_URL", Value: "https://api-test.example.com", Environment: []string{"preview"}},
			{Key: "DEBUG", Value: "false", Environment: []string{"preview"}}, // Adding preview with same value as production
			{Key: "NEW_VAR", Value: "hello", Environment: []string{"production", "preview", "development"}},
		}

		result := MergeEnvironmentVariables(parent, child)

		// Expected:
		// 1. API_URL = https://api.example.com for production
		// 2. API_URL = https://api-test.example.com for preview
		// 3. DEBUG = false for production, preview
		// 4. DEBUG = true for development
		// 5. NEW_VAR = hello for production, preview, development

		assert.Len(t, result, 5)

		// Find and verify each entry
		var apiUrlProd, apiUrlPreview, debugFalse, debugTrue, newVar ProjectEnvironmentVariable

		for _, entry := range result {
			switch {
			case entry.Key == "API_URL" && entry.Value == "https://api.example.com":
				apiUrlProd = entry
			case entry.Key == "API_URL" && entry.Value == "https://api-test.example.com":
				apiUrlPreview = entry
			case entry.Key == "DEBUG" && entry.Value == "false":
				debugFalse = entry
			case entry.Key == "DEBUG" && entry.Value == "true":
				debugTrue = entry
			case entry.Key == "NEW_VAR":
				newVar = entry
			}
		}

		assert.ElementsMatch(t, []string{"production"}, apiUrlProd.Environment)
		assert.ElementsMatch(t, []string{"preview"}, apiUrlPreview.Environment)
		assert.ElementsMatch(t, []string{"preview", "production"}, debugFalse.Environment)
		assert.ElementsMatch(t, []string{"development"}, debugTrue.Environment)
		assert.ElementsMatch(t, []string{"development", "preview", "production"}, newVar.Environment)
	})

	t.Run("only parent variables with empty environment", func(t *testing.T) {
		parent := []ProjectEnvironmentVariable{
			{Key: "API_URL", Value: "https://api.example.com", Environment: []string{}},
		}
		result := MergeEnvironmentVariables(parent, []ProjectEnvironmentVariable{})

		assert.Len(t, result, 1)
		assert.Equal(t, "API_URL", result[0].Key)
		assert.Equal(t, "https://api.example.com", result[0].Value)
		assert.ElementsMatch(t, []string{"development", "preview", "production"}, result[0].Environment)
	})

	t.Run("only child variables with empty environment", func(t *testing.T) {
		child := []ProjectEnvironmentVariable{
			{Key: "DEBUG", Value: "true", Environment: []string{}},
		}
		result := MergeEnvironmentVariables([]ProjectEnvironmentVariable{}, child)

		assert.Len(t, result, 1)
		assert.Equal(t, "DEBUG", result[0].Key)
		assert.Equal(t, "true", result[0].Value)
		assert.ElementsMatch(t, []string{"development", "preview", "production"}, result[0].Environment)
	})

	t.Run("child adds new environments to existing key with empty environment", func(t *testing.T) {
		parent := []ProjectEnvironmentVariable{
			{Key: "FEATURE_FLAG", Value: "true", Environment: []string{"production"}},
		}
		child := []ProjectEnvironmentVariable{
			{Key: "FEATURE_FLAG", Value: "true", Environment: []string{}},
		}

		result := MergeEnvironmentVariables(parent, child)

		assert.Len(t, result, 1)
		assert.Equal(t, "FEATURE_FLAG", result[0].Key)
		assert.Equal(t, "true", result[0].Value)
		assert.ElementsMatch(t, []string{"development", "preview", "production"}, result[0].Environment)
	})
}
