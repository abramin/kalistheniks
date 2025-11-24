package services

import (
	"context"
	"time"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/alexanderramin/kalistheniks/internal/repositories"
)

type SessionService struct {
	sessions repositories.SessionRepository
}

func NewSessionService(repo repositories.SessionRepository) *SessionService {
	return &SessionService{sessions: repo}
}

func (s *SessionService) CreateSession(ctx context.Context, userID string, performedAt *time.Time, sessionType *string, notes *string) (models.Session, error) {
	when := time.Now().UTC()
	if performedAt != nil {
		when = performedAt.UTC()
	}

	session := models.Session{
		UserID:      userID,
		PerformedAt: when,
		Notes:       notes,
		SessionType: sessionType,
	}

	return s.sessions.Create(ctx, session)
}

func (s *SessionService) AddSet(ctx context.Context, userID, sessionID, exerciseID string, setIndex, reps int, weight float64, rpe *int) (models.Set, error) {
	set := models.Set{
		SessionID:  sessionID,
		ExerciseID: exerciseID,
		SetIndex:   setIndex,
		Reps:       reps,
		WeightKG:   weight,
		RPE:        rpe,
	}
	// TODO: verify session belongs to user before inserting.
	return s.sessions.AddSet(ctx, set)
}

func (s *SessionService) ListSessions(ctx context.Context, userID string) ([]models.Session, error) {
	return s.sessions.ListWithSets(ctx, userID)
}
