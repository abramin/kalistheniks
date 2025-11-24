package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/alexanderramin/kalistheniks/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (models.User, error)
	FindByEmail(ctx context.Context, email string) (models.User, error)
	FindByID(ctx context.Context, id string) (models.User, error)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, email, passwordHash string) (models.User, error) {
	const q = `
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, password_hash, created_at, updated_at`

	var u models.User
	err := r.db.QueryRowContext(ctx, q, email, passwordHash).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (models.User, error) {
	const q = `
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE email = $1`

	var u models.User
	err := r.db.QueryRowContext(ctx, q, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, err
	}
	return u, err
}

func (r *userRepo) FindByID(ctx context.Context, id string) (models.User, error) {
	const q = `
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE id = $1`

	var u models.User
	err := r.db.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}
