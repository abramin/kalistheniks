package services

import (
	"context"
	"errors"
	"time"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(ctx context.Context, s *models.Session) (*models.Session, error)
	AddSet(ctx context.Context, set *models.Set) (*models.Set, error)
	ListWithSets(ctx context.Context, userID *uuid.UUID) ([]*models.Session, error)
	SessionBelongsToUser(ctx context.Context, sessionID *uuid.UUID, userID *uuid.UUID) (bool, error)
}

type SessionService struct {
	sessions SessionRepository
}

func NewSessionService(repo SessionRepository) *SessionService {
	return &SessionService{sessions: repo}
}

func (s *SessionService) CreateSession(ctx context.Context, userID *uuid.UUID, performedAt *time.Time, sessionType *string, notes *string) (*models.Session, error) {
	when := time.Now().UTC()
	if performedAt != nil {
		when = performedAt.UTC()
	}

	session := &models.Session{
		UserID:      userID,
		PerformedAt: when,
		Notes:       notes,
		SessionType: sessionType,
	}

	return s.sessions.Create(ctx, session)
}

func (s *SessionService) AddSet(ctx context.Context, userID *uuid.UUID, sessionID *uuid.UUID, exerciseID *uuid.UUID, setIndex, reps int, weight float64, rpe *int) (*models.Set, error) {
	owned, err := s.sessions.SessionBelongsToUser(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	if !owned {
		return nil, errors.New("session does not belong to user")
	}

	set := &models.Set{
		SessionID:  sessionID,
		ExerciseID: exerciseID,
		SetIndex:   setIndex,
		Reps:       reps,
		WeightKG:   weight,
		RPE:        rpe,
	}
	return s.sessions.AddSet(ctx, set)
}

func (s *SessionService) ListSessions(ctx context.Context, userID *uuid.UUID) ([]*models.Session, error) {
	return s.sessions.ListWithSets(ctx, userID)
}
