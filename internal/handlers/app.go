package handlers

import (
	"log"

	"github.com/alexanderramin/kalistheniks/internal/config"
	"github.com/alexanderramin/kalistheniks/internal/handlers/contracts"
)

// App wires shared dependencies for HTTP handlers.
type App struct {
	AuthService    contracts.AuthService
	SessionService contracts.SessionService
	PlanService    contracts.PlanService
	Logger         *log.Logger
	Config         config.Config
}
