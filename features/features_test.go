package features

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
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
	"golang.org/x/crypto/bcrypt"
)

var (
	testDB        *sql.DB
	testCleanup   func()
	testServer    *httptest.Server
	jwtSecret     = "test-jwt-secret-key-for-godog-tests"
	migrationsDir = "../migrations"
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
		Paths:  []string{"features"},
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

	// Auth-related steps
	ctx.Step(`^the database is empty$`, state.theDatabaseIsEmpty)
	ctx.Step(`^a user already exists with email "([^"]*)"$`, state.aUserAlreadyExistsWithEmail)
	ctx.Step(`^a user exists with email "([^"]*)" and password "([^"]*)"$`, state.aUserExistsWithEmailAndPassword)
	ctx.Step(`^I have a valid token from logging in as "([^"]*)"$`, state.iHaveAValidTokenFromLoggingInAs)
	ctx.Step(`^I POST /signup with body:$`, state.iPostSignupWithBody)
	ctx.Step(`^I POST /login with body:$`, state.iPostLoginWithBody)

	// Sessions-related steps
	ctx.Step(`^I POST /sessions with headers:$`, state.iPostSessionsWithHeaders)
	ctx.Step(`^I POST /sessions without an Authorization header$`, state.iPostSessionsWithoutAuthHeader)
	ctx.Step(`^I POST /sessions with body:$`, state.iPostSessionsWithBody)
	ctx.Step(`^I POST /sessions/([^/]+)/sets with headers:$`, state.iPostSessionSetsWithHeaders)
	ctx.Step(`^I POST /sessions/invalid-session-id/sets with headers:$`, state.iPostInvalidSessionSetsWithHeaders)
	ctx.Step(`^I GET /sessions with headers:$`, state.iGetSessionsWithHeaders)

	// Plan-related steps
	ctx.Step(`^I GET /plan/next with headers:$`, state.iGetPlanNextWithHeaders)
	ctx.Step(`^I GET /plan/next without an Authorization header$`, state.iGetPlanNextWithoutAuthHeader)

	// Assertion steps
	ctx.Step(`^the response status should be (\d+)$`, state.theResponseStatusShouldBe)
	ctx.Step(`^the response JSON should include "([^"]*)" and "([^"]*)"$`, state.theResponseJSONShouldIncludeFields)
	ctx.Step(`^the response JSON should include a non-empty "([^"]*)"$`, state.theResponseJSONShouldIncludeNonEmptyField)
	ctx.Step(`^the response JSON should include an "([^"]*)" explaining the email is taken$`, state.theResponseJSONShouldIncludeErrorAboutEmailTaken)
	ctx.Step(`^the response JSON should include "([^"]*)"$`, state.theResponseJSONShouldIncludeField)
	ctx.Step(`^the response JSON should include an "([^"]*)" about invalid request body$`, state.theResponseJSONShouldIncludeErrorAboutInvalidBody)
	ctx.Step(`^the response JSON should include default values:$`, state.theResponseJSONShouldIncludeDefaultValues)
	ctx.Step(`^the response JSON should include a list where:$`, state.theResponseJSONShouldIncludeList)

	// Data setup steps
	ctx.Step(`^my last recorded set has:$`, state.myLastRecordedSetHas)
	ctx.Step(`^I have added two sets to session "([^"]*)"$`, state.iHaveAddedTwoSetsToSession)
}

// ========== Database setup steps ==========

func (s *scenarioState) theDatabaseIsEmpty() error {
	ctx := context.Background()

	// Delete all data from tables
	tables := []string{"sets", "sessions", "users"}
	for _, table := range tables {
		if _, err := s.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			return fmt.Errorf("failed to clear table %s: %w", table, err)
		}
	}

	return nil
}

func (s *scenarioState) aUserAlreadyExistsWithEmail(email string) error {
	return s.aUserExistsWithEmailAndPassword(email, "SomePassword!1")
}

