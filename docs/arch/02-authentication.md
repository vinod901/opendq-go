# Authentication

OpenDQ uses OIDC (OpenID Connect) for authentication, supporting any OIDC-compliant identity provider such as Keycloak, Okta, Auth0, or Azure AD.

## Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│                        Authentication Flow                                │
│                                                                          │
│  ┌─────────┐    ┌──────────┐    ┌──────────┐    ┌───────────────────┐  │
│  │ Browser │───▶│ OpenDQ   │───▶│ Keycloak │───▶│ User Credentials  │  │
│  │         │    │ Frontend │    │  (OIDC)  │    │                   │  │
│  └─────────┘    └──────────┘    └──────────┘    └───────────────────┘  │
│       │                              │                                   │
│       │                              │ JWT Token (ID + Access)           │
│       │                              ▼                                   │
│       │              ┌──────────────────────────────┐                   │
│       └─────────────▶│    OpenDQ API Server         │                   │
│                      │    (Validates JWT)           │                   │
│                      └──────────────────────────────┘                   │
└──────────────────────────────────────────────────────────────────────────┘
```

## OIDC Configuration

### Environment Variables

```bash
# OIDC Configuration
OIDC_ISSUER=http://localhost:8180/realms/master
OIDC_CLIENT_ID=opendq-client
OIDC_CLIENT_SECRET=your_client_secret_here
OIDC_REDIRECT_URL=http://localhost:8080/auth/callback
```

### Configuration Structure

```go
// pkg/config/config.go
type OIDCConfig struct {
    Issuer       string  // OIDC issuer URL (e.g., Keycloak realm URL)
    ClientID     string  // OAuth2 client ID
    ClientSecret string  // OAuth2 client secret
    RedirectURL  string  // OAuth2 callback URL
}
```

## Auth Manager

The Auth Manager (`internal/auth/auth.go`) handles all authentication operations:

### Initialization

```go
// Create auth manager with OIDC configuration
authManager, err := auth.NewManager(ctx, auth.Config{
    Issuer:       cfg.OIDC.Issuer,
    ClientID:     cfg.OIDC.ClientID,
    ClientSecret: cfg.OIDC.ClientSecret,
    RedirectURL:  cfg.OIDC.RedirectURL,
})
```

### Key Methods

#### 1. GetAuthURL - Generate Login URL

```go
// Generate URL to redirect users to for authentication
func (m *Manager) GetAuthURL(state string) string {
    return m.oauth2Config.AuthCodeURL(state)
}

// Usage in login handler:
// GET /auth/login
func handleLogin(w http.ResponseWriter, r *http.Request) {
    state := generateRandomState()
    authURL := authManager.GetAuthURL(state)
    http.Redirect(w, r, authURL, http.StatusFound)
}
```

#### 2. ExchangeCode - Handle OAuth2 Callback

```go
// Exchange authorization code for tokens
func (m *Manager) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
    return m.oauth2Config.Exchange(ctx, code)
}

// Usage in callback handler:
// GET /auth/callback?code=xxx&state=xxx
func handleCallback(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    token, err := authManager.ExchangeCode(r.Context(), code)
    if err != nil {
        http.Error(w, "Authentication failed", http.StatusUnauthorized)
        return
    }
    // Extract ID token and create session
}
```

#### 3. VerifyIDToken - Validate Tokens

```go
// Verify and decode ID token
func (m *Manager) VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
    return m.verifier.Verify(ctx, rawIDToken)
}
```

#### 4. ValidateToken - Authenticate API Requests

```go
// Validate token from API request
func (m *Manager) ValidateToken(ctx context.Context, token string) (*Claims, error) {
    idToken, err := m.VerifyIDToken(ctx, token)
    if err != nil {
        return nil, fmt.Errorf("invalid token: %w", err)
    }
    return ExtractClaims(idToken)
}
```

## Claims Structure

```go
// Claims represents OIDC claims extracted from ID token
type Claims struct {
    Subject       string   `json:"sub"`               // User ID
    Email         string   `json:"email"`             // Email address
    EmailVerified bool     `json:"email_verified"`    // Email verification status
    Name          string   `json:"name"`              // Full name
    PreferredName string   `json:"preferred_username"` // Username
    Groups        []string `json:"groups"`            // Group memberships
}
```

## Auth Middleware

The Auth Middleware validates tokens on protected endpoints:

```go
// internal/middleware/middleware.go
type AuthMiddleware struct {
    authManager *auth.Manager
}

func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip auth for public endpoints
        if isPublicEndpoint(r.URL.Path) {
            next.ServeHTTP(w, r)
            return
        }

        // Authenticate request
        claims, err := m.authManager.AuthenticateRequest(r.Context(), r)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Add claims to context
        ctx := context.WithValue(r.Context(), contextKeyClaims, claims)
        ctx = context.WithValue(ctx, contextKeyUserID, claims.Subject)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Public Endpoints

These endpoints bypass authentication:

