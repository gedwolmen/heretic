// Package hookengine is the comprehensive hook engine for heretic Ultimate.
// It extends the basic hooks/ package with the 12 events and 5 tiers
// required by om-my-openagent's full feature set.
//
// The engine is layered: the Tier (Session/ToolGuard/Transform/
// Continuation/Skill) determines WHEN the hook fires; the Event
// determines WHAT data the hook receives.
package hookengine

// Tier is the hook's classification. The engine registers hooks by tier
// and fires them in a defined order within each tier.
type Tier string

const (
	// TierSession: hooks that fire once per session (start, end, idle,
	// error). Examples: agent-usage-reminder, auto-update-checker,
	// session-notification.
	TierSession Tier = "session"
	// TierToolGuard: hooks that guard tool execution. Examples:
	// write-existing-file-guard, comment-checker, bash-file-read-guard.
	TierToolGuard Tier = "toolguard"
	// TierTransform: hooks that transform the LLM's view of the world.
	// Examples: directory-agents-injector, category-skill-reminder.
	TierTransform Tier = "transform"
	// TierContinuation: hooks that keep work going across turns.
	// Examples: todo-continuation-enforcer, start-work-continuation.
	TierContinuation Tier = "continuation"
	// TierSkill: hooks that bridge skills into the agent loop.
	// Examples: keyword-detector, model-fallback.
	TierSkill Tier = "skill"
)

// AllTiers returns the 5 tiers in firing order.
func AllTiers() []Tier {
	return []Tier{TierSession, TierToolGuard, TierTransform, TierContinuation, TierSkill}
}

// Event is the hook event name. These mirror the OpenCode SDK events.
type Event string

const (
	// EventSessionStart fires when a new session is created.
	EventSessionStart Event = "session.start"
	// EventSessionEnd fires when a session ends cleanly.
	EventSessionEnd Event = "session.end"
	// EventSessionIdle fires when the session becomes idle.
	EventSessionIdle Event = "session.idle"
	// EventSessionError fires on a session-level error.
	EventSessionError Event = "session.error"
	// EventMessageReceived fires when a user message arrives.
	EventMessageReceived Event = "message.received"
	// EventPreToolUse fires before a tool call.
	EventPreToolUse Event = "tool.pre"
	// EventPostToolUse fires after a tool call.
	EventPostToolUse Event = "tool.post"
	// EventToolError fires when a tool call fails.
	EventToolError Event = "tool.error"
	// EventSystemTransform fires to transform the system prompt.
	EventSystemTransform Event = "system.transform"
	// EventChatTransform fires to transform chat messages.
	EventChatTransform Event = "chat.transform"
	// EventCompaction fires when context is being compacted.
	EventCompaction Event = "session.compacting"
	// EventAutoContinue fires after compaction to auto-resume.
	EventAutoContinue Event = "compaction.autocontinue"
)

// AllEvents returns the 12 supported events.
func AllEvents() []Event {
	return []Event{
		EventSessionStart, EventSessionEnd, EventSessionIdle, EventSessionError,
		EventMessageReceived,
		EventPreToolUse, EventPostToolUse, EventToolError,
		EventSystemTransform, EventChatTransform,
		EventCompaction, EventAutoContinue,
	}
}

// EventTiers maps an event to the tiers that may handle it.
func EventTiers(e Event) []Tier {
	switch e {
	case EventSessionStart, EventSessionEnd, EventSessionIdle, EventSessionError:
		return []Tier{TierSession}
	case EventMessageReceived:
		return []Tier{TierSession, TierTransform, TierSkill}
	case EventPreToolUse, EventPostToolUse, EventToolError:
		return []Tier{TierToolGuard, TierTransform, TierSkill}
	case EventSystemTransform, EventChatTransform:
		return []Tier{TierTransform, TierSkill}
	case EventCompaction, EventAutoContinue:
		return []Tier{TierContinuation}
	}
	return nil
}
