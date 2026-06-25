package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultTier1(t *testing.T) {
	t.Parallel()
	servers := DefaultTier1()
	require.Len(t, servers, 5)
}

func TestTier1_Client_UnknownServer(t *testing.T) {
	t.Parallel()
	t1 := NewTier1(DefaultTier1())
	_, err := t1.Client("nonexistent")
	require.Error(t, err)
}

func TestTier2_Load(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mcp.json")
	body := `{
		"mcpServers": {
			"alpha": {
				"type": "stdio",
				"command": "alpha-mcp",
				"env": {"TOKEN": "${TEST_TOKEN}"}
			}
		}
	}`
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	t2 := NewTier2(true)
	t.Setenv("TEST_TOKEN", "secret-123")
	servers, err := t2.Load(path)
	require.NoError(t, err)
	require.Len(t, servers, 1)
	require.Equal(t, "alpha", servers[0].Name)
	require.Equal(t, "secret-123", servers[0].Env["TOKEN"])
}

func TestTier2_Load_NoEnvExpansion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mcp.json")
	body := `{
		"mcpServers": {
			"alpha": {
				"type": "stdio",
				"command": "alpha-mcp",
				"env": {"TOKEN": "${TEST_TOKEN}"}
			}
		}
	}`
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	t2 := NewTier2(false)
	t.Setenv("TEST_TOKEN", "secret-123")
	servers, err := t2.Load(path)
	require.NoError(t, err)
	require.Equal(t, "${TEST_TOKEN}", servers[0].Env["TOKEN"])
}

func TestSkillMCPManager_RegisterAndGet(t *testing.T) {
	t.Parallel()
	mgr := NewSkillMCPManager()
	c, err := newStdioClient(ServerConfig{Name: "test", Type: "stdio", Command: "/bin/true"})
	if err != nil {
		t.Skip("stdio client requires /bin/true")
	}
	defer c.Close()
	mgr.Register("sess-1", "test", c, Token{AccessToken: "x"})
	got, ok := mgr.Get("sess-1", "test")
	require.True(t, ok)
	require.Equal(t, "test", got.Name())
	tok, ok := mgr.Token("sess-1", "test")
	require.True(t, ok)
	require.Equal(t, "x", tok.AccessToken)
}

func TestSkillMCPManager_Drop(t *testing.T) {
	t.Parallel()
	mgr := NewSkillMCPManager()
	mgr.Register("sess", "srv", &fakeClient{}, Token{AccessToken: "x"})
	require.Equal(t, 1, mgr.Count())
	mgr.Drop("sess", "srv")
	require.Equal(t, 0, mgr.Count())
}

type fakeClient struct{}

func (f *fakeClient) Name() string         { return "fake" }
func (f *fakeClient) Transport() Transport { return TransportStdio }
func (f *fakeClient) ListTools(ctx context.Context) ([]Tool, error) { return nil, nil }
func (f *fakeClient) CallTool(ctx context.Context, name string, args map[string]any) (any, error) { return nil, nil }
func (f *fakeClient) Close() error { return nil }

func TestPKCE(t *testing.T) {
	t.Parallel()
	pkce, err := NewPKCE()
	require.NoError(t, err)
	require.NotEmpty(t, pkce.Verifier)
	require.NotEmpty(t, pkce.Challenge)
	require.Equal(t, "S256", pkce.Method)
	// Challenge is SHA-256(Verifier), base64url-encoded, 43 chars.
	require.Len(t, pkce.Challenge, 43)
}

func TestOAuthClient_AuthorizeURL(t *testing.T) {
	t.Parallel()
	c := NewOAuthClient(OAuthConfig{
		AuthorizationEndpoint: "https://example.com/oauth/authorize",
		TokenEndpoint:         "https://example.com/oauth/token",
		ClientID:              "client-1",
		RedirectURL:           "https://app/cb",
		Scopes:                []string{"read", "write"},
	})
	pkce, _ := NewPKCE()
	u, err := c.AuthorizeURL(pkce, "state-1")
	require.NoError(t, err)
	require.Contains(t, u, "response_type=code")
	require.Contains(t, u, "client_id=client-1")
	require.Contains(t, u, "code_challenge="+pkce.Challenge)
	require.Contains(t, u, "state=state-1")
}

func TestToken_Expired(t *testing.T) {
	t.Parallel()
	tok := Token{ExpiresAt: timeExpired()}
	require.True(t, tok.Expired())
	tok = Token{ExpiresAt: timeNow().Add(2 * time.Hour)}
	require.False(t, tok.Expired())
}

// timeExpired and timeNow are tiny helpers to keep the test
// focused on the behavior, not on time literals.
func timeExpired() time.Time      { return time.Time{} } // zero is in the past
func timeNow() time.Time         { return time.Now() }
