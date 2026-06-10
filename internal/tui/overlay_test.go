package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestCompositeOverlay_CenteredOverBackground(t *testing.T) {
	bg := strings.TrimSuffix(strings.Repeat("0123456789\n", 9), "\n") // 9 lines, 10 wide
	fg := "AAAA\nBBBB"                                                // 2 lines, 4 wide

	out := compositeOverlay(fg, bg)
	lines := strings.Split(out, "\n")

	if len(lines) != 9 {
		t.Fatalf("canvas must keep background height, got %d lines", len(lines))
	}
	// Vertically centered: rows 3 and 4 of 9 hold the 2-line overlay.
	if !strings.Contains(lines[3], "AAAA") || !strings.Contains(lines[4], "BBBB") {
		t.Fatalf("overlay not vertically centered:\n%s", out)
	}
	// Horizontally centered at column (10-4)/2 = 3: background visible on both sides.
	if got := ansi.Strip(lines[3]); got != "012AAAA789" {
		t.Errorf("expected background visible around overlay, got %q", got)
	}
	// Rows outside the overlay untouched.
	if lines[0] != "0123456789" || lines[8] != "0123456789" {
		t.Errorf("rows outside overlay must be untouched")
	}
}

func TestCompositeOverlay_ANSIBackgroundSurvivesSplicing(t *testing.T) {
	red := "\x1b[31m" + strings.Repeat("x", 10) + "\x1b[0m"
	bg := strings.TrimSuffix(strings.Repeat(red+"\n", 5), "\n")
	fg := "AB"

	out := compositeOverlay(fg, bg)
	mid := strings.Split(out, "\n")[2]

	plain := ansi.Strip(mid)
	if plain != "xxxxABxxxx" {
		t.Errorf("expected overlay spliced into colored line, got %q", plain)
	}
	if ansi.StringWidth(mid) != 10 {
		t.Errorf("spliced line width = %d, want 10", ansi.StringWidth(mid))
	}
}

func TestCompositeOverlay_BackgroundShorterThanOverlay(t *testing.T) {
	bg := "ab"
	fg := "XXXX\nYYYY\nZZZZ"

	out := compositeOverlay(fg, bg)
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("canvas must grow to overlay height, got %d", len(lines))
	}
	for i, want := range []string{"XXXX", "YYYY", "ZZZZ"} {
		if !strings.Contains(lines[i], want) {
			t.Errorf("line %d missing %q: %q", i, want, lines[i])
		}
	}
}

func TestCompositeOverlay_ShortBackgroundLinesPadded(t *testing.T) {
	bg := "a\nb\nc\nd\ne"
	fg := "XX"

	out := compositeOverlay(fg, bg)
	lines := strings.Split(out, "\n")
	// Overlay centered on a 2-wide canvas? Canvas width is the background's
	// widest line (1) vs overlay width (2): overlay defines the minimum.
	if !strings.Contains(lines[2], "XX") {
		t.Errorf("expected overlay on middle row, got %q", lines[2])
	}
}
