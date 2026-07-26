package main

import (
	"encoding/json"
	"testing"
)

// parse is a small helper: unmarshal merged config and return the servers map
// under the client's key.
func servers(t *testing.T, b []byte, key string) map[string]any {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(b, &root); err != nil {
		t.Fatalf("merged config is not valid JSON: %v\n%s", err, b)
	}
	s, ok := root[key].(map[string]any)
	if !ok {
		t.Fatalf("no %q object in %s", key, b)
	}
	return s
}

func TestMergeMcpConfig_Fresh(t *testing.T) {
	out, err := mergeMcpConfig(nil, mcpClients["claude"])
	if err != nil {
		t.Fatal(err)
	}
	entry := servers(t, out, "mcpServers")["verifi"].(map[string]any)
	if entry["command"] != "verifi" {
		t.Errorf("command = %v, want verifi", entry["command"])
	}
	args := entry["args"].([]any)
	if len(args) != 1 || args[0] != "mcp" {
		t.Errorf("args = %v, want [mcp]", args)
	}
}

func TestMergeMcpConfig_PreservesOthers(t *testing.T) {
	existing := []byte(`{"mcpServers":{"other":{"command":"other-tool"}}}`)
	out, err := mergeMcpConfig(existing, mcpClients["cursor"])
	if err != nil {
		t.Fatal(err)
	}
	s := servers(t, out, "mcpServers")
	if _, ok := s["other"]; !ok {
		t.Error("merge dropped the existing 'other' server")
	}
	if _, ok := s["verifi"]; !ok {
		t.Error("merge did not add 'verifi'")
	}
}

func TestMergeMcpConfig_VscodeStdioTag(t *testing.T) {
	out, err := mergeMcpConfig(nil, mcpClients["vscode"])
	if err != nil {
		t.Fatal(err)
	}
	// VS Code uses "servers" and an explicit stdio type.
	entry := servers(t, out, "servers")["verifi"].(map[string]any)
	if entry["type"] != "stdio" {
		t.Errorf("vscode entry type = %v, want stdio", entry["type"])
	}
}

func TestMergeMcpConfig_InvalidJSON(t *testing.T) {
	if _, err := mergeMcpConfig([]byte("{not json"), mcpClients["claude"]); err == nil {
		t.Error("want an error on invalid existing config, got nil")
	}
}
