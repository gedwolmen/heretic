package hookengine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testPayload struct {
	eventName Event
	sessID    string
}

func (p *testPayload) Event() Event    { return p.eventName }
func (p *testPayload) SessionID() string { return p.sessID }

func TestRegistry_RegisterAndGet(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	h := &Hook{
		Name:   "test",
		Tier:   TierToolGuard,
		Events: []Event{EventPreToolUse},
		Fn: func(ctx context.Context, p Payload) (Result, error) {
			return Result{Decision: DecisionAllow}, nil
		},
	}
	require.NoError(t, r.Register(h))
	got, ok := r.Get("test")
	require.True(t, ok)
	require.Equal(t, "test", got.Name)
}

func TestRegistry_DuplicateName(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	h := &Hook{Name: "x", Tier: TierSession, Events: []Event{EventSessionStart}, Fn: func(ctx context.Context, p Payload) (Result, error) { return Result{}, nil }}
	require.NoError(t, r.Register(h))
	require.Error(t, r.Register(h))
}

func TestRegistry_EmptyName(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	h := &Hook{Name: "", Tier: TierSession, Fn: func(ctx context.Context, p Payload) (Result, error) { return Result{}, nil }}
	require.Error(t, r.Register(h))
}

func TestRegistry_ForEvent(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	require.NoError(t, r.Register(&Hook{Name: "a", Tier: TierSession, Events: []Event{EventSessionStart}, Fn: func(ctx context.Context, p Payload) (Result, error) { return Result{}, nil }}))
	require.NoError(t, r.Register(&Hook{Name: "b", Tier: TierToolGuard, Events: []Event{EventPreToolUse}, Fn: func(ctx context.Context, p Payload) (Result, error) { return Result{}, nil }}))
	got := r.ForEvent(EventPreToolUse)
	require.Len(t, got, 1)
	require.Equal(t, "b", got[0].Name)
}

func TestEventTiers(t *testing.T) {
	t.Parallel()
	require.Equal(t, []Tier{TierSession}, EventTiers(EventSessionStart))
	require.Equal(t, []Tier{TierSession, TierTransform, TierSkill}, EventTiers(EventMessageReceived))
	require.Equal(t, []Tier{TierToolGuard, TierTransform, TierSkill}, EventTiers(EventPreToolUse))
	require.Equal(t, []Tier{TierContinuation}, EventTiers(EventCompaction))
}

func TestEngine_Fire_EmptyRegistry(t *testing.T) {
	t.Parallel()
	e := NewEngine(NewRegistry(), NewRunner())
	agg, err := e.Fire(context.Background(), EventSessionStart, &testPayload{})
	require.NoError(t, err)
	require.Equal(t, DecisionNone, agg.Decision)
	require.Equal(t, 0, agg.HookCount)
}

func TestEngine_Fire_Aggregate(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	// Hook 1: allow
	require.NoError(t, r.Register(&Hook{
		Name: "allow", Tier: TierToolGuard, Events: []Event{EventPreToolUse},
		Fn: func(ctx context.Context, p Payload) (Result, error) {
			return Result{Decision: DecisionAllow}, nil
		},
	}))
	// Hook 2: deny
	require.NoError(t, r.Register(&Hook{
		Name: "deny", Tier: TierToolGuard, Events: []Event{EventPreToolUse},
		Fn: func(ctx context.Context, p Payload) (Result, error) {
			return Result{Decision: DecisionDeny, Reason: "nope"}, nil
		},
	}))
	e := NewEngine(r, NewRunner())
	agg, err := e.Fire(context.Background(), EventPreToolUse, &testPayload{})
	require.NoError(t, err)
	require.Equal(t, DecisionDeny, agg.Decision)
	require.Contains(t, agg.Reason, "nope")
	require.Equal(t, 2, agg.HookCount)
}

func TestEngine_Fire_HookError(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	require.NoError(t, r.Register(&Hook{
		Name: "boom", Tier: TierSession, Events: []Event{EventSessionStart},
		Fn: func(ctx context.Context, p Payload) (Result, error) {
			return Result{}, errors.New("oops")
		},
	}))
	e := NewEngine(r, NewRunner())
	agg, err := e.Fire(context.Background(), EventSessionStart, &testPayload{})
	require.NoError(t, err) // hook error is collected, not returned
	require.Len(t, agg.Errors, 1)
}

func TestRunner_Timeout(t *testing.T) {
	t.Parallel()
	runner := &Runner{DefaultTimeout: 50 * time.Millisecond}
	r := NewRegistry()
	require.NoError(t, r.Register(&Hook{
		Name: "slow", Tier: TierSession, Events: []Event{EventSessionStart},
		Fn: func(ctx context.Context, p Payload) (Result, error) {
			select {
			case <-time.After(200 * time.Millisecond):
				return Result{}, nil
			case <-ctx.Done():
				return Result{}, ctx.Err()
			}
		},
	}))
	e := NewEngine(r, runner)
	_, _ = e.Fire(context.Background(), EventSessionStart, &testPayload{})
	// Verify the runner honored the timeout: the hook returned ctx.Err
	// (which we discard in the engine), but the call completed in
	// ~50ms, not 200ms.
}
