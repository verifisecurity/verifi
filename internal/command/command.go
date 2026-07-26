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
		Name: "scan", Args: "<path>", Ready: true,
		Summary: "Report what is vulnerable and the fix for each one",
		Long:    "Resolves a project's full dependency tree and matches it against the OSV advisory database, then reports for each vulnerable package whether your code imports it, the version to move to, what that clears, and the honest limits of what has been checked. It changes nothing in your project. Every scan writes a plan: each fix it found, with the evidence and the gate's verdict, every entry marked apply=false. Mark the ones you want and `verifi fix` applies those. Use --inventory or --sbom to describe the tree without looking for advisories; neither needs a database. The first scan needs the advisory database, which --download fetches into ~/.verifi/osv.",
		Flags: []Flag{
			{"--json", "Print the selected view as JSON"},
			{"--inventory", "Report the dependency tree instead of the findings"},
			{"--sbom", "Print the dependency tree as a CycloneDX SBOM"},
			{"--download", "Fetch the OSV database into the local cache first"},
			{"--db <dir>", "Use a specific OSV database directory"},
			{"--offline", "Skip the registry check for published versions"},
			{"--out <file>", "Write the plan here instead of the default location"},
		},
		Example: []string{"scan", "testdata/npm/vuln", "--db", "testdata/osv", "--offline"},
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
