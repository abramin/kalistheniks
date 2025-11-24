package services

import (
	"context"
	"errors"
	"testing"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/alexanderramin/kalistheniks/internal/services/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

// go: mockgen -source=auth.go -destination=../services/mocks/auth_mock.go -package=mocks AuthService

type authDeps struct {
	ctrl      *gomock.Controller
	svc       *AuthService
	usersRepo *mocks.MockUserRepository
	auth      *mocks.MockAuth
}

func newAuthDeps(t *testing.T) authDeps {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	usersRepo := mocks.NewMockUserRepository(ctrl)
	auth := mocks.NewMockAuth(ctrl)
	return authDeps{
		ctrl:      ctrl,
		svc:       &AuthService{users: usersRepo, jwtSecret: "testsecret", auth: auth},
		usersRepo: usersRepo,
		auth:      auth,
	}
}

func TestAuthService_Signup(t *testing.T) {
	ctx := context.Background()

	email := "test@example.com"
	password := "password123"

	t.Run("successful signup", func(t *testing.T) {
		deps := newAuthDeps(t)
		deps.usersRepo.EXPECT().Create(ctx, email, gomock.Any()).Return(models.User{ID: "user-123", Email: email}, nil)
		deps.auth.EXPECT().GenerateToken("user-123", "testsecret").Return("signed-token", nil)

		user, token, err := deps.svc.Signup(ctx, email, password)
		require.NoError(t, err)
		require.Equal(t, "user-123", user.ID)
		require.Equal(t, "signed-token", token)
	})

	t.Run("repository error", func(t *testing.T) {
		deps := newAuthDeps(t)
		deps.usersRepo.EXPECT().Create(ctx, email, gomock.Any()).Return(models.User{}, errors.New("db error"))
		user, token, err := deps.svc.Signup(ctx, email, password)
		require.ErrorIs(t, err, ErrCreateUser)
		require.ErrorContains(t, err, "db error")
		require.Empty(t, user)
		require.Empty(t, token)

	})

	t.Run("token generation error", func(t *testing.T) {
		deps := newAuthDeps(t)
		deps.usersRepo.EXPECT().Create(ctx, email, gomock.Any()).Return(models.User{ID: "user-123"}, nil)
		deps.auth.EXPECT().GenerateToken("user-123", "testsecret").Return("", errors.New("token error"))

		user, token, err := deps.svc.Signup(ctx, email, password)
		require.ErrorIs(t, err, ErrGenerateToken)
		require.ErrorContains(t, err, "token error")
		require.Empty(t, user)
		require.Empty(t, token)
	})
}

func TestAuthService_Login(t *testing.T) {
	const hashedPassword = "$2a$10$nzAyuLjvw2JKqKETtpFyvukYMwsMoAByVcziZ7RGnZUlQvehEJ8qq"

	t.Run("successful login", func(t *testing.T) {
		deps := newAuthDeps(t)
		deps.usersRepo.EXPECT().FindByEmail(gomock.Any(), "test@example.com").Return(models.User{ID: "user-123", PasswordHash: hashedPassword}, nil)
		deps.auth.EXPECT().GenerateToken("user-123", "testsecret").Return("valid.token.string", nil)

		user, token, err := deps.svc.Login(context.Background(), "test@example.com", "password123")
		require.NoError(t, err)
		require.Equal(t, "user-123", user.ID)
		require.Equal(t, "valid.token.string", token)
	})

	t.Run("user not found", func(t *testing.T) {
		deps := newAuthDeps(t)
		deps.usersRepo.EXPECT().FindByEmail(gomock.Any(), "test@example.com").Return(models.User{}, errors.New("user not found"))
		user, token, err := deps.svc.Login(context.Background(), "test@example.com", "password123")
		require.ErrorIs(t, err, ErrFindUser)
		require.ErrorContains(t, err, "user not found")
		require.Empty(t, user)
		require.Empty(t, token)
	})
	t.Run("invalid password", func(t *testing.T) {
		deps := newAuthDeps(t)
		deps.usersRepo.EXPECT().FindByEmail(gomock.Any(), "test@example.com").Return(models.User{ID: "user-123", PasswordHash: hashedPassword}, nil)
		user, token, err := deps.svc.Login(context.Background(), "test@example.com", "wrongpassword")
		require.ErrorIs(t, err, ErrInvalidCredentials)
		require.Empty(t, user)
		require.Empty(t, token)

	})
	t.Run("token generation error", func(t *testing.T) {
		deps := newAuthDeps(t)
		deps.usersRepo.EXPECT().FindByEmail(gomock.Any(), "test@example.com").Return(models.User{ID: "user-123", PasswordHash: hashedPassword}, nil)
		deps.auth.EXPECT().GenerateToken("user-123", "testsecret").Return("", errors.New("token error"))

		user, token, err := deps.svc.Login(context.Background(), "test@example.com", "password123")
		require.ErrorIs(t, err, ErrGenerateToken)
		require.ErrorContains(t, err, "token error")
		require.Empty(t, user)
		require.Empty(t, token)
	})
}

func TestAuthService_VerifyToken(t *testing.T) {
	deps := newAuthDeps(t)
	authService := deps.svc
	mockAuth := deps.auth

	t.Run("valid token", func(t *testing.T) {
		mockAuth.EXPECT().ParseToken("valid.token.string", "testsecret").Return("user-123", nil)
		userID, err := authService.VerifyToken(context.Background(), "valid.token.string")
		require.NoError(t, err)
		require.Equal(t, "user-123", userID)
	})
	t.Run("invalid token", func(t *testing.T) {
		mockAuth.EXPECT().ParseToken("invalid.token.string", "testsecret").Return("", errors.New("invalid token"))
		userID, err := authService.VerifyToken(context.Background(), "invalid.token.string")
		require.ErrorIs(t, err, ErrParseToken)
		require.ErrorContains(t, err, "invalid token")
		require.Empty(t, userID)
	})
}
