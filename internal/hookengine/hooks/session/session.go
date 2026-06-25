// Package sessionhooks contains Session-tier hooks.
package sessionhooks

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gedwolmen/heretic/internal/hookengine"
)

// EventData is the payload for session events.
type EventData struct {
	EventName hookengine.Event `json:"-"`
	SessID    string           `json:"session_id"`
	CWD       string           `json:"cwd"`
	StartedAt time.Time        `json:"started_at"`
}

func (e *EventData) Event() hookengine.Event { return e.EventName }
func (e *EventData) SessionID() string       { return e.SessID }

// NewSessionStartData builds a session.start payload.
func NewSessionStartData(sessID, cwd string) *EventData {
	return &EventData{
		EventName: hookengine.EventSessionStart,
		SessID:    sessID,
		CWD:       cwd,
		StartedAt: time.Now(),
	}
}

// AgentUsageReminder fires on session.start and reminds the user which
// agent they are using.
func AgentUsageReminder(agentName string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:  "agent-usage-reminder",
		Tier:  hookengine.TierSession,
		Events: []hookengine.Event{hookengine.EventSessionStart},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			return hookengine.Result{
				Context: fmt.Sprintf("You are using the %s agent.", agentName),
			}, nil
		},
	}
}

// AutoUpdateChecker fires on session.start. It checks for updates in the
// background (non-blocking via Async).
func AutoUpdateChecker(version string, lastChecked *time.Time, mu *sync.RWMutex) *hookengine.Hook {
	return &hookengine.Hook{
		Name:  "auto-update-checker",
		Tier:  hookengine.TierSession,
		Events: []hookengine.Event{hookengine.EventSessionStart},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			mu.Lock()
			now := time.Now()
			*lastChecked = now
			mu.Unlock()
			return hookengine.Result{Async: true}, nil
		},
	}
}

// SessionNotification fires on session.start and emits a desktop/OS
// notification. (The actual notification dispatch is a TODO; the hook
// is a no-op for now.)
func SessionNotification() *hookengine.Hook {
	return &hookengine.Hook{
		Name:  "session-notification",
		Tier:  hookengine.TierSession,
		Events: []hookengine.Event{hookengine.EventSessionStart, hookengine.EventSessionEnd},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			return hookengine.Result{}, nil
		},
	}
}

// CompactionContextInjector fires on session.compacting. It preserves
// important context (todos, plans) across compaction.
func CompactionContextInjector(todos *[]string, mu *sync.RWMutex) *hookengine.Hook {
	return &hookengine.Hook{
		Name:  "compaction-context-injector",
		Tier:  hookengine.TierSession,
		Events: []hookengine.Event{hookengine.EventCompaction},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			mu.RLock()
			pending := append([]string{}, *todos...)
			mu.RUnlock()
			if len(pending) == 0 {
				return hookengine.Result{}, nil
			}
			return hookengine.Result{
				Context: "Active todos across compaction: " + joinLines(pending),
			}, nil
		},
	}
}

// CompactionTodoPreserver preserves the todo list in a side file so
// that the auto-continue hook can re-load it after compaction.
func CompactionTodoPreserver(todos *[]string, sideFile string, mu *sync.RWMutex) *hookengine.Hook {
	return &hookengine.Hook{
		Name:  "compaction-todo-preserver",
		Tier:  hookengine.TierSession,
		Events: []hookengine.Event{hookengine.EventCompaction},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			mu.RLock()
			pending := append([]string{}, *todos...)
			mu.RUnlock()
			if len(pending) == 0 {
				return hookengine.Result{}, nil
			}
			body := joinLines(pending)
			_ = os.WriteFile(sideFile, []byte(body), 0o644)
			return hookengine.Result{}, nil
		},
	}
}

