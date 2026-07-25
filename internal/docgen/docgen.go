// Package docgen keeps the drift-prone parts of the docs in sync with the code.
// It regenerates marker-delimited blocks from the command registry: the README
// (install, ecosystems, command reference) and the per-command docs pages
// (signature, flags, and a real captured example). The prose between markers is
// human owned and never touched. See ADR 0006.
package docgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/verifisecurity/verifi/internal/command"
	"github.com/verifisecurity/verifi/internal/ecosystem"
)

const installURL = "https://raw.githubusercontent.com/verifisecurity/verifi/main/install.sh"

// Generate regenerates the README and the per-command docs pages. examples maps
// a command name to its captured real output; commands without an example are
// left without one.
func Generate(readmePath, docsDir string, examples map[string]string) error {
	if err := updateReadme(readmePath); err != nil {
		return err
	}
	for _, c := range command.All {
		if !c.Ready {
			continue
		}
		page := filepath.Join(docsDir, "commands", c.Name+".md")
		if _, err := os.Stat(page); os.IsNotExist(err) {
			continue // no page authored for this command, skip
		}
		if err := updateCommandPage(page, c, examples[c.Name]); err != nil {
			return err
		}
	}
	return nil
}

func updateReadme(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out := string(data)
	for _, b := range []struct{ name, content string }{
		{"INSTALL", installBlock()},
		{"ECOSYSTEMS", ecosystemsBlock()},
		{"COMMANDS", commandsBlock()},
	} {
		out, err = replaceBlock(out, b.name, b.content, true)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	return writeIfChanged(path, data, out)
}

func updateCommandPage(path string, c command.Command, example string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out := string(data)
	blocks := []struct{ name, content string }{
		{"SIGNATURE", "```\nverifi " + c.Invocation() + "\n```"},
		{"FLAGS", flagsBlock(c)},
	}
	if example != "" {
		blocks = append(blocks, struct{ name, content string }{
			"EXAMPLE", "```\n" + strings.TrimRight(example, "\n") + "\n```",
		})
	}
	for _, b := range blocks {
		// Command-page blocks are optional: a page only carries the ones it needs.
		out, err = replaceBlock(out, b.name, b.content, false)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	return writeIfChanged(path, data, out)
}

// replaceBlock swaps the text between "<!-- BEGIN name -->" and "<!-- END name -->"
// for content. Idempotent. If required and the markers are missing, it errors;
// otherwise a missing block is left alone.
func replaceBlock(doc, name, content string, required bool) (string, error) {
	begin := "<!-- BEGIN " + name + " -->"
	end := "<!-- END " + name + " -->"
	i := strings.Index(doc, begin)
	j := strings.Index(doc, end)
	if i < 0 || j < 0 || j < i {
		if required {
			return "", fmt.Errorf("marker block %q not found", name)
		}
		return doc, nil
	}
	return doc[:i+len(begin)] + "\n" + content + "\n" + doc[j:], nil
}

func writeIfChanged(path string, old []byte, out string) error {
	if out == string(old) {
		return nil
	}
	return os.WriteFile(path, []byte(out), 0o644)
}

func flagsBlock(c command.Command) string {
	if len(c.Flags) == 0 {
		return "_None._"
	}
	w := 0
	for _, f := range c.Flags {
		if len(f.Name) > w {
			w = len(f.Name)
		}
	}
	var b strings.Builder
	b.WriteString("```\n")
	for _, f := range c.Flags {
		fmt.Fprintf(&b, "%-*s  %s\n", w, f.Name, f.Desc)
	}
	b.WriteString("```")
	return b.String()
}

func installBlock() string {
	return "```sh\ncurl -fsSL " + installURL + " | sh\n```"
}

func ecosystemsBlock() string {
	sup := ecosystem.Supported()
	w := 0
	for _, e := range sup {
		if len(e.Name) > w {
			w = len(e.Name)
		}
	}
	var b strings.Builder
	b.WriteString("```\n")
	for _, e := range sup {
		fmt.Fprintf(&b, "%-*s  %s\n", w, e.Name, e.Manifest)
	}
	b.WriteString("```")
	if ecosystem.HasPlanned() {
		b.WriteString("\n\nMore ecosystems are on the way.")
	}
	return b.String()
}

func commandsBlock() string {
	type row struct{ left, summary string }
	var rows []row
	var soon []string
	for _, c := range command.All {
		if !c.Ready {
			soon = append(soon, c.Name)
			continue
		}
		rows = append(rows, row{"verifi " + c.Invocation(), c.Summary})
	}
	w := 0
	for _, r := range rows {
		if len(r.left) > w {
			w = len(r.left)
		}
	}
	var b strings.Builder
	b.WriteString("```\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "%-*s  %s\n", w, r.left, r.summary)
	}
	b.WriteString("```")
	if len(soon) > 0 {
		fmt.Fprintf(&b, "\n\nComing soon: %s.", strings.Join(soon, ", "))
	}
	return b.String()
}
