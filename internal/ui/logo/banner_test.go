package logo

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/require"
	"charm.land/lipgloss/v2"
)

func TestCharmBannerLines(t *testing.T) {
	t.Parallel()
	lines := CharmBannerLines()
	require.Len(t, lines, 8)
	// First and last lines should be non-empty.
	require.NotEmpty(t, lines[0])
	require.NotEmpty(t, lines[len(lines)-1])
}

func TestRenderCharmBanner(t *testing.T) {
	t.Parallel()
	opts := CharmBannerOpts{
		ColorA: lipgloss.Color("#FF6B9D"),
		ColorB: lipgloss.Color("#FFD86F"),
		Base:   lipgloss.NewStyle(),
	}
	out := RenderCharmBanner(opts)
	require.NotEmpty(t, out)
	// Output should contain ANSI escape sequences (color codes).
	require.Contains(t, out, "\x1b[")
}

func TestInterpColor(t *testing.T) {
	t.Parallel()
	a := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	b := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	mid := interpColor(a, b, 0.5)
	r, g, bl, al := mid.RGBA()
	t.Logf("mid RGBA: r=%d g=%d b=%d a=%d", r, g, bl, al)
	// The 16-bit-per-channel values: red is in the top 8 bits.
	// 0x8000 is mid for the high byte.
	require.InDelta(t, 0x8000, int(r), 0x100)
	require.InDelta(t, 0x8000, int(g), 0x100)
	require.InDelta(t, 0x8000, int(bl), 0x100)
	require.Equal(t, uint32(0xFF00), al&0xFF00)
}
