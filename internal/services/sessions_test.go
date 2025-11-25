package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/alexanderramin/kalistheniks/internal/services/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

//go:generate mockgen -source=sessions.go -destination=../services/mocks/sessions_mock.go -package=mocks SessionService

func TestSessionService_CreateSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionRepository := mocks.NewMockSessionRepository(ctrl)
	ctx := context.Background()
	performedAt := time.Now().Add(-24 * time.Hour)
	sessionType := "workout"
	sessionID := uuid.New()
	userID := uuid.New()
	notes := "Felt great!"

	t.Run("creates session successfully", func(t *testing.T) {
		service := NewSessionService(mockSessionRepository)
		mockSessionRepository.EXPECT().Create(ctx, gomock.Any()).Return(&models.Session{
			ID:          &sessionID,
			UserID:      &userID,
			PerformedAt: performedAt.UTC(),
			SessionType: &sessionType,
			Notes:       &notes,
		}, nil)
		res, err := service.CreateSession(ctx, &userID, &performedAt, &sessionType, &notes)
		require.NoError(t, err)
		require.Equal(t, &sessionID, res.ID)

	})

	t.Run("handles repository error", func(t *testing.T) {
		service := NewSessionService(mockSessionRepository)
		mockSessionRepository.EXPECT().Create(ctx, gomock.Any()).Return(&models.Session{}, errors.New("db error"))
		res, err := service.CreateSession(ctx, &userID, &performedAt, &sessionType, &notes)
		require.Error(t, err)
		require.ErrorContains(t, err, "db error")
		require.Empty(t, res)
	})
}

func TestSessionService_AddSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionRepository := mocks.NewMockSessionRepository(ctrl)
	ctx := context.Background()
	sessionID := uuid.New()
	setID := uuid.New()
	exerciseID := uuid.New()
	userID := uuid.New()

	t.Run("adds set successfully", func(t *testing.T) {
		service := NewSessionService(mockSessionRepository)
		mockSessionRepository.EXPECT().SessionBelongsToUser(ctx, &sessionID, &userID).Return(true, nil)
		mockSessionRepository.EXPECT().AddSet(ctx, gomock.Any()).Return(&models.Set{
			ID:         &setID,
			SessionID:  &sessionID,
			ExerciseID: &exerciseID,
			SetIndex:   1,
			Reps:       10,
			WeightKG:   50.0,
		}, nil)
		res, err := service.AddSet(ctx, &userID, &sessionID, &exerciseID, 1, 10, 50.0, nil)
		require.NoError(t, err)
		require.Equal(t, &setID, res.ID)
	})

	t.Run("handles session ownership error", func(t *testing.T) {
		service := NewSessionService(mockSessionRepository)
		mockSessionRepository.EXPECT().SessionBelongsToUser(ctx, &sessionID, &userID).Return(false, nil)
		res, err := service.AddSet(ctx, &userID, &sessionID, &exerciseID, 1, 10, 50.0, nil)
		require.Error(t, err)
		require.ErrorContains(t, err, "session does not belong to user")
		require.Empty(t, res)
	})

	t.Run("handles repository error", func(t *testing.T) {
		service := NewSessionService(mockSessionRepository)
		mockSessionRepository.EXPECT().SessionBelongsToUser(ctx, &sessionID, &userID).Return(true, nil)
		mockSessionRepository.EXPECT().AddSet(ctx, gomock.Any()).Return(&models.Set{}, errors.New("db error"))
		res, err := service.AddSet(ctx, &userID, &sessionID, &exerciseID, 1, 10, 50.0, nil)
		require.Error(t, err)
		require.ErrorContains(t, err, "db error")
		require.Empty(t, res)
	})
}

func TestSessionService_ListSessions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionRepository := mocks.NewMockSessionRepository(ctrl)
	ctx := context.Background()
	userID := uuid.New()
	sessionID := uuid.New()

	t.Run("lists sessions successfully", func(t *testing.T) {
		service := NewSessionService(mockSessionRepository)
		mockSessionRepository.EXPECT().ListWithSets(ctx, &userID).Return([]*models.Session{
			{
				ID:     &sessionID,
				UserID: &userID,
				Sets:   []models.Set{},
			},
		}, nil)
		res, err := service.ListSessions(ctx, &userID)
		require.NoError(t, err)
		require.Len(t, res, 1)
		require.Equal(t, &sessionID, res[0].ID)
		require.NotNil(t, res[0].Sets)
		require.Empty(t, res[0].Sets)
	})

}
