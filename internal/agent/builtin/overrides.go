package builtin

// AgentOverrideConfig lets users override built-in agent properties
// via heretic.json. Mirrors AgentOverrideConfig from
// packages/omo-opencode/src/agents/types.ts.
type AgentOverrideConfig struct {
	// Model overrides the agent's model preference.
	Model string `json:"model,omitempty"`
	// Variant overrides the model variant.
	Variant string `json:"variant,omitempty"`
	// PromptAppend is appended to the agent's system prompt.
	PromptAppend string `json:"prompt_append,omitempty"`
	// Skills lists the skills to inject for this agent.
	Skills []string `json:"skills,omitempty"`
	// Tools maps tool names to enabled/disabled.
	Tools map[string]bool `json:"tools,omitempty"`
	// Category maps this agent to a default category.
	Category string `json:"category,omitempty"`
	// FallbackModels is the ordered list of fallback models.
	FallbackModels []string `json:"fallback_models,omitempty"`
}

// AgentOverrides is a map of agent name to its override config.
type AgentOverrides map[Name]AgentOverrideConfig

// OverrideApplier applies AgentOverrideConfig to a base Agent.
type OverrideApplier struct {
	Base      Agent
	Override  AgentOverrideConfig
}

// Apply returns a new Agent with the override applied. The base agent's
// prompt, model, and tool list are mutated according to the override.
func (o OverrideApplier) Apply() Agent {
	return &overriddenAgent{
		base:     o.Base,
		override: o.Override,
	}
}

// overriddenAgent wraps a base Agent and applies overrides lazily.
type overriddenAgent struct {
	base     Agent
	override AgentOverrideConfig
}

func (o *overriddenAgent) AgentName() Name { return o.base.AgentName() }
func (o *overriddenAgent) DisplayName() string { return o.base.DisplayName() }
func (o *overriddenAgent) Mode() Mode { return o.base.Mode() }

func (o *overriddenAgent) ModelPreference() string {
	if o.override.Model != "" {
		return o.override.Model
	}
	return o.base.ModelPreference()
}

func (o *overriddenAgent) AllowedTools() []ToolRef {
	if len(o.override.Tools) == 0 {
		return o.base.AllowedTools()
	}
	base := o.base.AllowedTools()
	out := make([]ToolRef, 0, len(base))
	for _, t := range base {
		if override, ok := o.override.Tools[t.Name]; ok {
			enabled := override
			out = append(out, ToolRef{Name: t.Name, Enabled: &enabled})
		} else {
			out = append(out, t)
		}
	}
	return out
}

func (o *overriddenAgent) SystemPrompt() string {
	if o.override.PromptAppend == "" {
		return o.base.SystemPrompt()
	}
	return o.base.SystemPrompt() + "\n\n" + o.override.PromptAppend
}

func (o *overriddenAgent) Description() string { return o.base.Description() }

// ApplyOverrides returns a new Registry with overrides applied.
func ApplyOverrides(r *Registry, overrides AgentOverrides) *Registry {
	for name, ov := range overrides {
		if base, ok := r.Get(name); ok {
			r.Register(OverrideApplier{Base: base, Override: ov}.Apply())
		}
	}
	return r
}
