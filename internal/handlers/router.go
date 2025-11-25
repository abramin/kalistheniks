package handlers

import (
	"net/http"

	apiHandlers "github.com/alexanderramin/kalistheniks/internal/handlers/api"
	authHandlers "github.com/alexanderramin/kalistheniks/internal/handlers/auth"
	handlerMiddleware "github.com/alexanderramin/kalistheniks/internal/handlers/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router registers HTTP routes for the API.
func Router(app *App) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	auth := authHandlers.New(app.AuthService)
	api := apiHandlers.New(app.SessionService, app.PlanService)
	authMw := handlerMiddleware.NewAuth(app.AuthService)

	r.Get("/health", auth.Health)
	r.Post("/signup", auth.Signup)
	r.Post("/login", auth.Login)

	r.Group(func(protected chi.Router) {
		protected.Use(authMw.RequireAuth)
		protected.Get("/sessions", api.ListSessions)
		protected.Post("/sessions", api.CreateSession)
		protected.Post("/sessions/{id}/sets", api.CreateSet)
		protected.Get("/plan/next", api.NextPlan)
	})

	return r
}
