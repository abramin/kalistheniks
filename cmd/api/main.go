package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexanderramin/kalistheniks/internal/config"
	"github.com/alexanderramin/kalistheniks/internal/handlers"
	"github.com/alexanderramin/kalistheniks/internal/repositories"
	"github.com/alexanderramin/kalistheniks/internal/services"
	"github.com/alexanderramin/kalistheniks/internal/services/plan"
	"github.com/alexanderramin/kalistheniks/pkg/db"
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
	planService := plan.NewPlanService(sessionRepo)

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
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Start server in a goroutine
	go func() {
		logger.Printf("starting server on %s", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	logger.Println("shutting down server gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("server shutdown error: %v", err)
	}

	// Close database connection
	if err := database.Close(); err != nil {
		logger.Printf("database close error: %v", err)
	}

	logger.Println("server stopped")
}