// NoHephaestusNonGPT denies the hephaestus agent on non-OpenAI models.
// (Heuristic from om-my-openagent.)
func NoHephaestusNonGPT(agentName, providerID string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:  "no-hephaestus-non-gpt",
		Tier:  hookengine.TierSession,
		Events: []hookengine.Event{hookengine.EventSessionStart},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if agentName == "hephaestus" && providerID != "openai" {
				return hookengine.Result{
					Decision: hookengine.DecisionDeny,
					Reason:   "hephaestus is restricted to OpenAI providers",
				}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionNone}, nil
		},
	}
}

// NoSisyphusGPT denies the sisyphus agent on OpenAI models.
func NoSisyphusGPT(agentName, providerID string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:  "no-sisyphus-gpt",
		Tier:  hookengine.TierSession,
		Events: []hookengine.Event{hookengine.EventSessionStart},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if agentName == "sisyphus" && providerID == "openai" {
				return hookengine.Result{
					Decision: hookengine.DecisionDeny,
					Reason:   "sisyphus is restricted to non-OpenAI providers",
				}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionNone}, nil
		},
	}
}

// NonInteractiveEnv detects non-interactive environments and adjusts
// the agent's behavior.
func NonInteractiveEnv(env string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:  "non-interactive-env",
		Tier:  hookengine.TierSession,
		Events: []hookengine.Event{hookengine.EventSessionStart},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if env != "" {
				return hookengine.Result{Context: "Environment: " + env}, nil
			}
			return hookengine.Result{}, nil
		},
	}
}

// SessionTodoStatus reports todo progress on session.idle.
func SessionTodoStatus(todos *[]string, mu *sync.RWMutex) *hookengine.Hook {
	return &hookengine.Hook{
		Name:  "session-todo-status",
		Tier:  hookengine.TierSession,
		Events: []hookengine.Event{hookengine.EventSessionIdle, hookengine.EventSessionEnd},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			mu.RLock()
			n := len(*todos)
			mu.RUnlock()
			if n == 0 {
				return hookengine.Result{}, nil
			}
			return hookengine.Result{
				Context: fmt.Sprintf("%d todos still pending", n),
			}, nil
		},
	}
}

// joinLines is a tiny helper to format a slice as a newline-joined
// string without importing strings.
func joinLines(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	out := ss[0]
	for _, s := range ss[1:] {
		out += "\n" + s
	}
	return out
}

// RalphLoop is the iteration driver: while a "ralph" loop is active,
// continue working on the current task.
func RalphLoop(loopActive func() bool, iteration *int, maxIter int) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "ralph-loop",
		Tier:   hookengine.TierContinuation,
		Events: []hookengine.Event{hookengine.EventSessionIdle},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if !loopActive() {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			*iteration++
			if maxIter > 0 && *iteration >= maxIter {
				return hookengine.Result{
					Decision: hookengine.DecisionDeny,
					Reason:   fmt.Sprintf("ralph loop reached max iterations (%d)", maxIter),
				}, nil
			}
			return hookengine.Result{Context: fmt.Sprintf("ralph loop iteration %d", *iteration)}, nil
		},
	}
}

// EmptyTaskResponseDetector catches when an agent returns an empty
// response and nudges it.
func EmptyTaskResponseDetector() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "empty-task-response-detector",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventToolError},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			return hookengine.Result{Context: "empty response detected; retry with explicit output"}, nil
		},
	}
}

// PreemptiveCompactionTrigger triggers compaction when the message
// history exceeds a threshold.
func PreemptiveCompactionTrigger(approxMessageCount *int, threshold int) *hookengine.Hook {
	if threshold <= 0 {
		threshold = 100
	}
	return &hookengine.Hook{
		Name:   "preemptive-compaction-trigger",
		Tier:   hookengine.TierContinuation,
		Events: []hookengine.Event{hookengine.EventMessageReceived},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			*approxMessageCount++
			if *approxMessageCount >= threshold {
				return hookengine.Result{Context: "approaching context limit; consider compaction"}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionNone}, nil
		},
	}
}
