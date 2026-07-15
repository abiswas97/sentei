package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// The motion clock: one gated tick counter from which every animation —
// star frames, star colors, shimmer band positions — derives as a pure
// function. One clock means one vocabulary and deterministic tests.
const (
	// motionResolution is the clock period; shimmer needs ~15fps to read
	// as a sweep rather than steps.
	motionResolution = 60 * time.Millisecond

	// starInterval is the twinkle's frame duration (Claude Code's pace).
	starInterval = 120 * time.Millisecond

	// shimmerPeriod is how long the band takes to cross a line once.
	shimmerPeriod = 2500 * time.Millisecond

	// shimmerBandHalf is the band's half-width in cells: runes this close
	// to the band center blend toward the ramp peak.
	shimmerBandHalf = 6.0
)

// starFrames is the working twinkle: a star growing out of the pending dot
// and collapsing back. Every frame is one cell; done is the crystallized ✦.
var starFrames = []string{"·", "✢", "✳", "✻", "✽", "✻", "✳", "✢"}

type MotionPreference uint8

const (
	MotionFull MotionPreference = iota
	MotionOff
)

func motionPreference(getenv func(string) string) MotionPreference {
	if strings.EqualFold(getenv("SENTEI_MOTION"), "off") || strings.EqualFold(getenv("TERM"), "dumb") {
		return MotionOff
	}
	return MotionFull
}

// motionTickMsg advances the motion clock.
type motionTickMsg struct{}

// motionTickCmd schedules the next clock tick.
func motionTickCmd() tea.Cmd {
	return tea.Tick(motionResolution, func(time.Time) tea.Msg {
		return motionTickMsg{}
	})
}

// shimmerRamp is one base→peak color pair; ramps are palette data so both
// themes declare their own.
type shimmerRamp struct {
	base string // hex
	peak string // hex
}

// starFrame returns the twinkle frame for a tick.
func starFrame(tick int) string {
	ticksPerFrame := int(starInterval / motionResolution)
	return starFrames[(tick/ticksPerFrame)%len(starFrames)]
}

// shimmerLine renders text with a gradient band sweeping across it: each
// rune's color blends from ramp.base toward ramp.peak by its distance to
// the moving band center. Output is bold; stripped text equals the input.
func shimmerLine(text string, ramp shimmerRamp, tick int) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}
	// The band travels from past the right edge to past the left edge
	// (matching a leftward reading sweep) and wraps every shimmerPeriod.
	travel := float64(len(runes)) + 2*shimmerBandHalf
	elapsed := float64(tick) * float64(motionResolution)
	progress := elapsed / float64(shimmerPeriod)
	progress -= float64(int(progress))
	center := travel*(1-progress) - shimmerBandHalf

	var b strings.Builder
	for i, r := range runes {
		d := float64(i) - center
		if d < 0 {
			d = -d
		}
		intensity := 1 - d/shimmerBandHalf
		if intensity < 0 {
			intensity = 0
		}
		color := lerpHex(ramp.base, ramp.peak, intensity)
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(color)).Render(string(r)))
	}
	return b.String()
}

// lerpHex interpolates two #rrggbb colors; t is clamped to [0,1].
func lerpHex(a, b string, t float64) string {
	t = min(max(t, 0), 1)
	ar, ag, ab := splitHex(a)
	br, bg, bb := splitHex(b)
	lerp := func(x, y int) int { return x + int(float64(y-x)*t) }
	return fmt.Sprintf("#%02x%02x%02x", lerp(ar, br), lerp(ag, bg), lerp(ab, bb))
}

func splitHex(h string) (r, g, b int) {
	h = strings.TrimPrefix(h, "#")
	if len(h) != 6 {
		return 0, 0, 0
	}
	parse := func(s string) int {
		v, _ := strconv.ParseInt(s, 16, 32)
		return int(v)
	}
	return parse(h[0:2]), parse(h[2:4]), parse(h[4:6])
}

// Motion is the presentation bundle the model injects into pure layouts:
// the current star frame plus shimmer closures bound to the current tick.
// Nil Motion means static fallbacks (✻, plain styles).
type Motion struct {
	Frame  string
	Accent func(string) string
	Body   func(string) string
}

// motion builds the presentation bundle for the current clock tick.
func (m Model) motion() *Motion {
	tick := m.motionTick
	return &Motion{
		Frame:  starFrame(tick),
		Accent: func(s string) string { return shimmerLine(s, rampAccent, tick) },
		Body:   func(s string) string { return shimmerLine(s, rampBody, tick) },
	}
}
