package middleware

import (
	"net/http"

	"github.com/MarkyGitHub/e-commerce/internal/auth"
	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// RequireRole erstellt Middleware die eine bestimmte Rolle erfordert
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				shared.WriteError(w, http.StatusUnauthorized, "not authenticated")
				return
			}

			if !hasRole(user, role) {
				shared.WriteError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin erstellt Middleware die Admin-Rolle erfordert
func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole(auth.RoleAdmin)(next)
}

// hasRole prüft ob Benutzer die erforderliche Rolle hat
// Admin hat Zugriff auf alles, Kunde nur auf Kunden-Ressourcen
func hasRole(user *auth.User, requiredRole string) bool {
	// Admin hat alle Berechtigungen
	if user.Role == auth.RoleAdmin {
		return true
	}
	// Prüfen ob Benutzer die erforderliche Rolle hat
	return user.Role == requiredRole
}
