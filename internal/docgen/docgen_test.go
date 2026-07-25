package docgen

import (
	"strings"
	"testing"
)

func TestReplaceBlock(t *testing.T) {
	doc := "top\n<!-- BEGIN X -->\nold content\n<!-- END X -->\nbottom"
	want := "top\n<!-- BEGIN X -->\nnew\n<!-- END X -->\nbottom"

	got, err := replaceBlock(doc, "X", "new", true)
	if err != nil {
		t.Fatalf("replaceBlock: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// Idempotent: replacing again with the same content is a no-op.
	again, err := replaceBlock(got, "X", "new", true)
	if err != nil || again != want {
		t.Errorf("not idempotent: got %q err %v", again, err)
	}
}

func TestReplaceBlock_MissingMarker(t *testing.T) {
	// Required: a missing marker is an error.
	if _, err := replaceBlock("no markers here", "X", "y", true); err != nil {
		// expected
	} else {
		t.Error("expected an error when a required marker block is absent")
	}
	// Optional: a missing marker leaves the doc unchanged.
	if got, err := replaceBlock("no markers", "X", "y", false); err != nil || got != "no markers" {
		t.Errorf("optional missing marker should be a no-op, got %q err %v", got, err)
	}
}

func TestCommandsBlock(t *testing.T) {
	b := commandsBlock()
	// A ready command appears in the reference.
	if !strings.Contains(b, "verifi status <path>") {
		t.Errorf("commands block missing status:\n%s", b)
	}
	// A not-ready command is not in the reference, only in "Coming soon".
	if strings.Contains(b, "verifi fix <path>  ") {
		t.Errorf("placeholder command should not be in the reference:\n%s", b)
	}
	if !strings.Contains(b, "Coming soon: fix.") {
		t.Errorf("commands block should list fix as coming soon:\n%s", b)
	}
}

func TestEcosystemsBlock(t *testing.T) {
	b := ecosystemsBlock()
	if !strings.Contains(b, "npm") {
		t.Errorf("ecosystems block should list npm:\n%s", b)
	}
	// Planned ecosystems are summarised, not listed by name.
	if strings.Contains(b, "Maven") || strings.Contains(b, "PyPI") {
		t.Errorf("planned ecosystems should not be named in the public block:\n%s", b)
	}
	if !strings.Contains(b, "on the way") {
		t.Errorf("block should note more are coming:\n%s", b)
	}
}

func TestInstallBlock(t *testing.T) {
	if !strings.Contains(installBlock(), "curl -fsSL") {
		t.Error("install block should contain the curl one-liner")
	}
}
