package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/alexanderramin/kalistheniks/internal/handlers/contracts"
	"github.com/alexanderramin/kalistheniks/internal/handlers/middleware"
	"github.com/alexanderramin/kalistheniks/internal/handlers/response"
	"github.com/alexanderramin/kalistheniks/internal/validation"
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

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

	var payload struct {
		PerformedAt *time.Time `json:"performed_at"`
		SessionType *string    `json:"session_type"`
		Notes       *string    `json:"notes"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		handleJSONError(w, err)
		return
	}

	// Validate optional fields
	if payload.SessionType != nil {
		if err := validation.ValidateStringLength(*payload.SessionType, 1, 50, "session_type"); err != nil {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if payload.Notes != nil {
		if err := validation.ValidateStringLength(*payload.Notes, 0, 1000, "notes"); err != nil {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
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

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

	var payload struct {
		ExerciseID string  `json:"exercise_id"`
		SetIndex   int     `json:"set_index"`
		Reps       int     `json:"reps"`
		WeightKG   float64 `json:"weight_kg"`
		RPE        *int    `json:"rpe"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		handleJSONError(w, err)
		return
	}

	// Validate exercise ID
	exerciseUUID, err := uuid.Parse(payload.ExerciseID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid exercise ID")
		return
	}

	// Validate numeric fields
	if err := validation.ValidateNonNegativeInt(payload.SetIndex, "set_index"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validation.ValidatePositiveInt(payload.Reps, "reps"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validation.ValidateIntRange(payload.Reps, 1, 1000, "reps"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validation.ValidatePositiveFloat(payload.WeightKG, "weight_kg"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validation.ValidateFloatRange(payload.WeightKG, 0.1, 1000, "weight_kg"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if payload.RPE != nil {
		if err := validation.ValidateIntRange(*payload.RPE, 1, 10, "rpe"); err != nil {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
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

// handleJSONError provides consistent error handling for JSON decoding errors
func handleJSONError(w http.ResponseWriter, err error) {
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
}
