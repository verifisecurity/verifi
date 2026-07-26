package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/verifisecurity/verifi/internal/gate"
	"github.com/verifisecurity/verifi/internal/plan"
)

// runFix implements `verifi fix <path> [--plan <file>]`: read the plan a scan
// wrote and run the fixes a human marked in it.
//
// It analyses nothing. Everything it needs is already in the plan, including the
// exact command per package, so what runs is what you reviewed rather than a
// fresh decision made at apply time. That is also why there is no preview flag:
// the plan is the preview.
//
// Two things must agree before anything is written. The gate has to have
// authorised the fix, and a human has to have marked it. Either alone does
// nothing.
func runFix(args []string) error {
	path, planPath := "", ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--plan":
			if i+1 >= len(args) {
				return fmt.Errorf("--plan needs a file path")
			}
			i++
			planPath = args[i]
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
	if planPath == "" {
		planPath = plan.Path(verifiHome(), path)
	}

	p, err := plan.Read(planPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("no plan at %s\nrun `verifi scan %s` first", planPath, path)
	}
	if err != nil {
		return err
	}

	// Marked by a human, then split on what the gate allows. A fix the gate only
	// proposes is reported with its reason rather than run: marking it says you
	// want it, but the gate is saying verifi cannot deliver it safely yet.
	var run, blocked []plan.Candidate
	for _, c := range p.Selected() {
		if c.Gate.Authorization == gate.Confirm {
			run = append(run, c)
		} else {
			blocked = append(blocked, c)
		}
	}

	if len(run) == 0 && len(blocked) == 0 {
		fmt.Printf("Nothing marked to apply in %s\n", planPath)
		if n := actionableCount(p); n > 0 {
			fmt.Printf("It has %d fix(es) available. Set \"apply\": true on the ones you want, then run this again.\n", n)
		}
		return nil
	}

	for _, c := range blocked {
		fmt.Printf("Skipping %s: marked, but verifi can only propose this one.\n", c.Package)
		for _, r := range c.Gate.Reasons {
			fmt.Printf("   %s\n", r)
		}
	}
	if len(run) == 0 {
		return nil
	}
	return applyCandidates(os.Stdout, os.Stderr, path, run)
}

func actionableCount(p plan.Plan) int {
	n := 0
	for _, c := range p.Candidates {
		if c.Actionable() {
			n++
		}
	}
	return n
}

// applyCandidates runs each candidate's recorded package-manager command in the
// workspace, streaming output. It takes writers so a caller can capture the run
// instead of streaming it. It writes to the project; everything that decides
// whether that is allowed happened before this point.
func applyCandidates(out, errw io.Writer, path string, cands []plan.Candidate) error {
	for _, c := range cands {
		fmt.Fprintf(out, "%s\n  $ %s\n", actionLine(c), join(c.Command))
		for _, w := range c.Gate.Warnings {
			fmt.Fprintf(errw, "verifi: %s\n", w)
		}
		cmd := exec.Command(c.Command[0], c.Command[1:]...)
		cmd.Dir = path
		cmd.Stdout = out
		cmd.Stderr = errw
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("apply %s: %w", c.Package, err)
		}
	}
	fmt.Fprintln(out, "\nDone. Re-run `verifi scan` to confirm, and run your tests.")
	return nil
}

func actionLine(c plan.Candidate) string {
	if c.Action == "remove" {
		return "remove " + c.Package
	}
	return fmt.Sprintf("upgrade %s: %s -> %s", c.Package, c.Current, c.Target)
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
