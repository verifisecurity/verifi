package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/verifisecurity/verifi/internal/fix"
)

// runFix implements `verifi fix <path> [--apply] [--db <dir>] [--offline]`:
// build the fix plan from what status recommends, and preview it (default) or
// apply it. Previewing writes nothing. Applying runs the package manager, and
// at today's advisory confidence it is an explicit opt-in, never silent.
func runFix(args []string) error {
	path, dbDir := "", ""
	apply, offline := false, false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--apply":
			apply = true
		case "--offline":
			offline = true
		case "--db":
			if i+1 >= len(args) {
				return fmt.Errorf("--db needs a directory")
			}
			i++
			dbDir = args[i]
		default:
			if len(a) > 0 && a[0] == '-' {
				return fmt.Errorf("unknown flag %q", a)
			}
			path = a
		}
	}
	if path == "" {
		path = "."
	}

	res, err := analyze(path, dbDir, offline)
	if err != nil {
		return err
	}
	if w := staleWarning(res.dbMeta); w != "" {
		fmt.Fprintln(os.Stderr, w)
	}
	plan := fix.Build(res.recs)

	if plan.Empty() {
		fmt.Println("Nothing to fix.")
		if len(plan.Skipped) > 0 {
			fmt.Printf("%d vulnerable package(s) have no applicable fix yet.\n", len(plan.Skipped))
		}
		return nil
	}

	if !apply {
		printPlan(plan)
		return nil
	}
	return applyPlan(path, plan)
}

func printPlan(plan fix.Plan) {
	fmt.Printf("Planned fixes (%d). Nothing is written without --apply.\n\n", len(plan.Actions))
	for _, a := range plan.Actions {
		fmt.Printf("  %s\n", actionLine(a))
		fmt.Printf("      %s\n", a.Reason)
		fmt.Printf("      $ %s\n\n", join(a.Command))
	}
	if len(plan.Skipped) > 0 {
		fmt.Printf("No applicable fix yet: %s\n\n", join(plan.Skipped))
	}
	fmt.Println("Apply with:  verifi fix <path> --apply")
}

func actionLine(a fix.Action) string {
	if a.Kind == "remove" {
		return "remove " + a.Name
	}
	return fmt.Sprintf("upgrade %s: %s -> %s", a.Name, a.From, a.To)
}

func applyPlan(path string, plan fix.Plan) error {
	return applyPlanTo(os.Stdout, os.Stderr, path, plan)
}

// applyPlanTo runs each action's package-manager command in path, streaming
// output to out and errw. It is the shared apply core, taking writers so a
// caller can capture the result instead of streaming it to the terminal. It
// writes to the project; the caller decides whether that is allowed.
func applyPlanTo(out, errw io.Writer, path string, plan fix.Plan) error {
	for _, a := range plan.Actions {
		fmt.Fprintf(out, "%s\n  $ %s\n", actionLine(a), join(a.Command))
		cmd := exec.Command(a.Command[0], a.Command[1:]...)
		cmd.Dir = path
		cmd.Stdout = out
		cmd.Stderr = errw
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("apply %s: %w", a.Name, err)
		}
	}
	fmt.Fprintln(out, "\nDone. Re-run `verifi status` to confirm, and run your tests.")
	return nil
}

func join(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += " "
		}
		out += p
	}
	return out
}
