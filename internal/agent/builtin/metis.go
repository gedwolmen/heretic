package builtin

import (
	_ "embed"
)

//go:embed prompts/metis.md
var metisPrompt string

// NewMetisAgent returns the Metis agent: pre-planning clarifier.
func NewMetisAgent() Agent {
	return &baseAgent{
		name:         NameMetis,
		display:      "Metis",
		mode:         ModeSubagent,
		modelPref:    "large",
		systemPrompt: metisPrompt,
		description:  "Pre-planning intent clarifier. Surfaces assumptions and open questions.",
		allowedTools: readOnlyTools(),
	}
}
