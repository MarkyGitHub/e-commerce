package middleware

import (
	"context"
	"net/http"

	"github.com/mschl/ecommerce-api/internal/auth"
	"github.com/mschl/ecommerce-api/internal/shared"
)

// UserContextKey ist aus shared importiert um Import-Zyklen zu vermeiden

// AuthMiddleware verarbeitet Authentifizierung für geschützte Endpunkte
type AuthMiddleware struct {
	authService auth.AuthService
}

// NewAuthMiddleware erstellt eine neue AuthMiddleware
func NewAuthMiddleware(authService auth.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// Authenticate validiert Session-Token und hängt Benutzer an Kontext
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

		// Benutzer an Kontext anhängen
		ctx := context.WithValue(r.Context(), shared.UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext holt den Benutzer aus dem Request-Kontext
func GetUserFromContext(ctx context.Context) (*auth.User, bool) {
	user, ok := ctx.Value(shared.UserContextKey).(*auth.User)
	return user, ok
}
