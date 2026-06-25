// Package toolregistry is the tool catalog with gating flags. It decides
// which tools are exposed to a given session based on configuration.
package toolregistry

import (
	"fmt"
	"sort"
	"sync"

	"charm.land/fantasy"
)

// GatingFlag is a bit-set of conditions that gate a tool.
type GatingFlag uint32

const (
	// GateAlways: tool is always available.
	GateAlways GatingFlag = 0
	// GateTeamMode: tool requires team_mode.enabled.
	GateTeamMode GatingFlag = 1 << iota
	// GateTaskSystem: tool requires experimental.task_system.
	GateTaskSystem
	// GateHashlineEdit: tool requires hashline_edit to be on.
	GateHashlineEdit
	// GateInteractiveBash: tool requires the `tmux` binary on PATH.
	GateInteractiveBash
	// GateLookAt: tool requires the multimodal-looker to be enabled.
	GateLookAt
	// GateBackgroundAgent: tool requires background agents.
	GateBackgroundAgent
)

// String returns the human-readable list of gates.
func (g GatingFlag) String() string {
	if g == GateAlways {
		return "always"
	}
	var parts []string
	if g&GateTeamMode != 0 {
		parts = append(parts, "team_mode")
	}
	if g&GateTaskSystem != 0 {
		parts = append(parts, "task_system")
	}
	if g&GateHashlineEdit != 0 {
		parts = append(parts, "hashline_edit")
	}
	if g&GateInteractiveBash != 0 {
		parts = append(parts, "interactive_bash")
	}
	if g&GateLookAt != 0 {
		parts = append(parts, "look_at")
	}
	if g&GateBackgroundAgent != 0 {
		parts = append(parts, "background_agent")
	}
	return fmt.Sprintf("gated(%v)", parts)
}

// Spec describes one tool in the catalog.
type Spec struct {
	Name        string
	Description string
	Gates       GatingFlag
	Factory     func() fantasy.AgentTool
}

// Registry is the tool catalog.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Spec
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Spec)}
}

// Register adds a tool to the catalog.
func (r *Registry) Register(s Spec) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[s.Name] = s
}

// Get returns the spec for a tool.
func (r *Registry) Get(name string) (Spec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.tools[name]
	return s, ok
}

// All returns all registered tools, sorted by name.
func (r *Registry) All() []Spec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Spec, 0, len(r.tools))
	for _, s := range r.tools {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// GatingContext is the runtime state used to decide gating.
type GatingContext struct {
	TeamMode          bool
	TaskSystem        bool
	HashlineEdit      bool
	InteractiveBash   bool
	LookAt            bool
	BackgroundAgent   bool
}

// IsEnabled returns true if the spec's gates are all satisfied.
func (s Spec) IsEnabled(ctx GatingContext) bool {
	if s.Gates == GateAlways {
		return true
	}
	if s.Gates&GateTeamMode != 0 && !ctx.TeamMode {
		return false
	}
	if s.Gates&GateTaskSystem != 0 && !ctx.TaskSystem {
		return false
	}
	if s.Gates&GateHashlineEdit != 0 && !ctx.HashlineEdit {
		return false
	}
	if s.Gates&GateInteractiveBash != 0 && !ctx.InteractiveBash {
		return false
	}
	if s.Gates&GateLookAt != 0 && !ctx.LookAt {
		return false
	}
	if s.Gates&GateBackgroundAgent != 0 && !ctx.BackgroundAgent {
		return false
	}
	return true
}

// BuildTools instantiates every tool that passes gating. The factory
// may be nil for tools that are registered but not yet implemented.
func (r *Registry) BuildTools(ctx GatingContext) []fantasy.AgentTool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []fantasy.AgentTool
	for _, s := range r.tools {
		if !s.IsEnabled(ctx) {
			continue
		}
		if s.Factory == nil {
			continue
		}
		out = append(out, s.Factory())
	}
	return out
}
