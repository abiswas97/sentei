package tui

import (
	"image/color"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestApplyPalette_SwitchesTokensAndRebuildStyles(t *testing.T) {
	defer applyPalette(darkPalette)

	applyPalette(lightPalette)
	if colorAccent != lightPalette.accent {
		t.Errorf("colorAccent = %v, want light %v", colorAccent, lightPalette.accent)
	}
	if got := styleAccent.GetForeground(); got != lightPalette.accent {
		t.Errorf("styleAccent foreground = %v, want light %v", got, lightPalette.accent)
	}
	if got := styleTitle.GetForeground(); got != lightPalette.emphasis {
		t.Errorf("styleTitle foreground = %v, want light emphasis %v", got, lightPalette.emphasis)
	}

	applyPalette(darkPalette)
	if colorAccent != darkPalette.accent {
		t.Errorf("colorAccent = %v after restore, want dark %v", colorAccent, darkPalette.accent)
	}
	if got := styleDim.GetForeground(); got != darkPalette.dim {
		t.Errorf("styleDim foreground = %v after restore, want dark %v", got, darkPalette.dim)
	}
}

func TestUpdate_LightBackgroundAppliesLightPalette(t *testing.T) {
	defer applyPalette(darkPalette)
	m := NewModel(nil, nil, "/repo")

	m.Update(tea.BackgroundColorMsg{Color: color.White})

	if colorEmphasis != lightPalette.emphasis {
		t.Errorf("light background must apply the light palette, emphasis = %v", colorEmphasis)
	}
}

func TestUpdate_DarkBackgroundKeepsDarkPalette(t *testing.T) {
	defer applyPalette(darkPalette)
	m := NewModel(nil, nil, "/repo")

	m.Update(tea.BackgroundColorMsg{Color: color.Black})

	if colorEmphasis != darkPalette.emphasis {
		t.Errorf("dark background must keep the dark palette, emphasis = %v", colorEmphasis)
	}
}

func TestInit_RequestsBackgroundColor(t *testing.T) {
	m := NewModel(nil, nil, "/repo")
	if m.Init() == nil {
		t.Error("Init must return a command (background color request)")
	}
}
