package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexanderramin/kalistheniks/internal/auth"
	"github.com/alexanderramin/kalistheniks/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (models.User, error)
	FindByEmail(ctx context.Context, email string) (models.User, error)
}

type AuthService struct {
	users     UserRepository
	jwtSecret string
}

// TODO: move errors to relevant packages
var (
	ErrHashPassword       = errors.New("failed to hash password")
	ErrCreateUser         = errors.New("failed to create user")
	ErrFindUser           = errors.New("failed to find user")
	ErrGenerateToken      = errors.New("failed to generate token")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrParseToken         = errors.New("failed to parse token")
)

func NewAuthService(users UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Signup(ctx context.Context, email, password string) (models.User, string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, "", fmt.Errorf("%w: %v", ErrHashPassword, err)
	}

	user, err := s.users.Create(ctx, email, string(hashed))
	if err != nil {
		return models.User{}, "", fmt.Errorf("%w: %v", ErrCreateUser, err)
	}

	token, err := auth.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return models.User{}, "", fmt.Errorf("%w: %v", ErrGenerateToken, err)
	}
	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (models.User, string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return models.User{}, "", fmt.Errorf("%w: %v", ErrFindUser, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return models.User{}, "", ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return models.User{}, "", fmt.Errorf("%w: %v", ErrGenerateToken, err)
	}
	return user, token, nil
}

func (s *AuthService) VerifyToken(_ context.Context, token string) (string, error) {
	userID, err := auth.ParseToken(token, s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrParseToken, err)
	}
	return userID, nil
}
