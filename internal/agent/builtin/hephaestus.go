package builtin

import (
	_ "embed"
)

//go:embed prompts/hephaestus.md
var hephaestusPrompt string

// NewHephaestusAgent returns the Hephaestus agent: plan-driven builder.
func NewHephaestusAgent() Agent {
	return &baseAgent{
		name:         NameHephaestus,
		display:      "Hephaestus",
		mode:         ModePrimary,
		modelPref:    "large",
		systemPrompt: hephaestusPrompt,
		description:  "Plan-driven builder. Reads a plan, ships the plan. Does not improvise.",
		allowedTools: allTools(),
	}
}
