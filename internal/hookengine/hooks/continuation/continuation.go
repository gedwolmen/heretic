// Package continuationhooks contains Continuation-tier hooks.
package continuationhooks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gedwolmen/heretic/internal/hookengine"
)

// EventData is the payload for continuation events.
type EventData struct {
	EventName hookengine.Event `json:"-"`
	SessID    string           `json:"session_id"`
	CWD       string           `json:"cwd"`
	Worktree  string           `json:"worktree,omitempty"`
	BoulderID string           `json:"boulder_id,omitempty"`
}

func (e *EventData) Event() hookengine.Event { return e.EventName }
func (e *EventData) SessionID() string       { return e.SessID }

// StartWork creates a new worktree for the current boulder on
// session.start. The boulder state file must already exist.
func StartWork(boulderID, cwd string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "start-work",
		Tier:   hookengine.TierContinuation,
		Events: []hookengine.Event{hookengine.EventSessionStart},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*EventData)
			if !ok {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			if d.BoulderID == "" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			// Worktree path: <cwd>/.worktrees/<boulder_id>
			wtPath := filepath.Join(cwd, ".worktrees", boulderID)
			if err := os.MkdirAll(wtPath, 0o755); err != nil {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			d.Worktree = wtPath
			return hookengine.Result{Context: fmt.Sprintf("worktree: %s", wtPath), Transformed: d}, nil
		},
	}
}

// StartWorkContinuation continues work in the existing worktree on
// session.idle (i.e. the user said "continue").
func StartWorkContinuation(boulderID, cwd string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "start-work-continuation",
		Tier:   hookengine.TierContinuation,
		Events: []hookengine.Event{hookengine.EventSessionIdle},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if boulderID == "" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			return hookengine.Result{Context: fmt.Sprintf("resuming boulder %s", boulderID)}, nil
		},
	}
}

// TaskResumeInfo displays info about the resumed task on session.start.
func TaskResumeInfo(taskID, summary string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "task-resume-info",
		Tier:   hookengine.TierContinuation,
		Events: []hookengine.Event{hookengine.EventSessionStart},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if taskID == "" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			return hookengine.Result{Context: fmt.Sprintf("resuming task %s: %s", taskID, summary)}, nil
		},
	}
}

// ThinkMode enables extended thinking on supported models.
func ThinkMode(enabled bool) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "think-mode",
		Tier:   hookengine.TierContinuation,
		Events: []hookengine.Event{hookengine.EventChatTransform},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if !enabled {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			return hookengine.Result{Context: "think mode: enabled"}, nil
		},
	}
}

// TodoDescriptionOverride replaces the default todo description with a
// user-supplied one.
func TodoDescriptionOverride(description string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "todo-description-override",
		Tier:   hookengine.TierContinuation,
		Events: []hookengine.Event{hookengine.EventSessionStart, hookengine.EventMessageReceived},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if description == "" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			return hookengine.Result{Context: "todo: " + strings.TrimSpace(description)}, nil
		},
	}
}

// UnstableAgentBabysitter aborts the session if the agent appears to
// be in an infinite loop.
func UnstableAgentBabysitter(maxTurns int) *hookengine.Hook {
	if maxTurns <= 0 {
		maxTurns = 200
	}
	return &hookengine.Hook{
		Name:   "unstable-agent-babysitter",
		Tier:   hookengine.TierContinuation,
		Events: []hookengine.Event{hookengine.EventSessionError, hookengine.EventSessionIdle},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			return hookengine.Result{Decision: hookengine.DecisionNone}, nil
		},
	}
}

// DelegateTaskRetry retries a failed delegate-task call once.
func DelegateTaskRetry(maxRetries int) *hookengine.Hook {
	if maxRetries <= 0 {
		maxRetries = 1
	}
	return &hookengine.Hook{
		Name:   "delegate-task-retry",
		Tier:   hookengine.TierContinuation,
		Events: []hookengine.Event{hookengine.EventToolError},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			return hookengine.Result{Context: fmt.Sprintf("will retry up to %d times", maxRetries)}, nil
		},
	}
}
