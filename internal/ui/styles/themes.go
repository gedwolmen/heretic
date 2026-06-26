package styles

import (
	"github.com/charmbracelet/x/exp/charmtone"
)

// ThemeForProvider returns the Styles associated with the given provider
// ID. Unknown or empty provider IDs yield the default Charmtone Pantera
// theme.
func ThemeForProvider(providerID string) Styles {
	switch providerID {
	case "hyper":
		return HyperhereticObsidiana()
	default:
		return CharmtonePantera()
	}
}

// CharmtonePantera returns the Charmtone dark theme. It's the default style
// for the UI.
func CharmtonePantera() Styles {
	s := quickStyle(quickStyleOpts{
		primary:   charmtone.Guac,
		secondary: charmtone.Lichen,
		accent:    charmtone.Orchid,
		keyword:   charmtone.Lilac,

		fgBase:       charmtone.Salt,
		fgMoreSubtle: charmtone.Steep,
		fgSubtle:     charmtone.Smoke,
		fgMostSubtle: charmtone.Squid,

		onPrimary: charmtone.Pepper,

		bgBase:         charmtone.Pepper,
		bgLeastVisible: charmtone.BBQ,
		bgLessVisible:  charmtone.Char,
		bgMostVisible:  charmtone.Iron,

		separator: charmtone.Squid,

		destructive:       charmtone.Chili,
		error:             charmtone.Pom,
		warningSubtle:     charmtone.Butter,
		warning:           charmtone.Tang,
		denied:            charmtone.Salmon,
		busy:              charmtone.Zest,
		info:              charmtone.Lichen,
		infoMoreSubtle:    charmtone.Turtle,
		infoMostSubtle:    charmtone.Zinc,
		success:           charmtone.Julep,
		successMoreSubtle: charmtone.Guac,
		successMostSubtle: charmtone.Gator,

		// ANSI 16-color palette for remapping raw terminal output
		// (e.g. bang-mode shell commands) onto legible Charmtone colors.
		ansiBlack:   charmtone.BBQ,
		ansiRed:     charmtone.Pom,
		ansiGreen:   charmtone.Guac,
		ansiYellow:  charmtone.Tang,
		ansiBlue:    charmtone.Malibu,
		ansiMagenta: charmtone.Lilac,
		ansiCyan:    charmtone.Lichen,
		ansiWhite:   charmtone.Smoke,

		ansiBrightBlack:   charmtone.Iron,
		ansiBrightRed:     charmtone.Coral,
		ansiBrightGreen:   charmtone.Julep,
		ansiBrightYellow:  charmtone.Zest,
		ansiBrightBlue:    charmtone.Sardine,
		ansiBrightMagenta: charmtone.Blush,
		ansiBrightCyan:    charmtone.Turtle,
		ansiBrightWhite:   charmtone.Salt,
	})

	// Bang ! prompt overrides - use Pepper/Lichen/Guac colors.
	s.Editor.PromptBangIconFocused = s.Editor.PromptBangIconFocused.
		Foreground(charmtone.Pepper).
		Background(charmtone.Lichen)
	s.Editor.PromptBangDotsFocused = s.Editor.PromptBangDotsFocused.
		Foreground(charmtone.Lichen)
	s.Editor.PromptBangDotsBlurred = s.Editor.PromptBangDotsBlurred.
		Foreground(charmtone.Guac)

	// Shell bar/prompt overrides - use Guac/Iron/Lichen colors.
	s.Messages.ShellBarFocused = s.Messages.ShellBarFocused.
		BorderForeground(charmtone.Guac)
	s.Messages.ShellBarBlurred = s.Messages.ShellBarBlurred.
		BorderForeground(charmtone.Iron)
	s.Messages.ShellPrompt = s.Messages.ShellPrompt.
		Foreground(charmtone.Lichen)
	s.Messages.ShellPromptBlurred = s.Messages.ShellPromptBlurred.
		Foreground(charmtone.Lichen)

	return s
}

// HyperhereticObsidiana returns the Hyperheretic dark theme.
func HyperhereticObsidiana() Styles {
	return CharmtonePantera()
}
