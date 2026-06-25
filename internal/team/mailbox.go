package team

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Message is one entry in a member's mailbox.
type Message struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Subject   string    `json:"subject,omitempty"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	Read      bool      `json:"read"`
}

// Mailbox manages a team's mailbox directory.
type Mailbox struct {
	teamDir string
	mu      sync.Mutex
}

// NewMailbox returns a Mailbox for the given team directory.
func NewMailbox(teamDir string) *Mailbox {
	return &Mailbox{teamDir: teamDir}
}

// memberDir returns the per-member mailbox directory.
func (m *Mailbox) memberDir(member string) string {
	return filepath.Join(m.teamDir, "mailbox", member)
}

// Send appends a message to the recipient's mailbox.
func (m *Mailbox) Send(msg Message) error {
	if msg.To == "" {
		return errors.New("mailbox: To is required")
	}
	if msg.From == "" {
		return errors.New("mailbox: From is required")
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("%d-%s", msg.CreatedAt.UnixNano(), msg.From)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	dir := m.memberDir(msg.To)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, msg.ID+".json")
	b, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// List returns the messages in a member's mailbox, sorted by CreatedAt.
func (m *Mailbox) List(member string) ([]Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	dir := m.memberDir(member)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []Message
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		var msg Message
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		if err := json.Unmarshal(b, &msg); err != nil {
			continue
		}
		out = append(out, msg)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

// Unread returns only the unread messages.
func (m *Mailbox) Unread(member string) ([]Message, error) {
	all, err := m.List(member)
	if err != nil {
		return nil, err
	}
	var out []Message
	for _, msg := range all {
		if !msg.Read {
			out = append(out, msg)
		}
	}
	return out, nil
}

// MarkRead marks a message as read (by ID).
func (m *Mailbox) MarkRead(member, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path := filepath.Join(m.memberDir(member), id+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var msg Message
	if err := json.Unmarshal(b, &msg); err != nil {
		return err
	}
	msg.Read = true
	out, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

// Delete removes a message.
func (m *Mailbox) Delete(member, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return os.Remove(filepath.Join(m.memberDir(member), id+".json"))
}
