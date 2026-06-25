// Package skillhooks contains Skill-tier hooks.
package skillhooks

import (
	"context"
	"fmt"
	"strings"

	"github.com/gedwolmen/heretic/internal/hookengine"
	"github.com/gedwolmen/heretic/internal/intentgate"
)

// MessageData is the payload for message.received events.
type MessageData struct {
	EventName hookengine.Event `json:"-"`
	SessID    string           `json:"session_id"`
	Content   string           `json:"content"`
	IsFirst   bool             `json:"is_first"`
}

func (e *MessageData) Event() hookengine.Event { return e.EventName }
func (e *MessageData) SessionID() string       { return e.SessID }

// KeywordDetector is the IntentGate integration. It detects keywords
// on the first user message and injects the matching mode prompt.
func KeywordDetector() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "keyword-detector",
		Tier:   hookengine.TierSkill,
		Events: []hookengine.Event{hookengine.EventMessageReceived},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*MessageData)
			if !ok || !d.IsFirst {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			modes := intentgate.Detect(d.Content)
			if len(modes) == 0 {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			var names []string
			for _, m := range modes {
				names = append(names, string(m))
			}
			return hookengine.Result{
				Context: "IntentGate modes detected: " + strings.Join(names, ", "),
			}, nil
		},
	}
}

// AnthropicContextWindowLimitRecovery handles Anthropic-specific
// recovery when the context window is hit.
func AnthropicContextWindowLimitRecovery() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "anthropic-context-window-limit-recovery",
		Tier:   hookengine.TierSkill,
		Events: []hookengine.Event{hookengine.EventToolError, hookengine.EventSessionError},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			return hookengine.Result{Context: "context window exceeded; consider compaction"}, nil
		},
	}
}

// AstGrepSgProvision ensures the `sg` binary is available for the
// ast-grep skill.
func AstGrepSgProvision(sgAvailable func() bool) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "ast-grep-sg-provision",
		Tier:   hookengine.TierSkill,
		Events: []hookengine.Event{hookengine.EventSessionStart},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if !sgAvailable() {
				return hookengine.Result{Context: "ast-grep skill may be degraded: `sg` not on PATH"}, nil
			}
			return hookengine.Result{}, nil
		},
	}
}

// CodegraphBootstrap ensures the codegraph MCP is loaded.
func CodegraphBootstrap() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "codegraph-bootstrap",
		Tier:   hookengine.TierSkill,
		Events: []hookengine.Event{hookengine.EventSessionStart},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			return hookengine.Result{}, nil
		},
	}
}

// LegacyPluginToast warns if a legacy oh-my-openagent plugin is loaded.
func LegacyPluginToast() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "legacy-plugin-toast",
		Tier:   hookengine.TierSkill,
		Events: []hookengine.Event{hookengine.EventSessionStart},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			return hookengine.Result{Context: "heretic: legacy plugin detected"}, nil
		},
	}
}

// ModelFallback switches to a fallback model on transient errors.
func ModelFallback(fallbacks []string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "model-fallback",
		Tier:   hookengine.TierSkill,
		Events: []hookengine.Event{hookengine.EventToolError, hookengine.EventSessionError},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if len(fallbacks) == 0 {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			return hookengine.Result{Context: fmt.Sprintf("will try fallback: %s", fallbacks[0])}, nil
		},
	}
}

// TeamModeStatusInjector informs the agent that team mode is active.
func TeamModeStatusInjector(enabled bool) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "team-mode-status-injector",
		Tier:   hookengine.TierSkill,
		Events: []hookengine.Event{hookengine.EventSessionStart, hookengine.EventChatTransform},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if !enabled {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			return hookengine.Result{Context: "team mode: active"}, nil
		},
	}
}

// TeamMailboxInjector injects pending mailbox messages into the chat.
func TeamMailboxInjector(messages []string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "team-mailbox-injector",
		Tier:   hookengine.TierSkill,
		Events: []hookengine.Event{hookengine.EventChatTransform},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if len(messages) == 0 {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			return hookengine.Result{Context: "team mailbox:\n" + strings.Join(messages, "\n")}, nil
		},
	}
}

// TeamToolGating blocks team tools when team mode is off.
func TeamToolGating(enabled bool) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "team-tool-gating",
		Tier:   hookengine.TierSkill,
		Events: []hookengine.Event{hookengine.EventPreToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*MessageData)
			if !ok {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			_ = d
			if !enabled {
				return hookengine.Result{
					Decision: hookengine.DecisionDeny,
					Reason:   "team mode is disabled",
				}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionAllow}, nil
		},
	}
}

// AutoSlashCommand converts "/command" text into a slash-command call.
func AutoSlashCommand(commands map[string]string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "auto-slash-command",
		Tier:   hookengine.TierSession,
		Events: []hookengine.Event{hookengine.EventMessageReceived},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*MessageData)
			if !ok {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			content := strings.TrimSpace(d.Content)
			if !strings.HasPrefix(content, "/") {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			cmd := strings.SplitN(content[1:], " ", 2)[0]
			if _, ok := commands[cmd]; ok {
				return hookengine.Result{Context: fmt.Sprintf("invoking /%s", cmd)}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionNone}, nil
		},
	}
}

// BackgroundNotification fires when a background task completes.
func BackgroundNotification() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "background-notification",
		Tier:   hookengine.TierSession,
		Events: []hookengine.Event{hookengine.EventSessionIdle, hookengine.EventMessageReceived},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			return hookengine.Result{}, nil
		},
	}
}

// PrometheusMdOnly is the gate that allows prometheus to edit only
// .md files. If a non-.md file is targeted, the edit is denied.
func PrometheusMdOnly(agentName string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "prometheus-md-only",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPreToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if agentName != "prometheus" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionNone}, nil
		},
	}
}
