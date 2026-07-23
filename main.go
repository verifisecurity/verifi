// Command verifi is the Verifi CLI: the open-source fix layer for the software
// supply chain. It takes what is flagged, decides what matters by policy, and
// drives the fix. This is a pre-release build, so most commands are placeholders
// while the tool is built in the open. Running it shows a welcome splash.
package main

import (
	"fmt"
	"os"

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
	case "", "welcome":
		loop := hasFlag(args, "--loop")
		static := hasFlag(args, "--static")
		if static {
			splash.Static()
			return
		}
		splash.Show(loop)

	case "version", "-v", "--version":
		fmt.Printf("verifi %s (%s), built %s\n", version, commit, date)

	case "inspect":
		if err := runInspect(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "verifi:", err)
			os.Exit(1)
		}

	case "status":
		if err := runStatus(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "verifi:", err)
			os.Exit(1)
		}

	case "scan", "fix":
		fmt.Printf("`verifi %s` is coming in a future release. This is a pre-release build.\n", cmd)
		fmt.Println("Follow along: https://github.com/verifisecurity/verifi")

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
	fmt.Print(`verifi: the fix layer for your software supply chain (pre-release)

Usage:
  verifi [command]

Commands:
  welcome        Show the welcome splash (default)
  inspect <path> Resolve the project's dependencies (--json, --sbom)
  status <path>  Show what is vulnerable, matched against OSV (--json, --db <dir>)
  fix <path>     Decide what matters, open fixes, gate the rest (coming soon)
  version        Print the version
  help           Show this help

Flags (welcome):
  --static       Print the banner without animation
  --loop         Replay the animation until interrupted

Learn more: https://verifisecurity.com
`)
}
