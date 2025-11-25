package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alexanderramin/kalistheniks/internal/handlers/contracts"
	"github.com/alexanderramin/kalistheniks/internal/handlers/middleware"
	"github.com/alexanderramin/kalistheniks/internal/handlers/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	Sessions contracts.SessionService
	Plans    contracts.PlanService
}

func New(sessions contracts.SessionService, plans contracts.PlanService) *Handler {
	return &Handler{
		Sessions: sessions,
		Plans:    plans,
	}
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.CurrentUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var payload struct {
		PerformedAt *time.Time `json:"performed_at"`
		SessionType *string    `json:"session_type"`
		Notes       *string    `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	session, err := h.Sessions.CreateSession(r.Context(), userID, payload.PerformedAt, payload.SessionType, payload.Notes)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	response.JSON(w, http.StatusCreated, session)
}

func (h *Handler) CreateSet(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.CurrentUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	sessionID := chi.URLParam(r, "id")
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid session ID")
		return
	}
	var payload struct {
		ExerciseID string  `json:"exercise_id"`
		SetIndex   int     `json:"set_index"`
		Reps       int     `json:"reps"`
		WeightKG   float64 `json:"weight_kg"`
		RPE        *int    `json:"rpe"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	exerciseUUID, err := uuid.Parse(payload.ExerciseID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid exercise ID")
		return
	}
	set, err := h.Sessions.AddSet(r.Context(), userID, sessionUUID, exerciseUUID, payload.SetIndex, payload.Reps, payload.WeightKG, payload.RPE)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to add set")
		return
	}
	response.JSON(w, http.StatusCreated, set)
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.CurrentUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	sessions, err := h.Sessions.ListSessions(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}
	response.JSON(w, http.StatusOK, sessions)
}

func (h *Handler) NextPlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.CurrentUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	suggestion, err := h.Plans.NextSuggestion(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to compute next plan")
		return
	}
	response.JSON(w, http.StatusOK, suggestion)
}
