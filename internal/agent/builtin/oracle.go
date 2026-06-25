package builtin

import (
	_ "embed"
)

//go:embed prompts/oracle.md
var oraclePrompt string

// NewOracleAgent returns the Oracle agent: architecture advisor.
func NewOracleAgent() Agent {
	return &baseAgent{
		name:         NameOracle,
		display:      "Oracle",
		mode:         ModeSubagent,
		modelPref:    "large",
		systemPrompt: oraclePrompt,
		description:  "Architecture / strategy advisor. Surfaces options, does not decide.",
		allowedTools: readOnlyTools(),
	}
}
