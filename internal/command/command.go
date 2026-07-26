// Package command is the single source of truth for the CLI's commands. The
// help text and the generated docs both render from it, so they cannot drift.
// Adding a command, flag, or example here is what makes it appear in
// `verifi --help` and in the docs, in step.
package command

// Flag is one command-line flag.
type Flag struct {
	Name string
	Desc string
}

// Command is one CLI command.
type Command struct {
	Name    string
	Args    string // argument placeholder, e.g. "<path>", or "" for none
	Summary string // one line, used in help and the README reference
	Ready   bool   // false = placeholder, shown as coming soon, hidden from docs

	// For the detailed docs pages.
	Long    string   // a paragraph explaining what the command does
	Flags   []Flag   // flags to document
	Example []string // args to run to capture a real example, nil = no live example
}

// All is every command, in display order. Running verifi with no command shows
// the splash, so "welcome" is not a command, it is the default.
var All = []Command{
	{
		Name: "inspect", Args: "<path>", Ready: true,
		Summary: "Resolve the project's dependencies (--json, --sbom)",
		Long:    "Resolves a project's full dependency tree, direct and transitive, into a structured inventory. It reads the lockfile and does not install or run anything. Use it to see what you actually depend on, or to export a CycloneDX SBOM.",
		Flags: []Flag{
			{"--json", "Print the inventory as JSON"},
			{"--sbom", "Print a CycloneDX SBOM"},
		},
		Example: []string{"inspect", "testdata/npm/vuln"},
	},
	{
		Name: "update", Ready: true,
		Summary: "Download the OSV database into the local cache",
		Long:    "Downloads the OSV advisory database for an ecosystem into a local cache (~/.verifi/osv). Run it once, from anywhere; it is machine-global, not per-project. After it, status matches offline. Refresh it occasionally to pick up new advisories.",
		Flags: []Flag{
			{"--ecosystem <name>", "Ecosystem to download (default npm)"},
		},
		Example: nil, // network download, not captured
	},
	{
		Name: "status", Args: "<path>", Ready: true,
		Summary: "Show what is vulnerable and the fix (--json, --db, --offline)",
		Long:    "Scans a project against the OSV database and prints, for each vulnerable package, whether your code imports it, the version to upgrade to, what that clears, and the honest limits of what has been checked. Read-only, it never changes your project. Run `verifi update` first.",
		Flags: []Flag{
			{"--json", "Print findings as JSON"},
			{"--db <dir>", "Use a specific OSV database directory"},
			{"--offline", "Skip the registry check for published versions"},
		},
		Example: []string{"status", "testdata/npm/vuln", "--db", "testdata/osv", "--offline"},
	},
	{
		Name: "fix", Args: "<path>", Ready: true,
		Summary: "Apply the recommended fix, or preview it (--apply, --db, --offline)",
		Long:    "Applies the fix that status recommends: upgrade to a safe version, or remove a direct dependency your code never imports. It previews by default and writes nothing; --apply runs the package manager to make the change. At today's advisory confidence, applying is an explicit opt-in, never silent.",
		Flags: []Flag{
			{"--apply", "Write the change via the package manager (default is preview)"},
			{"--db <dir>", "Use a specific OSV database directory"},
			{"--offline", "Skip the registry check for published versions"},
		},
		Example: []string{"fix", "testdata/npm/vuln", "--db", "testdata/osv", "--offline"},
	},
	{
		Name: "version", Ready: true,
		Summary: "Print the version",
		Long:    "Prints the CLI version, commit, and build date.",
	},
	{
		Name: "help", Ready: true,
		Summary: "Show this help",
		Long:    "Prints the command list and usage.",
	},
}

// Invocation is "name" or "name args", e.g. "status <path>".
func (c Command) Invocation() string {
	if c.Args == "" {
		return c.Name
	}
	return c.Name + " " + c.Args
}
