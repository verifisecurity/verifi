// Command verifi is the Verifi CLI: an open-source tool to find and fix risky
// dependencies. This is a pre-release build — most commands are placeholders
// while the tool is being built in the open. Running it shows a welcome splash.
package main

import (
	"fmt"
	"os"

	"github.com/verifi-security-platform/verifi-cli/internal/splash"
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

	case "scan", "status", "fix":
		fmt.Printf("`verifi %s` is coming in a future release. This is a pre-release build.\n", cmd)
		fmt.Println("Follow along: https://github.com/verifi-security-platform/verifi-cli")

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
	fmt.Print(`verifi — find and fix risky dependencies (pre-release)

Usage:
  verifi [command]

Commands:
  welcome        Show the welcome splash (default)
  scan <path>    Scan a project's dependencies (coming soon)
  status         Show findings for the current project (coming soon)
  version        Print the version
  help           Show this help

Flags (welcome):
  --static       Print the banner without animation
  --loop         Replay the animation until interrupted

Learn more: https://verifisecurity.com
`)
}
