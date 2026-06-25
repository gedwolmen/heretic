package mcp

import (
	"sync"
)

// SkillMCPManager manages per-session skill-embedded MCPs.
// Each session gets its own isolated client keyed by sessionID.
type SkillMCPManager struct {
	mu      sync.Mutex
	clients map[string]MCPClient // key = sessionID:serverName
	tokens  map[string]Token     // key = sessionID:serverName
}

// NewSkillMCPManager returns a new manager.
func NewSkillMCPManager() *SkillMCPManager {
	return &SkillMCPManager{
		clients: make(map[string]MCPClient),
		tokens:  make(map[string]Token),
	}
}

func (m *SkillMCPManager) key(sessionID, serverName string) string {
	return sessionID + ":" + serverName
}

// Register stores a client + token for a (session, server) pair.
// Replaces any existing entry.
func (m *SkillMCPManager) Register(sessionID, serverName string, c MCPClient, t Token) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[m.key(sessionID, serverName)] = c
	m.tokens[m.key(sessionID, serverName)] = t
}

// Get returns the client for a (session, server) pair.
func (m *SkillMCPManager) Get(sessionID, serverName string) (MCPClient, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.clients[m.key(sessionID, serverName)]
	return c, ok
}

// Token returns the stored token for a (session, server) pair.
func (m *SkillMCPManager) Token(sessionID, serverName string) (Token, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tokens[m.key(sessionID, serverName)]
	return t, ok
}

// Drop removes a (session, server) entry.
func (m *SkillMCPManager) Drop(sessionID, serverName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(sessionID, serverName)
	if c, ok := m.clients[k]; ok {
		_ = c.Close()
	}
	delete(m.clients, k)
	delete(m.tokens, k)
}

// CloseAll closes every registered client.
func (m *SkillMCPManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.clients {
		_ = c.Close()
	}
	m.clients = make(map[string]MCPClient)
	m.tokens = make(map[string]Token)
}

// Count returns the number of registered clients (for diagnostics).
func (m *SkillMCPManager) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.clients)
}
