package handlers

import (
	"log"

	"github.com/alexanderramin/kalistheniks/internal/config"
	"github.com/alexanderramin/kalistheniks/internal/services"
)

// App wires shared dependencies for HTTP handlers.
type App struct {
	AuthService    *services.AuthService
	SessionService *services.SessionService
	PlanService    *services.PlanService
	Logger         *log.Logger
	Config         config.Config
}
