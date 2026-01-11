package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Manager handles OIDC authentication
type Manager struct {
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
	oauth2Config oauth2.Config
}

// Config contains OIDC configuration
type Config struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// NewManager creates a new authentication manager
func NewManager(ctx context.Context, cfg Config) (*Manager, error) {
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.ClientID,
	})

	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	return &Manager{
		provider:     provider,
		verifier:     verifier,
		oauth2Config: oauth2Config,
	}, nil
}

// GetAuthURL returns the URL to redirect users to for authentication
func (m *Manager) GetAuthURL(state string) string {
	return m.oauth2Config.AuthCodeURL(state)
}

// ExchangeCode exchanges an authorization code for tokens
func (m *Manager) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return m.oauth2Config.Exchange(ctx, code)
}

// VerifyIDToken verifies an ID token and returns the claims
func (m *Manager) VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return m.verifier.Verify(ctx, rawIDToken)
}

// UserInfo retrieves user information from the OIDC provider
func (m *Manager) UserInfo(ctx context.Context, tokenSource oauth2.TokenSource) (*oidc.UserInfo, error) {
	return m.provider.UserInfo(ctx, tokenSource)
}

// Claims represents OIDC claims
type Claims struct {
	Subject       string   `json:"sub"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Name          string   `json:"name"`
	PreferredName string   `json:"preferred_username"`
	Groups        []string `json:"groups"`
}

// ExtractClaims extracts claims from an ID token
func ExtractClaims(idToken *oidc.IDToken) (*Claims, error) {
	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}
	return &claims, nil
}

// ValidateToken validates a token from the Authorization header
func (m *Manager) ValidateToken(ctx context.Context, token string) (*Claims, error) {
	idToken, err := m.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, err := ExtractClaims(idToken)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// AuthenticateRequest authenticates an HTTP request
func (m *Manager) AuthenticateRequest(ctx context.Context, r *http.Request) (*Claims, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	return m.ValidateToken(ctx, token)
}
