package handlers

import (
	"net/http"
	"time"

	apiHandlers "github.com/alexanderramin/kalistheniks/internal/handlers/api"
	authHandlers "github.com/alexanderramin/kalistheniks/internal/handlers/auth"
	handlerMiddleware "github.com/alexanderramin/kalistheniks/internal/handlers/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/unrolled/secure"
)

// Router registers HTTP routes for the API.
func Router(app *App) http.Handler {
	r := chi.NewRouter()

<<<<<<< HEAD
	// Security middleware
	r.Use(handlerMiddleware.SecurityHeaders)
	r.Use(handlerMiddleware.CORS)

	// Chi built-in middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
=======
	// Chi built-in middleware
	r.Use(middleware.RealIP)                   // Extract real IP from proxies
	r.Use(middleware.RequestID)                 // Generate unique request IDs
	r.Use(middleware.Logger)                    // Log HTTP requests
	r.Use(middleware.Recoverer)                 // Recover from panics
	r.Use(middleware.Timeout(60 * time.Second)) // Request timeout protection
	r.Use(middleware.Compress(5))               // gzip compression for responses

	// Security headers middleware
	secureMiddleware := secure.New(secure.Options{
		FrameDeny:             true,                          // Prevent clickjacking (X-Frame-Options: DENY)
		ContentTypeNosniff:    true,                          // Prevent MIME type sniffing
		BrowserXssFilter:      true,                          // Enable XSS filter
		ContentSecurityPolicy: "default-src 'self'",          // CSP policy
		ReferrerPolicy:        "strict-origin-when-cross-origin", // Referrer policy
		IsDevelopment:         app.Config.Env == "development", // Disable in dev for easier debugging
	})
	r.Use(secureMiddleware.Handler)

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"}, // Adjust for your frontend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Cache preflight requests for 5 minutes
	}))

	// Global rate limiting: 100 requests per minute per IP
	r.Use(httprate.LimitByIP(100, 1*time.Minute))
>>>>>>> 7948578 (Add production-ready middleware using ecosystem packages)

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
