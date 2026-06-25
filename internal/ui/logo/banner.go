package logo

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// CharmBanner is the 8-line colorful "Charm" block art used as the
// splash banner. Each line has its own color so the result reads as
// a vertical gradient (pink → orange → yellow → orange → pink).
//
// The art is the same one used by charmbracelet/crush; we renamed the
// wordmark "Heretic" but kept the Charm art as the visual identity
// banner (heretic is a crush fork).
const CharmBanner = ` **      ** ******** *******   ******** ********** **   ****** 
/**     /**/**///// /**////** /**///// /////**/// /**  **////**
/**     /**/**      /**   /** /**          /**    /** **    // 
/**********/******* /*******  /*******     /**    /**/**       
/**//////**/**////  /**///**  /**////      /**    /**/**       
/**     /**/**      /**  //** /**          /**    /**//**    **
/**     /**/********/**   //**/********    /**    /** //****** 
//      // //////// //     // ////////     //     //   //////  `

// CharmBannerLines returns the banner as a slice of lines so the
// caller can style each line independently.
func CharmBannerLines() []string {
	lines := strings.Split(CharmBanner, "\n")
	// Remove trailing empty line if present.
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	return lines
}

// CharmBannerOpts controls per-line coloring.
type CharmBannerOpts struct {
	// ColorA is the color at the top of the gradient.
	ColorA color.Color
	// ColorB is the color at the bottom.
	ColorB color.Color
	// Style applied to every line (used to set width, alignment, etc).
	Base lipgloss.Style
}

// RenderCharmBanner returns the banner with each line colored as a
// vertical gradient from ColorA to ColorB. Width is auto-sized to
// the longest line; alignment defaults to Center.
func RenderCharmBanner(opts CharmBannerOpts) string {
	lines := CharmBannerLines()
	if len(lines) == 0 {
		return ""
	}
	out := make([]string, len(lines))
	for i, line := range lines {
		// Compute the gradient position for this line.
		t := 0.0
		if len(lines) > 1 {
			t = float64(i) / float64(len(lines)-1)
		}
		c := interpColor(opts.ColorA, opts.ColorB, t)
		style := opts.Base.Foreground(c)
		out[i] = style.Render(line)
	}
	return strings.Join(out, "\n")
}

// RenderCharmBannerInline is like RenderCharmBanner but applies
// per-line colors to a single string with embedded "\n" line
// separators (used when the caller has a single styled string).
func RenderCharmBannerInline(input string, opts CharmBannerOpts) string {
	lines := strings.Split(input, "\n")
	out := make([]string, len(lines))
	for i, line := range lines {
		t := 0.0
		if len(lines) > 1 {
			t = float64(i) / float64(len(lines)-1)
		}
		c := interpColor(opts.ColorA, opts.ColorB, t)
		style := opts.Base.Foreground(c)
		out[i] = style.Render(line)
	}
	return strings.Join(out, "\n")
}

// interpColor blends two colors by t in [0, 1] and returns the result
// as a color.Color. The blend is done in the 16-bit-per-channel RGBA
// space to avoid the uint8 quantization that would otherwise occur if
// we round-tripped through color.RGBA.
func interpColor(a, b color.Color, t float64) color.Color {
	ar, ag, ab, aa := channels16(a)
	br, bg, bb, ba := channels16(b)
	r := uint16(float64(ar) + t*(float64(br)-float64(ar)))
	g := uint16(float64(ag) + t*(float64(bg)-float64(ag)))
	bl := uint16(float64(ab) + t*(float64(bb)-float64(ab)))
	al := uint16(float64(aa) + t*(float64(ba)-float64(aa)))
	return color.RGBA64{R: r, G: g, B: bl, A: al}
}

// channels16 returns the 16-bit-per-channel RGBA values of c.
func channels16(c color.Color) (r, g, b, a uint16) {
	if c64, ok := c.(color.RGBA64); ok {
		return c64.R, c64.G, c64.B, c64.A
	}
	if cr, ok := c.(color.RGBA); ok {
		// color.RGBA has 8-bit channels; expand to 16-bit.
		return uint16(cr.R) * 0x101, uint16(cr.G) * 0x101, uint16(cr.B) * 0x101, uint16(cr.A) * 0x101
	}
	if nr, ok := c.(color.NRGBA); ok {
		return uint16(nr.R) * 0x101, uint16(nr.G) * 0x101, uint16(nr.B) * 0x101, uint16(nr.A) * 0x101
	}
	vr, vg, vb, va := c.RGBA()
	return uint16(vr), uint16(vg), uint16(vb), uint16(va)
}
