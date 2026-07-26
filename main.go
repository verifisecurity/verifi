// Command verifi is the Verifi CLI: the open-source fix layer for the software
// supply chain. It takes what is flagged, decides what matters by policy, and
// drives the fix. This is a pre-release build, so most commands are placeholders
// while the tool is built in the open. Running it shows a welcome splash.
package main

import (
	"fmt"
	"os"

	"github.com/verifisecurity/verifi/internal/command"
	"github.com/verifisecurity/verifi/internal/splash"
)

// Set at build time via -ldflags (see .goreleaser.yaml).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	args := os.Args[1:]
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}

	switch cmd {
	case "", "--static", "--loop":
		// No command (optionally with a splash flag): say hello with the splash.
		// It renders the same version this binary reports, so the banner and
		// `verifi version` cannot disagree.
		splash.Version = version
		loop := hasFlag(args, "--loop")
		static := hasFlag(args, "--static")
		if static {
			splash.Static()
			return
		}
		splash.Show(loop)

	case "version", "-v", "--version":
		fmt.Printf("verifi %s (%s), built %s\n", version, commit, date)

	case "scan":
		if err := runScan(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "verifi:", err)
			os.Exit(1)
		}

	case "fix":
		if err := runFix(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "verifi:", err)
			os.Exit(1)
		}

	case "help", "-h", "--help":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "verifi: unknown command %q\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func usage() {
	fmt.Println("verifi: the fix layer for your software supply chain (pre-release)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  verifi [command]")
	fmt.Println()
	fmt.Println("Commands:")
	w := 0
	for _, c := range command.All {
		if n := len(c.Invocation()); n > w {
			w = n
		}
	}
	for _, c := range command.All {
		summary := c.Summary
		if !c.Ready {
			summary += " (coming soon)"
		}
		fmt.Printf("  %-*s  %s\n", w, c.Invocation(), summary)
	}
	fmt.Println()
	fmt.Println("Run verifi with no command to see the splash. Flags for it:")
	fmt.Println("  --static       Print the banner without animation")
	fmt.Println("  --loop         Replay the animation until interrupted")
	fmt.Println()
	fmt.Println("Learn more: https://verifisecurity.com")
}
