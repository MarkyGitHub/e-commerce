package middleware

import (
	"net/http"

	"github.com/mschl/ecommerce-api/internal/auth"
	"github.com/mschl/ecommerce-api/internal/shared"
)

// RequireRole creates middleware that requires a specific role
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

// RequireAdmin creates middleware that requires admin role
func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole(auth.RoleAdmin)(next)
}

// hasRole checks if a user has the required role
// Admin has access to everything, customer only has access to customer resources
func hasRole(user *auth.User, requiredRole string) bool {
	// Admin has all permissions
	if user.Role == auth.RoleAdmin {
		return true
	}
	// Check if user has the required role
	return user.Role == requiredRole
}
