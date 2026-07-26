package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/verifisecurity/verifi/internal/capability"
	"github.com/verifisecurity/verifi/internal/command"
)

var update = flag.Bool("update", false, "regenerate golden files")

// TestMcpServerMatchesCapabilities is the anti-drift guard: the MCP tools are
// exactly the capability registry, every capability has a handler, and every
// capability names a Ready CLI command, so both surfaces exist for each one. A
// new capability flows into MCP automatically; one missing a handler, or a CLI
// command, fails here rather than shipping broken.
func TestMcpServerMatchesCapabilities(t *testing.T) {
	var want []string
	for _, c := range capability.All {
		want = append(want, c.Name)
		if _, ok := capabilityHandlers[c.Name]; !ok {
			t.Errorf("capability %q has no handler in capabilityHandlers", c.Name)
		}
		if !readyCommand(c.CLICommand) {
			t.Errorf("capability %q names CLI command %q, which is missing or not Ready", c.Name, c.CLICommand)
		}
	}
	got := newMcpServer().ToolNames()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("MCP tools = %v, want the capabilities %v", got, want)
	}
}

func readyCommand(name string) bool {
	for _, c := range command.All {
		if c.Name == name {
			return c.Ready
		}
	}
	return false
}

// TestMcpServe_Golden drives the real server end to end over the npm fixture: a
// full session (initialize, tools/list, then scan_workspace and
// list_dependencies calls) asserted against a golden transcript. scan_workspace
// runs offline so it is deterministic and needs no network. Regenerate with:
// go test . -update
func TestMcpServe_Golden(t *testing.T) {
	reqs := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_dependencies","arguments":{"workspace":"testdata/npm/vuln"}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"scan_workspace","arguments":{"workspace":"testdata/npm/vuln","db":"testdata/osv","offline":true}}}`,
	}
	var out bytes.Buffer
	if err := newMcpServer().Serve(strings.NewReader(strings.Join(reqs, "\n")+"\n"), &out); err != nil {
		t.Fatalf("serve: %v", err)
	}
	got := out.Bytes()

	golden := filepath.Join("testdata", "mcp", "expected.transcript.jsonl")
	if *update {
		if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden (run: go test . -update): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("transcript mismatch\n got: %s\nwant: %s", got, want)
	}
}
