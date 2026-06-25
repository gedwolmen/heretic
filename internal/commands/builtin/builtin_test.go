package builtin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_ReturnsAllBuiltinCommands(t *testing.T) {
	t.Parallel()
	cmds, err := Load()
	require.NoError(t, err)
	require.Len(t, cmds, 9, "expected 9 builtin commands (ralph-loop, ulw-loop, cancel-ralph, refactor, start-work, stop-continuation, remove-ai-slops, handoff, hyperplan)")
}

func TestLoad_AllCommandsHaveBuiltinFlag(t *testing.T) {
	t.Parallel()
	cmds, err := Load()
	require.NoError(t, err)
	for _, c := range cmds {
		require.True(t, c.Builtin, "command %q should have Builtin=true", c.Name)
		require.True(t, strings.HasPrefix(c.ID, "builtin:"), "command %q should have builtin: ID prefix", c.Name)
	}
}

func TestLoad_AllCommandsHaveNonEmptyContent(t *testing.T) {
	t.Parallel()
	cmds, err := Load()
	require.NoError(t, err)
	for _, c := range cmds {
		require.NotEmpty(t, c.Content, "command %q has empty content", c.Name)
	}
}

func TestLoad_KnownCommandNames(t *testing.T) {
	t.Parallel()
	cmds, err := Load()
	require.NoError(t, err)
	names := make(map[string]bool)
	for _, c := range cmds {
		names[c.Name] = true
	}
	expected := []string{
		"ralph-loop",
		"ulw-loop",
		"cancel-ralph",
		"refactor",
		"start-work",
		"stop-continuation",
		"remove-ai-slops",
		"handoff",
		"hyperplan",
	}
	for _, name := range expected {
		require.True(t, names[name], "expected builtin command %q to be present", name)
	}
}

func TestParse_Frontmatter(t *testing.T) {
	t.Parallel()
	in := `---
description: (builtin) test command
argument-hint: '"task" [--flag=val]'
---

# Body

This is the body of the test command.`
	c := parse("test", in)
	require.Equal(t, "test", c.Name)
	require.Contains(t, c.Content, "# Body")
	require.Contains(t, c.Content, "This is the body of the test command.")
}

func TestParse_NoFrontmatter(t *testing.T) {
	t.Parallel()
	in := "Just a body, no frontmatter."
	c := parse("plain", in)
	require.Equal(t, "plain", c.Name)
	require.Equal(t, in, c.Content)
}

func TestParse_UnterminatedFrontmatter(t *testing.T) {
	t.Parallel()
	in := "---\ndescription: never closes\n\n# Body"
	c := parse("bad", in)
	require.Equal(t, "bad", c.Name)
	require.Equal(t, in, c.Content)
}

func TestParse_ExtractsArguments(t *testing.T) {
	t.Parallel()
	in := `---
description: x
---

Use $ARGUMENT1 and $ARGUMENT2 here. $ARGUMENT1 again.`
	c := parse("withargs", in)
	require.Len(t, c.Arguments, 2)
	// Deduplicated; order is the first-seen order.
	require.Equal(t, "ARGUMENT1", c.Arguments[0].ID)
	require.Equal(t, "ARGUMENT2", c.Arguments[1].ID)
}
