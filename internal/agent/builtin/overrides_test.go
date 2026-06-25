package builtin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOverrideApplier_Apply(t *testing.T) {
	t.Parallel()
	base := NewSisyphusAgent()
	override := AgentOverrideConfig{
		Model:       "small",
		PromptAppend: "ALWAYS use tabs for indentation.",
		Tools:       map[string]bool{"bash": false},
	}
	applied := OverrideApplier{Base: base, Override: override}.Apply()
	require.Equal(t, "small", applied.ModelPreference())
	require.Contains(t, applied.SystemPrompt(), "ALWAYS use tabs")
	require.Contains(t, applied.SystemPrompt(), "Sisyphus") // base prompt preserved

	// Tools override: bash should be disabled.
	tools := applied.AllowedTools()
	var bash *ToolRef
	for i, t := range tools {
		if t.Name == "bash" {
			bash = &tools[i]
			break
		}
	}
	require.NotNil(t, bash)
	require.NotNil(t, bash.Enabled)
	require.False(t, *bash.Enabled)
}

func TestOverrideApplier_NoChangesWhenEmpty(t *testing.T) {
	t.Parallel()
	base := NewSisyphusAgent()
	applied := OverrideApplier{Base: base, Override: AgentOverrideConfig{}}.Apply()
	require.Equal(t, base.ModelPreference(), applied.ModelPreference())
	require.Equal(t, base.SystemPrompt(), applied.SystemPrompt())
}

func TestApplyOverrides_Registry(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	overrides := AgentOverrides{
		NameExplore: {Model: "large"},
		NameSisyphus: {PromptAppend: "Be concise."},
	}
	ApplyOverrides(r, overrides)
	exp, err := r.GetOrError(NameExplore)
	require.NoError(t, err)
	require.Equal(t, "large", exp.ModelPreference())
	sis, err := r.GetOrError(NameSisyphus)
	require.NoError(t, err)
	require.Contains(t, sis.SystemPrompt(), "Be concise.")
}

func TestApplyOverrides_UnknownAgentIgnored(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	overrides := AgentOverrides{
		"nonexistent": {Model: "large"},
	}
	ApplyOverrides(r, overrides) // should not panic
}
