package middleware

import (
	"context"
	"net/http"

	"github.com/enkaigaku/dvd-rental/pkg/auth"
)

// ContextKey is the type for context keys used by middleware.
type ContextKey string

const (
	// ClaimsContextKey is the context key for JWT claims.
	ClaimsContextKey ContextKey = "claims"
)

// AuthMiddleware verifies JWT tokens and injects claims into context.
type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

// Require returns middleware that requires a valid JWT in the Authorization header.
func (m *AuthMiddleware) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			WriteJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
			return
		}

		tokenStr, err := auth.ExtractToken(authHeader)
		if err != nil {
			WriteJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header format")
			return
		}

		claims, err := m.jwtManager.VerifyToken(tokenStr)
		if err != nil {
			WriteJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns middleware that checks the authenticated user has the given role.
// Must be used after Require.
func (m *AuthMiddleware) RequireRole(role auth.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				WriteJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
				return
			}

			if claims.Role != role {
				WriteJSONError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClaims extracts JWT claims from request context.
func GetClaims(ctx context.Context) *auth.Claims {
	claims, ok := ctx.Value(ClaimsContextKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}
