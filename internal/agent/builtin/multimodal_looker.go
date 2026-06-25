package builtin

import (
	_ "embed"
)

//go:embed prompts/multimodal_looker.md
var multimodalLookerPrompt string

// NewMultimodalLookerAgent returns the Multimodal Looker agent.
func NewMultimodalLookerAgent() Agent {
	return &baseAgent{
		name:         NameMultimodalLooker,
		display:      "Multimodal Looker",
		mode:         ModeSubagent,
		modelPref:    "small",
		systemPrompt: multimodalLookerPrompt,
		description:  "Image analysis. Screenshots, mocks, diagrams. Read-only.",
		allowedTools: readOnlyTools(),
	}
}
