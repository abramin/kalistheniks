package handlers

import (
	"log"

	"github.com/alexanderramin/kalistheniks/internal/config"
	"github.com/alexanderramin/kalistheniks/internal/db"
	"github.com/alexanderramin/kalistheniks/internal/rules"
)

// App wires shared dependencies for HTTP handlers.
type App struct {
	DB     *db.DB
	Logger *log.Logger
	Config config.Config
	Rules  *rules.RuleEngine
}
