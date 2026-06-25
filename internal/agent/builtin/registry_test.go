package builtin

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllNames(t *testing.T) {
	t.Parallel()
	names := AllNames()
	require.Len(t, names, 11)
}

func TestDefaultAgents_AllUnique(t *testing.T) {
	t.Parallel()
	agents := DefaultAgents()
	seen := make(map[Name]bool, len(agents))
	for _, a := range agents {
		require.False(t, seen[a.AgentName()], "duplicate agent %q", a.AgentName())
		seen[a.AgentName()] = true
	}
	require.Len(t, agents, 11)
}

func TestDefaultAgents_AllHavePrompts(t *testing.T) {
	t.Parallel()
	for _, a := range DefaultAgents() {
		require.NotEmpty(t, a.SystemPrompt(), "agent %q has empty system prompt", a.AgentName())
		require.NotEmpty(t, a.Description(), "agent %q has empty description", a.AgentName())
		require.NotEmpty(t, a.DisplayName(), "agent %q has empty display name", a.AgentName())
		require.NotEmpty(t, a.AllowedTools(), "agent %q has empty tool list", a.AgentName())
	}
}

func TestRegistry_GetAndMiss(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	a, ok := r.Get(NameSisyphus)
	require.True(t, ok)
	require.Equal(t, NameSisyphus, a.AgentName())
	_, ok = r.Get("nonexistent")
	require.False(t, ok)
}

func TestRegistry_GetOrError(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	_, err := r.GetOrError("nope")
	require.Error(t, err)
	a, err := r.GetOrError(NameExplore)
	require.NoError(t, err)
	require.Equal(t, NameExplore, a.AgentName())
}

func TestRegistry_Names(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	names := r.Names()
	require.Len(t, names, 11)
	require.True(t, sort.StringsAreSorted(toStrings(names)))
}

func TestRegistry_PrimariesAndSubagents(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	primaries := r.Primaries()
	subagents := r.Subagents()
	// sisyphus, hephaestus, atlas, prometheus → 4 primaries
	require.Len(t, primaries, 4)
	// oracle, librarian, explore, multimodal-looker, metis, momus, sisyphus-junior → 7 subagents
	require.Len(t, subagents, 7)
	// Total: 11
	require.Equal(t, 11, len(primaries)+len(subagents))
}

func TestRegistry_Register(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	custom := &baseAgent{
		name:         "custom",
		display:      "Custom",
		mode:         ModeSubagent,
		modelPref:    "small",
		systemPrompt: "x",
		description:  "x",
		allowedTools: []ToolRef{{Name: "view"}},
	}
	r.Register(custom)
	got, err := r.GetOrError("custom")
	require.NoError(t, err)
	require.Equal(t, "Custom", got.DisplayName())
}

func toStrings(ns []Name) []string {
	out := make([]string, len(ns))
	for i, n := range ns {
		out[i] = string(n)
	}
	return out
}

func TestMode_Constants(t *testing.T) {
	t.Parallel()
	require.Equal(t, Mode("primary"), ModePrimary)
	require.Equal(t, Mode("subagent"), ModeSubagent)
	require.Equal(t, Mode("all"), ModeAll)
}
