package builtin

import (
	_ "embed"
)

//go:embed prompts/sisyphus_junior.md
var sisyphusJuniorPrompt string

// NewSisyphusJuniorAgent returns the Sisyphus Junior agent: lightweight coder.
func NewSisyphusJuniorAgent() Agent {
	return &baseAgent{
		name:         NameSisyphusJunior,
		display:      "Sisyphus Junior",
		mode:         ModeSubagent,
		modelPref:    "small",
		systemPrompt: sisyphusJuniorPrompt,
		description:  "Lightweight coder for delegation. Focused tasks only. Does not spawn further subagents.",
		allowedTools: codingTools(),
	}
}
