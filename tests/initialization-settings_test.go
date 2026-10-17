package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestInitializationEnvironment(t *testing.T) {
	for _, invalid := range []string{"null", "[]", "true", "bad JSON"} {
		t.Setenv("JAVA_DEV_TOOLS_JDTLS_SETTINGS", invalid)
		if _, err := readInitializationSettings(); err == nil {
			t.Fatalf("accepted invalid defaults %q", invalid)
		}
	}
	t.Setenv("JAVA_DEV_TOOLS_JDTLS_SETTINGS", "")
	defaults, err := readInitializationSettings()
	if err != nil || defaults != nil {
		t.Fatal("ordinary Java launch must have no injected settings")
	}
	t.Setenv("JAVA_DEV_TOOLS_JDTLS_SETTINGS", `{"java.import.gradle.arguments":["--init-script","/adapter.gradle"]}`)
	defaults, err = readInitializationSettings()
	if err != nil || len(defaults) != 1 {
		t.Fatal("valid defaults were not read", err)
	}
}

func TestAndroidInitializationSettings(t *testing.T) {
	var defaults map[string]json.RawMessage
	_ = json.Unmarshal([]byte(`{"java.import.gradle.arguments":["--init-script","/adapter.gradle"],"java.configuration.updateBuildConfiguration":"automatic"}`), &defaults)
	for _, client := range []string{
		`{"java":{"import":{"gradle":{"arguments":["--offline"]}},"configuration":{"updateBuildConfiguration":"interactive"},"completion":{"enabled":false}}}`,
		`{"java.import.gradle.arguments":["--offline"],"java.configuration.updateBuildConfiguration":"interactive","java.completion.enabled":false}`,
	} {
		var request map[string]json.RawMessage
		_ = json.Unmarshal([]byte(`{"method":"initialize","id":37,"params":{"rootUri":"file:///project","initializationOptions":{"unknown":"keep","extendedClientCapabilities":{"other":true},"settings":`+client+`}}}`), &request)
		if !enableSourceCapability(request, defaults) {
			t.Fatal("initialization not handled")
		}
		body, _ := json.Marshal(request)
		var got map[string]any
		_ = json.Unmarshal(body, &got)
		params := got["params"].(map[string]any)
		options := params["initializationOptions"].(map[string]any)
		settings := options["settings"].(map[string]any)
		if !reflect.DeepEqual(settings["java.import.gradle.arguments"], []any{"--init-script", "/adapter.gradle", "--offline"}) {
			t.Fatal("required adapter or client arguments lost", settings)
		}
		if settings["java.configuration.updateBuildConfiguration"] != "interactive" || settings["java.completion.enabled"] != false {
			t.Fatal("client preferences did not override defaults", settings)
		}
		caps := options["extendedClientCapabilities"].(map[string]any)
		if got["id"] != float64(37) || params["rootUri"] != "file:///project" || options["unknown"] != "keep" || caps["other"] != true || caps["classFileContentsSupport"] != true {
			t.Fatal("unrelated initialization fields changed", got)
		}
	}
	for _, settings := range []json.RawMessage{nil, json.RawMessage("null"), json.RawMessage("{} ")} {
		var got map[string]any
		if err := json.Unmarshal(mergeInitializationSettings(defaults, settings), &got); err != nil || len(got) != 2 {
			t.Fatal("defaults missing for empty client options", got, err)
		}
	}
}

func TestDottedSettingsPrecedence(t *testing.T) {
	var settings map[string]json.RawMessage
	_ = json.Unmarshal([]byte(`{"java":{"completion":{"enabled":true}},"java.completion.enabled":false,"java.import.gradle":{"arguments":["--offline"]}}`), &settings)
	flat := flattenSettings(settings)
	if string(flat["java.completion.enabled"]) != "false" || string(flat["java.import.gradle.arguments"]) != `["--offline"]` {
		t.Fatal("mixed nested/dotted preferences were not flattened consistently", flat)
	}
}

func TestSettingsUpdatesRetainAdapter(t *testing.T) {
	var defaults map[string]json.RawMessage
	_ = json.Unmarshal([]byte(`{"java.import.gradle.arguments":["--init-script","/adapter.gradle"]}`), &defaults)
	for _, arguments := range []string{`["--offline"]`, `["--init-script","/adapter.gradle","--offline"]`} {
		var request map[string]json.RawMessage
		_ = json.Unmarshal([]byte(`{"method":"workspace/didChangeConfiguration","params":{"unknown":true,"settings":{"java.import.gradle.arguments":`+arguments+`}}}`), &request)
		if !configureSettingsNotification(request, defaults) {
			t.Fatal("settings notification not handled")
		}
		var params map[string]json.RawMessage
		_ = json.Unmarshal(request["params"], &params)
		var settings map[string]any
		_ = json.Unmarshal(params["settings"], &settings)
		if !reflect.DeepEqual(settings["java.import.gradle.arguments"], []any{"--init-script", "/adapter.gradle", "--offline"}) || string(params["unknown"]) != "true" {
			t.Fatal("adapter duplicated/lost or fields changed", settings)
		}
		before, _ := json.Marshal(request)
		if configureSettingsNotification(request, nil) {
			t.Fatal("ordinary Java notification modified")
		}
		after, _ := json.Marshal(request)
		if string(before) != string(after) {
			t.Fatal("ordinary Java notification modified")
		}
	}
}
