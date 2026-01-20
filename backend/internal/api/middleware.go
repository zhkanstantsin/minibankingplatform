package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"minibankingplatform/pkg/jwt"
)

// publicPaths are endpoints that don't require authentication.
var publicPaths = map[string]bool{
	"/auth/login":    true,
	"/auth/register": true,
}

// AuthMiddleware creates a middleware that validates JWT tokens and injects claims into context.
func AuthMiddleware(tm *jwt.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for public paths
			if publicPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeUnauthorized(w, r.URL.Path, "Missing authorization header")
				return
			}

			// Check Bearer prefix
			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				writeUnauthorized(w, r.URL.Path, "Invalid authorization header format")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
			if tokenString == "" {
				writeUnauthorized(w, r.URL.Path, "Empty token")
				return
			}

			// Validate token
			claims, err := tm.ValidateToken(tokenString)
			if err != nil {
				writeUnauthorized(w, r.URL.Path, "Invalid or expired token")
				return
			}

			// Add claims to context and continue
			ctx := ContextWithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// writeUnauthorized writes a 401 response with ProblemDetails.
func writeUnauthorized(w http.ResponseWriter, instance string, detail string) {
	problem := ProblemDetails{
		Type:     problemBaseURL + "unauthorized",
		Title:    "Unauthorized",
		Status:   http.StatusUnauthorized,
		Detail:   ptr(detail),
		Instance: ptr(instance),
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(problem)
}
