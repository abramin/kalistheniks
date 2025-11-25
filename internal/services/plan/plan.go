package plan

import (
	"context"
	"database/sql"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/google/uuid"
)

type SessionRepository interface {
	GetLastSet(ctx context.Context, userID uuid.UUID) (*models.Set, error)
	GetLastSession(ctx context.Context, userID uuid.UUID) (*models.Session, error)
}

// PlanService holds simple V1 progression logic.
type PlanService struct {
	sessions SessionRepository
}

func NewPlanService(repo SessionRepository) *PlanService {
	return &PlanService{sessions: repo}
}

// TODO: replace with a proper rule engine integration.
// NextSuggestion returns a naive progression recommendation based on the last recorded set.
func (p *PlanService) NextSuggestion(ctx context.Context, userID uuid.UUID) (*models.PlanSuggestion, error) {
	lastSet, err := p.sessions.GetLastSet(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No history: start with a default.
			id := uuid.New()
			return &models.PlanSuggestion{
				ExerciseID: id,
				WeightKG:   20,
				Reps:       8,
				Notes:      "No history found; starting default weight and reps.",
			}, nil
		}
		return nil, err
	}

	const upperRepRange = 12
	const lowerRepRange = 6
	suggestion := &models.PlanSuggestion{
		ExerciseID: lastSet.ExerciseID,
		WeightKG:   lastSet.WeightKG,
		Reps:       lastSet.Reps,
	}

	switch {
	case lastSet.Reps >= upperRepRange:
		suggestion.WeightKG = lastSet.WeightKG + 2.5
		suggestion.Notes = "Hit upper range; increase weight."
	case lastSet.Reps <= lowerRepRange:
		suggestion.Reps = lastSet.Reps - 1
		suggestion.Notes = "Fell short; keep weight, reduce reps."
	default:
		suggestion.Notes = "Maintain weight and rep target."
	}

	if lastSession, err := p.sessions.GetLastSession(ctx, userID); err == nil {
		if lastSession.SessionType != nil {
			switch *lastSession.SessionType {
			case "upper":
				suggestion.Notes += " Next: switch to lower body."
			case "lower":
				suggestion.Notes += " Next: switch to upper body."
			}
		}
	}

	return suggestion, nil
}
