package contracts

import (
	"context"
	"time"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/google/uuid"
)

type AuthService interface {
	Signup(ctx context.Context, email, password string) (*models.User, string, error)
	Login(ctx context.Context, email, password string) (*models.User, string, error)
	VerifyToken(ctx context.Context, token string) (string, error)
}

type SessionService interface {
	CreateSession(ctx context.Context, userID uuid.UUID, performedAt *time.Time, sessionType *string, notes *string) (*models.Session, error)
	AddSet(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, exerciseID uuid.UUID, setIndex, reps int, weight float64, rpe *int) (*models.Set, error)
	ListSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
}

type PlanService interface {
	NextSuggestion(ctx context.Context, userID uuid.UUID) (*models.PlanSuggestion, error)
}