func (s *scenarioState) aUserExistsWithEmailAndPassword(email, password string) error {
	ctx := context.Background()

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert user into database
	const q = `INSERT INTO users (email, password_hash) VALUES ($1, $2)`
	if _, err := s.db.ExecContext(ctx, q, email, string(hash)); err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (s *scenarioState) iHaveAValidTokenFromLoggingInAs(email string) error {
	// First ensure user exists
	if err := s.aUserExistsWithEmailAndPassword(email, "TestPassword!1"); err != nil {
		return err
	}

	// Login to get token
	loginBody := map[string]string{
		"email":    email,
		"password": "TestPassword!1",
	}
	bodyBytes, _ := json.Marshal(loginBody)

	req, err := http.NewRequest("POST", s.baseURL+"/login", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	token, ok := result["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("no token in login response")
	}

	s.token = token
	return nil
}

// ========== HTTP request steps ==========

func (s *scenarioState) iPostSignupWithBody(body *godog.DocString) error {
	return s.doPostRequest("/signup", body.Content, "")
}

func (s *scenarioState) iPostLoginWithBody(body *godog.DocString) error {
	return s.doPostRequest("/login", body.Content, "")
}

func (s *scenarioState) iPostSessionsWithHeaders(table *godog.Table) error {
	return godog.ErrPending
}

func (s *scenarioState) iPostSessionsWithoutAuthHeader() error {
	return godog.ErrPending
}

func (s *scenarioState) iPostSessionsWithBody(body *godog.DocString) error {
	return godog.ErrPending
}

func (s *scenarioState) iPostSessionSetsWithHeaders(sessionID string, table *godog.Table) error {
	return godog.ErrPending
}

func (s *scenarioState) iPostInvalidSessionSetsWithHeaders(table *godog.Table) error {
	return godog.ErrPending
}

func (s *scenarioState) iGetSessionsWithHeaders(table *godog.Table) error {
	return godog.ErrPending
}

func (s *scenarioState) iGetPlanNextWithHeaders(table *godog.Table) error {
	return godog.ErrPending
}

func (s *scenarioState) iGetPlanNextWithoutAuthHeader() error {
	return godog.ErrPending
}

// ========== Assertion steps ==========

func (s *scenarioState) theResponseStatusShouldBe(expectedStatus int) error {
	if s.lastResponse == nil {
		return fmt.Errorf("no response received")
	}

	if s.lastResponse.StatusCode != expectedStatus {
		return fmt.Errorf("expected status %d, got %d. Response body: %s",
			expectedStatus, s.lastResponse.StatusCode, string(s.lastResponseBody))
	}

	return nil
}

func (s *scenarioState) theResponseJSONShouldIncludeFields(field1, field2 string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Handle nested fields using dot notation
	if !hasNestedField(result, field1) {
		return fmt.Errorf("field %q not found in response: %s", field1, string(s.lastResponseBody))
	}

	if !hasNestedField(result, field2) {
		return fmt.Errorf("field %q not found in response: %s", field2, string(s.lastResponseBody))
	}

	return nil
}

func (s *scenarioState) theResponseJSONShouldIncludeNonEmptyField(field string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	value, exists := result[field]
	if !exists {
		return fmt.Errorf("field %q not found in response", field)
	}

	// Check if value is non-empty
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("field %q is not a string", field)
	}

	if strValue == "" {
		return fmt.Errorf("field %q is empty", field)
	}

	return nil
}

func (s *scenarioState) theResponseJSONShouldIncludeErrorAboutEmailTaken(field string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	errorMsg, exists := result[field]
	if !exists {
		return fmt.Errorf("field %q not found in response", field)
	}

	errorStr, ok := errorMsg.(string)
	if !ok {
		return fmt.Errorf("field %q is not a string", field)
	}

	// Check if error message mentions email or taken/exists
	lower := strings.ToLower(errorStr)
	if !strings.Contains(lower, "email") && !strings.Contains(lower, "exists") && !strings.Contains(lower, "taken") {
		return fmt.Errorf("error message does not mention email being taken: %q", errorStr)
	}

	return nil
}

func (s *scenarioState) theResponseJSONShouldIncludeField(expectedJSON string) error {
	// Parse expected JSON (e.g., "error":"invalid credentials")
	parts := strings.SplitN(expectedJSON, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid expected JSON format: %q", expectedJSON)
	}

	field := strings.Trim(parts[0], `"`)
	expectedValue := strings.Trim(parts[1], `"`)

	var result map[string]interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	actualValue, exists := result[field]
	if !exists {
		return fmt.Errorf("field %q not found in response", field)
	}

	actualStr := fmt.Sprintf("%v", actualValue)
	if actualStr != expectedValue {
		return fmt.Errorf("field %q has value %q, expected %q", field, actualStr, expectedValue)
	}

	return nil
}

func (s *scenarioState) theResponseJSONShouldIncludeErrorAboutInvalidBody(field string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	errorMsg, exists := result[field]
	if !exists {
		return fmt.Errorf("field %q not found in response", field)
	}

	errorStr, ok := errorMsg.(string)
	if !ok {
		return fmt.Errorf("field %q is not a string", field)
	}

	// Check if error message mentions invalid, validation, or required
	lower := strings.ToLower(errorStr)
	if !strings.Contains(lower, "invalid") && !strings.Contains(lower, "validation") &&
		!strings.Contains(lower, "required") && !strings.Contains(lower, "missing") {
		return fmt.Errorf("error message does not mention validation error: %q", errorStr)
	}

	return nil
}

func (s *scenarioState) theResponseJSONShouldIncludeDefaultValues(table *godog.Table) error {
	return godog.ErrPending
}

func (s *scenarioState) theResponseJSONShouldIncludeList(table *godog.Table) error {
	return godog.ErrPending
}

// ========== Data setup steps ==========

func (s *scenarioState) myLastRecordedSetHas(table *godog.Table) error {
	return godog.ErrPending
}

func (s *scenarioState) iHaveAddedTwoSetsToSession(sessionID string) error {
	return godog.ErrPending
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
