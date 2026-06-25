package logo

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// HereticBlockArt is the 5-line HERETIC block art used as the wordmark
// in the splash and header. Lines are pre-padded so the rendered
// art is rectangular regardless of terminal width.
//
// Visual:
//
//	█   █ █████ ████  █████ █████ ███  ███
//	█   █ █     █   █ █       █    █  █
//	█████ ████  ████  ████    █    █  █
//	█   █ █     █  █  █       █    █  █
//	█   █ █████ █   █ █████   █   ███  ███
const HereticBlockArt = `█   █ █████ ████  █████ █████ ███  ███
█   █ █     █   █ █       █    █  █
█████ ████  ████  ████    █    █  █
█   █ █     █  █  █       █    █  █
█   █ █████ █   █ █████   █   ███  ███`

// HereticBlockArtLines returns the art as a slice of lines.
func HereticBlockArtLines() []string {
	lines := strings.Split(HereticBlockArt, "\n")
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	return lines
}

// HereticBlockArtOpts controls per-line coloring.
type HereticBlockArtOpts struct {
	ColorA color.Color
	ColorB color.Color
	Base   lipgloss.Style
}

// RenderHereticBlockArt returns the HERETIC wordmark with each line
// colored as a vertical gradient from ColorA to ColorB.
func RenderHereticBlockArt(opts HereticBlockArtOpts) string {
	lines := HereticBlockArtLines()
	if len(lines) == 0 {
		return ""
	}
	out := make([]string, len(lines))
	for i, line := range lines {
		t := 0.0
		if len(lines) > 1 {
			t = float64(i) / float64(len(lines)-1)
		}
		c := interpColor(opts.ColorA, opts.ColorB, t)
		out[i] = opts.Base.Foreground(c).Render(line)
	}
	return strings.Join(out, "\n")
}
