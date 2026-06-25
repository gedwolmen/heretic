package builtin

import (
	_ "embed"

	"charm.land/fantasy"
)

//go:embed prompts/sisyphus.md
var sisyphusPrompt string

// NewSisyphusAgent returns the Sisyphus agent: the master orchestrator.
// Primary mode; the default agent for interactive sessions.
func NewSisyphusAgent() Agent {
	return &baseAgent{
		name:         NameSisyphus,
		display:      "Sisyphus",
		mode:         ModePrimary,
		modelPref:    "large",
		systemPrompt: sisyphusPrompt,
		description:  "Master orchestrator. Takes a request and ships a result. Delegates aggressively.",
		allowedTools: allTools(),
	}
}

// Compile-time check that Agent interface is satisfied.
var _ Agent = (*baseAgent)(nil)

// allTools returns the full tool set, mirroring the crush tool registry.
// This is the "no restrictions" set; agents restrict by omitting tools
// from their allowedTools list.
func allTools() []ToolRef {
	return []ToolRef{
		{Name: "bash"},
		{Name: "edit"},
		{Name: "write"},
		{Name: "multi_edit"},
		{Name: "view"},
		{Name: "glob"},
		{Name: "grep"},
		{Name: "read_image"},
		{Name: "task"},
		{Name: "skill"},
		{Name: "skill_mcp"},
		{Name: "heretic_info"},
		{Name: "heretic_logs"},
		{Name: "todos"},
		{Name: "background_output"},
		{Name: "background_cancel"},
		{Name: "fetch"},
		{Name: "download"},
		{Name: "web_fetch"},
		{Name: "web_search"},
		{Name: "sourcegraph"},
		{Name: "interactive_bash"},
		{Name: "hashline_edit"},
	}
}

// readOnlyTools returns the read-only subset for subagents that should
// not modify the project.
func readOnlyTools() []ToolRef {
	return []ToolRef{
		{Name: "view"},
		{Name: "glob"},
		{Name: "grep"},
		{Name: "ls"},
		{Name: "skill"},
		{Name: "read_image"},
	}
}

// codingTools returns the read + write subset for subagents that code.
func codingTools() []ToolRef {
	return []ToolRef{
		{Name: "view"},
		{Name: "glob"},
		{Name: "grep"},
		{Name: "edit"},
		{Name: "write"},
		{Name: "skill"},
	}
}

// FantasyToolNames returns the names of the tools as fantasy.AgentTool
// references would expect. (Convenience for the registry consumer.)
func FantasyToolNames(refs []ToolRef) []string {
	out := make([]string, 0, len(refs))
	for _, r := range refs {
		out = append(out, r.Name)
	}
	return out
}

// Compiles-only: confirm fantasy.AgentTool is reachable.
var _ fantasy.AgentTool
