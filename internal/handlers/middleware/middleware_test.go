package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuthService is a mock implementation for testing
type mockAuthService struct {
	verifyTokenFunc func(ctx context.Context, token string) (string, error)
}

func (m *mockAuthService) VerifyToken(ctx context.Context, token string) (string, error) {
	if m.verifyTokenFunc != nil {
		return m.verifyTokenFunc(ctx, token)
	}
	return "", errors.New("not implemented")
}

func (m *mockAuthService) Signup(ctx context.Context, email, password string) (*models.User, string, error) {
	return nil, "", errors.New("not implemented")
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	return nil, "", errors.New("not implemented")
}

func TestRequireAuth(t *testing.T) {
	userID := uuid.New().String()

	tests := []struct {
		name           string
		authHeader     string
		verifyFunc     func(ctx context.Context, token string) (string, error)
		expectedStatus int
		expectUserID   bool
	}{
		{
			name:           "missing token",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "invalid bearer format",
			authHeader:     "InvalidFormat token",
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:       "valid token",
			authHeader: "Bearer validtoken",
			verifyFunc: func(ctx context.Context, token string) (string, error) {
				return userID, nil
			},
			expectedStatus: http.StatusOK,
			expectUserID:   true,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalidtoken",
			verifyFunc: func(ctx context.Context, token string) (string, error) {
				return "", errors.New("invalid token")
			},
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "empty bearer token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "bearer only",
			authHeader:     "Bearer",
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := &mockAuthService{
				verifyTokenFunc: tt.verifyFunc,
			}
			middleware := NewAuth(mockAuth)

			// Create test handler that checks for user ID in context
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectUserID {
					id, ok := CurrentUserID(r)
					assert.True(t, ok, "expected user ID in context")
					assert.Equal(t, userID, id.String())
				}
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.RequireAuth(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestCurrentUserID(t *testing.T) {
	tests := []struct {
		name        string
		contextVal  interface{}
		expectValid bool
	}{
		{
			name:        "valid UUID",
			contextVal:  uuid.New().String(),
			expectValid: true,
		},
		{
			name:        "invalid UUID string",
			contextVal:  "not-a-uuid",
			expectValid: false,
		},
		{
			name:        "empty string",
			contextVal:  "",
			expectValid: false,
		},
		{
			name:        "wrong type",
			contextVal:  123,
			expectValid: false,
		},
		{
			name:        "nil context value",
			contextVal:  nil,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.contextVal != nil {
				ctx := context.WithValue(req.Context(), userIDContextKey, tt.contextVal)
				req = req.WithContext(ctx)
			}

			id, ok := CurrentUserID(req)

			if tt.expectValid {
				assert.True(t, ok)
				assert.NotEqual(t, uuid.Nil, id)
			} else {
				assert.False(t, ok)
				assert.Equal(t, uuid.Nil, id)
			}
		})
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		expectedToken string
	}{
		{
			name:          "valid bearer token",
			authHeader:    "Bearer mytoken123",
			expectedToken: "mytoken123",
		},
		{
			name:          "token with spaces",
			authHeader:    "Bearer token with spaces",
			expectedToken: "token with spaces",
		},
		{
			name:          "empty header",
			authHeader:    "",
			expectedToken: "",
		},
		{
			name:          "bearer lowercase",
			authHeader:    "bearer mytoken",
			expectedToken: "",
		},
		{
			name:          "no bearer prefix",
			authHeader:    "mytoken",
			expectedToken: "",
		},
		{
			name:          "bearer only",
			authHeader:    "Bearer",
			expectedToken: "",
		},
		{
			name:          "bearer with space only",
			authHeader:    "Bearer ",
			expectedToken: "",
		},
		{
			name:          "basic auth",
			authHeader:    "Basic dXNlcjpwYXNz",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			token := ExtractBearerToken(req)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify all security headers are set
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "default-src 'self'", rec.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
	assert.Equal(t, "geolocation=(), microphone=(), camera=()", rec.Header().Get("Permissions-Policy"))
}

func TestCORS(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		origin         string
		expectHeaders  bool
		expectedStatus int
	}{
		{
			name:           "preflight request with origin",
			method:         http.MethodOptions,
			origin:         "https://example.com",
			expectHeaders:  true,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "preflight without origin",
			method:         http.MethodOptions,
			origin:         "",
			expectHeaders:  false,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "regular request with origin",
			method:         http.MethodGet,
			origin:         "https://example.com",
			expectHeaders:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "regular request without origin",
			method:         http.MethodGet,
			origin:         "",
			expectHeaders:  false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectHeaders {
				assert.Equal(t, tt.origin, rec.Header().Get("Access-Control-Allow-Origin"))
				assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Methods"))
				assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Headers"))
			} else {
				assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestCurrentUserIDFromRequest(t *testing.T) {
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), userIDContextKey, userID.String())
	req = req.WithContext(ctx)

	extractedID, ok := CurrentUserID(req)
	require.True(t, ok)
	assert.Equal(t, userID, extractedID)
}
