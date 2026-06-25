package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTeamModeConfig_IsEnabled(t *testing.T) {
	t.Parallel()
	var c TeamModeConfig
	require.False(t, c.IsEnabled())
	c.Enabled = true
	require.True(t, c.IsEnabled())
}

func TestTeamModeConfig_EffectiveBounds(t *testing.T) {
	t.Parallel()
	c := TeamModeConfig{
		MaxParallelMembers:      -1,
		MaxMembers:              999,
		MessagePayloadMaxBytes:  0,
		RecipientUnreadMaxBytes: 100,
		MailboxPollIntervalMs:   100,
		MaxWallClockMinutes:     0,
		MaxMemberTurns:          -5,
	}
	require.Equal(t, 4, c.EffectiveMaxParallelMembers())
	require.Equal(t, 8, c.EffectiveMaxMembers())
	require.Equal(t, 32768, c.EffectiveMessagePayloadMaxBytes())
	require.Equal(t, 262144, c.EffectiveRecipientUnreadMaxBytes())
	require.Equal(t, 3000, c.EffectiveMailboxPollInterval())
	require.Equal(t, 120, c.EffectiveMaxWallClockMinutes())
	require.Equal(t, 500, c.EffectiveMaxMemberTurns())
}

func TestTeamModeConfig_JSONRoundtrip(t *testing.T) {
	t.Parallel()
	in := TeamModeConfig{
		Enabled:                 true,
		MaxParallelMembers:      3,
		MaxMembers:              6,
		MessagePayloadMaxBytes:  16384,
		RecipientUnreadMaxBytes: 131072,
		MailboxPollIntervalMs:   1500,
	}
	b, err := json.Marshal(in)
	require.NoError(t, err)
	var out TeamModeConfig
	require.NoError(t, json.Unmarshal(b, &out))
	require.Equal(t, in, out)
}

func TestBackgroundAgentConfig_EffectiveModelConcurrency(t *testing.T) {
	t.Parallel()
	var b BackgroundAgentConfig
	require.Equal(t, 5, b.EffectiveModelConcurrency())
	b.ModelConcurrency = 8
	require.Equal(t, 8, b.EffectiveModelConcurrency())
}

func TestBackgroundAgentConfig_ProviderOverride(t *testing.T) {
	t.Parallel()
	b := BackgroundAgentConfig{
		ModelConcurrency:      5,
		ProviderConcurrency: map[string]int{"anthropic": 2},
	}
	require.Equal(t, 2, b.EffectiveConcurrencyFor("anthropic"))
	require.Equal(t, 5, b.EffectiveConcurrencyFor("openai"))
}

func TestBackgroundAgentConfig_EffectiveParentWakeTimeout(t *testing.T) {
	t.Parallel()
	var b BackgroundAgentConfig
	require.Equal(t, 30000, b.EffectiveParentWakeTimeoutMs())
	b.ParentWakeTimeoutMs = 5000
	require.Equal(t, 5000, b.EffectiveParentWakeTimeoutMs())
}

func TestBoulderStateConfig_EffectiveMaxHistory(t *testing.T) {
	t.Parallel()
	var b BoulderStateConfig
	require.Equal(t, 100, b.EffectiveMaxHistory())
	b.MaxHistory = 50
	require.Equal(t, 50, b.EffectiveMaxHistory())
}
