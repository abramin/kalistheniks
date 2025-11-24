package services

import (
	"context"
	"errors"

	"github.com/alexanderramin/kalistheniks/internal/auth"
	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/alexanderramin/kalistheniks/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users     repositories.UserRepository
	jwtSecret string
}

func NewAuthService(users repositories.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Signup(ctx context.Context, email, password string) (models.User, string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, "", err
	}

	user, err := s.users.Create(ctx, email, string(hashed))
	if err != nil {
		return models.User{}, "", err
	}

	token, err := auth.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return models.User{}, "", err
	}
	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (models.User, string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return models.User{}, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return models.User{}, "", errors.New("invalid credentials")
	}

	token, err := auth.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return models.User{}, "", err
	}
	return user, token, nil
}

func (s *AuthService) VerifyToken(_ context.Context, token string) (string, error) {
	return auth.ParseToken(token, s.jwtSecret)
}
