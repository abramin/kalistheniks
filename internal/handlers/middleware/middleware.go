package middleware

import (
	"context"
	"net/http"

	"github.com/alexanderramin/kalistheniks/internal/handlers/response"
	"github.com/alexanderramin/kalistheniks/internal/services"
	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "userID"

type Auth struct {
	auth *services.AuthService
}

func NewAuth(auth *services.AuthService) *Auth {
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
func CurrentUserID(r *http.Request) (*uuid.UUID, bool) {
	id, ok := r.Context().Value(userIDContextKey).(string)
	uid, err := uuid.Parse(id)
	if !ok || err != nil {
		return nil, false
	}
	return &uid, true
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
