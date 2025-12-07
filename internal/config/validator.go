package config

import (
	"fmt"
	"os"
	"strings"
)

// EnvVarDefinition defines the rules for an environment variable
type EnvVarDefinition struct {
	Name        string
	Required    bool
	FeatureFlag string // If set, this variable is required only if the FeatureFlag env var is "true"
}

// Validator handles environment variable validation
type Validator struct {
	definitions []EnvVarDefinition
}

// NewValidator creates a new Validator with the given definitions
func NewValidator(definitions []EnvVarDefinition) *Validator {
	return &Validator{
		definitions: definitions,
	}
}

// Validate checks if the environment variables satisfy the definitions
func (v *Validator) Validate() error {
	var missingVars []string

	for _, def := range v.definitions {
		// If not required at all, skip
		if !def.Required {
			continue
		}

		// Check if feature flag is involved
		if def.FeatureFlag != "" {
			flagValue := os.Getenv(def.FeatureFlag)
			// We treat "true" (case-insensitive) as enabled
			if strings.ToLower(flagValue) != "true" {
				// Feature flag is not enabled, so this variable is not required
				continue
			}
		}

		// Check if the variable is set and not empty
		if os.Getenv(def.Name) == "" {
			if def.FeatureFlag != "" {
				missingVars = append(missingVars, fmt.Sprintf("%s (required by %s=true)", def.Name, def.FeatureFlag))
			} else {
				missingVars = append(missingVars, def.Name)
			}
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}

	return nil
}
