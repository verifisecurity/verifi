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
	{
		Name:        "propose_fix",
		Summary:     "Propose a gated fix for each vulnerable dependency, without applying it.",
		Description: "For each vulnerable dependency, return the recommended fix (upgrade or remove), the evidence behind it, and the gate decision: whether verifi may apply it once you confirm, or can only propose it. Read-only, writes nothing. npm. Needs the local OSV database (run `verifi update`).",
		CLICommand:  "fix",
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
		Name:        "apply_fix",
		Summary:     "Apply a confirmed, gate-approved fix to the workspace.",
		Description: "Apply the recommended fix by running the package manager, but only for fixes the gate authorizes to apply, and only when confirm is true. A fix the gate can only propose is never applied, and confirm must come from the human, not the model. Optionally restrict to one package. Not read-only: it changes the workspace. npm.",
		CLICommand:  "fix",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workspace": map[string]any{"type": "string", "description": "Project directory to fix"},
				"package":   map[string]any{"type": "string", "description": "Restrict the apply to this one package (default: all confirmable fixes)"},
				"confirm":   map[string]any{"type": "boolean", "description": "Must be true to write; the human's go-ahead, not the model's"},
				"db":        map[string]any{"type": "string", "description": "OSV database directory (default: the local ~/.verifi cache)"},
				"offline":   map[string]any{"type": "boolean", "description": "Skip the live registry check for published versions"},
			},
			"required": []string{"workspace", "confirm"},
		},
	},
}
