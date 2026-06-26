package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gedwolmen/heretic/internal/home"
)

// mcpServersFile is the shared shape of Claude Desktop's
// claude_desktop_config.json, Cursor's mcp.json, and Claude Code's
// .mcp.json / ~/.claude.json: a top-level "mcpServers" object.
type mcpServersFile struct {
	MCPServers map[string]json.RawMessage `json:"mcpServers"`
}

// discoveredServer is the union of fields used by Claude Desktop, Cursor, and
// Claude Code to describe a single MCP server. "type" is Claude Desktop's
// transport field; Cursor uses "transport" for the same purpose.
type discoveredServer struct {
	Type      string            `json:"type"`
	Transport string            `json:"transport"`
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	Env       map[string]string `json:"env"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
}

// normalizeDiscovered converts a raw discovered server entry into a
// [MCPConfig]. Returns ok=false when the entry cannot be classified as a
// stdio command or a remote (URL) server.
func normalizeDiscovered(d discoveredServer) (MCPConfig, bool) {
	switch {
	case d.URL != "":
		t := strings.ToLower(strings.TrimSpace(d.Type))
		if t == "" {
			t = strings.ToLower(strings.TrimSpace(d.Transport))
		}
		var mt MCPType
		switch t {
		case "http":
			mt = MCPHttp
		case "sse":
			mt = MCPSSE
		default:
			// Remote servers without an explicit transport default to
			// SSE, which is the legacy remote default for Claude and
			// Cursor configs.
			mt = MCPSSE
		}
		return MCPConfig{
			Type:    mt,
			URL:     d.URL,
			Headers: d.Headers,
			Env:     d.Env,
		}, true
	case d.Command != "":
		return MCPConfig{
			Type:    MCPStdio,
			Command: d.Command,
			Args:    d.Args,
			Env:     d.Env,
		}, true
	}
	return MCPConfig{}, false
}

// parseMCPServers parses the "mcpServers" object from a Claude/Cursor-style
// config file. Unparseable individual entries are skipped rather than failing
// the whole file.
func parseMCPServers(data []byte) map[string]MCPConfig {
	var f mcpServersFile
	if err := json.Unmarshal(data, &f); err != nil {
		slog.Debug("Failed to parse MCP config file", "error", err)
		return nil
	}
	out := make(map[string]MCPConfig)
	for name, raw := range f.MCPServers {
		var d discoveredServer
		if err := json.Unmarshal(raw, &d); err != nil {
			slog.Debug("Failed to parse MCP server entry", "name", name, "error", err)
			continue
		}
		if m, ok := normalizeDiscovered(d); ok {
			out[name] = m
		}
	}
	return out
}

// opencodeMCPFile is the shape of opencode's opencode.json: a top-level "mcp"
// object (not "mcpServers" as in Claude/Cursor).
type opencodeMCPFile struct {
	MCP map[string]json.RawMessage `json:"mcp"`
}

// opencodeServer is the per-server shape in opencode's config. Opencode uses
// "command" as an array (not command+args), "environment" (not env), "type"
// is "local" (stdio) or "remote", and an "enabled" flag.
type opencodeServer struct {
	Type        string            `json:"type"`
	Command     []string          `json:"command"`
	URL         string            `json:"url"`
	Environment map[string]string `json:"environment"`
	Headers     map[string]string `json:"headers"`
	Enabled     *bool             `json:"enabled"`
}

// normalizeOpenCode converts an opencode server entry into an [MCPConfig].
// Returns ok=false for disabled or unclassifiable entries.
func normalizeOpenCode(d opencodeServer) (MCPConfig, bool) {
	if d.Enabled != nil && !*d.Enabled {
		return MCPConfig{}, false
	}
	switch strings.ToLower(strings.TrimSpace(d.Type)) {
	case "remote":
		if d.URL == "" {
			return MCPConfig{}, false
		}
		return MCPConfig{
			Type:    MCPSSE,
			URL:     d.URL,
			Headers: d.Headers,
			Env:     d.Environment,
		}, true
	default: // "local" or empty
		if len(d.Command) == 0 {
			return MCPConfig{}, false
		}
		cmd := d.Command[0]
		var args []string
		if len(d.Command) > 1 {
			args = d.Command[1:]
		}
		return MCPConfig{
			Type:    MCPStdio,
			Command: cmd,
			Args:    args,
			Env:     d.Environment,
		}, true
	}
}

// parseOpenCodeMCP parses the "mcp" object from an opencode-style config
// file. Unparseable individual entries are skipped rather than failing the
// whole file.
func parseOpenCodeMCP(data []byte) map[string]MCPConfig {
	var f opencodeMCPFile
	if err := json.Unmarshal(data, &f); err != nil {
		slog.Debug("Failed to parse opencode MCP config file", "error", err)
		return nil
	}
	out := make(map[string]MCPConfig)
	for name, raw := range f.MCP {
		var d opencodeServer
		if err := json.Unmarshal(raw, &d); err != nil {
			slog.Debug("Failed to parse opencode MCP server entry", "name", name, "error", err)
			continue
		}
		if m, ok := normalizeOpenCode(d); ok {
			out[name] = m
		}
	}
	return out
}

// mcpConfigPaths returns the well-known MCP config file locations on the host,
// in precedence order (first source wins on name collisions). workingDir adds
// project-level Claude Code (.mcp.json), Cursor (.cursor/mcp.json), and
// opencode (opencode.json) files.
func mcpConfigPaths(workingDir string) []string {
	var paths []string

	// Claude Desktop.
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths, filepath.Join(home.Dir(), "Library", "Application Support", "Claude", "claude_desktop_config.json"))
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			paths = append(paths, filepath.Join(appdata, "Claude", "claude_desktop_config.json"))
		}
	default:
		paths = append(paths, filepath.Join(home.Config(), "Claude", "claude_desktop_config.json"))
	}

	// Cursor (global).
	paths = append(paths, filepath.Join(home.Dir(), ".cursor", "mcp.json"))

	// Claude Code (global ~/.claude.json carries a top-level mcpServers).
	paths = append(paths, filepath.Join(home.Dir(), ".claude.json"))

	// opencode (global ~/.config/opencode/opencode.json uses "mcp" key).
	paths = append(paths, filepath.Join(home.Config(), "opencode", "opencode.json"))

	// Project-level.
	if workingDir != "" {
		paths = append(paths,
			filepath.Join(workingDir, ".mcp.json"),
			filepath.Join(workingDir, ".cursor", "mcp.json"),
			filepath.Join(workingDir, "opencode.json"),
		)
	}

	return paths
}

// DiscoverInstalledMCPs scans well-known MCP configuration files on the host
// (Claude Desktop, Cursor, Claude Code, opencode, and the project's .mcp.json)
// and returns the servers normalized to [MCPConfig]. On name collisions between
// sources, the first source encountered wins. Missing or unparseable files are
// skipped silently.
func DiscoverInstalledMCPs(workingDir string) map[string]MCPConfig {
	out := make(map[string]MCPConfig)
	for _, p := range mcpConfigPaths(workingDir) {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var servers map[string]MCPConfig
		if filepath.Base(p) == "opencode.json" {
			servers = parseOpenCodeMCP(data)
		} else {
			servers = parseMCPServers(data)
		}
		for name, m := range servers {
			if _, exists := out[name]; exists {
				continue
			}
			out[name] = m
		}
	}
	return out
}

// mergeMCPs merges discovered servers into dst. Entries already present in dst
// are left untouched so user-declared configuration always wins. The dst map
// is mutated in place; the return value is the number of servers added.
func mergeMCPs(dst MCPs, discovered map[string]MCPConfig) int {
	if dst == nil {
		return 0
	}
	added := 0
	for name, m := range discovered {
		if _, exists := dst[name]; exists {
			continue
		}
		dst[name] = m
		added++
	}
	return added
}

// MergeDiscoveredMCPs scans the host for installed MCP servers (see
// [DiscoverInstalledMCPs]) and merges any that are not already declared in the
// user's config into the in-memory config. Discovered servers are not persisted
// to disk; they are re-merged on each startup. User-declared servers always
// take precedence.
//
// Auto-detection is enabled by default; set options.auto_detect_mcp to false to
// disable it. Returns the number of servers added.
func (s *ConfigStore) MergeDiscoveredMCPs() int {
	if opt := s.config.Options; opt != nil && opt.AutoDetectMCP != nil && !*opt.AutoDetectMCP {
		slog.Debug("MCP auto-detection disabled by config")
		return 0
	}
	discovered := DiscoverInstalledMCPs(s.workingDir)
	if len(discovered) == 0 {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.config.MCP == nil {
		s.config.MCP = make(MCPs)
	}
	return mergeMCPs(s.config.MCP, discovered)
}
