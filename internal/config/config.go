package config

import "os"

// Config holds environment-driven configuration.
type Config struct {
	DBDSN     string
	JWTSecret string
	Addr      string
	Env       string // Environment: development, staging, production
}

// Load reads configuration from environment variables.
// Real validation and defaults will be added in a later stage.
func Load() (Config, error) {
	cfg := Config{
		DBDSN:     getenvOrDefault("DB_DSN", "postgres://kalistheniks:kalistheniks@localhost:5432/kalistheniks?sslmode=disable"),
		JWTSecret: getenvOrDefault("JWT_SECRET", "replace-me"),
		Addr:      getenvOrDefault("ADDR", ":8080"),
		Env:       getenvOrDefault("ENV", "development"),
	}

	// TODO: add validation (e.g., ensure secrets and DSNs are set in production).
	return cfg, nil
}

func getenvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
