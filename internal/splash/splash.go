// Package splash renders the `verifi` welcome screen: muted boxes of varying
// sizes fall down a blackish gradient (the website hero motif), then fade out
// as the solid VERIFI wordmark and a "coming soon · beta" line settle in.
//
// Stdlib only, no dependencies, by design: this is a security tool's own CLI,
// so keeping its dependency surface at zero is a feature, not a limitation.
package splash

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// ---------------------------- TUNABLES ----------------------------
// Tweak these to change the look; everything visual lives here.

const (
	canvasW = 56 // panel width in cells
	canvasH = 18 // panel height in rows

	fps           = 30
	fallSeconds   = 2.2 // boxes rain down
	settleSeconds = 1.1 // boxes fade out, wordmark fades in
	holdSeconds   = 1.6 // hold the finished banner

	spawnChance = 0.55 // 0..1, probability of a new box each frame
	blockDim    = 0.55 // dim the bright web palette toward the "low opacity" look
	bottomFade  = 0.66 // boxes start fading below this fraction of the height
)

// The hero palette (verifi-web tetris-background), dimmed at render time.
var palette = []rgb{
	{255, 68, 68}, {255, 102, 68}, {255, 170, 68},
	{68, 255, 68}, {68, 255, 170}, {68, 136, 255}, {170, 68, 255},
}

var (
	bgTop    = rgb{20, 17, 13} // blackish gradient, top
	bgBottom = rgb{11, 9, 7}   // ...darker at the bottom
	inkWord  = rgb{245, 242, 238}
	inkTeal  = rgb{35, 176, 190} // brand teal for the subtitle
	inkMuted = rgb{150, 140, 130}
)

// ------------------------------------------------------------------

// Show renders the welcome. It animates when stdout is a real terminal;
// otherwise (piped, CI) it prints the static banner. Pass loop to replay.
func Show(loop bool) {
	if !isTTY() {
		Static()
		return
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sig; showCursor(); os.Exit(0) }()

	hideCursor()
	defer showCursor()

	for {
		playOnce()
		if !loop {
			break
		}
		fmt.Print(cursorUp(canvasH))
	}
	fmt.Println()
}

// Static prints just the settled banner, no animation.
func Static() { fmt.Print(renderFrame(nil, 0, 1)) }

type rgb struct{ r, g, b float64 }

func lerp(a, b, t float64) float64 { return a + (b-a)*t }

func mix(a, b rgb, t float64) rgb {
	return rgb{lerp(a.r, b.r, t), lerp(a.g, b.g, t), lerp(a.b, b.b, t)}
}

func (c rgb) dim(f float64) rgb { return rgb{c.r * f, c.g * f, c.b * f} }

func (c rgb) fg() string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", int(c.r), int(c.g), int(c.b))
}
func (c rgb) bg() string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", int(c.r), int(c.g), int(c.b))
}

// A falling box of a given size (the "different sizes, wave form" look).
type box struct {
	x, y  float64
	w, h  int
	speed float64
	col   rgb
}

type cell struct {
	ch rune
	fg rgb
	bg rgb
}

// ---- the VERIFI wordmark, a 5-row block font ----

var glyphs = map[rune][]string{
	'V': {"#   #", "#   #", "#   #", " # # ", "  #  "},
	'E': {"#####", "#    ", "#### ", "#    ", "#####"},
	'R': {"#### ", "#   #", "#### ", "#  # ", "#   #"},
	'I': {"#####", "  #  ", "  #  ", "  #  ", "#####"},
	'F': {"#####", "#    ", "#### ", "#    ", "#    "},
}

// wordmarkMask returns the set of lit (col,row) cells for "VERIFI", and its
// pixel width, so we can centre it on the canvas.
func wordmarkMask() (map[[2]int]bool, int) {
	word := "VERIFI"
	mask := map[[2]int]bool{}
	x := 0
	for _, r := range word {
		g := glyphs[r]
		for row := 0; row < 5; row++ {
			for col := 0; col < len(g[row]); col++ {
				if g[row][col] == '#' {
					mask[[2]int{x + col, row}] = true
				}
			}
		}
		x += 5 + 1 // glyph width + gap
	}
	return mask, x - 1
}

func bgAt(y int) rgb { return mix(bgTop, bgBottom, float64(y)/float64(canvasH-1)) }

