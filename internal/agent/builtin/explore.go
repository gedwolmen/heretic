package builtin

import (
	_ "embed"
)

//go:embed prompts/explore.md
var explorePrompt string

// NewExploreAgent returns the Explore agent: codebase investigator.
func NewExploreAgent() Agent {
	return &baseAgent{
		name:         NameExplore,
		display:      "Explore",
		mode:         ModeSubagent,
		modelPref:    "small",
		systemPrompt: explorePrompt,
		description:  "Codebase investigator. Read-only, file:line references.",
		allowedTools: readOnlyTools(),
	}
}
