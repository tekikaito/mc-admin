package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ValidationFunc is a function that validates an environment variable value
type ValidationFunc func(value string) error

// Chain combines multiple validation functions into one.
// They are executed in order, and the first error stops the chain.
func Chain(validators ...ValidationFunc) ValidationFunc {
	return func(value string) error {
		for _, v := range validators {
			if err := v(value); err != nil {
				return err
			}
		}
		return nil
	}
}

// EnvVarDefinition defines the rules for an environment variable
type EnvVarDefinition struct {
	Name           string
	Required       bool
	FeatureFlag    string         // If set, this variable is required only if the FeatureFlag env var is "true"
	ValidationFunc ValidationFunc // Optional validation function
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
	var errors []string

	for _, def := range v.definitions {
		val := os.Getenv(def.Name)

		// Determine if the variable is effectively required
		isRequired := def.Required
		if def.FeatureFlag != "" {
			flagValue := os.Getenv(def.FeatureFlag)
			// We treat "true" (case-insensitive) as enabled
			if strings.ToLower(flagValue) != "true" {
				isRequired = false
			}
		}

		// Check for presence if required
		if val == "" {
			if isRequired {
				if def.FeatureFlag != "" {
					errors = append(errors, fmt.Sprintf("%s (required by %s=true)", def.Name, def.FeatureFlag))
				} else {
					errors = append(errors, def.Name)
				}
			}
			// If not present and not required, skip validation
			continue
		}

		// If present, run validation if defined
		if def.ValidationFunc != nil {
			if err := def.ValidationFunc(val); err != nil {
				errors = append(errors, fmt.Sprintf("%s invalid: %v", def.Name, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("environment validation failed: %s", strings.Join(errors, ", "))
	}

	return nil
}

// Common validators

// IsInteger checks if the value is a valid integer
func IsInteger(value string) error {
	_, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("must be an integer")
	}
	return nil
}

// IsNotEmpty checks if the value is not empty (after trimming whitespace)
func IsNotEmpty(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("must not be empty")
	}
	return nil
}

// MaxLength returns a validator that checks if the value length is within the limit
func MaxLength(max int) ValidationFunc {
	return func(value string) error {
		if len(value) > max {
			return fmt.Errorf("must be at most %d characters", max)
		}
		return nil
	}
}
