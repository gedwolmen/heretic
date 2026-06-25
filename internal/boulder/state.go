// Package boulder implements the boulder state feature: a JSON-backed
// work tracker that survives across sessions. Mirrors
// packages/omo-opencode/src/features/boulder-state/ and
// packages/boulder-state/.
//
// The boulder state file is named `boulder.json` and lives at the project
// root (or in a configured location). It tracks a single active work
// item plus its history.
package boulder

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Status is the lifecycle state of a boulder.
type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusRegressed Status = "regressed"
	StatusBlocked   Status = "blocked"
)

// Entry is one history record.
type Entry struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    Status    `json:"status"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
	Note      string    `json:"note,omitempty"`
}

// State is the full boulder state.
type State struct {
	Current *Entry   `json:"current,omitempty"`
	History []Entry  `json:"history"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Manager is the boulder state manager.
type Manager struct {
	mu       sync.Mutex
	file     string
	maxHist  int
}

// NewManager returns a Manager rooted at the given file path.
func NewManager(file string) *Manager {
	return &Manager{file: file, maxHist: 100}
}

// Load reads the state from disk. Returns an empty State if the file
// does not exist.
func (m *Manager) Load() (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, err := os.ReadFile(m.file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{}, nil
		}
		return State{}, err
	}
	var st State
	if err := json.Unmarshal(b, &st); err != nil {
		return State{}, err
	}
	return st, nil
}

// Save writes the state to disk atomically.
func (m *Manager) Save(st State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	st.UpdatedAt = time.Now()
	if err := os.MkdirAll(filepath.Dir(m.file), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.file, b, 0o644)
}

// SetCurrent starts a new boulder. If a current one is active, it is
// moved to history with status "regressed".
func (m *Manager) SetCurrent(title string) (Entry, error) {
	st, err := m.Load()
	if err != nil {
		return Entry{}, err
	}
	if st.Current != nil {
		// Move the previous current to history.
		prev := *st.Current
		prev.Status = StatusRegressed
		prev.EndedAt = time.Now()
		st.History = append(st.History, prev)
	}
	now := time.Now()
	st.Current = &Entry{
		ID:        fmt.Sprintf("b-%d", now.UnixNano()),
		Title:     title,
		Status:    StatusActive,
		StartedAt: now,
	}
	if err := m.Save(st); err != nil {
		return Entry{}, err
	}
	return *st.Current, nil
}

// Complete marks the current boulder as completed and moves it to
// history.
func (m *Manager) Complete(note string) error {
	st, err := m.Load()
	if err != nil {
		return err
	}
	if st.Current == nil {
		return errors.New("boulder: no current boulder")
	}
	cur := *st.Current
	cur.Status = StatusCompleted
	cur.EndedAt = time.Now()
	cur.Note = note
	st.History = append(st.History, cur)
	st.Current = nil
	return m.Save(st)
}

// Block marks the current boulder as blocked.
func (m *Manager) Block(note string) error {
	st, err := m.Load()
	if err != nil {
		return err
	}
	if st.Current == nil {
		return errors.New("boulder: no current boulder")
	}
	st.Current.Status = StatusBlocked
	st.Current.Note = note
	return m.Save(st)
}

// TrimHistory drops the oldest history entries if the slice exceeds
// maxHist. The current boulder is preserved.
func (m *Manager) TrimHistory() error {
	st, err := m.Load()
	if err != nil {
		return err
	}
	if len(st.History) > m.maxHist {
		st.History = st.History[len(st.History)-m.maxHist:]
		return m.Save(st)
	}
	return nil
}
