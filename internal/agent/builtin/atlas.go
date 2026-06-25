package builtin

import (
	_ "embed"
)

//go:embed prompts/atlas.md
var atlasPrompt string

// NewAtlasAgent returns the Atlas agent: multi-agent orchestrator.
func NewAtlasAgent() Agent {
	return &baseAgent{
		name:         NameAtlas,
		display:      "Atlas",
		mode:         ModePrimary,
		modelPref:    "large",
		systemPrompt: atlasPrompt,
		description:  "Multi-agent orchestrator. Decomposes complex tasks, spawns a team, integrates results.",
		allowedTools: allTools(),
	}
}
