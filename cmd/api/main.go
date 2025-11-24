package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexanderramin/kalistheniks/internal/config"
	"github.com/alexanderramin/kalistheniks/internal/db"
	"github.com/alexanderramin/kalistheniks/internal/handlers"
	"github.com/alexanderramin/kalistheniks/internal/rules"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := log.New(os.Stdout, "api ", log.LstdFlags|log.Lshortfile)

	app := &handlers.App{
		DB:     db.New(),
		Logger: logger,
		Config: cfg,
		Rules:  rules.New(),
	}

	router := handlers.Router(app)

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Printf("starting server on %s", cfg.Addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server error: %v", err)
	}
}
