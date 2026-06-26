package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeDiscovered(t *testing.T) {
	t.Parallel()

	t.Run("stdio server from command", func(t *testing.T) {
		t.Parallel()
		m, ok := normalizeDiscovered(discoveredServer{
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
			Env:     map[string]string{"ROOT": "/tmp"},
		})
		require.True(t, ok)
		require.Equal(t, MCPStdio, m.Type)
		require.Equal(t, "npx", m.Command)
		require.Equal(t, []string{"-y", "@modelcontextprotocol/server-filesystem"}, m.Args)
		require.Equal(t, map[string]string{"ROOT": "/tmp"}, m.Env)
	})

	t.Run("remote sse from type field", func(t *testing.T) {
		t.Parallel()
		m, ok := normalizeDiscovered(discoveredServer{
			URL:  "https://mcp.example.com/sse",
			Type: "sse",
		})
		require.True(t, ok)
		require.Equal(t, MCPSSE, m.Type)
		require.Equal(t, "https://mcp.example.com/sse", m.URL)
	})

	t.Run("remote http from transport field", func(t *testing.T) {
		t.Parallel()
		m, ok := normalizeDiscovered(discoveredServer{
			URL:       "https://mcp.example.com/mcp",
			Transport: "http",
			Headers:   map[string]string{"Authorization": "Bearer x"},
		})
		require.True(t, ok)
		require.Equal(t, MCPHttp, m.Type)
		require.Equal(t, "https://mcp.example.com/mcp", m.URL)
		require.Equal(t, map[string]string{"Authorization": "Bearer x"}, m.Headers)
	})

	t.Run("remote url with no type defaults to sse", func(t *testing.T) {
		t.Parallel()
		m, ok := normalizeDiscovered(discoveredServer{
			URL: "https://mcp.example.com/sse",
		})
		require.True(t, ok)
		require.Equal(t, MCPSSE, m.Type)
	})

	t.Run("type takes precedence over transport", func(t *testing.T) {
		t.Parallel()
		m, ok := normalizeDiscovered(discoveredServer{
			URL:       "https://mcp.example.com/mcp",
			Type:      "http",
			Transport: "sse",
		})
		require.True(t, ok)
		require.Equal(t, MCPHttp, m.Type)
	})

	t.Run("unclassifiable entry is skipped", func(t *testing.T) {
		t.Parallel()
		_, ok := normalizeDiscovered(discoveredServer{})
		require.False(t, ok)
	})
}

func TestParseMCPServers(t *testing.T) {
	t.Parallel()

	t.Run("parses mixed stdio and remote servers", func(t *testing.T) {
		t.Parallel()
		data := []byte(`{
			"mcpServers": {
				"filesystem": {
					"command": "npx",
					"args": ["-y", "@modelcontextprotocol/server-filesystem"]
				},
				"remote": {
					"type": "http",
					"url": "https://mcp.example.com/mcp"
				}
			}
		}`)
		got := parseMCPServers(data)
		require.Len(t, got, 2)
		require.Equal(t, MCPStdio, got["filesystem"].Type)
		require.Equal(t, "npx", got["filesystem"].Command)
		require.Equal(t, MCPHttp, got["remote"].Type)
		require.Equal(t, "https://mcp.example.com/mcp", got["remote"].URL)
	})

	t.Run("skips entry with wrong field type but keeps valid ones", func(t *testing.T) {
		t.Parallel()
		data := []byte(`{
			"mcpServers": {
				"good": {"command": "npx"},
				"bad": {"command": 123}
			}
		}`)
		got := parseMCPServers(data)
		require.Len(t, got, 1)
		require.Contains(t, got, "good")
	})

	t.Run("missing mcpServers key returns empty", func(t *testing.T) {
		t.Parallel()
		data := []byte(`{"other": 123}`)
		got := parseMCPServers(data)
		require.Empty(t, got)
	})

	t.Run("invalid top-level json returns nil", func(t *testing.T) {
		t.Parallel()
		got := parseMCPServers([]byte(`{not json`))
		require.Nil(t, got)
	})
}

func TestMergeMCPs(t *testing.T) {
	t.Parallel()

	t.Run("user-declared server is not overwritten", func(t *testing.T) {
		t.Parallel()
		dst := MCPs{
			"shared": {Type: MCPStdio, Command: "user-command"},
		}
		discovered := map[string]MCPConfig{
			"shared": {Type: MCPStdio, Command: "discovered-command"},
			"new":    {Type: MCPHttp, URL: "https://mcp.example.com"},
		}
		added := mergeMCPs(dst, discovered)
		require.Equal(t, 1, added)
		require.Equal(t, "user-command", dst["shared"].Command)
		require.Contains(t, dst, "new")
	})

	t.Run("nil dst returns zero", func(t *testing.T) {
		t.Parallel()
		added := mergeMCPs(nil, map[string]MCPConfig{
			"x": {Type: MCPStdio, Command: "npx"},
		})
		require.Equal(t, 0, added)
	})

	t.Run("adds all new servers and returns count", func(t *testing.T) {
		t.Parallel()
		dst := MCPs{}
		discovered := map[string]MCPConfig{
			"a": {Type: MCPStdio, Command: "a"},
			"b": {Type: MCPSSE, URL: "https://a.example.com"},
			"c": {Type: MCPHttp, URL: "https://b.example.com"},
		}
		added := mergeMCPs(dst, discovered)
		require.Equal(t, 3, added)
		require.Len(t, dst, 3)
	})
}

