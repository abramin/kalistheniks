package services

import (
	"context"
	"database/sql"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/alexanderramin/kalistheniks/internal/repositories"
)

// PlanService holds simple V1 progression logic.
type PlanService struct {
	sessions repositories.SessionRepository
}

func NewPlanService(repo repositories.SessionRepository) *PlanService {
	return &PlanService{sessions: repo}
}

// NextSuggestion returns a naive progression recommendation based on the last recorded set.
func (p *PlanService) NextSuggestion(ctx context.Context, userID string) (models.PlanSuggestion, error) {
	lastSet, err := p.sessions.GetLastSet(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No history: start with a default.
			return models.PlanSuggestion{
				ExerciseID: "",
				WeightKG:   20,
				Reps:       8,
				Notes:      "No history found; starting default weight and reps.",
			}, nil
		}
		return models.PlanSuggestion{}, err
	}

	const upperRepRange = 12
	const lowerRepRange = 6
	suggestion := models.PlanSuggestion{
		ExerciseID: lastSet.ExerciseID,
		WeightKG:   lastSet.WeightKG,
		Reps:       lastSet.Reps,
	}

	switch {
	case lastSet.Reps >= upperRepRange:
		suggestion.WeightKG = lastSet.WeightKG + 2.5
		suggestion.Notes = "Hit upper range; increase weight."
	case lastSet.Reps <= lowerRepRange:
		suggestion.Reps = lowerRepRange - 1
		suggestion.Notes = "Fell short; keep weight, reduce reps."
	default:
		suggestion.Notes = "Maintain weight and rep target."
	}

	if lastSession, err := p.sessions.GetLastSession(ctx, userID); err == nil {
		if lastSession.SessionType != nil {
			if *lastSession.SessionType == "upper" {
				suggestion.Notes += " Next: switch to lower body."
			} else if *lastSession.SessionType == "lower" {
				suggestion.Notes += " Next: switch to upper body."
			}
		}
	}

	return suggestion, nil
}
