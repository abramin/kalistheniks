package handlers

import (
	"encoding/json"
	"net/http"
)

func (a *App) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *App) signup(w http.ResponseWriter, _ *http.Request) {
	a.notImplemented(w)
}

func (a *App) login(w http.ResponseWriter, _ *http.Request) {
	a.notImplemented(w)
}

func (a *App) sessions(w http.ResponseWriter, _ *http.Request) {
	a.notImplemented(w)
}

func (a *App) sessionSets(w http.ResponseWriter, _ *http.Request) {
	// TODO: parse session ID from URL when implementing logic.
	a.notImplemented(w)
}

func (a *App) nextPlan(w http.ResponseWriter, _ *http.Request) {
	a.notImplemented(w)
}

func (a *App) notImplemented(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "not implemented"})
}
