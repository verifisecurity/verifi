// Package docgen keeps the drift-prone parts of the docs in sync with the code.
// It regenerates marker-delimited blocks (the command reference, the install
// snippet) from the command registry, so a command added in code shows up in
// the README without anyone editing prose. The prose between markers is human
// owned and never touched.
package docgen

import (
	"fmt"
	"os"
	"strings"

	"github.com/verifisecurity/verifi/internal/command"
)

// installURL is the canonical one-line installer.
const installURL = "https://raw.githubusercontent.com/verifisecurity/verifi/main/install.sh"

// UpdateReadme regenerates the generated blocks in the README at path, in place.
func UpdateReadme(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out := string(data)
	for _, b := range []struct {
		name, content string
	}{
		{"INSTALL", installBlock()},
		{"COMMANDS", commandsBlock()},
	} {
		out, err = replaceBlock(out, b.name, b.content)
		if err != nil {
			return err
		}
	}
	if out == string(data) {
		return nil
	}
	return os.WriteFile(path, []byte(out), 0o644)
}

// replaceBlock swaps the text between "<!-- BEGIN name -->" and "<!-- END name -->"
// for content. It is idempotent and errors if the markers are missing.
func replaceBlock(doc, name, content string) (string, error) {
	begin := "<!-- BEGIN " + name + " -->"
	end := "<!-- END " + name + " -->"
	i := strings.Index(doc, begin)
	j := strings.Index(doc, end)
	if i < 0 || j < 0 || j < i {
		return "", fmt.Errorf("marker block %q not found", name)
	}
	return doc[:i+len(begin)] + "\n" + content + "\n" + doc[j:], nil
}

func installBlock() string {
	return "```sh\ncurl -fsSL " + installURL + " | sh\n```"
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
