package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/alexanderramin/kalistheniks/internal/handlers/contracts"
	"github.com/alexanderramin/kalistheniks/internal/handlers/response"
	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/alexanderramin/kalistheniks/internal/validation"
)

type Handler struct {
	AuthService contracts.AuthService
}

func New(auth contracts.AuthService) *Handler {
	return &Handler{AuthService: auth}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	// Limit request body size to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxErr):
			response.Error(w, http.StatusBadRequest, "malformed JSON")
		case errors.As(err, &unmarshalErr):
			response.Error(w, http.StatusBadRequest, "invalid field type")
		case errors.Is(err, io.ErrUnexpectedEOF):
			response.Error(w, http.StatusBadRequest, "malformed JSON")
		case err.Error() == "http: request body too large":
			response.Error(w, http.StatusRequestEntityTooLarge, "request body too large")
		default:
			response.Error(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}

	// Validate email format
	if err := validation.ValidateEmail(payload.Email); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid email address")
		return
	}

	// Validate password strength
	if err := validation.ValidatePassword(payload.Password); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	user, token, err := h.AuthService.Signup(r.Context(), payload.Email, payload.Password)
	if err != nil {
		// Don't expose internal error details
		response.Error(w, http.StatusBadRequest, "failed to create account")
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{
		"user":  userResponse(user),
		"token": token,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	// Limit request body size to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxErr):
			response.Error(w, http.StatusBadRequest, "malformed JSON")
		case errors.As(err, &unmarshalErr):
			response.Error(w, http.StatusBadRequest, "invalid field type")
		case errors.Is(err, io.ErrUnexpectedEOF):
			response.Error(w, http.StatusBadRequest, "malformed JSON")
		case err.Error() == "http: request body too large":
			response.Error(w, http.StatusRequestEntityTooLarge, "request body too large")
		default:
			response.Error(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}

	// Basic validation - don't give away whether email exists
	if payload.Email == "" || payload.Password == "" {
		response.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	user, token, err := h.AuthService.Login(r.Context(), payload.Email, payload.Password)
	if err != nil {
		// Use generic error message to avoid user enumeration
		response.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"user":  userResponse(user),
		"token": token,
	})
}

func userResponse(u *models.User) map[string]any {
	return map[string]any{
		"id":         u.ID,
		"email":      u.Email,
		"created_at": u.CreatedAt,
	}
}
