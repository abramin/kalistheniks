package repositories

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/alexanderramin/kalistheniks/pkg/test_helpers"
)

var (
	testDB  *sql.DB
	cleanup func()
)

func TestMain(m *testing.M) {
	db, stop, err := test_helpers.StartPostgresContainer(context.Background())
	if err != nil {
		panic(err)
	}
	testDB = db
	cleanup = stop

	code := m.Run()

	if cleanup != nil {
		cleanup()
	}

	os.Exit(code)
}
