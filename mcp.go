package main

import (
	"os"

	"github.com/verifisecurity/verifi/internal/capability"
	"github.com/verifisecurity/verifi/internal/mcp"
)

// runMcp implements `verifi mcp`: serve verifi's capabilities as MCP tools over
// stdio so an editor or agent can call them, or `verifi mcp install <client>` to
// wire it into an editor. It is a thin wrapper: each tool runs the same core
// path as the matching CLI command, so the CLI and the MCP surface cannot drift.
func runMcp(args []string) error {
	if len(args) > 0 && args[0] == "install" {
		return runMcpInstall(args[1:])
	}
	return newMcpServer().Serve(os.Stdin, os.Stdout)
}

// newMcpServer builds the server from the capability registry
// (internal/capability), the single source of truth for what verifi exposes.
// Handlers bind by name (capability.go). Kept separate from runMcp so tests can
// drive it without touching os.Stdin/os.Stdout.
func newMcpServer() *mcp.Server {
	s := mcp.NewServer("verifi", version)
	for _, c := range capability.All {
		h, ok := capabilityHandlers[c.Name]
		if !ok {
			continue // wiring bug, caught by TestMcpServerMatchesCapabilities
		}
		s.AddTool(mcp.Tool{
			Name:        c.Name,
			Description: c.Description,
			InputSchema: c.InputSchema,
			Handler:     h,
		})
	}
	return s
}
