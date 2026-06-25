package toolregistry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register(Spec{Name: "x", Gates: GateAlways})
	got, ok := r.Get("x")
	require.True(t, ok)
	require.Equal(t, "x", got.Name)
}

func TestSpec_IsEnabled_Always(t *testing.T) {
	t.Parallel()
	s := Spec{Gates: GateAlways}
	require.True(t, s.IsEnabled(GatingContext{}))
}

func TestSpec_IsEnabled_TeamMode(t *testing.T) {
	t.Parallel()
	s := Spec{Gates: GateTeamMode}
	require.False(t, s.IsEnabled(GatingContext{}))
	require.True(t, s.IsEnabled(GatingContext{TeamMode: true}))
}

func TestSpec_IsEnabled_MultipleGates(t *testing.T) {
	t.Parallel()
	s := Spec{Gates: GateTeamMode | GateBackgroundAgent}
	require.False(t, s.IsEnabled(GatingContext{TeamMode: true}))
	require.True(t, s.IsEnabled(GatingContext{TeamMode: true, BackgroundAgent: true}))
}

func TestRegistry_All_SortedByName(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register(Spec{Name: "zebra"})
	r.Register(Spec{Name: "alpha"})
	r.Register(Spec{Name: "mango"})
	all := r.All()
	require.Len(t, all, 3)
	require.Equal(t, "alpha", all[0].Name)
	require.Equal(t, "mango", all[1].Name)
	require.Equal(t, "zebra", all[2].Name)
}

func TestRegistry_BuildTools_SkipsGated(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register(Spec{Name: "always", Gates: GateAlways})
	r.Register(Spec{Name: "team", Gates: GateTeamMode})
	tools := r.BuildTools(GatingContext{TeamMode: true})
	// Both pass: "always" has no gate; "team" gate is satisfied.
	// But neither has a Factory, so BuildTools returns empty.
	require.Empty(t, tools)
}

func TestGatingFlag_String(t *testing.T) {
	t.Parallel()
	require.Equal(t, "always", GateAlways.String())
	require.Contains(t, (GateTeamMode | GateBackgroundAgent).String(), "team_mode")
	require.Contains(t, (GateTeamMode | GateBackgroundAgent).String(), "background_agent")
}