// renderFrame paints one frame: gradient backdrop, boxes, then wordmark.
func renderFrame(boxes []box, blockAlpha, wordAlpha float64) string {
	grid := make([][]cell, canvasH)
	for y := 0; y < canvasH; y++ {
		grid[y] = make([]cell, canvasW)
		bg := bgAt(y)
		for x := 0; x < canvasW; x++ {
			grid[y][x] = cell{' ', bg, bg}
		}
	}

	// Boxes.
	for _, b := range boxes {
		a := blockAlpha
		if fy := float64(canvasH) * bottomFade; b.y > fy {
			a *= math.Max(0, 1-(b.y-fy)/(float64(canvasH)-fy))
		}
		if a <= 0.02 {
			continue
		}
		for dy := 0; dy < b.h; dy++ {
			for dx := 0; dx < b.w; dx++ {
				px, py := int(b.x)+dx, int(b.y)+dy
				if px < 0 || px >= canvasW || py < 0 || py >= canvasH {
					continue
				}
				bg := grid[py][px].bg
				grid[py][px] = cell{'█', mix(bg, b.col, a), bg}
			}
		}
	}

	// Wordmark + subtitle, centred, drawn on top so it "fades to solid".
	if wordAlpha > 0.01 {
		mask, w := wordmarkMask()
		ox := (canvasW - w) / 2
		oy := canvasH/2 - 4
		for p := range mask {
			px, py := ox+p[0], oy+p[1]
			if px < 0 || px >= canvasW || py < 0 || py >= canvasH {
				continue
			}
			bg := grid[py][px].bg
			grid[py][px] = cell{'█', mix(bg, inkWord, wordAlpha), bg}
		}
		drawText(grid, oy+7, center("coming soon · beta", canvasW), mix(bgAt(oy+7), inkTeal, wordAlpha))
		drawText(grid, oy+9, center("github.com/verifisecurity/verifi", canvasW), mix(bgAt(oy+9), inkMuted, wordAlpha*0.9))
	}

	// Serialise with truecolor escapes.
	var b strings.Builder
	for y := 0; y < canvasH; y++ {
		for x := 0; x < canvasW; x++ {
			c := grid[y][x]
			b.WriteString(c.bg.bg())
			b.WriteString(c.fg.fg())
			b.WriteRune(c.ch)
		}
		b.WriteString("\033[0m\033[K\n")
	}
	return b.String()
}

func drawText(grid [][]cell, row int, s string, col rgb) {
	if row < 0 || row >= canvasH {
		return
	}
	x := 0
	for _, r := range s {
		if x >= canvasW {
			break
		}
		grid[row][x].ch = r
		grid[row][x].fg = col
		x++
	}
}

func center(s string, width int) string {
	n := len([]rune(s))
	if n >= width {
		return s
	}
	return strings.Repeat(" ", (width-n)/2) + s
}

func spawnBox() box {
	col := palette[rand.Intn(len(palette))].dim(blockDim)
	// Bias toward small boxes, with the occasional bigger one → wave form.
	w, h := 1, 1
	switch rand.Intn(5) {
	case 0, 1:
		w, h = 2, 1
	case 2:
		w, h = 2, 2
	case 3:
		w, h = 3, 2
	}
	return box{
		x:     rand.Float64() * float64(canvasW-3),
		y:     -float64(h),
		w:     w,
		h:     h,
		speed: 0.25 + rand.Float64()*0.5,
		col:   col,
	}
}

func playOnce() {
	boxes := []box{}
	frame := time.Second / fps
	total := fallSeconds + settleSeconds + holdSeconds
	steps := int(total * float64(fps))
	first := true

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(fps)

		blockAlpha, wordAlpha := 1.0, 0.0
		switch {
		case t < fallSeconds:
			if rand.Float64() < spawnChance {
				boxes = append(boxes, spawnBox())
			}
		case t < fallSeconds+settleSeconds:
			p := (t - fallSeconds) / settleSeconds
			blockAlpha = 1 - p
			wordAlpha = p
		default:
			blockAlpha = 0
			wordAlpha = 1
		}

		kept := boxes[:0]
		for _, b := range boxes {
			b.y += b.speed
			if b.y < float64(canvasH)+2 {
				kept = append(kept, b)
			}
		}
		boxes = kept

		out := renderFrame(boxes, blockAlpha, wordAlpha)
		if !first {
			out = cursorUp(canvasH) + out
		}
		first = false
		fmt.Print(out)
		time.Sleep(frame)
	}
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func hideCursor()          { fmt.Print("\033[?25l") }
func showCursor()          { fmt.Print("\033[?25h") }
func cursorUp(n int) string { return fmt.Sprintf("\033[%dA", n) }
