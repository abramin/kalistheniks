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

	"github.com/alexanderramin/kalistheniks/internal/db"
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
