package model

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/gedwolmen/heretic/internal/commands"
)

// inlineCommandState holds a slash command that was inlined into the
// editor so the user can type its argument inline (e.g.
// "/ralph-loop fix the bug"). On send, the typed text after the
// invocation is injected into content and dispatched as the prompt.
type inlineCommandState struct {
	// invocation is the command token placed into the editor without a
	// trailing space (e.g. "/ralph-loop").
	invocation string
	// content is the command body, with optional $ARG placeholders.
	content string
	// arguments are the named $ARG placeholders, if any.
	arguments []commands.Argument
}

// startInlineCommand places a slash command invocation into the editor
// so the user can type its argument inline. The pending command is
// remembered in m.inlineCommand; on send the typed text is injected
// into the command body and dispatched.
func (m *UI) startInlineCommand(invocation, content string, arguments []commands.Argument) tea.Cmd {
	prevHeight := m.textarea.Height()
	m.inlineCommand = &inlineCommandState{
		invocation: invocation,
		content:    content,
		arguments:  arguments,
	}
	m.focus = uiFocusEditor
	m.textarea.SetValue(invocation + " ")
	m.textarea.MoveToEnd()
	cmds := []tea.Cmd{m.textarea.Focus()}
	if cmd := m.handleTextareaHeightChange(prevHeight); cmd != nil {
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

// buildInlineCommandContent injects the user-supplied argument text
// into a slash command body. When the body contains $ARG placeholders
// they are replaced with the argument text; otherwise the argument
// text is appended to the body so the agent receives the task context.
// An empty argument yields the original body unchanged.
func buildInlineCommandContent(content, userInput string, arguments []commands.Argument) string {
	userInput = strings.TrimSpace(userInput)
	if userInput == "" {
		return content
	}
	if len(arguments) > 0 {
		args := make(map[string]string, len(arguments))
		for _, arg := range arguments {
			args[arg.ID] = userInput
		}
		return substituteArgs(content, args)
	}
	return content + "\n\n" + userInput
}
