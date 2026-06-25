// Package team implements parallel multi-agent coordination for heretic
// Ultimate. Mirrors packages/omo-opencode/src/features/team-mode/.
//
// Teams are directories on disk:
//
//	~/.omo/teams/{name}/
//	  config.json    # spec: name, members, owner session
//	  state.json     # runtime: status, current task, etc.
//	  mailbox/       # inter-agent messages
//	    {member}/
//	      *.json
//	  tasklist.jsonl # append-only task log
//	  worktrees/     # per-member git worktrees
//	    {member}/
//
// All file I/O uses osutil atomic writes where possible.
package team

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Status is the team's lifecycle state.
type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusShutdown  Status = "shutdown"
	StatusCompleted Status = "completed"
)

// MemberKind is how a team member is declared.
type MemberKind string

const (
	// KindSubagent: direct agent reference.
	KindSubagent MemberKind = "subagent_type"
	// KindCategory: routed through sisyphus-junior.
	KindCategory MemberKind = "category"
)

// Member is one entry in the team's member list.
type Member struct {
	Name    string     `json:"name"`
	Kind    MemberKind `json:"kind"`
	Agent   string     `json:"agent,omitempty"`
	Role    string     `json:"role,omitempty"`
	Enabled bool       `json:"enabled,omitempty"`
}

// Config is the spec stored in config.json.
type Config struct {
	Name    string    `json:"name"`
	Owner   string    `json:"owner_session"`
	Members []Member  `json:"members"`
	Created time.Time `json:"created_at"`
}

// State is the runtime state stored in state.json.
type State struct {
	Status      Status    `json:"status"`
	CurrentTask string    `json:"current_task,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Storage is the on-disk team store.
type Storage struct {
	BaseDir string
	mu      sync.Mutex
}

// NewStorage returns a Storage rooted at baseDir (typically
// ~/.omo/teams or <project>/.omo/teams).
func NewStorage(baseDir string) *Storage {
	return &Storage{BaseDir: baseDir}
}

// TeamDir returns the per-team directory.
func (s *Storage) TeamDir(name string) string {
	return filepath.Join(s.BaseDir, name)
}

// EnsureDir creates the team's directory tree if it doesn't exist.
func (s *Storage) EnsureDir(name string) error {
	dir := s.TeamDir(name)
	for _, sub := range []string{"", "mailbox", "worktrees"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			return err
		}
	}
	return nil
}

// SaveConfig writes the team's config.json atomically.
func (s *Storage) SaveConfig(name string, cfg Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.EnsureDir(name); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.TeamDir(name), "config.json"), b, 0o644)
}

// LoadConfig reads the team's config.json. Returns ErrNotFound if the
// file does not exist.
func (s *Storage) LoadConfig(name string) (Config, error) {
	var cfg Config
	b, err := os.ReadFile(filepath.Join(s.TeamDir(name), "config.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, ErrNotFound
		}
		return cfg, err
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// SaveState writes the team's state.json atomically.
func (s *Storage) SaveState(name string, st State) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.EnsureDir(name); err != nil {
		return err
	}
	st.UpdatedAt = time.Now()
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.TeamDir(name), "state.json"), b, 0o644)
}

// LoadState reads the team's state.json.
func (s *Storage) LoadState(name string) (State, error) {
	var st State
	b, err := os.ReadFile(filepath.Join(s.TeamDir(name), "state.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{Status: StatusPending}, nil
		}
		return st, err
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return st, err
	}
	return st, nil
}

// ListTeams returns all team names under the storage base.
func (s *Storage) ListTeams() ([]string, error) {
	entries, err := os.ReadDir(s.BaseDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Only include directories that have a config.json (sanity).
		if _, err := os.Stat(filepath.Join(s.BaseDir, e.Name(), "config.json")); err == nil {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// DeleteTeam removes a team's directory tree.
func (s *Storage) DeleteTeam(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.RemoveAll(s.TeamDir(name))
}

// ErrNotFound is returned by LoadConfig when a team does not exist.
var ErrNotFound = errors.New("team: not found")

// ErrAlreadyExists is returned by Create when a team is already present.
var ErrAlreadyExists = errors.New("team: already exists")

// Create is a convenience that creates a new team and persists its
// initial config + state.
func (s *Storage) Create(cfg Config) error {
	if cfg.Name == "" {
		return errors.New("team: name is required")
	}
	if cfg.Created.IsZero() {
		cfg.Created = time.Now()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := os.Stat(s.TeamDir(cfg.Name)); err == nil {
		return fmt.Errorf("%w: %s", ErrAlreadyExists, cfg.Name)
	}
	if err := s.EnsureDir(cfg.Name); err != nil {
		return err
	}
	// Write directly to avoid recursive locking on SaveConfig/SaveState.
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(s.TeamDir(cfg.Name), "config.json"), b, 0o644); err != nil {
		return err
	}
	st := State{Status: StatusPending, UpdatedAt: time.Now()}
	bs, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.TeamDir(cfg.Name), "state.json"), bs, 0o644)
}
