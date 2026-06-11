package tui

import (
	"testing"

	"charm.land/lipgloss/v2"
)

func TestLightPalette_ContrastTokens(t *testing.T) {
	if lightPalette.dim != lipgloss.Color("243") {
		t.Errorf("light dim must be 243, got %v", lightPalette.dim)
	}
	if lightPalette.warning != lipgloss.Color("130") {
		t.Errorf("light warning must be 130, got %v", lightPalette.warning)
	}
}