func TestDiscoverInstalledMCPs_ProjectFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	const serverName = "heretic-test-discover-fs"
	err := os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(`{
		"mcpServers": {
			"`+serverName+`": {
				"command": "npx",
				"args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
			}
		}
	}`), 0o644)
	require.NoError(t, err)

	got := DiscoverInstalledMCPs(dir)
	m, ok := got[serverName]
	require.True(t, ok, "project .mcp.json server should be discovered")
	require.Equal(t, MCPStdio, m.Type)
	require.Equal(t, "npx", m.Command)
	require.Equal(t, []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"}, m.Args)
}

func TestNormalizeOpenCode(t *testing.T) {
	t.Parallel()

	t.Run("local server from command array", func(t *testing.T) {
		t.Parallel()
		enabled := true
		m, ok := normalizeOpenCode(opencodeServer{
			Type:        "local",
			Command:     []string{"uvx", "minimax-coding-plan-mcp", "-y"},
			Environment: map[string]string{"API_KEY": "secret"},
			Enabled:     &enabled,
		})
		require.True(t, ok)
		require.Equal(t, MCPStdio, m.Type)
		require.Equal(t, "uvx", m.Command)
		require.Equal(t, []string{"minimax-coding-plan-mcp", "-y"}, m.Args)
		require.Equal(t, map[string]string{"API_KEY": "secret"}, m.Env)
	})

	t.Run("local server with single command element", func(t *testing.T) {
		t.Parallel()
		m, ok := normalizeOpenCode(opencodeServer{
			Command: []string{"codegraph"},
		})
		require.True(t, ok)
		require.Equal(t, MCPStdio, m.Type)
		require.Equal(t, "codegraph", m.Command)
		require.Nil(t, m.Args)
	})

	t.Run("remote server", func(t *testing.T) {
		t.Parallel()
		m, ok := normalizeOpenCode(opencodeServer{
			Type:    "remote",
			URL:     "https://mcp.example.com/sse",
			Headers: map[string]string{"Authorization": "Bearer x"},
		})
		require.True(t, ok)
		require.Equal(t, MCPSSE, m.Type)
		require.Equal(t, "https://mcp.example.com/sse", m.URL)
		require.Equal(t, map[string]string{"Authorization": "Bearer x"}, m.Headers)
	})

	t.Run("disabled server is skipped", func(t *testing.T) {
		t.Parallel()
		disabled := false
		_, ok := normalizeOpenCode(opencodeServer{
			Command: []string{"npx"},
			Enabled: &disabled,
		})
		require.False(t, ok)
	})

	t.Run("nil enabled defaults to enabled", func(t *testing.T) {
		t.Parallel()
		m, ok := normalizeOpenCode(opencodeServer{
			Command: []string{"npx"},
		})
		require.True(t, ok)
		require.Equal(t, MCPStdio, m.Type)
	})

	t.Run("empty command is skipped", func(t *testing.T) {
		t.Parallel()
		_, ok := normalizeOpenCode(opencodeServer{Type: "local"})
		require.False(t, ok)
	})

	t.Run("remote without url is skipped", func(t *testing.T) {
		t.Parallel()
		_, ok := normalizeOpenCode(opencodeServer{Type: "remote"})
		require.False(t, ok)
	})
}

func TestParseOpenCodeMCP(t *testing.T) {
	t.Parallel()

	t.Run("parses local and remote servers", func(t *testing.T) {
		t.Parallel()
		data := []byte(`{
			"mcp": {
				"local-srv": {
					"type": "local",
					"command": ["uvx", "some-mcp"],
					"enabled": true
				},
				"remote-srv": {
					"type": "remote",
					"url": "https://mcp.example.com/sse"
				}
			}
		}`)
		got := parseOpenCodeMCP(data)
		require.Len(t, got, 2)
		require.Equal(t, MCPStdio, got["local-srv"].Type)
		require.Equal(t, "uvx", got["local-srv"].Command)
		require.Equal(t, []string{"some-mcp"}, got["local-srv"].Args)
		require.Equal(t, MCPSSE, got["remote-srv"].Type)
		require.Equal(t, "https://mcp.example.com/sse", got["remote-srv"].URL)
	})

	t.Run("skips disabled server", func(t *testing.T) {
		t.Parallel()
		data := []byte(`{
			"mcp": {
				"off": {"type": "local", "command": ["npx"], "enabled": false},
				"on": {"type": "local", "command": ["npx"]}
			}
		}`)
		got := parseOpenCodeMCP(data)
		require.Len(t, got, 1)
		require.Contains(t, got, "on")
	})

	t.Run("skips entry with wrong field type", func(t *testing.T) {
		t.Parallel()
		data := []byte(`{
			"mcp": {
				"good": {"type": "local", "command": ["npx"]},
				"bad": {"command": 123}
			}
		}`)
		got := parseOpenCodeMCP(data)
		require.Len(t, got, 1)
		require.Contains(t, got, "good")
	})

	t.Run("missing mcp key returns empty", func(t *testing.T) {
		t.Parallel()
		data := []byte(`{"other": 123}`)
		got := parseOpenCodeMCP(data)
		require.Empty(t, got)
	})
}

func TestDiscoverInstalledMCPs_OpenCodeFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	const serverName = "heretic-test-opencode-srv"
	err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(`{
		"mcp": {
			"`+serverName+`": {
				"type": "local",
				"command": ["codegraph", "serve", "--mcp"],
				"enabled": true
			}
		}
	}`), 0o644)
	require.NoError(t, err)

	got := DiscoverInstalledMCPs(dir)
	m, ok := got[serverName]
	require.True(t, ok, "project opencode.json server should be discovered")
	require.Equal(t, MCPStdio, m.Type)
	require.Equal(t, "codegraph", m.Command)
	require.Equal(t, []string{"serve", "--mcp"}, m.Args)
}
