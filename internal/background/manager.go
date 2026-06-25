// Package background implements background agent spawning and parent-wake
// notification. Mirrors packages/omo-opencode/src/features/background-agent/.
//
// A background agent is an LLM call that runs in parallel with the main
// session. Concurrency is bounded per provider/model.
package background

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Status is the lifecycle state of a background task.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// Task is a single background agent invocation.
type Task struct {
	ID         string
	ProviderID string
	ModelID    string
	Prompt     string
	StartedAt  time.Time
	EndedAt    time.Time
	Status     Status
	Output     string
	Error      string
}

// Manager runs background tasks with per-provider/model concurrency.
type Manager struct {
	mu       sync.Mutex
	tasks    map[string]*Task
	limits   map[string]chan struct{} // keyed by provider/model
	maxPer   int
	maxQueue int
	// wakeCh is the channel the parent listens on for completion
	// notifications. The Manager SENDS a task ID on wakeCh each time
	// a task transitions to completed/failed/cancelled.
	wakeCh chan string
}

// Config configures a Manager.
type Config struct {
	MaxPerKey   int // default 5
	MaxQueue    int // 0 = unlimited
	WakeBufSize int // default 64
}

// NewManager returns a Manager with the given config.
func NewManager(cfg Config) *Manager {
	if cfg.MaxPerKey <= 0 {
		cfg.MaxPerKey = 5
	}
	if cfg.WakeBufSize <= 0 {
		cfg.WakeBufSize = 64
	}
	return &Manager{
		tasks:    make(map[string]*Task),
		limits:   make(map[string]chan struct{}),
		maxPer:   cfg.MaxPerKey,
		maxQueue: cfg.MaxQueue,
		wakeCh:   make(chan string, cfg.WakeBufSize),
	}
}

// WakeCh returns the channel the parent listens on. Buffered; callers
// should drain it.
func (m *Manager) WakeCh() <-chan string { return m.wakeCh }

// slot returns the semaphore for a (provider, model) pair.
func (m *Manager) slot(providerID, modelID string) chan struct{} {
	key := providerID + "/" + modelID
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.limits[key]
	if !ok {
		s = make(chan struct{}, m.maxPer)
		m.limits[key] = s
	}
	return s
}

// Spawn starts a new background task. The function is run in a
// goroutine; the task is registered immediately and updated on
// completion.
//
// queueOnFull controls what happens when the concurrency limit is hit.
// If true, the call blocks until a slot is free (FIFO). If false, an
// error is returned.
func (m *Manager) Spawn(ctx context.Context, t *Task, fn func(ctx context.Context) (string, error), queueOnFull bool) error {
	if t.ID == "" {
		return errors.New("background: task ID is required")
	}
	if t.ProviderID == "" || t.ModelID == "" {
		return errors.New("background: provider and model are required")
	}
	// Take an internal snapshot. The caller's Task pointer is NOT
	// updated; this avoids races between the caller and the
	// goroutine writing to the same Task.
	internal := *t
	s := m.slot(internal.ProviderID, internal.ModelID)
	if !queueOnFull {
		select {
		case s <- struct{}{}:
			// got a slot
		default:
			return fmt.Errorf("background: concurrency limit reached for %s/%s", internal.ProviderID, internal.ModelID)
		}
	} else {
		select {
		case s <- struct{}{}:
			// got a slot
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	m.mu.Lock()
	internal.Status = StatusRunning
	internal.StartedAt = time.Now()
	m.tasks[internal.ID] = &internal
	m.mu.Unlock()

	go func() {
		defer func() { <-s }()
		ctx2, cancel := context.WithCancel(context.Background())
		go func() {
			<-ctx.Done()
			cancel()
		}()
		out, err := fn(ctx2)
		m.mu.Lock()
		internal.EndedAt = time.Now()
		if err != nil {
			internal.Status = StatusFailed
			internal.Error = err.Error()
		} else {
			internal.Status = StatusCompleted
			internal.Output = out
		}
		m.mu.Unlock()
		// Non-blocking send to wakeCh.
		select {
		case m.wakeCh <- internal.ID:
		default:
		}
	}()
	return nil
}

// Cancel marks a task as cancelled. Cancellation of an in-flight fn
// requires the caller to wire up context cancellation; this is a
// status update only.
func (m *Manager) Cancel(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return fmt.Errorf("background: task %q not found", id)
	}
	if t.Status != StatusRunning && t.Status != StatusPending {
		return fmt.Errorf("background: task %q is not running (status=%s)", id, t.Status)
	}
	t.Status = StatusCancelled
	t.EndedAt = time.Now()
	return nil
}

// Get returns a copy of the task by ID. The copy is a snapshot; the
// caller may not assume the returned value updates as the task runs.
func (m *Manager) Get(id string) (Task, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return Task{}, false
	}
	return *t, true
}

// List returns copies of all tasks.
func (m *Manager) List() []Task {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		out = append(out, *t)
	}
	return out
}

// Output returns the output of a completed task, or "" if the task
// is not completed.
func (m *Manager) Output(id string) (string, error) {
	t, ok := m.Get(id)
	if !ok {
		return "", fmt.Errorf("background: task %q not found", id)
	}
	if t.Status == StatusRunning || t.Status == StatusPending {
		return "", fmt.Errorf("background: task %q is not complete (status=%s)", id, t.Status)
	}
	if t.Status == StatusFailed {
		return "", errors.New(t.Error)
	}
	return t.Output, nil
}
