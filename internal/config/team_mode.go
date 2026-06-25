package config

// TeamModeConfig configures parallel multi-agent coordination.
// Mirrors packages/omo-opencode/src/config/schema/team-mode.ts.
//
// OFF by default. Enable via heretic.json:
//
//	{
//	  "team_mode": {
//	    "enabled": true,
//	    "tmux_visualization": false,
//	    "max_parallel_members": 4,
//	    "max_members": 8,
//	    "max_messages_per_run": 10000,
//	    "max_wall_clock_minutes": 120,
//	    "max_member_turns": 500,
//	    "base_dir": null,
//	    "message_payload_max_bytes": 32768,
//	    "recipient_unread_max_bytes": 262144,
//	    "mailbox_poll_interval_ms": 3000
//	  }
//	}
type TeamModeConfig struct {
	Enabled                  bool   `json:"enabled,omitempty" jsonschema:"description=Enable team mode (parallel multi-agent coordination),default=false"`
	TmuxVisualization        bool   `json:"tmux_visualization,omitempty" jsonschema:"description=Visualize team members in tmux panes,default=false"`
	MaxParallelMembers       int    `json:"max_parallel_members,omitempty" jsonschema:"description=Maximum concurrent member spawns (1..8),default=4,minimum=1,maximum=8"`
	MaxMembers               int    `json:"max_members,omitempty" jsonschema:"description=Maximum total members per team (1..8),default=8,minimum=1,maximum=8"`
	MaxMessagesPerRun        int    `json:"max_messages_per_run,omitempty" jsonschema:"description=Maximum messages any single member may process per run,default=10000,minimum=1"`
	MaxWallClockMinutes      int    `json:"max_wall_clock_minutes,omitempty" jsonschema:"description=Hard wall-clock cap for any member run in minutes,default=120,minimum=1"`
	MaxMemberTurns           int    `json:"max_member_turns,omitempty" jsonschema:"description=Maximum LLM turns any single member may consume,default=500,minimum=1"`
	BaseDir                  string `json:"base_dir,omitempty" jsonschema:"description=Override default teams directory (default: ~/.omo/teams or <project>/.omo/teams)"`
	MessagePayloadMaxBytes   int    `json:"message_payload_max_bytes,omitempty" jsonschema:"description=Maximum size of a single inter-agent message payload in bytes,default=32768,minimum=1024"`
	RecipientUnreadMaxBytes  int    `json:"recipient_unread_max_bytes,omitempty" jsonschema:"description=Maximum unread mailbox size per recipient in bytes,default=262144,minimum=1024"`
	MailboxPollIntervalMs    int    `json:"mailbox_poll_interval_ms,omitempty" jsonschema:"description=How often members poll their mailbox in milliseconds,default=3000,minimum=500"`
}

// IsEnabled returns true if team mode is on.
func (t TeamModeConfig) IsEnabled() bool { return t.Enabled }

// EffectiveMaxParallelMembers returns MaxParallelMembers with a sane default
// and bounds check.
func (t TeamModeConfig) EffectiveMaxParallelMembers() int {
	if t.MaxParallelMembers < 1 {
		return 4
	}
	if t.MaxParallelMembers > 8 {
		return 8
	}
	return t.MaxParallelMembers
}

// EffectiveMaxMembers returns MaxMembers with a sane default and bounds check.
func (t TeamModeConfig) EffectiveMaxMembers() int {
	if t.MaxMembers < 1 {
		return 8
	}
	if t.MaxMembers > 8 {
		return 8
	}
	return t.MaxMembers
}

// EffectiveMessagePayloadMaxBytes returns the max with bounds.
func (t TeamModeConfig) EffectiveMessagePayloadMaxBytes() int {
	if t.MessagePayloadMaxBytes < 1024 {
		return 32768
	}
	return t.MessagePayloadMaxBytes
}

// EffectiveRecipientUnreadMaxBytes returns the max with bounds.
func (t TeamModeConfig) EffectiveRecipientUnreadMaxBytes() int {
	if t.RecipientUnreadMaxBytes < 1024 {
		return 262144
	}
	return t.RecipientUnreadMaxBytes
}

// EffectiveMailboxPollInterval returns the poll interval as a time.Duration.
func (t TeamModeConfig) EffectiveMailboxPollInterval() int {
	if t.MailboxPollIntervalMs < 500 {
		return 3000
	}
	return t.MailboxPollIntervalMs
}

// EffectiveMaxWallClockMinutes returns wall-clock cap with default.
func (t TeamModeConfig) EffectiveMaxWallClockMinutes() int {
	if t.MaxWallClockMinutes < 1 {
		return 120
	}
	return t.MaxWallClockMinutes
}

// EffectiveMaxMemberTurns returns the turn cap with default.
func (t TeamModeConfig) EffectiveMaxMemberTurns() int {
	if t.MaxMemberTurns < 1 {
		return 500
	}
	return t.MaxMemberTurns
}
