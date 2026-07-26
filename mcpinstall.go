package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/verifisecurity/verifi/internal/capability"
)

// mcpClient is an editor that speaks MCP, and where its project-scoped config
// lives. Paths are relative to the current directory: install wires verifi for
// this workspace, the IDE-first default, and never touches global state.
type mcpClient struct {
	key      string // the JSON object that holds servers
	path     string // config file, relative to the workspace
	stdioTag bool   // the client wants an explicit "type": "stdio"
}

var mcpClients = map[string]mcpClient{
	"claude": {key: "mcpServers", path: ".mcp.json"},
	"cursor": {key: "mcpServers", path: ".cursor/mcp.json"},
	"vscode": {key: "servers", path: ".vscode/mcp.json", stdioTag: true},
}

// runMcpInstall wires verifi into an editor by adding its stdio server to the
// client's project config, keeping any servers already there.
func runMcpInstall(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: verifi mcp install <client> (one of: %s)", strings.Join(clientNames(), ", "))
	}
	name := args[0]
	cl, ok := mcpClients[name]
	if !ok {
		return fmt.Errorf("unknown client %q, want one of: %s", name, strings.Join(clientNames(), ", "))
	}

	existing, err := os.ReadFile(cl.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	merged, err := mergeMcpConfig(existing, cl)
	if err != nil {
		return fmt.Errorf("%s: %w", cl.path, err)
	}
	if dir := filepath.Dir(cl.path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	if err := os.WriteFile(cl.path, merged, 0o644); err != nil {
		return err
	}

	fmt.Printf("Wrote %s. Restart %s to pick up verifi.\n", cl.path, name)
	fmt.Println("Tools your editor will discover:")
	for _, c := range capability.All {
		fmt.Printf("  %-18s %s\n", c.Name, c.Summary)
	}
	return nil
}

// mergeMcpConfig adds verifi's stdio server entry to an existing config, or a
// fresh one, preserving any other servers already present.
func mergeMcpConfig(existing []byte, cl mcpClient) ([]byte, error) {
	root := map[string]any{}
	if len(strings.TrimSpace(string(existing))) > 0 {
		if err := json.Unmarshal(existing, &root); err != nil {
			return nil, fmt.Errorf("existing config is not valid JSON: %w", err)
		}
	}
	servers, _ := root[cl.key].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	entry := map[string]any{"command": "verifi", "args": []string{"mcp"}}
	if cl.stdioTag {
		entry["type"] = "stdio"
	}
	servers["verifi"] = entry
	root[cl.key] = servers

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

func clientNames() []string {
	names := make([]string, 0, len(mcpClients))
	for n := range mcpClients {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
