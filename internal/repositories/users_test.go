package repositories

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexanderramin/kalistheniks/internal/db"
	"github.com/alexanderramin/kalistheniks/pkg/test_helpers"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testDB *sql.DB
var migrationsDir = filepath.Join("..", "..", "migrations")

func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, err := postgres.Run(ctx, "postgres:16-alpine")
	if err != nil {
		panic(err)
	}

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	testDB, err = db.New(dsn)
	if err != nil {
		panic(err)
	}

	if err := test_helpers.WaitForDB(ctx, testDB, 10*time.Second); err != nil {
		panic(err)
	}

	if err := test_helpers.RunMigrations(ctx, testDB, migrationsDir); err != nil {
		panic(err)
	}

	code := m.Run()

	_ = testDB.Close()
	_ = pg.Terminate(ctx)

	os.Exit(code)
}

func TestUserRepository_Create(t *testing.T) {
	t.Run("creates user successfully", func(t *testing.T) {
		require := require.New(t)
		repo := NewUserRepository(testDB)
		email := "test@example.com"
		password := "hashedpassword"

		user, err := repo.Create(context.Background(), email, password)
		require.NoError(err)
		require.Equal(email, user.Email)
		require.NotEmpty(user.ID)
		truncateUsers(t)
	})

	t.Run("handles duplicate email", func(t *testing.T) {
		require := require.New(t)
		repo := NewUserRepository(testDB)
		email := "test@example.com"
		password := "hashedpassword"

		_, err := repo.Create(context.Background(), email, password)
		require.NoError(err)

		_, err = repo.Create(context.Background(), email, password)
		require.Error(err)
		truncateUsers(t)
	})

}

func TestUserRepository_FindByEmail(t *testing.T) {
	t.Run("finds user successfully", func(t *testing.T) {
		require := require.New(t)
		repo := NewUserRepository(testDB)
		email := "test@example.com"
		password := "hashedpassword"
		createdUser, err := repo.Create(context.Background(), email, password)
		require.NoError(err)

		foundUser, err := repo.FindByEmail(context.Background(), email)
		require.NoError(err)
		require.Equal(createdUser.ID, foundUser.ID)
		require.Equal(createdUser.Email, foundUser.Email)
		truncateUsers(t)
	})

	t.Run("handles user not found", func(t *testing.T) {
		require := require.New(t)
		repo := NewUserRepository(testDB)
		_, err := repo.FindByEmail(context.Background(), "nonexistent@example.com")
		require.Error(err)
		truncateUsers(t)
	})

}

func TestUserRepository_FindByID(t *testing.T) {
	t.Run("finds user successfully", func(t *testing.T) {
		require := require.New(t)
		repo := NewUserRepository(testDB)
		email := "test@example.com"
		password := "hashedpassword"
		createdUser, err := repo.Create(context.Background(), email, password)
		require.NoError(err)

		foundUser, err := repo.FindByID(context.Background(), createdUser.ID)
		require.NoError(err)
		require.Equal(createdUser.ID, foundUser.ID)
		require.Equal(createdUser.Email, foundUser.Email)
		truncateUsers(t)
	})

	t.Run("handles user not found", func(t *testing.T) {
		require := require.New(t)
		repo := NewUserRepository(testDB)
		_, err := repo.FindByID(context.Background(), "nonexistent-id")
		require.Error(err)
		truncateUsers(t)
	})
}

func truncateUsers(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}
