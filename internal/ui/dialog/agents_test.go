package dialog

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestMainAgents(t *testing.T) {
	t.Parallel()
	agents := MainAgents()
	require.Len(t, agents, 4)
	require.Equal(t, AgentSisyphus, agents[0])
	require.Equal(t, AgentHephaestus, agents[1])
	require.Equal(t, AgentAtlas, agents[2])
	require.Equal(t, AgentPrometheus, agents[3])
}

func TestAgentDescription(t *testing.T) {
	t.Parallel()
	require.NotEmpty(t, AgentDescription(AgentSisyphus))
	require.NotEmpty(t, AgentDescription(AgentHephaestus))
	require.NotEmpty(t, AgentDescription(AgentAtlas))
	require.NotEmpty(t, AgentDescription(AgentPrometheus))
	require.Empty(t, AgentDescription("unknown"))
}

func TestAgentSwitch_ID(t *testing.T) {
	t.Parallel()
	a := &AgentSwitch{}
	require.Equal(t, AgentSwitchID, a.ID())
}

func newTestAgentSwitch(current AgentName) *AgentSwitch {
	a := &AgentSwitch{current: current, selectedIx: indexOf(current)}
	a.keyMap.Up = key.NewBinding(key.WithKeys("up"))
	a.keyMap.Down = key.NewBinding(key.WithKeys("down"))
	a.keyMap.Tab = key.NewBinding(key.WithKeys("tab"))
	a.keyMap.ShiftTab = key.NewBinding(key.WithKeys("shift+tab"))
	a.keyMap.Left = key.NewBinding(key.WithKeys("left"))
	a.keyMap.Right = key.NewBinding(key.WithKeys("right"))
	a.keyMap.Select = key.NewBinding(key.WithKeys("enter"))
	a.keyMap.Close = key.NewBinding(key.WithKeys("esc"))
	return a
}

func indexOf(a AgentName) int {
	for i, name := range MainAgents() {
		if name == a {
			return i
		}
	}
	return 0
}

func TestAgentSwitch_HandleMsg_Number(t *testing.T) {
	t.Parallel()
	a := newTestAgentSwitch(AgentSisyphus)
	// Pressing "1" returns Select with sisyphus.
	action := a.HandleMsg(tea.KeyPressMsg{Text: "1"})
	selectAction, ok := action.(ActionSelectAgent)
	require.True(t, ok)
	require.Equal(t, AgentSisyphus, selectAction.Agent)
}

func TestAgentSwitch_HandleMsg_Number4(t *testing.T) {
	t.Parallel()
	a := newTestAgentSwitch(AgentSisyphus)
	action := a.HandleMsg(tea.KeyPressMsg{Text: "4"})
	selectAction, ok := action.(ActionSelectAgent)
	require.True(t, ok)
	require.Equal(t, AgentPrometheus, selectAction.Agent)
}

func TestAgentSwitch_HandleMsg_Close(t *testing.T) {
	t.Parallel()
	a := newTestAgentSwitch(AgentSisyphus)
	action := a.HandleMsg(tea.KeyPressMsg{Text: "esc"})
	_, ok := action.(ActionClose)
	require.True(t, ok)
}

func TestAgentSwitch_HandleMsg_Select(t *testing.T) {
	t.Parallel()
	a := newTestAgentSwitch(AgentHephaestus)
	action := a.HandleMsg(tea.KeyPressMsg{Text: "enter"})
	selectAction, ok := action.(ActionSelectAgent)
	require.True(t, ok)
	require.Equal(t, AgentHephaestus, selectAction.Agent)
}

func TestAgentSwitch_Navigation(t *testing.T) {
	t.Parallel()
	a := newTestAgentSwitch(AgentSisyphus)
	// Pressing down moves to next.
	a.HandleMsg(tea.KeyPressMsg{Code: tea.KeyDown})
	require.Equal(t, AgentHephaestus, a.Selected())
	// Wrap around.
	a.selectedIx = 3
	a.HandleMsg(tea.KeyPressMsg{Code: tea.KeyDown})
	require.Equal(t, AgentSisyphus, a.Selected())
	// Up wraps backward.
	a.selectedIx = 0
	a.HandleMsg(tea.KeyPressMsg{Code: tea.KeyUp})
	require.Equal(t, AgentPrometheus, a.Selected())
}
