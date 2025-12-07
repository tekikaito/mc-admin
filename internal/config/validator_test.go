package config

import (
	"os"
	"testing"
)

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		definitions []EnvVarDefinition
		envVars     map[string]string
		wantErr     bool
	}{
		{
			name: "Required variable present",
			definitions: []EnvVarDefinition{
				{Name: "TEST_VAR", Required: true},
			},
			envVars: map[string]string{
				"TEST_VAR": "value",
			},
			wantErr: false,
		},
		{
			name: "Required variable missing",
			definitions: []EnvVarDefinition{
				{Name: "TEST_VAR", Required: true},
			},
			envVars: map[string]string{},
			wantErr: true,
		},
		{
			name: "Feature flag disabled, dependent var missing",
			definitions: []EnvVarDefinition{
				{Name: "DEPENDENT_VAR", Required: true, FeatureFlag: "FEATURE_FLAG"},
			},
			envVars: map[string]string{
				"FEATURE_FLAG": "false",
			},
			wantErr: false,
		},
		{
			name: "Feature flag enabled, dependent var present",
			definitions: []EnvVarDefinition{
				{Name: "DEPENDENT_VAR", Required: true, FeatureFlag: "FEATURE_FLAG"},
			},
			envVars: map[string]string{
				"FEATURE_FLAG":  "true",
				"DEPENDENT_VAR": "value",
			},
			wantErr: false,
		},
		{
			name: "Feature flag enabled, dependent var missing",
			definitions: []EnvVarDefinition{
				{Name: "DEPENDENT_VAR", Required: true, FeatureFlag: "FEATURE_FLAG"},
			},
			envVars: map[string]string{
				"FEATURE_FLAG": "true",
			},
			wantErr: true,
		},
		{
			name: "Feature flag missing (defaults to false), dependent var missing",
			definitions: []EnvVarDefinition{
				{Name: "DEPENDENT_VAR", Required: true, FeatureFlag: "FEATURE_FLAG"},
			},
			envVars: map[string]string{},
			wantErr: false,
		},
		{
			name: "Validation function passes",
			definitions: []EnvVarDefinition{
				{Name: "INT_VAR", Required: true, ValidationFunc: IsInteger},
			},
			envVars: map[string]string{
				"INT_VAR": "123",
			},
			wantErr: false,
		},
		{
			name: "Validation function fails",
			definitions: []EnvVarDefinition{
				{Name: "INT_VAR", Required: true, ValidationFunc: IsInteger},
			},
			envVars: map[string]string{
				"INT_VAR": "abc",
			},
			wantErr: true,
		},
		{
			name: "Optional variable with validation (present and valid)",
			definitions: []EnvVarDefinition{
				{Name: "OPTIONAL_INT", Required: false, ValidationFunc: IsInteger},
			},
			envVars: map[string]string{
				"OPTIONAL_INT": "456",
			},
			wantErr: false,
		},
		{
			name: "Optional variable with validation (present and invalid)",
			definitions: []EnvVarDefinition{
				{Name: "OPTIONAL_INT", Required: false, ValidationFunc: IsInteger},
			},
			envVars: map[string]string{
				"OPTIONAL_INT": "not-an-int",
			},
			wantErr: true,
		},
		{
			name: "Max length validation passes",
			definitions: []EnvVarDefinition{
				{Name: "SHORT_VAR", Required: true, ValidationFunc: MaxLength(5)},
			},
			envVars: map[string]string{
				"SHORT_VAR": "12345",
			},
			wantErr: false,
		},
		{
			name: "Max length validation fails",
			definitions: []EnvVarDefinition{
				{Name: "SHORT_VAR", Required: true, ValidationFunc: MaxLength(5)},
			},
			envVars: map[string]string{
				"SHORT_VAR": "123456",
			},
			wantErr: true,
		},
		{
			name: "Chained validation passes",
			definitions: []EnvVarDefinition{
				{Name: "CHAINED_VAR", Required: true, ValidationFunc: Chain(IsNotEmpty, MaxLength(5))},
			},
			envVars: map[string]string{
				"CHAINED_VAR": "123",
			},
			wantErr: false,
		},
		{
			name: "Chained validation fails first check",
			definitions: []EnvVarDefinition{
				{Name: "CHAINED_VAR", Required: true, ValidationFunc: Chain(IsNotEmpty, MaxLength(5))},
			},
			envVars: map[string]string{
				"CHAINED_VAR": "   ",
			},
			wantErr: true,
		},
		{
			name: "Chained validation fails second check",
			definitions: []EnvVarDefinition{
				{Name: "CHAINED_VAR", Required: true, ValidationFunc: Chain(IsNotEmpty, MaxLength(5))},
			},
			envVars: map[string]string{
				"CHAINED_VAR": "123456",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			v := NewValidator(tt.definitions)
			if err := v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
