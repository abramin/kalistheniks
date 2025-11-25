package features

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/alexanderramin/kalistheniks/internal/config"
	"github.com/alexanderramin/kalistheniks/internal/handlers"
	"github.com/alexanderramin/kalistheniks/internal/repositories"
	"github.com/alexanderramin/kalistheniks/internal/services"
	"github.com/alexanderramin/kalistheniks/internal/services/plan"
	"github.com/alexanderramin/kalistheniks/pkg/test_helpers"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var (
	testDB      *sql.DB
	testCleanup func()
	testServer  *httptest.Server
	jwtSecret   = "test-jwt-secret-key-for-godog-tests"
)

type scenarioState struct {
	db               *sql.DB
	client           *http.Client
	baseURL          string
	lastResponse     *http.Response
	lastResponseBody []byte
	token            string
}

func TestFeatures(t *testing.T) {
	opts := godog.Options{
		Output: colors.Colored(os.Stdout),
		Paths:  []string{"."},
		Strict: true,
		Format: "pretty",
	}

	status := godog.TestSuite{
		Name:                 "bdd",
		ScenarioInitializer:  InitializeScenario,
		TestSuiteInitializer: InitializeTestSuite,
		Options:              &opts,
	}.Run()

	if status != 0 {
		t.Fatalf("godog suite failed with status: %d", status)
	}
}

// InitializeTestSuite sets up the database and HTTP server once for all scenarios.
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		// Start PostgreSQL container
		db, cleanup, err := test_helpers.StartPostgresContainer(context.Background())
		if err != nil {
			panic(fmt.Sprintf("failed to start postgres container: %v", err))
		}
		testDB = db
		testCleanup = cleanup

		// Create application dependencies
		userRepo := repositories.NewUserRepository(testDB)
		sessionRepo := repositories.NewSessionRepository(testDB)
		authService := services.NewAuthService(userRepo, jwtSecret)
		sessionService := services.NewSessionService(sessionRepo)
		planService := plan.NewPlanService(sessionRepo)

		cfg := config.Config{
			Addr:      ":8080",
			DBDSN:     "",
			JWTSecret: jwtSecret,
			Env:       "test",
		}

		app := &handlers.App{
			AuthService:    authService,
			SessionService: sessionService,
			PlanService:    planService,
			Logger:         log.New(os.Stdout, "test ", log.LstdFlags),
			Config:         cfg,
		}

		// Create test HTTP server
		router := handlers.Router(app)
		testServer = httptest.NewServer(router)
	})

	ctx.AfterSuite(func() {
		if testServer != nil {
			testServer.Close()
		}
		if testCleanup != nil {
			testCleanup()
		}
	})
}

// InitializeScenario registers step definitions for all feature files.
func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &scenarioState{
		db:     testDB,
		client: http.DefaultClient,
	}

	ctx.Before(func(context.Context, *godog.Scenario) (context.Context, error) {
		state.baseURL = testServer.URL
		state.token = ""
		state.lastResponse = nil
		state.lastResponseBody = nil
		return context.Background(), nil
	})

	// Register step definitions from separate files
	registerAuthSteps(ctx, state)
	registerSessionsSteps(ctx, state)
	registerPlanSteps(ctx, state)
	registerAssertionSteps(ctx, state)
	registerDataSetupSteps(ctx, state)
}

// ========== Helper methods ==========

func (s *scenarioState) doPostRequest(path, body, token string) error {
	req, err := http.NewRequest("POST", s.baseURL+path, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}

	s.lastResponse = resp
	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	s.lastResponseBody = bodyBytes

	return nil
}

func (s *scenarioState) doGetRequest(path, token string) error {
	req, err := http.NewRequest("GET", s.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}

	s.lastResponse = resp
	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	s.lastResponseBody = bodyBytes

	return nil
}

// hasNestedField checks if a nested field exists using dot notation (e.g., "user.id")
func hasNestedField(data map[string]interface{}, field string) bool {
	parts := strings.Split(field, ".")
	current := data

	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return false
		}

		// If this is the last part, we found the field
		if i == len(parts)-1 {
			return true
		}

		// Otherwise, traverse deeper
		nested, ok := value.(map[string]interface{})
		if !ok {
			return false
		}
		current = nested
	}

	return false
}
