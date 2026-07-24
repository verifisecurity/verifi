package docgen

import (
	"strings"
	"testing"
)

func TestReplaceBlock(t *testing.T) {
	doc := "top\n<!-- BEGIN X -->\nold content\n<!-- END X -->\nbottom"
	want := "top\n<!-- BEGIN X -->\nnew\n<!-- END X -->\nbottom"

	got, err := replaceBlock(doc, "X", "new")
	if err != nil {
		t.Fatalf("replaceBlock: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// Idempotent: replacing again with the same content is a no-op.
	again, err := replaceBlock(got, "X", "new")
	if err != nil || again != want {
		t.Errorf("not idempotent: got %q err %v", again, err)
	}
}

func TestReplaceBlock_MissingMarker(t *testing.T) {
	if _, err := replaceBlock("no markers here", "X", "y"); err == nil {
		t.Error("expected an error when the marker block is absent")
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

func TestInstallBlock(t *testing.T) {
	if !strings.Contains(installBlock(), "curl -fsSL") {
		t.Error("install block should contain the curl one-liner")
	}
}
