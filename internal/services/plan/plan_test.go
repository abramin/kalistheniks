package plan

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/alexanderramin/kalistheniks/internal/services/plan/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

// go:generate mockgen -source=plan.go -destination=../services/plan/mocks/plan_mock.go -package=mocks SessionRepository
func TestPlanService_NextSuggestion(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	t.Run("generates suggestion successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSessionRepository := mocks.NewMockSessionRepository(ctrl)

		service := NewPlanService(mockSessionRepository)
		mockSessionRepository.EXPECT().GetLastSet(ctx, userID).Return(models.Set{
			ExerciseID: "push-ups",
			WeightKG:   0.0,
			Reps:       10,
		}, nil)
		stype := "workout"
		mockSessionRepository.EXPECT().GetLastSession(ctx, userID).Return(models.Session{
			SessionType: &stype,
		}, nil)
		suggestion, err := service.NextSuggestion(ctx, userID)
		require.NoError(t, err)
		require.Contains(t, suggestion.Notes, "Maintain weight and rep target")
		require.Equal(t, "push-ups", suggestion.ExerciseID)
		require.Equal(t, 0.0, suggestion.WeightKG)
		require.Equal(t, 10, suggestion.Reps)
	})
	t.Run("handles no history case", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSessionRepository := mocks.NewMockSessionRepository(ctrl)

		service := NewPlanService(mockSessionRepository)
		mockSessionRepository.EXPECT().GetLastSet(ctx, userID).Return(models.Set{}, sql.ErrNoRows)

		suggestion, err := service.NextSuggestion(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, 20.0, suggestion.WeightKG)
		require.Equal(t, 8, suggestion.Reps)
		require.Contains(t, suggestion.Notes, "No history found")
	})
	t.Run("increases weight on high reps", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSessionRepository := mocks.NewMockSessionRepository(ctrl)

		service := NewPlanService(mockSessionRepository)
		mockSessionRepository.EXPECT().GetLastSet(ctx, userID).Return(models.Set{
			ExerciseID: "squats",
			WeightKG:   40.0,
			Reps:       13,
		}, nil)

		suggestion, err := service.NextSuggestion(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, 42.5, suggestion.WeightKG)
		require.Equal(t, 13, suggestion.Reps)
		require.Contains(t, suggestion.Notes, "increase weight")
	})

	t.Run("decreases reps on low reps", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSessionRepository := mocks.NewMockSessionRepository(ctrl)

		service := NewPlanService(mockSessionRepository)
		mockSessionRepository.EXPECT().GetLastSet(ctx, userID).Return(models.Set{
			ExerciseID: "deadlifts",
			WeightKG:   60.0,
			Reps:       5,
		}, nil)

		suggestion, err := service.NextSuggestion(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, 60.0, suggestion.WeightKG)
		require.Equal(t, 4, suggestion.Reps)
		require.Contains(t, suggestion.Notes, "decrease reps")
	})

	t.Run("handles repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSessionRepository := mocks.NewMockSessionRepository(ctrl)

		mockSessionRepository.EXPECT().GetLastSet(ctx, "").Return(models.Set{}, errors.New("invalid user ID"))

		service := NewPlanService(mockSessionRepository)
		_, err := service.NextSuggestion(ctx, "")
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid user ID")
	})
}