```go
func isPublicEndpoint(path string) bool {
    publicPaths := []string{
        "/health",
        "/metrics",
        "/auth/login",
        "/auth/callback",
    }
    for _, publicPath := range publicPaths {
        if strings.HasPrefix(path, publicPath) {
            return true
        }
    }
    return false
}
```

## Login/Signup Flow

### Login Flow (Existing Users)

```
1. User visits /auth/login
2. Backend generates OAuth2 authorization URL with state
3. User is redirected to Keycloak login page
4. User enters credentials
5. Keycloak validates credentials
6. Keycloak redirects back to /auth/callback with authorization code
7. Backend exchanges code for tokens
8. Backend validates ID token
9. Backend creates session (stored in Redis/cookie)
10. User is redirected to application
```

### Signup Flow (New Users)

```
1. User visits /auth/signup (or Keycloak registration page)
2. User fills registration form in Keycloak
3. Keycloak creates user account
4. Same login flow as above
5. On first login, backend creates user record in database
6. User is associated with default tenant (or tenant from invite)
```

### Implementation Example

```go
// Auth routes to add to handler.go
mux.HandleFunc("/auth/login", h.handleLogin)
mux.HandleFunc("/auth/callback", h.handleCallback)
mux.HandleFunc("/auth/logout", h.handleLogout)
mux.HandleFunc("/auth/me", h.handleCurrentUser)

// Login handler
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
    state := generateState()
    // Store state in session/cookie for CSRF protection
    authURL := h.authManager.GetAuthURL(state)
    http.Redirect(w, r, authURL, http.StatusFound)
}

// Callback handler
func (h *Handler) handleCallback(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    state := r.URL.Query().Get("state")
    
    // Validate state for CSRF protection
    if !validateState(state) {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }
    
    // Exchange code for tokens
    token, err := h.authManager.ExchangeCode(r.Context(), code)
    if err != nil {
        http.Error(w, "Authentication failed", http.StatusUnauthorized)
        return
    }
    
    // Extract ID token
    rawIDToken, ok := token.Extra("id_token").(string)
    if !ok {
        http.Error(w, "No ID token", http.StatusInternalServerError)
        return
    }
    
    // Verify ID token
    idToken, err := h.authManager.VerifyIDToken(r.Context(), rawIDToken)
    if err != nil {
        http.Error(w, "Invalid ID token", http.StatusUnauthorized)
        return
    }
    
    // Extract claims
    claims, err := auth.ExtractClaims(idToken)
    if err != nil {
        http.Error(w, "Failed to extract claims", http.StatusInternalServerError)
        return
    }
    
    // Create or update user in database
    // Create session
    // Redirect to application
}
```

## Keycloak Setup

### 1. Create Realm

```bash
# Access Keycloak admin console
http://localhost:8180/admin
# Login: admin / admin
```

### 2. Create Client

1. Go to Clients → Create
2. Client ID: `opendq-client`
3. Client Protocol: `openid-connect`
4. Root URL: `http://localhost:8080`
5. Save

### 3. Configure Client

1. Access Type: `confidential`
2. Valid Redirect URIs: `http://localhost:8080/auth/callback`
3. Web Origins: `http://localhost:8080`
4. Save

### 4. Get Client Secret

1. Go to Credentials tab
2. Copy the Secret value
3. Set in environment: `OIDC_CLIENT_SECRET=xxx`

### 5. Create Users

1. Go to Users → Add User
2. Fill in username, email, etc.
3. Set password in Credentials tab

## Security Best Practices

1. **Token Validation**: Always validate tokens server-side
2. **HTTPS**: Use TLS in production
3. **State Parameter**: Use for CSRF protection
4. **Token Storage**: Store tokens securely (HttpOnly cookies)
5. **Token Refresh**: Implement token refresh mechanism
6. **Logout**: Properly invalidate sessions on logout
7. **Scope Limiting**: Request only necessary scopes

## Frontend Integration

### Using Access Token in API Requests

```javascript
// Frontend (SvelteKit)
async function fetchAPI(endpoint, options = {}) {
    const token = getAccessToken(); // From cookie/localStorage
    
    const response = await fetch(`/api/v1${endpoint}`, {
        ...options,
        headers: {
            ...options.headers,
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
    });
    
    if (response.status === 401) {
        // Token expired, redirect to login
        redirectToLogin();
    }
    
    return response;
}
```

## Troubleshooting

### Common Issues

1. **Invalid Issuer**: Ensure OIDC_ISSUER matches Keycloak realm URL exactly
2. **Invalid Redirect URI**: Must match configured redirect URI in Keycloak
3. **Token Expired**: Implement token refresh or re-authentication
4. **CORS Issues**: Configure Keycloak Web Origins correctly

### Debugging

```bash
# Check Keycloak is running
curl http://localhost:8180/realms/master/.well-known/openid-configuration

# Verify token (decode JWT)
echo $TOKEN | cut -d'.' -f2 | base64 -d | jq
```
