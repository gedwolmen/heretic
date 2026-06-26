package model

import (
	"testing"

	"github.com/gedwolmen/heretic/internal/commands"
	"github.com/stretchr/testify/require"
)

func TestBuildInlineCommandContent_NoArgumentsAppends(t *testing.T) {
	t.Parallel()
	content := "You are starting a Ralph Loop.\nWork on the task."
	got := buildInlineCommandContent(content, "fix the bug", nil)
	require.Equal(t, content+"\n\nfix the bug", got)
}

func TestBuildInlineCommandContent_NoArgumentsEmptyInputReturnsBody(t *testing.T) {
	t.Parallel()
	content := "Body of the command."
	require.Equal(t, content, buildInlineCommandContent(content, "", nil))
	require.Equal(t, content, buildInlineCommandContent(content, "   ", nil))
}

func TestBuildInlineCommandContent_PlaceholdersSubstituted(t *testing.T) {
	t.Parallel()
	content := "Refactor $NAME now. Target is $NAME."
	args := []commands.Argument{{ID: "NAME"}}
	got := buildInlineCommandContent(content, "extract validation", args)
	require.Equal(t, "Refactor extract validation now. Target is extract validation.", got)
}

func TestBuildInlineCommandContent_PlaceholderEmptyInputReturnsBody(t *testing.T) {
	t.Parallel()
	content := "Refactor $NAME."
	args := []commands.Argument{{ID: "NAME"}}
	require.Equal(t, content, buildInlineCommandContent(content, "", args))
}
