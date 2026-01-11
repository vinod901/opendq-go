package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/vinod901/opendq-go/internal/auth"
	"github.com/vinod901/opendq-go/internal/authorization"
	"github.com/vinod901/opendq-go/internal/tenant"
)

// Context keys for storing request-scoped data
type contextKey string

const (
	contextKeyClaims contextKey = "claims"
	contextKeyUserID contextKey = "user_id"
)

// AuthMiddleware handles OIDC authentication
type AuthMiddleware struct {
	authManager *auth.Manager
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authManager *auth.Manager) *AuthMiddleware {
	return &AuthMiddleware{
		authManager: authManager,
	}
}

// Handle authenticates requests
func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for public endpoints
		if isPublicEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

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

// TenantMiddleware handles tenant resolution
type TenantMiddleware struct {
	tenantManager *tenant.Manager
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(tenantManager *tenant.Manager) *TenantMiddleware {
	return &TenantMiddleware{
		tenantManager: tenantManager,
	}
}

// Handle resolves tenant from request
func (m *TenantMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract tenant from subdomain or header
		tenantSlug := extractTenantSlug(r)
		if tenantSlug == "" {
			http.Error(w, "Tenant not found", http.StatusBadRequest)
			return
		}

		// Add tenant to context
		ctx := tenant.WithTenantSlug(r.Context(), tenantSlug)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthzMiddleware handles OpenFGA authorization
type AuthzMiddleware struct {
	authzManager *authorization.Manager
}

// NewAuthzMiddleware creates a new authorization middleware
func NewAuthzMiddleware(authzManager *authorization.Manager) *AuthzMiddleware {
	return &AuthzMiddleware{
		authzManager: authzManager,
	}
}

// Handle checks authorization
func (m *AuthzMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authz for public endpoints
		if isPublicEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get user from context
		userID, ok := r.Context().Value(contextKeyUserID).(string)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get tenant from context
		tenantSlug, err := tenant.GetTenantSlug(r.Context())
		if err != nil {
			http.Error(w, "Tenant not found", http.StatusBadRequest)
			return
		}

		// Check if user has access to tenant
		allowed, err := m.authzManager.CheckTenantAccess(
			r.Context(),
			userID,
			tenantSlug,
			authorization.RelationMember,
		)
		if err != nil {
			http.Error(w, "Authorization check failed", http.StatusInternalServerError)
			return
		}

		if !allowed {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS
type CORSMiddleware struct {
	allowedOrigins []string
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(allowedOrigins []string) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: allowedOrigins,
	}
}

// Handle adds CORS headers
func (m *CORSMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if m.isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *CORSMiddleware) isAllowedOrigin(origin string) bool {
	for _, allowed := range m.allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// Helper functions

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

func extractTenantSlug(r *http.Request) string {
	// Try to get from header first
	if tenantHeader := r.Header.Get("X-Tenant"); tenantHeader != "" {
		return tenantHeader
	}

	// Try to get from subdomain
	host := r.Host
	parts := strings.Split(host, ".")
	if len(parts) > 2 {
		return parts[0]
	}

	return ""
}
