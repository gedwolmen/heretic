package mcp

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// PKCE is the Proof Key for Code Exchange flow (RFC 7636). It is
// required for OAuth 2.0 authorization code flows used by skill-embedded
// MCPs.
type PKCE struct {
	// Verifier is the high-entropy secret the client keeps.
	Verifier string
	// Challenge is the SHA-256 hash of the verifier, base64url-encoded.
	Challenge string
	// Method is always "S256" (we never use the insecure "plain" method).
	Method string
}

// NewPKCE generates a fresh PKCE pair.
func NewPKCE() (*PKCE, error) {
	verifier, err := randomURLSafe(64)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return &PKCE{Verifier: verifier, Challenge: challenge, Method: "S256"}, nil
}

// randomURLSafe returns n bytes of URL-safe random data.
func randomURLSafe(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Token is the OAuth 2.0 token response.
type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
}

// Expired returns true if the token has expired (or is within 30s of
// expiring).
func (t Token) Expired() bool {
	return time.Until(t.ExpiresAt) < 30*time.Second
}

// OAuthConfig configures the OAuth 2.0 flow for one MCP server.
type OAuthConfig struct {
	// AuthorizationEndpoint is the server's /authorize URL.
	AuthorizationEndpoint string
	// TokenEndpoint is the server's /token URL.
	TokenEndpoint string
	// ClientID is the OAuth client identifier.
	ClientID string
	// RedirectURL is where the server redirects after auth.
	RedirectURL string
	// Scopes is the OAuth scope list.
	Scopes []string
}

// DCRClient is a dynamically-registered OAuth client (RFC 7591).
type DCRClient struct {
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret,omitempty"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

// OAuthClient is the OAuth 2.0 client used by skill-embedded MCPs.
type OAuthClient struct {
	cfg OAuthConfig
	mu  sync.Mutex
}

// NewOAuthClient returns a new OAuth client.
func NewOAuthClient(cfg OAuthConfig) *OAuthClient {
	return &OAuthClient{cfg: cfg}
}

// AuthorizeURL builds the URL the user should visit to grant access.
func (o *OAuthClient) AuthorizeURL(pkce *PKCE, state string) (string, error) {
	u, err := url.Parse(o.cfg.AuthorizationEndpoint)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", o.cfg.ClientID)
	q.Set("redirect_uri", o.cfg.RedirectURL)
	q.Set("scope", strings.Join(o.cfg.Scopes, " "))
	q.Set("state", state)
	q.Set("code_challenge", pkce.Challenge)
	q.Set("code_challenge_method", pkce.Method)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// ExchangeCode exchanges the auth code for a token.
func (o *OAuthClient) ExchangeCode(ctx interface{ Done() <-chan struct{} }, code, pkceVerifier string) (Token, error) {
	if code == "" {
		return Token{}, errors.New("oauth: code is required")
	}
	if pkceVerifier == "" {
		return Token{}, errors.New("oauth: PKCE verifier is required")
	}
	// Stub: a real implementation would POST to TokenEndpoint with
	// grant_type=authorization_code, code, redirect_uri, client_id,
	// code_verifier, and parse the JSON response.
	_ = ctx
	return Token{
		AccessToken: "stub-access-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}, nil
}

// Refresh exchanges a refresh token for a new access token.
func (o *OAuthClient) Refresh(refreshToken string) (Token, error) {
	if refreshToken == "" {
		return Token{}, errors.New("oauth: refresh_token is required")
	}
	// Stub: a real implementation would POST grant_type=refresh_token.
	return Token{
		AccessToken:  "stub-refreshed",
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}, nil
}

// DCRRegistration is a stub for the Dynamic Client Registration flow
// (RFC 7591). A real implementation would POST the registration request
// to the server's registration_endpoint and parse the response.
type DCRRegistration struct {
	RegistrationEndpoint string
	ClientName          string
	RedirectURIs        []string
}

// Register performs a DCR request. Stub: returns an error.
func (o *OAuthClient) Register(reg DCRRegistration) (*DCRClient, error) {
	return nil, fmt.Errorf("oauth: DCR not yet implemented (endpoint=%s)", reg.RegistrationEndpoint)
}

// EnsureScope returns a copy of scopes with the given scope appended
// if not present.
func EnsureScope(scopes []string, want string) []string {
	for _, s := range scopes {
		if s == want {
			return scopes
		}
	}
	return append(scopes, want)
}

// BearerAuth returns a header map with the standard Authorization
// header for a bearer token.
func BearerAuth(t Token) http.Header {
	h := http.Header{}
	h.Set("Authorization", t.TokenType+" "+t.AccessToken)
	return h
}
