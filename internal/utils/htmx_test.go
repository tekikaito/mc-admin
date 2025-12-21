package utils

import (
	"encoding/json"
	"testing"
)

func TestBuildToastTrigger(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		toastType   string
		wantMessage string
		wantType    string
	}{
		{
			name:        "simple success message",
			message:     "Operation successful",
			toastType:   "success",
			wantMessage: "Operation successful",
			wantType:    "success",
		},
		{
			name:        "simple error message",
			message:     "Operation failed",
			toastType:   "error",
			wantMessage: "Operation failed",
			wantType:    "error",
		},
		{
			name:        "message with double quotes",
			message:     `Failed to set time: error with "quotes"`,
			toastType:   "error",
			wantMessage: `Failed to set time: error with "quotes"`,
			wantType:    "error",
		},
		{
			name:        "message with single quotes",
			message:     "Failed to set time: error with 'single quotes'",
			toastType:   "error",
			wantMessage: "Failed to set time: error with 'single quotes'",
			wantType:    "error",
		},
		{
			name:        "message with backslash",
			message:     `Failed to set time: error with \backslash`,
			toastType:   "error",
			wantMessage: `Failed to set time: error with \backslash`,
			wantType:    "error",
		},
		{
			name:        "message with newline",
			message:     "Failed to set time: error with newline\nand more",
			toastType:   "error",
			wantMessage: "Failed to set time: error with newline\nand more",
			wantType:    "error",
		},
		{
			name:        "message with curly braces and brackets",
			message:     "Failed to set time: error with {curly} and [brackets]",
			toastType:   "error",
			wantMessage: "Failed to set time: error with {curly} and [brackets]",
			wantType:    "error",
		},
		{
			name:        "message with special characters",
			message:     `Failed: <script>alert("xss")</script>`,
			toastType:   "error",
			wantMessage: `Failed: <script>alert("xss")</script>`,
			wantType:    "error",
		},
		{
			name:        "empty message",
			message:     "",
			toastType:   "info",
			wantMessage: "",
			wantType:    "info",
		},
		{
			name:        "unicode characters",
			message:     "Success! 🎉 操作成功",
			toastType:   "success",
			wantMessage: "Success! 🎉 操作成功",
			wantType:    "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildToastTrigger(tt.message, tt.toastType)

			// Verify the result is valid JSON
			var parsed map[string]map[string]string
			if err := json.Unmarshal([]byte(result), &parsed); err != nil {
				t.Fatalf("BuildToastTrigger() returned invalid JSON: %v\nGot: %s", err, result)
			}

			// Verify the structure
			showToast, ok := parsed["showToast"]
			if !ok {
				t.Fatalf("BuildToastTrigger() missing 'showToast' key in result: %s", result)
			}

			// Verify the message
			if gotMessage := showToast["message"]; gotMessage != tt.wantMessage {
				t.Errorf("BuildToastTrigger() message = %q, want %q", gotMessage, tt.wantMessage)
			}

			// Verify the type
			if gotType := showToast["type"]; gotType != tt.wantType {
				t.Errorf("BuildToastTrigger() type = %q, want %q", gotType, tt.wantType)
			}
		})
	}
}

func TestBuildToastTrigger_ValidJSON(t *testing.T) {
	// Test with various potentially problematic inputs to ensure valid JSON
	problemInputs := []struct {
		message   string
		toastType string
	}{
		{`"`, "error"},
		{`""`, "error"},
		{`\"`, "error"},
		{`\n\r\t`, "error"},
		{`{"nested": "json"}`, "error"},
		{`</script><script>alert("xss")</script>`, "error"},
		{`\u0000`, "error"}, // null byte
	}

	for _, input := range problemInputs {
		t.Run("valid_json_for_"+input.message, func(t *testing.T) {
			result := BuildToastTrigger(input.message, input.toastType)
			
			var parsed map[string]map[string]string
			if err := json.Unmarshal([]byte(result), &parsed); err != nil {
				t.Errorf("BuildToastTrigger(%q, %q) returned invalid JSON: %v\nGot: %s", 
					input.message, input.toastType, err, result)
			}
		})
	}
}

func TestBuildToastTrigger_NoInjection(t *testing.T) {
	// Test that injection attempts are properly escaped
	injectionAttempts := []string{
		`", "type": "injected"}}`,
		`\", "type": "injected"}}`,
		`","type":"injected"}},"extra":{"key":"value`,
	}

	for _, attempt := range injectionAttempts {
		t.Run("injection_attempt", func(t *testing.T) {
			result := BuildToastTrigger(attempt, "error")
			
			var parsed map[string]map[string]string
			if err := json.Unmarshal([]byte(result), &parsed); err != nil {
				t.Fatalf("BuildToastTrigger() returned invalid JSON: %v", err)
			}

			// Verify the injection attempt is in the message field, not breaking the structure
			if parsed["showToast"]["message"] != attempt {
				t.Errorf("Injection attempt not properly escaped: got %q, want %q", 
					parsed["showToast"]["message"], attempt)
			}

			// Verify type is still "error"
			if parsed["showToast"]["type"] != "error" {
				t.Errorf("Type was changed by injection: got %q, want %q", 
					parsed["showToast"]["type"], "error")
			}

			// Verify no extra keys were added
			if len(parsed) != 1 {
				t.Errorf("Extra keys added to result: got %d keys, want 1", len(parsed))
			}
			if len(parsed["showToast"]) != 2 {
				t.Errorf("Extra keys added to showToast: got %d keys, want 2", len(parsed["showToast"]))
			}
		})
	}
}
