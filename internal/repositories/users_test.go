package repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	t.Run("creates user successfully", func(t *testing.T) {
		repo := NewUserRepository(testDB)
		email := "test@example.com"
		password := "hashedpassword"

		user, err := repo.Create(context.Background(), email, password)
		require.NoError(t, err)
		require.Equal(t, email, user.Email)
		require.NotEmpty(t, user.ID)
		truncateUsers(t)
	})

	t.Run("duplicate email raises error", func(t *testing.T) {
		repo := NewUserRepository(testDB)
		email := "test@example.com"
		password := "hashedpassword"

		_, err := repo.Create(context.Background(), email, password)
		require.NoError(t, err)

		_, err = repo.Create(context.Background(), email, password)
		require.Error(t, err)
		truncateUsers(t)
	})

}

func TestUserRepository_FindByEmail(t *testing.T) {
	t.Run("finds user successfully", func(t *testing.T) {
		repo := NewUserRepository(testDB)
		email := "test@example.com"
		password := "hashedpassword"
		createdUser, err := repo.Create(context.Background(), email, password)
		require.NoError(t, err)

		foundUser, err := repo.FindByEmail(context.Background(), email)
		require.NoError(t, err)
		require.Equal(t, createdUser.ID, foundUser.ID)
		require.Equal(t, createdUser.Email, foundUser.Email)
		truncateUsers(t)
	})

	t.Run("handles user not found", func(t *testing.T) {
		repo := NewUserRepository(testDB)
		_, err := repo.FindByEmail(context.Background(), "nonexistent@example.com")
		require.Error(t, err)
		truncateUsers(t)
	})

}

func TestUserRepository_FindByID(t *testing.T) {
	t.Run("finds user successfully", func(t *testing.T) {
		repo := NewUserRepository(testDB)
		email := "test@example.com"
		password := "hashedpassword"
		createdUser, err := repo.Create(context.Background(), email, password)
		require.NoError(t, err)

		foundUser, err := repo.FindByID(context.Background(), createdUser.ID)
		require.NoError(t, err)
		require.Equal(t, createdUser.ID, foundUser.ID)
		require.Equal(t, createdUser.Email, foundUser.Email)
		truncateUsers(t)
	})

	t.Run("handles user not found", func(t *testing.T) {
		repo := NewUserRepository(testDB)
		userID := uuid.New()
		_, err := repo.FindByID(context.Background(), &userID)
		require.Error(t, err)
		truncateUsers(t)
	})
}

func truncateUsers(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}
