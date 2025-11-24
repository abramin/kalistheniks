package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/go-chi/chi/v5"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func (a *App) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) signup(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, token, err := a.AuthService.Signup(r.Context(), payload.Email, payload.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"user":  userResponse(user),
		"token": token,
	})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, token, err := a.AuthService.Login(r.Context(), payload.Email, payload.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user":  userResponse(user),
		"token": token,
	})
}

func (a *App) createSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var payload struct {
		PerformedAt *time.Time `json:"performed_at"`
		SessionType *string    `json:"session_type"`
		Notes       *string    `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	session, err := a.SessionService.CreateSession(r.Context(), userID, payload.PerformedAt, payload.SessionType, payload.Notes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (a *App) createSet(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	sessionID := chi.URLParam(r, "id")
	var payload struct {
		ExerciseID string  `json:"exercise_id"`
		SetIndex   int     `json:"set_index"`
		Reps       int     `json:"reps"`
		WeightKG   float64 `json:"weight_kg"`
		RPE        *int    `json:"rpe"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	set, err := a.SessionService.AddSet(r.Context(), userID, sessionID, payload.ExerciseID, payload.SetIndex, payload.Reps, payload.WeightKG, payload.RPE)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add set")
		return
	}
	writeJSON(w, http.StatusCreated, set)
}

func (a *App) listSessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	sessions, err := a.SessionService.ListSessions(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (a *App) nextPlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	suggestion, err := a.PlanService.NextSuggestion(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute next plan")
		return
	}
	writeJSON(w, http.StatusOK, suggestion)
}

func (a *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing token")
			return
		}
		userID, err := a.AuthService.VerifyToken(r.Context(), token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func userResponse(u models.User) map[string]any {
	return map[string]any{
		"id":         u.ID,
		"email":      u.Email,
		"created_at": u.CreatedAt,
	}
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(authHeader) > len(prefix) && authHeader[:len(prefix)] == prefix {
		return authHeader[len(prefix):]
	}
	return ""
}

func currentUserID(r *http.Request) (string, bool) {
	id, ok := r.Context().Value(userIDContextKey).(string)
	return id, ok && id != ""
}
