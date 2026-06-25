package builtin

import (
	_ "embed"
)

//go:embed prompts/librarian.md
var librarianPrompt string

// NewLibrarianAgent returns the Librarian agent: external research.
func NewLibrarianAgent() Agent {
	return &baseAgent{
		name:         NameLibrarian,
		display:      "Librarian",
		mode:         ModeSubagent,
		modelPref:    "large",
		systemPrompt: librarianPrompt,
		description:  "External research. Library docs, APIs, RFCs. Citable summaries.",
		allowedTools: readOnlyTools(),
	}
}
