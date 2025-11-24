package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexanderramin/kalistheniks/internal/config"
	"github.com/alexanderramin/kalistheniks/internal/db"
	"github.com/alexanderramin/kalistheniks/internal/handlers"
	"github.com/alexanderramin/kalistheniks/internal/repositories"
	"github.com/alexanderramin/kalistheniks/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := log.New(os.Stdout, "api ", log.LstdFlags|log.Lshortfile)

	database, err := db.New(cfg.DBDSN)
	if err != nil {
		logger.Fatalf("failed to connect to db: %v", err)
	}
	if err := db.Ping(context.Background(), database); err != nil {
		logger.Fatalf("db ping failed: %v", err)
	}

	userRepo := repositories.NewUserRepository(database)
	sessionRepo := repositories.NewSessionRepository(database)
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	sessionService := services.NewSessionService(sessionRepo)
	planService := services.NewPlanService(sessionRepo)

	app := &handlers.App{
		AuthService:    authService,
		SessionService: sessionService,
		PlanService:    planService,
		Logger:         logger,
		Config:         cfg,
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
