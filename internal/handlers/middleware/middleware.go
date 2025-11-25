package middleware

import (
	"context"
	"net/http"

	"github.com/alexanderramin/kalistheniks/internal/handlers/contracts"
	"github.com/alexanderramin/kalistheniks/internal/handlers/response"
	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "userID"

type Auth struct {
	auth contracts.AuthService
}

func NewAuth(auth contracts.AuthService) *Auth {
	return &Auth{auth: auth}
}

// RequireAuth ensures requests include a valid bearer token and injects the user ID into the request context.
func (m *Auth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ExtractBearerToken(r)
		if token == "" {
			response.Error(w, http.StatusUnauthorized, "missing token")
			return
		}
		userID, err := m.auth.VerifyToken(r.Context(), token)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CurrentUserID extracts the authenticated user ID from the request context.
func CurrentUserID(r *http.Request) (uuid.UUID, bool) {
	id, ok := r.Context().Value(userIDContextKey).(string)
	uid, err := uuid.Parse(id)
	if !ok || err != nil {
		return uuid.UUID{}, false
	}
	return uid, true
}

// ExtractBearerToken pulls a bearer token from the Authorization header.
func ExtractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(authHeader) > len(prefix) && authHeader[:len(prefix)] == prefix {
		return authHeader[len(prefix):]
	}
	return ""
}

// SecurityHeaders adds security-related HTTP headers to responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Content Security Policy - adjust as needed for your app
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions policy (replaces Feature-Policy)
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// CORS adds CORS headers for cross-origin requests
// For production, customize allowed origins
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow specific origins in production, "*" only for development
		origin := r.Header.Get("Origin")
		if origin != "" {
			// In production, validate origin against allowed list
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
