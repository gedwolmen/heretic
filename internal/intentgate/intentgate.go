// Package intentgate implements keyword-based mode injection for heretic.
//
// On the first user message of each session, IntentGate scans the message
// for keywords (`ultrawork` / `ulw`, `search`, `analyze`, `team`) and
// appends a mode-specific prompt to the system message before sending it
// to the LLM.
//
// Design notes:
//   - This is a Go-idiomatic concept port of oh-my-opencode's keyword
//     detector. We do NOT copy code; we reimplement the concept in Go
//     using only stdlib (strings, regexp) + embed for the prompt files.
//   - Detection is keyword + word-boundary aware. `ultrawork` matches
//     `ultrawork build X` but NOT `ultraworks`.
//   - Multiple keywords may be detected in one message; prompts stack.
//   - Subsequent messages do not re-trigger detection.
package intentgate

import (
	_ "embed"
	"regexp"
	"strings"
)

// Mode is a keyword-detected mode that gets injected into the system prompt.
type Mode string

const (
	ModeUltrawork Mode = "ultrawork"
	ModeSearch    Mode = "search"
	ModeAnalyze   Mode = "analyze"
	ModeTeam      Mode = "team"
)

//go:embed prompts/ultrawork.md
var ultraworkPrompt string

//go:embed prompts/search.md
var searchPrompt string

//go:embed prompts/analyze.md
var analyzePrompt string

//go:embed prompts/team.md
var teamPrompt string

// promptFor returns the mode-specific prompt text.
func promptFor(m Mode) string {
	switch m {
	case ModeUltrawork:
		return ultraworkPrompt
	case ModeSearch:
		return searchPrompt
	case ModeAnalyze:
		return analyzePrompt
	case ModeTeam:
		return teamPrompt
	}
	return ""
}

// detection maps each mode to the keyword(s) that trigger it. The first
// match wins per mode. Aliases are listed alongside the canonical keyword.
var detection = map[Mode][]string{
	ModeUltrawork: {"ultrawork", "ulw"},
	ModeSearch:    {"search"},
	ModeAnalyze:   {"analyze"},
	ModeTeam:      {"team"},
}

// buildRegex returns a case-insensitive word-boundary regex for a keyword.
func buildRegex(keyword string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(keyword) + `\b`)
}

// Detect scans the input for mode keywords. It returns the set of detected
// modes in canonical (insertion) order: ultrawork, search, analyze, team.
func Detect(input string) []Mode {
	var out []Mode
	order := []Mode{ModeUltrawork, ModeSearch, ModeAnalyze, ModeTeam}
	for _, m := range order {
		for _, kw := range detection[m] {
			if buildRegex(kw).MatchString(input) {
				out = append(out, m)
				break
			}
		}
	}
	return out
}

// Inject appends the mode-specific prompts to the system message. The
// prompts are concatenated in the order the modes were detected.
func Inject(systemPrompt string, modes []Mode) string {
	if len(modes) == 0 {
		return systemPrompt
	}
	var sb strings.Builder
	sb.WriteString(systemPrompt)
	for _, m := range modes {
		p := promptFor(m)
		if p == "" {
			continue
		}
		sb.WriteString("\n\n")
		sb.WriteString(p)
	}
	return sb.String()
}
