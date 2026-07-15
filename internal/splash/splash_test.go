package splash

import (
	"regexp"
	"strings"
	"testing"
)

var ansi = regexp.MustCompile("\033\\[[0-9;]*[a-zA-Z]")

// The static banner should render the settled wordmark and the subtitle.
// Cells are wrapped in per-cell ANSI codes, so strip those before asserting.
func TestStaticFrameContainsWordmarkAndSubtitle(t *testing.T) {
	frame := ansi.ReplaceAllString(renderFrame(nil, 0, 1), "")
	if !strings.Contains(frame, "coming soon · beta") {
		t.Error("static frame is missing the subtitle")
	}
	if strings.Count(frame, "█") == 0 {
		t.Error("static frame has no wordmark blocks")
	}
}

// wordmarkMask must place all six letters (V E R I F I) and report a width
// that fits the canvas.
func TestWordmarkFitsCanvas(t *testing.T) {
	_, w := wordmarkMask()
	if w <= 0 || w > canvasW {
		t.Errorf("wordmark width %d does not fit canvas width %d", w, canvasW)
	}
}
