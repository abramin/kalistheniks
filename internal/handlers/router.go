package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router registers HTTP routes for the API.
func Router(app *App) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", app.health)
	r.Post("/signup", app.signup)
	r.Post("/login", app.login)

	r.Group(func(protected chi.Router) {
		protected.Use(app.authMiddleware)
		protected.Get("/sessions", app.listSessions)
		protected.Post("/sessions", app.createSession)
		protected.Post("/sessions/{id}/sets", app.createSet)
		protected.Get("/plan/next", app.nextPlan)
	})

	return r
}
