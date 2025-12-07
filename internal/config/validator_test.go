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
