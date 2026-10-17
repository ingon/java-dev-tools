package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const fixtureURI = "jdt://contents/totipo-proxy-test/Fixture.class?source=fixture"
const fixtureSource = "class Fixture { String s = \"literal \\n and \\t\"; }\n"

// Separate processes exercise the real proxy's pipes, framing, and shutdown.
func TestProxyProcess(t *testing.T) {
	if os.Getenv("TOTIPO_PROXY_PROCESS") != "1" {
		return
	}
	if err := run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestServerProcess(t *testing.T) {
	if os.Getenv("TOTIPO_SERVER_PROCESS") != "1" {
		return
	}
	r := bufio.NewReader(os.Stdin)
	var classRequest, probe map[string]any
	send := func(id, result any) {
		body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
		fmt.Fprintf(os.Stdout, "Content-Length: %d\r\n\r\n%s", len(body), body)
	}
	for {
		body, err := readLSPMessage(r)
		if err != nil {
			os.Exit(1)
		}
		var request map[string]any
		if json.Unmarshal(body, &request) != nil {
			os.Exit(1)
		}
		switch request["method"] {
		case "initialize":
			send(request["id"], map[string]any{"capabilities": map[string]any{}})
		case "textDocument/definition":
			point := map[string]any{"line": 0, "character": 0}
			send(request["id"], []any{map[string]any{"uri": fixtureURI, "range": map[string]any{"start": point, "end": point}}})
		case "java/classFileContents":
			classRequest = request
		case "test/probe":
			probe = request
		case "test/echo":
			send(request["id"], request["params"])
		case "shutdown":
			send(request["id"], nil)
		case "exit":
			os.Exit(0)
		}
		// Deliberately overlap an editor request with the proxy's internal one.
		if classRequest != nil && probe != nil {
			send(probe["id"], map[string]any{"probe": "preserved"})
			send(classRequest["id"], fixtureSource)
			classRequest, probe = nil, nil
		}
	}
}

func TestVirtualSourceNavigation(t *testing.T) {
	dir := t.TempDir()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	shell := os.Getenv("TEST_SHELL")
	if shell == "" {
		t.Fatal("TEST_SHELL must name the available shell")
	}
	stub := "#!" + shell + "\nexport TOTIPO_SERVER_PROCESS=1\nexec '" + strings.ReplaceAll(exe, "'", "'\\''") + "' -test.run=^TestServerProcess$\n"
	if err := os.WriteFile(filepath.Join(dir, "jdtls"), []byte(stub), 0700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=^TestProxyProcess$")
	cmd.Env = append(os.Environ(), "TOTIPO_PROXY_PROCESS=1", "PATH="+dir+":"+os.Getenv("PATH"))
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })
	timer := time.AfterFunc(10*time.Second, func() { _ = cmd.Process.Kill() })
	defer timer.Stop()
	r := bufio.NewReader(stdout)
	send := func(id int, method string, params any) {
		body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
		if _, err := fmt.Fprintf(stdin, "Content-Length: %d\r\n\r\n%s", len(body), body); err != nil {
			t.Fatal(err)
		}
	}
	receive := func() map[string]any {
		body, err := readLSPMessage(r)
		if err != nil {
			t.Fatal(err)
		}
		var response map[string]any
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatal(err)
		}
		return response
	}
	send(0, "initialize", map[string]any{})
	if receive()["id"] != float64(0) {
		t.Fatal("initialization was not forwarded")
	}
	send(10, "textDocument/definition", map[string]any{})
	send(11, "test/probe", map[string]any{})
	responses := map[float64]map[string]any{}
	for i := 0; i < 2; i++ {
		response := receive()
		responses[response["id"].(float64)] = response
	}
	if !reflect.DeepEqual(responses[11]["result"], map[string]any{"probe": "preserved"}) {
		t.Fatal("proxy and editor request IDs collided")
	}
	results := responses[10]["result"].([]any)
	uri := results[0].(map[string]any)["targetUri"].(string)
	path := strings.TrimPrefix(uri, "file://")
	t.Cleanup(func() { _ = os.Remove(path) })
	content, err := os.ReadFile(path)
	if err != nil || string(content) != fixtureSource {
		t.Fatalf("source text changed: %q, %v", content, err)
	}
	params := map[string]any{"textDocument": map[string]any{"uri": uri, "version": float64(3), "text": fixtureSource}, "position": map[string]any{"line": float64(2), "character": float64(4)}, "extra": "preserve"}
	send(12, "test/echo", params)
	params["textDocument"].(map[string]any)["uri"] = fixtureURI
	if !reflect.DeepEqual(receive()["result"], params) {
		t.Fatal("cached-source request lost fields or did not restore its jdt URI")
	}
	send(20, "shutdown", nil)
	_ = receive()
	send(21, "exit", nil)
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestReadOnlySourceReplacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "package", "Source.java")
	if err := writeSource(path, "first"); err != nil {
		t.Fatal(err)
	}
	if err := writeSource(path, "replacement"); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil || string(content) != "replacement" {
		t.Fatalf("stale cached source: %q, %v", content, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0400 {
		t.Fatal("cached source must be read-only")
	}
}

func TestAutomaticSourceCapability(t *testing.T) {
	for _, input := range []string{
		`{"method":"initialize"}`,
		`{"method":"initialize","params":{"initializationOptions":null}}`,
		`{"method":"initialize","id":17,"params":{"rootUri":"file:///example","initializationOptions":{"settings":{"java":{"import":{"gradle":{"arguments":["--offline"]}}}},"extendedClientCapabilities":{"classFileContentsSupport":false,"other":true},"unknown":"preserve"}}}`,
	} {
		var request map[string]json.RawMessage
		if err := json.Unmarshal([]byte(input), &request); err != nil {
			t.Fatal(err)
		}
		if !enableSourceCapability(request) {
			t.Fatal("initialization capability was not added")
		}
		body, _ := json.Marshal(request)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatal(err)
		}
		options := got["params"].(map[string]any)["initializationOptions"].(map[string]any)
		capabilities := options["extendedClientCapabilities"].(map[string]any)
		if capabilities["classFileContentsSupport"] != true {
			t.Fatal("class-file content support is missing")
		}
		if got["id"] != nil {
			if got["id"] != float64(17) || got["params"].(map[string]any)["rootUri"] != "file:///example" || options["unknown"] != "preserve" || capabilities["other"] != true {
				t.Fatal("initialization fields were lost")
			}
			settings := options["settings"].(map[string]any)
			arguments := settings["java"].(map[string]any)["import"].(map[string]any)["gradle"].(map[string]any)["arguments"]
			if !reflect.DeepEqual(arguments, []any{"--offline"}) {
				t.Fatal("Java settings were changed")
			}
		}
	}
	var ordinary map[string]json.RawMessage
	_ = json.Unmarshal([]byte(`{"id":20,"method":"shutdown"}`), &ordinary)
	before, _ := json.Marshal(ordinary)
	if enableSourceCapability(ordinary) {
		t.Fatal("non-initialization request was changed")
	}
	after, _ := json.Marshal(ordinary)
	if string(before) != string(after) {
		t.Fatal("non-initialization request was changed")
	}
}
