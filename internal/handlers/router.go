package handlers

import (
	"net/http"
)

// Router registers HTTP routes for the API.
func Router(app *App) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", app.health)
	mux.HandleFunc("/signup", app.signup)
	mux.HandleFunc("/login", app.login)
	mux.HandleFunc("/sessions", app.sessions)
	mux.HandleFunc("/sessions/", app.sessionSets) // expects /sessions/{id}/sets
	mux.HandleFunc("/plan/next", app.nextPlan)

	return mux
}
