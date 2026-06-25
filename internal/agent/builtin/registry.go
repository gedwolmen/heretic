// Package builtin provides the 11 built-in agents that ship with heretic
// Ultimate. These mirror oh-my-openagent's agent registry but are
// reimplemented in Go against the heretic runtime.
//
// The 11 agents:
//
//	primary:      sisyphus, hephaestus, atlas, prometheus
//	subagent:     oracle, librarian, explore, multimodal-looker,
//	              metis, momus, sisyphus-junior
//
// Each agent is a struct implementing the Agent interface. The Registry
// stores the catalog; consumers iterate it when building a session.
package builtin

import (
	_ "embed"
	"fmt"
	"sort"
	"sync"
)

// Mode is the agent's invocation mode.
type Mode string

const (
	// ModePrimary: the agent is the main session's agent. Exactly one
	// per session.
	ModePrimary Mode = "primary"
	// ModeSubagent: the agent is invoked from another agent via the task
	// tool (subagent delegation). Multiple per session.
	ModeSubagent Mode = "subagent"
	// ModeAll: the agent can be invoked in either role.
	ModeAll Mode = "all"
)

// Name is the registry key for a built-in agent.
type Name string

const (
	NameSisyphus         Name = "sisyphus"
	NameHephaestus       Name = "hephaestus"
	NameOracle           Name = "oracle"
	NameLibrarian        Name = "librarian"
	NameExplore          Name = "explore"
	NameMultimodalLooker Name = "multimodal-looker"
	NameMetis            Name = "metis"
	NameMomus            Name = "momus"
	NameAtlas            Name = "atlas"
	NameSisyphusJunior   Name = "sisyphus-junior"
	NamePrometheus       Name = "prometheus"
)

// AllNames returns the canonical list of all 11 built-in agent names.
func AllNames() []Name {
	return []Name{
		NameSisyphus, NameHephaestus, NameOracle, NameLibrarian,
		NameExplore, NameMultimodalLooker, NameMetis, NameMomus,
		NameAtlas, NameSisyphusJunior, NamePrometheus,
	}
}

// ToolRef is a reference to a tool by name. A nil value means "no access".
// A non-nil bool pointer allows runtime override.
type ToolRef struct {
	Name string
	// Enabled may be nil (use default) or *bool (explicit override).
	Enabled *bool
}

// Agent is the common interface for all built-in agents.
type Agent interface {
	// AgentName returns the registry name.
	AgentName() Name
	// DisplayName returns a human-friendly name for UI rendering.
	DisplayName() string
	// Mode returns the agent's invocation mode.
	Mode() Mode
	// ModelPreference returns the model-selection preference string
	// ("large" / "small" / "max" / etc.). Used by the config-driven
	// model resolver.
	ModelPreference() string
	// AllowedTools returns the list of tools this agent may invoke.
	AllowedTools() []ToolRef
	// SystemPrompt returns the agent's main system prompt. The
	// hook engine may append additional fragments to this.
	SystemPrompt() string
	// Description returns a short blurb shown in the UI / agent
	// delegation table.
	Description() string
}

// baseAgent is a shared partial implementation.
type baseAgent struct {
	name            Name
	display         string
	mode            Mode
	modelPref       string
	systemPrompt    string
	description     string
	allowedTools    []ToolRef
}

// AgentName implements Agent.
func (b *baseAgent) AgentName() Name { return b.name }

// DisplayName implements Agent.
func (b *baseAgent) DisplayName() string { return b.display }

// Mode implements Agent.
func (b *baseAgent) Mode() Mode { return b.mode }

// ModelPreference implements Agent.
func (b *baseAgent) ModelPreference() string { return b.modelPref }

// AllowedTools implements Agent.
func (b *baseAgent) AllowedTools() []ToolRef { return b.allowedTools }

// SystemPrompt implements Agent.
func (b *baseAgent) SystemPrompt() string { return b.systemPrompt }

// Description implements Agent.
func (b *baseAgent) Description() string { return b.description }

// Registry holds the catalog of all built-in agents. Concurrent-safe.
type Registry struct {
	mu     sync.RWMutex
	agents map[Name]Agent
}

// DefaultAgents returns the 11 default built-in agents.
func DefaultAgents() []Agent {
	return []Agent{
		NewSisyphusAgent(),
		NewHephaestusAgent(),
		NewOracleAgent(),
		NewLibrarianAgent(),
		NewExploreAgent(),
		NewMultimodalLookerAgent(),
		NewMetisAgent(),
		NewMomusAgent(),
		NewAtlasAgent(),
		NewSisyphusJuniorAgent(),
		NewPrometheusAgent(),
	}
}

// NewRegistry returns a Registry populated with the 11 default agents.
func NewRegistry() *Registry {
	r := &Registry{agents: make(map[Name]Agent)}
	for _, a := range DefaultAgents() {
		r.agents[a.AgentName()] = a
	}
	return r
}

// Register adds or replaces an agent in the registry.
func (r *Registry) Register(a Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[a.AgentName()] = a
}

// Get returns the agent with the given name.
func (r *Registry) Get(name Name) (Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.agents[name]
	return a, ok
}

// GetOrError is like Get but returns an error on miss.
func (r *Registry) GetOrError(name Name) (Agent, error) {
	a, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("builtin: unknown agent %q", name)
	}
	return a, nil
}

// Names returns all registered agent names, sorted alphabetically.
func (r *Registry) Names() []Name {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Name, 0, len(r.agents))
	for n := range r.agents {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Primaries returns the agents whose Mode is primary or all.
func (r *Registry) Primaries() []Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Agent
	for _, a := range r.agents {
		if a.Mode() == ModePrimary || a.Mode() == ModeAll {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AgentName() < out[j].AgentName() })
	return out
}

// Subagents returns the agents whose Mode is subagent or all.
func (r *Registry) Subagents() []Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Agent
	for _, a := range r.agents {
		if a.Mode() == ModeSubagent || a.Mode() == ModeAll {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AgentName() < out[j].AgentName() })
	return out
}
