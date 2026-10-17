package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Launchers supply defaults only for integrations that opt into them.
func readInitializationSettings() (map[string]json.RawMessage, error) {
	raw := os.Getenv("JAVA_DEV_TOOLS_JDTLS_SETTINGS")
	if raw == "" {
		return nil, nil
	}
	var settings map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &settings); err != nil || settings == nil {
		return nil, fmt.Errorf("JAVA_DEV_TOOLS_JDTLS_SETTINGS must be a JSON object")
	}
	return settings, nil
}

// JDTLS accepts dotted settings keys. Flatten both forms so a client's nested
// preferences and its dotted keys have the same precedence over defaults.
func flattenSettings(settings map[string]json.RawMessage) map[string]json.RawMessage {
	flat := make(map[string]json.RawMessage)
	var visit func(string, map[string]json.RawMessage)
	visit = func(prefix string, values map[string]json.RawMessage) {
		// Explicit dotted keys win over the corresponding nested representation.
		for _, dotted := range []bool{false, true} {
			for key, value := range values {
				if strings.Contains(key, ".") != dotted {
					continue
				}
				var nested map[string]json.RawMessage
				if json.Unmarshal(value, &nested) == nil && len(nested) > 0 {
					visit(prefix+key+".", nested)
				} else {
					flat[prefix+key] = value
				}
			}
		}
	}
	visit("", settings)
	return flat
}

func mergeInitializationSettings(defaults map[string]json.RawMessage, client json.RawMessage) json.RawMessage {
	var supplied map[string]json.RawMessage
	_ = json.Unmarshal(client, &supplied)
	merged := flattenSettings(defaults)
	for key, value := range flattenSettings(supplied) {
		// Keep the adapter's required init script and the client's extra arguments.
		if key == "java.import.gradle.arguments" && merged[key] != nil {
			var required, extra []string
			if json.Unmarshal(merged[key], &required) == nil && json.Unmarshal(value, &extra) == nil {
				present := false
				for start := 0; start+len(required) <= len(extra); start++ {
					matches := true
					for index := range required {
						matches = matches && required[index] == extra[start+index]
					}
					present = present || matches
				}
				if !present {
					extra = append(required, extra...)
				}
				value, _ = json.Marshal(extra)
			}
		}
		merged[key] = value
	}
	result, _ := json.Marshal(merged)
	return result
}

// Clients can resend preferences after initialize; retain the opted-in adapter.
func configureSettingsNotification(request map[string]json.RawMessage, defaults map[string]json.RawMessage) bool {
	var method string
	_ = json.Unmarshal(request["method"], &method)
	if len(defaults) == 0 || method != "workspace/didChangeConfiguration" {
		return false
	}
	var params map[string]json.RawMessage
	_ = json.Unmarshal(request["params"], &params)
	if params == nil {
		params = make(map[string]json.RawMessage)
	}
	params["settings"] = mergeInitializationSettings(defaults, params["settings"])
	request["params"], _ = json.Marshal(params)
	return true
}
