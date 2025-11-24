package handlers

import (
	"log"

	"github.com/alexanderramin/kalistheniks/internal/config"
	"github.com/alexanderramin/kalistheniks/internal/services"
	"github.com/alexanderramin/kalistheniks/internal/services/plan"
)

// App wires shared dependencies for HTTP handlers.
type App struct {
	AuthService    *services.AuthService
	SessionService *services.SessionService
	PlanService    *plan.PlanService
	Logger         *log.Logger
	Config         config.Config
}
