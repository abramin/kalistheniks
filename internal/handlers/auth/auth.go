package auth

import (
	"encoding/json"
	"net/http"

	"github.com/alexanderramin/kalistheniks/internal/handlers/response"
	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/alexanderramin/kalistheniks/internal/services"
)

type Handler struct {
	AuthService *services.AuthService
}

func New(auth *services.AuthService) *Handler {
	return &Handler{AuthService: auth}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, token, err := h.AuthService.Signup(r.Context(), payload.Email, payload.Password)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{
		"user":  userResponse(user),
		"token": token,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, token, err := h.AuthService.Login(r.Context(), payload.Email, payload.Password)
	if err != nil {
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
