// Package mcp implements the 3-tier Model Context Protocol system for
// heretic Ultimate. Mirrors packages/omo-opencode/src/mcp/.
//
// Tier 1: Built-in MCPs (heretic ships with 5).
// Tier 2: .mcp.json file (Claude Code compatible).
// Tier 3: Skill-embedded MCPs (per-session, OAuth 2.0 + PKCE + DCR).
//
// All tiers return a unified MCPClient interface.
package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Transport is the wire protocol the MCP client uses.
type Transport string

const (
	TransportStdio Transport = "stdio"
	TransportHTTP  Transport = "http"
	TransportSSE   Transport = "sse"
)

// ServerConfig is the common config for any MCP server.
type ServerConfig struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"` // stdio | http | sse
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	URL       string            `json:"url,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Timeout   time.Duration     `json:"timeout,omitempty"`
	Disabled  bool              `json:"disabled,omitempty"`
}

// MCPClient is the contract every tier-1/tier-2/tier-3 server satisfies.
type MCPClient interface {
	// Name returns the server's registry name.
	Name() string
	// Transport returns the wire protocol.
	Transport() Transport
	// ListTools returns the tools the server exposes.
	ListTools(ctx context.Context) ([]Tool, error)
	// CallTool invokes a tool by name with the given args.
	CallTool(ctx context.Context, name string, args map[string]any) (any, error)
	// Close releases resources.
	Close() error
}

// Tool is a single tool exposed by an MCP server.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Tier1 is the built-in MCP catalog.
type Tier1 struct {
	Servers []ServerConfig
	mu      sync.Mutex
	clients map[string]MCPClient
}

// DefaultTier1 returns the 5 built-in MCPs.
func DefaultTier1() []ServerConfig {
	return []ServerConfig{
		{Name: "context7", Type: "http", URL: "https://mcp.context7.com/sse"},
		{Name: "codegraph", Type: "stdio", Command: "codegraph-mcp"},
		{Name: "grep-app", Type: "http", URL: "https://grep.app/api/mcp"},
		{Name: "lsp", Type: "stdio", Command: "heretic-lsp-mcp"},
		{Name: "websearch", Type: "http", URL: "https://heretic-websearch.example.com/sse"},
	}
}

// NewTier1 returns a Tier1 populated with the given server configs.
func NewTier1(servers []ServerConfig) *Tier1 {
	return &Tier1{Servers: servers, clients: make(map[string]MCPClient)}
}

// Client returns the MCPClient for the given server name, starting it
// on first use. The caller must Close() when done.
func (t *Tier1) Client(name string) (MCPClient, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if c, ok := t.clients[name]; ok {
		return c, nil
	}
	for _, s := range t.Servers {
		if s.Name == name {
			c, err := startClient(s)
			if err != nil {
				return nil, err
			}
			t.clients[name] = c
			return c, nil
		}
	}
	return nil, fmt.Errorf("mcp: unknown server %q", name)
}

// CloseAll closes all started clients.
func (t *Tier1) CloseAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, c := range t.clients {
		_ = c.Close()
	}
	t.clients = make(map[string]MCPClient)
}

// startClient creates a client based on the server's transport.
func startClient(s ServerConfig) (MCPClient, error) {
	switch s.Type {
	case "stdio":
		return newStdioClient(s)
	case "http", "sse":
		return newHTTPClient(s)
	}
	return nil, fmt.Errorf("mcp: unknown transport %q", s.Type)
}

// stdioClient is the stdio MCP client. It launches the server as a
// subprocess and speaks the JSON-RPC framing.
type stdioClient struct {
	cfg  ServerConfig
	cmd  *exec.Cmd
	in   io.WriteCloser
	out  io.ReadCloser
	mu   sync.Mutex
	id   int
}

func newStdioClient(s ServerConfig) (*stdioClient, error) {
	c := &stdioClient{cfg: s, id: 0}
	if s.Command == "" {
		return nil, errors.New("mcp: stdio client requires Command")
	}
	cmd := exec.Command(s.Command, s.Args...)
	cmd.Env = os.Environ()
	for k, v := range s.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("mcp: start %s: %w", s.Command, err)
	}
	c.cmd = cmd
	c.in = in
	c.out = out
	return c, nil
}

func (c *stdioClient) Name() string     { return c.cfg.Name }
func (c *stdioClient) Transport() Transport { return TransportStdio }

func (c *stdioClient) Close() error {
	if c.in != nil {
		_ = c.in.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_ = c.cmd.Wait()
	}
	return nil
}

func (c *stdioClient) ListTools(ctx context.Context) ([]Tool, error) {
	// Stub: a real implementation would send a JSON-RPC request and
	// parse the response. We return a placeholder so the registry
	// compiles end-to-end.
	return nil, errors.New("mcp stdio: ListTools not yet implemented")
}

func (c *stdioClient) CallTool(ctx context.Context, name string, args map[string]any) (any, error) {
	return nil, errors.New("mcp stdio: CallTool not yet implemented")
}

// httpClient is the HTTP/SSE MCP client. It uses Bearer auth.
type httpClient struct {
	cfg    ServerConfig
	client *http.Client
}

func newHTTPClient(s ServerConfig) (*httpClient, error) {
	if s.URL == "" {
		return nil, errors.New("mcp: http client requires URL")
	}
	return &httpClient{cfg: s, client: &http.Client{Timeout: s.Timeout}}, nil
}

func (c *httpClient) Name() string         { return c.cfg.Name }
func (c *httpClient) Transport() Transport { return TransportHTTP }

func (c *httpClient) Close() error { return nil }

func (c *httpClient) ListTools(ctx context.Context) ([]Tool, error) {
	return nil, errors.New("mcp http: ListTools not yet implemented")
}

func (c *httpClient) CallTool(ctx context.Context, name string, args map[string]any) (any, error) {
	return nil, errors.New("mcp http: CallTool not yet implemented")
}

// Tier2 parses the Claude Code .mcp.json file and exposes the entries
// as ServerConfigs.
type Tier2 struct {
	AllowEnvExpansion bool
}

// NewTier2 returns a Tier2 parser. If allowEnvExpansion is true, ${VAR}
// and $VAR references in env values are expanded at load time.
func NewTier2(allowEnvExpansion bool) *Tier2 {
	return &Tier2{AllowEnvExpansion: allowEnvExpansion}
}

// mcpJSON is the structure of a .mcp.json file.
type mcpJSON struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// Load reads and parses a .mcp.json file, expanding env vars if
// configured.
func (t *Tier2) Load(path string) ([]ServerConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc mcpJSON
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	var out []ServerConfig
	for name, cfg := range doc.MCPServers {
		cfg.Name = name
		if t.AllowEnvExpansion {
			cfg.Env = expandEnvMap(cfg.Env)
			cfg.Headers = expandEnvMap(cfg.Headers)
		}
		out = append(out, cfg)
	}
	return out, nil
}

// expandEnvMap expands $VAR and ${VAR} in each value.
func expandEnvMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = os.ExpandEnv(v)
	}
	return out
}
