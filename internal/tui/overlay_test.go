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
	// Overlay rows are fully claimed: cleared flanks, no orphan background
	// glyphs beside the box. Background stays visible above and below.
	if got := strings.TrimRight(ansi.Strip(lines[3]), " "); got != "   AAAA" {
		t.Errorf("expected cleared flanks beside overlay, got %q", got)
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

	plain := strings.TrimRight(ansi.Strip(mid), " ")
	if plain != "    AB" {
		t.Errorf("expected overlay row with cleared flanks, got %q", plain)
	}
	// The colored background survives untouched on non-overlay rows.
	top := strings.Split(out, "\n")[0]
	if !strings.Contains(top, "\x1b[31m") {
		t.Errorf("background styling must survive above the overlay, got %q", top)
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
