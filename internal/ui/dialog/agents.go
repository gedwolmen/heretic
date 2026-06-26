package dialog

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/gedwolmen/heretic/internal/ui/common"
)

// AgentSwitchID is the identifier for the agent switch dialog.
const AgentSwitchID = "agent-switch"

// AgentName is the registry name of a built-in agent.
type AgentName string

// The 4 main agents exposed in the switcher. Subagent-only agents
// (oracle, librarian, explore, multimodal-looker, metis, momus,
// sisyphus-junior) are intentionally not switchable from the UI; they
// are reached via the task tool from a main agent.
const (
	AgentSisyphus   AgentName = "sisyphus"
	AgentHephaestus AgentName = "hephaestus"
	AgentAtlas      AgentName = "atlas"
	AgentPrometheus AgentName = "prometheus"
)

// MainAgents returns the 4 main agents in display order.
func MainAgents() []AgentName {
	return []AgentName{AgentSisyphus, AgentHephaestus, AgentAtlas, AgentPrometheus}
}

// AgentDescription returns a short blurb for an agent.
func AgentDescription(a AgentName) string {
	switch a {
	case AgentSisyphus:
		return "Master orchestrator"
	case AgentHephaestus:
		return "Plan-driven builder"
	case AgentAtlas:
		return "Multi-agent orchestrator"
	case AgentPrometheus:
		return "Plan generator (.md only)"
	}
	return ""
}

// ActionSelectAgent is returned when the user picks an agent.
type ActionSelectAgent struct {
	Agent AgentName
}

// AgentSwitch is a dialog that lets the user pick one of the 4 main
// agents. It uses 4 buttons in a 2x2 grid (Tab / Shift+Tab to move,
// Enter to confirm, Esc to cancel).
type AgentSwitch struct {
	com        *common.Common
	current    AgentName
	selectedIx int
	keyMap     struct {
		LeftRight,
		UpDown,
		Left,
		Right,
		Up,
		Down,
		Tab,
		ShiftTab,
		Select,
		Close,
		Number key.Binding
	}
	help help.Model
}

var _ Dialog = (*AgentSwitch)(nil)

// NewAgentSwitch creates a new agent switch dialog with the given
// currently-active agent.
func NewAgentSwitch(com *common.Common, current AgentName) *AgentSwitch {
	a := &AgentSwitch{
		com:     com,
		current: current,
	}
	// Default to the current agent's index.
	for i, name := range MainAgents() {
		if name == current {
			a.selectedIx = i
			break
		}
	}

	a.help = help.New()
	a.help.Styles = com.Styles.DialogHelpStyles()

	a.keyMap.LeftRight = key.NewBinding(
		key.WithKeys("left", "right"),
		key.WithHelp("←/→", "switch column"),
	)
	a.keyMap.Left = key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "left"),
	)
	a.keyMap.Right = key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "right"),
	)
	a.keyMap.UpDown = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑/↓", "switch agent"),
	)
	a.keyMap.Up = key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "prev"),
	)
	a.keyMap.Down = key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "next"),
	)
	a.keyMap.Tab = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next"),
	)
	a.keyMap.ShiftTab = key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev"),
	)
	a.keyMap.Select = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	)
	a.keyMap.Close = CloseKey
	a.keyMap.Number = key.NewBinding(
		key.WithKeys("1", "2", "3", "4"),
		key.WithHelp("1-4", "quick pick"),
	)
	return a
}

// ID implements Dialog.
func (a *AgentSwitch) ID() string {
	return AgentSwitchID
}

// Selected returns the currently-selected agent (for preview).
func (a *AgentSwitch) Selected() AgentName {
	return MainAgents()[a.selectedIx]
}

// Current returns the agent that was active when the dialog opened.
func (a *AgentSwitch) Current() AgentName {
	return a.current
}

// HandleMsg implements Dialog.
func (a *AgentSwitch) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		text := msg.Text
		_ = text
		// Number key quick-pick.
		switch text {
		case "1":
			return ActionSelectAgent{Agent: MainAgents()[0]}
		case "2":
			return ActionSelectAgent{Agent: MainAgents()[1]}
		case "3":
			return ActionSelectAgent{Agent: MainAgents()[2]}
		case "4":
			return ActionSelectAgent{Agent: MainAgents()[3]}
		}
		switch {
		case key.Matches(msg, a.keyMap.Close):
			return ActionClose{}
		case key.Matches(msg, a.keyMap.Select):
			return ActionSelectAgent{Agent: a.Selected()}
		case key.Matches(msg, a.keyMap.Right, a.keyMap.LeftRight):
			a.selectedIx = (a.selectedIx + 1) % len(MainAgents())
		case key.Matches(msg, a.keyMap.Left):
			a.selectedIx = (a.selectedIx - 1 + len(MainAgents())) % len(MainAgents())
		case key.Matches(msg, a.keyMap.Down):
			a.selectedIx = (a.selectedIx + 1) % len(MainAgents())
		case key.Matches(msg, a.keyMap.Up):
			a.selectedIx = (a.selectedIx - 1 + len(MainAgents())) % len(MainAgents())
		case key.Matches(msg, a.keyMap.Tab):
			a.selectedIx = (a.selectedIx + 1) % len(MainAgents())
		case key.Matches(msg, a.keyMap.ShiftTab):
			a.selectedIx = (a.selectedIx - 1 + len(MainAgents())) % len(MainAgents())
		}
	}
	return nil
}

