package db

import "context"

// DB is a placeholder for the database connection/handle.
// It will be expanded to manage a real connection pool later.
type DB struct{}

// New creates a placeholder DB handle without connecting anywhere yet.
func New() *DB {
	return &DB{}
}

// Ping is a stub that will eventually check connectivity.
func (db *DB) Ping(_ context.Context) error {
	// TODO: implement health check against the real database.
	return nil
}
