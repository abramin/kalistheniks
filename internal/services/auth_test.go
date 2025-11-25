package services

import (
	"context"
	"errors"
	"testing"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/alexanderramin/kalistheniks/internal/services/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

//go:generate mockgen -source=auth.go -destination=./mocks/services_mock.go -package=mocks AuthService

type authDeps struct {
	ctrl      *gomock.Controller
	svc       *AuthService
	usersRepo *mocks.MockUserRepository
}

func newAuthDeps(t *testing.T) authDeps {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	usersRepo := mocks.NewMockUserRepository(ctrl)
	return authDeps{
		ctrl:      ctrl,
		svc:       &AuthService{users: usersRepo, jwtSecret: "testsecret"},
		usersRepo: usersRepo,
	}
}

func TestAuthService_Signup(t *testing.T) {
	ctx := context.Background()

	email := "test@example.com"
	password := "password123"
	userID := uuid.New()

	t.Run("successful signup", func(t *testing.T) {
		deps := newAuthDeps(t)
		deps.usersRepo.EXPECT().Create(ctx, email, gomock.Any()).Return(&models.User{ID: userID, Email: email}, nil)
		user, token, err := deps.svc.Signup(ctx, email, password)
		require.NoError(t, err)
		require.Equal(t, userID, user.ID)
		require.NotEmpty(t, token)
	})

	t.Run("repository error", func(t *testing.T) {
		deps := newAuthDeps(t)
		deps.usersRepo.EXPECT().Create(ctx, email, gomock.Any()).Return(nil, errors.New("db error"))
		user, token, err := deps.svc.Signup(ctx, email, password)
		require.ErrorIs(t, err, ErrCreateUser)
		require.ErrorContains(t, err, "db error")
		require.Empty(t, user)
		require.Empty(t, token)

	})
}

func TestAuthService_Login(t *testing.T) {
	const hashedPassword = "$2a$10$nzAyuLjvw2JKqKETtpFyvukYMwsMoAByVcziZ7RGnZUlQvehEJ8qq"
	userID := uuid.New()
	t.Run("successful login", func(t *testing.T) {
		deps := newAuthDeps(t)
		deps.usersRepo.EXPECT().FindByEmail(gomock.Any(), "test@example.com").Return(&models.User{ID: userID, PasswordHash: hashedPassword}, nil)

		user, token, err := deps.svc.Login(context.Background(), "test@example.com", "password123")
		require.NoError(t, err)
		require.Equal(t, userID, user.ID)
		require.NotEmpty(t, token)
	})

	t.Run("user not found", func(t *testing.T) {
		deps := newAuthDeps(t)
		deps.usersRepo.EXPECT().FindByEmail(gomock.Any(), "test@example.com").Return(nil, errors.New("user not found"))
		user, token, err := deps.svc.Login(context.Background(), "test@example.com", "password123")
		require.ErrorIs(t, err, ErrFindUser)
		require.ErrorContains(t, err, "user not found")
		require.Empty(t, user)
		require.Empty(t, token)
	})
	t.Run("invalid password", func(t *testing.T) {
		deps := newAuthDeps(t)
		deps.usersRepo.EXPECT().FindByEmail(gomock.Any(), "test@example.com").Return(&models.User{ID: userID, PasswordHash: hashedPassword}, nil)
		user, token, err := deps.svc.Login(context.Background(), "test@example.com", "wrongpassword")
		require.ErrorIs(t, err, ErrInvalidCredentials)
		require.Empty(t, user)
		require.Empty(t, token)

	})
}
