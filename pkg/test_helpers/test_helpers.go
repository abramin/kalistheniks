package test_helpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alexanderramin/kalistheniks/pkg/db"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func WaitForDB(ctx context.Context, database *sql.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if err := db.Ping(ctx, database); err == nil {
			return nil
		} else if time.Now().After(deadline) {
			return err
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func RunMigrations(ctx context.Context, database *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if _, err := database.ExecContext(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("executing %s: %w", path, err)
		}
	}

	return nil
}

var migrationsDir = filepath.Join("..", "..", "migrations")

func StartPostgresContainer(ctx context.Context) (*sql.DB, func(), error) {
	pg, err := postgres.Run(ctx, "postgres:16-alpine")
	if err != nil {
		return nil, nil, err
	}

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pg.Terminate(ctx)
		return nil, nil, err
	}

	database, err := db.New(dsn)
	if err != nil {
		_ = pg.Terminate(ctx)
		return nil, nil, err
	}

	if err := WaitForDB(ctx, database, 10*time.Second); err != nil {
		_ = database.Close()
		_ = pg.Terminate(ctx)
		return nil, nil, err
	}

	if err := RunMigrations(ctx, database, migrationsDir); err != nil {
		_ = database.Close()
		_ = pg.Terminate(ctx)
		return nil, nil, err
	}

	cleanup := func() {
		_ = database.Close()
		_ = pg.Terminate(context.Background())
	}

	return database, cleanup, nil
}
