package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) (*models.User, error) {
	const q = `
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, password_hash, created_at, updated_at`

	var u models.User
	err := r.db.QueryRowContext(ctx, q, email, passwordHash).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	const q = `
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE email = $1`

	var u models.User
	err := r.db.QueryRowContext(ctx, q, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return &u, err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const q = `
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE id = $1`

	var u models.User
	err := r.db.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}
