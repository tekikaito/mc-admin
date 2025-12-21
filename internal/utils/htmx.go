package utils

import "encoding/json"

// BuildToastTrigger creates a properly escaped HX-Trigger header for toast notifications.
// It uses json.Marshal to ensure all special characters are safely escaped,
// preventing JSON injection vulnerabilities.
func BuildToastTrigger(message string, toastType string) string {
	trigger := map[string]map[string]string{
		"showToast": {
			"message": message,
			"type":    toastType,
		},
	}
	jsonBytes, err := json.Marshal(trigger)
	if err != nil {
		// Fallback to a safe default if marshaling fails
		return `{"showToast": {"message": "An error occurred", "type": "error"}}`
	}
	return string(jsonBytes)
}
