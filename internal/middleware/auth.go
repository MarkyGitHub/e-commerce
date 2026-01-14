package middleware

import (
	"context"
	"net/http"

	"github.com/mschl/ecommerce-api/internal/auth"
	"github.com/mschl/ecommerce-api/internal/shared"
)

// UserContextKey is imported from shared to avoid import cycles

// AuthMiddleware handles authentication for protected endpoints
type AuthMiddleware struct {
	authService auth.AuthService
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(authService auth.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// Authenticate validates the session token and attaches the user to the context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Session-Token")
		if token == "" {
			shared.WriteError(w, http.StatusUnauthorized, "missing session token")
			return
		}

		user, err := m.authService.ValidateSession(r.Context(), token)
		if err != nil {
			shared.WriteError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}

		// Attach user to context
		ctx := context.WithValue(r.Context(), shared.UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) (*auth.User, bool) {
	user, ok := ctx.Value(shared.UserContextKey).(*auth.User)
	return user, ok
}
