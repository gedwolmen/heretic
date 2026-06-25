package background

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestManager_SpawnAndComplete(t *testing.T) {
	t.Parallel()
	m := NewManager(Config{MaxPerKey: 2})
	task := &Task{ID: "t1", ProviderID: "anthropic", ModelID: "claude-sonnet-4"}
	err := m.Spawn(context.Background(), task, func(ctx context.Context) (string, error) {
		return "hello", nil
	}, false)
	require.NoError(t, err)

	// Wait for completion via WakeCh.
	select {
	case id := <-m.WakeCh():
		require.Equal(t, "t1", id)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for wake")
	}

	got, ok := m.Get("t1")
	require.True(t, ok)
	require.Equal(t, StatusCompleted, got.Status)
	require.Equal(t, "hello", got.Output)
}

func TestManager_SpawnAndFail(t *testing.T) {
	t.Parallel()
	m := NewManager(Config{MaxPerKey: 1})
	task := &Task{ID: "t1", ProviderID: "anthropic", ModelID: "claude-sonnet-4"}
	err := m.Spawn(context.Background(), task, func(ctx context.Context) (string, error) {
		return "", context.DeadlineExceeded
	}, false)
	require.NoError(t, err)

	select {
	case <-m.WakeCh():
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}

	got, _ := m.Get("t1")
	require.Equal(t, StatusFailed, got.Status)
	require.Contains(t, got.Error, "deadline")
}

func TestManager_ConcurrencyLimit(t *testing.T) {
	t.Parallel()
	m := NewManager(Config{MaxPerKey: 1})
	// Saturate the slot.
	hold := make(chan struct{})
	started := make(chan struct{}, 1)
	first := &Task{ID: "first", ProviderID: "p", ModelID: "m"}
	err := m.Spawn(context.Background(), first, func(ctx context.Context) (string, error) {
		started <- struct{}{}
		<-hold
		return "ok", nil
	}, false)
	require.NoError(t, err)
	<-started

	// Second task must be rejected (no queue).
	second := &Task{ID: "second", ProviderID: "p", ModelID: "m"}
	err = m.Spawn(context.Background(), second, func(ctx context.Context) (string, error) {
		return "ok", nil
	}, false)
	require.Error(t, err)

	// Release the first; it should complete.
	close(hold)
	<-m.WakeCh()
}

func TestManager_Cancel(t *testing.T) {
	t.Parallel()
	m := NewManager(Config{MaxPerKey: 1})
	task := &Task{ID: "t1", ProviderID: "p", ModelID: "m"}
	err := m.Spawn(context.Background(), task, func(ctx context.Context) (string, error) {
		time.Sleep(500 * time.Millisecond)
		return "ok", nil
	}, false)
	require.NoError(t, err)
	require.NoError(t, m.Cancel("t1"))
	got, _ := m.Get("t1")
	require.Equal(t, StatusCancelled, got.Status)
}

func TestManager_Output(t *testing.T) {
	t.Parallel()
	m := NewManager(Config{MaxPerKey: 1})
	task := &Task{ID: "t1", ProviderID: "p", ModelID: "m"}
	err := m.Spawn(context.Background(), task, func(ctx context.Context) (string, error) {
		return "result", nil
	}, false)
	require.NoError(t, err)
	<-m.WakeCh()
	out, err := m.Output("t1")
	require.NoError(t, err)
	require.Equal(t, "result", out)
}

func TestManager_Output_NotFound(t *testing.T) {
	t.Parallel()
	m := NewManager(Config{})
	_, err := m.Output("nope")
	require.Error(t, err)
}
