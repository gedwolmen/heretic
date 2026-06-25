package builtin

import (
	_ "embed"
)

//go:embed prompts/momus.md
var momusPrompt string

// NewMomusAgent returns the Momus agent: plan reviewer.
func NewMomusAgent() Agent {
	return &baseAgent{
		name:         NameMomus,
		display:      "Momus",
		mode:         ModeSubagent,
		modelPref:    "large",
		systemPrompt: momusPrompt,
		description:  "Plan reviewer. Approves / rejects / requests changes. Does not write code.",
		allowedTools: readOnlyTools(),
	}
}