// Cursor implements Dialog.
func (a *AgentSwitch) Cursor() *tea.Cursor { return nil }

// Draw implements Dialog.
//
// Layout: the 4 main agents are laid out in a single horizontal row.
// Buttons are padded to a fixed width so the row stays the same width
// regardless of which agent is selected. Below the row: the agent's
// description, wrapped to a fixed width so it doesn't reflow.
//
// The dialog's width is fixed at 60 columns; the body is centered
// horizontally within that. Navigation is Left/Right (and 1-4 number
// keys for quick pick). Up/Down wrap.
func (a *AgentSwitch) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := a.com.Styles
	agents := MainAgents()

	width := 60
	if area.Dx() < width {
		width = area.Dx()
	}
	if area.Dx() < 30 {
		// Pathological: tiny terminal. Just bail.
		return nil
	}

	// All 4 agents in a single row, padded to a fixed width so the
	// row stays the same width regardless of which is selected.
	// "hephaestus" is the longest at 10 chars; pad to 10.
	btns := make([]common.ButtonOpts, 0, len(agents))
	underlineIdx := map[AgentName]int{
		AgentSisyphus:   0, // S
		AgentHephaestus: 0, // H
		AgentAtlas:      0, // A
		AgentPrometheus: 0, // P
	}
	for i, name := range agents {
		label := string(name)
		for len(label) < 10 {
			label += " "
		}
		btns = append(btns, common.ButtonOpts{
			Text:          label,
			Selected:      i == a.selectedIx,
			Padding:       1,
			UnderlineIndex: underlineIdx[name],
		})
	}
	row := common.ButtonGroup(t, btns, "  ")

	// Description for the selected agent. Rendered as a fixed-width
	// pane so it doesn't reflow when the selection moves. We also
	// show the current vs pending state on a second line.
	desc := AgentDescription(a.Selected())
	stateLabel := ""
	if a.Selected() == a.Current() {
		stateLabel = "current agent"
	} else {
		stateLabel = "switch to this agent on Enter"
	}
	descInnerWidth := width - 8 // account for dialog frame + padding
	if descInnerWidth < 10 {
		descInnerWidth = 10
	}
	// Two lines: name + state. Padded to width so lipgloss doesn't
	// auto-wrap. Both lines are rendered with the same width so the
	// dialog never reflows.
	descLine := desc
	stateLine := stateLabel
	for len(descLine) < descInnerWidth {
		descLine += " "
	}
	for len(stateLine) < descInnerWidth {
		stateLine += " "
	}
	descStyled := t.Dialog.HelpView.Width(descInnerWidth).Render(descLine)
	stateStyled := t.Dialog.HelpView.Width(descInnerWidth).Render(stateLine)

	body := lipgloss.JoinVertical(
		lipgloss.Center,
		row,
		"",
		"",
		descStyled,
		stateStyled,
	)
	rc := NewRenderContext(t, width)
	rc.Title = "Select Agent"
	rc.AddPart(body)
	view := rc.Render()
	DrawCenter(scr, area, view)
	return nil
}

// wrapText is kept for reference; the description is now padded to
// the inner width instead of wrapped (see Draw).
func wrapText(s string, maxWidth int) string {
	if maxWidth <= 0 || len(s) <= maxWidth {
		return s
	}
	// Find a space to break on at or before maxWidth.
	breakAt := -1
	for i := maxWidth; i > 0; i-- {
		if i < len(s) && s[i] == ' ' {
			breakAt = i
			break
		}
	}
	if breakAt < 0 {
		// No space — hard break at maxWidth.
		return s[:maxWidth]
	}
	return s[:breakAt] + "\n" + wrapText(s[breakAt+1:], maxWidth)
}

// ShortHelp implements help.KeyMap.
func (a *AgentSwitch) ShortHelp() []key.Binding {
	return []key.Binding{
		a.keyMap.UpDown,
		a.keyMap.Select,
		a.keyMap.Close,
	}
}

// FullHelp implements help.KeyMap.
func (a *AgentSwitch) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{a.keyMap.UpDown, a.keyMap.LeftRight, a.keyMap.Tab, a.keyMap.ShiftTab},
		{a.keyMap.Number, a.keyMap.Select, a.keyMap.Close},
	}
}
