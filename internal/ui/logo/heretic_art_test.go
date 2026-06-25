package logo

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"
)

func TestHereticBlockArtLines(t *testing.T) {
	t.Parallel()
	lines := HereticBlockArtLines()
	require.Len(t, lines, 5)
	// All lines should be non-empty.
	for i, l := range lines {
		require.NotEmpty(t, l, "line %d is empty", i)
	}
}

func TestHereticBlockArt_FirstAndLastLineContainH(t *testing.T) {
	t.Parallel()
	lines := HereticBlockArtLines()
	// First line should contain "█████" (the H shape in block art).
	require.Contains(t, lines[0], "█████")
	// Last line should also contain "█████".
	require.Contains(t, lines[len(lines)-1], "█████")
}

func TestRenderHereticBlockArt(t *testing.T) {
	t.Parallel()
	opts := HereticBlockArtOpts{
		ColorA: lipgloss.Color("#FF6B9D"),
		ColorB: lipgloss.Color("#FFD86F"),
		Base:   lipgloss.NewStyle(),
	}
	out := RenderHereticBlockArt(opts)
	require.NotEmpty(t, out)
	// Output should have ANSI color codes.
	require.True(t, strings.Contains(out, "\x1b["))
	// 5 lines joined with \n produces 4 newlines.
	require.Equal(t, 4, strings.Count(out, "\n"))
}
