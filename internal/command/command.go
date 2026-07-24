// Package command is the single source of truth for the CLI's command list.
// Both the help text and the generated docs render from it, so they cannot
// drift. Adding a command here is what makes it appear in `verifi --help` and
// in the README, in step.
package command

// Command is one CLI command.
type Command struct {
	Name    string
	Args    string // argument placeholder, e.g. "<path>", or "" for none
	Summary string
	Ready   bool // false = placeholder, shown as coming soon, hidden from the README reference
}

// All is every command, in display order.
var All = []Command{
	{"welcome", "", "Show the welcome splash (default)", true},
	{"inspect", "<path>", "Resolve the project's dependencies (--json, --sbom)", true},
	{"update", "", "Download the OSV database into the local cache", true},
	{"status", "<path>", "Show what is vulnerable and the fix (--json, --db, --offline)", true},
	{"fix", "<path>", "Apply a safe fix or open a PR for review", false},
	{"version", "", "Print the version", true},
	{"help", "", "Show this help", true},
}

// Invocation is "name" or "name args", e.g. "status <path>".
func (c Command) Invocation() string {
	if c.Args == "" {
		return c.Name
	}
	return c.Name + " " + c.Args
}
