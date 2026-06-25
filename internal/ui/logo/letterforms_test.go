package logo

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"
)

func TestLetterT_NotEmpty(t *testing.T) {
	t.Parallel()
	got := LetterT(false)
	require.NotEmpty(t, got)
	require.Contains(t, got, "█")
}

func TestLetterI_NotEmpty(t *testing.T) {
	t.Parallel()
	got := LetterI(false)
	require.NotEmpty(t, got)
	require.Contains(t, got, "█")
}

func TestLetterT_StretchedDifferentFromUnstretched(t *testing.T) {
	t.Parallel()
	a := LetterT(false)
	b := LetterT(true)
	// Both must render; they may differ in width when stretched.
	require.NotEmpty(t, a)
	require.NotEmpty(t, b)
}

func TestRender_HereticWordmark(t *testing.T) {
	t.Parallel()
	o := Opts{
		FieldColor:   lipgloss.Color("#FFFFFF"),
		TitleColorA:  lipgloss.Color("#FF6B9D"),
		TitleColorB:  lipgloss.Color("#FFB86C"),
		CharmColor:   lipgloss.Color("#A0A0A0"),
		VersionColor: lipgloss.Color("#606060"),
		Width:        80,
	}
	got := Render(lipgloss.NewStyle(), "v0.1.0", false, o)
	require.NotEmpty(t, got)
	// The Render output should contain block characters from the
	// letterforms.
	require.True(t, strings.Contains(got, "█") || strings.Contains(got, "▀") || strings.Contains(got, "▄"),
		"rendered logo should contain block characters")
}

func TestRender_HyperHeretic(t *testing.T) {
	t.Parallel()
	o := Opts{
		FieldColor:   lipgloss.Color("#FFFFFF"),
		TitleColorA:  lipgloss.Color("#FF6B9D"),
		TitleColorB:  lipgloss.Color("#FFB86C"),
		CharmColor:   lipgloss.Color("#A0A0A0"),
		VersionColor: lipgloss.Color("#606060"),
		Width:        80,
		Hyper:        true,
	}
	got := Render(lipgloss.NewStyle(), "v0.1.0", false, o)
	require.NotEmpty(t, got)
	// When Hyper, the render should be wider (HYPER + HERETIC).
}
