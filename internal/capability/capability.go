// Package capability is the single source of truth for what verifi can do,
// named by developer intent rather than by CLI verb. The MCP server, the editor
// install command, and the generated docs all derive from this one list, so a
// change here flows to every surface and they cannot drift. Handlers live with
// the command layer (package main) and bind by Name; this package is pure
// metadata so any tool, including the docs generator, can import it.
package capability

// Capability is one thing verifi exposes.
type Capability struct {
	Name        string         // intent name, and the MCP tool name
	Summary     string         // one line, for docs and help
	Description string         // detailed, shown to an MCP client
	CLICommand  string         // the verifi command that runs the same core path
	InputSchema map[string]any // JSON Schema for the tool's arguments
}

// All is every capability, in display order. Add a capability here and it
// appears as an MCP tool, in the install listing, and in the docs, in step.
var All = []Capability{
	{
		Name:        "scan_workspace",
		Summary:     "Scan the project for vulnerable dependencies and their fixes.",
		Description: "Scan the project for vulnerable dependencies. Returns each finding: the package, its advisories, and the recommended fix. Read-only, npm. Needs the local OSV database (run `verifi update`).",
		CLICommand:  "status",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workspace": map[string]any{"type": "string", "description": "Project directory to scan"},
				"db":        map[string]any{"type": "string", "description": "OSV database directory (default: the local ~/.verifi cache)"},
				"offline":   map[string]any{"type": "boolean", "description": "Skip the live registry check for published versions"},
			},
			"required": []string{"workspace"},
		},
	},
	{
		Name:        "list_dependencies",
		Summary:     "Resolve the full dependency tree, as inventory or SBOM.",
		Description: "Resolve the project's full dependency tree, direct and transitive, into a structured inventory, or a CycloneDX SBOM when sbom is true. Read-only, npm.",
		CLICommand:  "inspect",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workspace": map[string]any{"type": "string", "description": "Project directory containing package-lock.json"},
				"sbom":      map[string]any{"type": "boolean", "description": "Return a CycloneDX SBOM instead of the inventory JSON"},
			},
			"required": []string{"workspace"},
		},
	},
}
