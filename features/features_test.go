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
	pendingHeaders   map[string]string
	createdSessionID string
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
		state.pendingHeaders = make(map[string]string)
		state.createdSessionID = ""
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
	ctx.Step(`^body:$`, state.andBody)

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
	ctx.Step(`^I have created a session with id "([^"]*)"$`, state.iHaveCreatedASessionWithID)
	ctx.Step(`^I have no recorded sessions or sets$`, state.iHaveNoRecordedSessionsOrSets)
	ctx.Step(`^the response JSON should include:$`, state.theResponseJSONShouldIncludeTable)
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
	if err := s.doPostRequest("/login", body.Content, ""); err != nil {
		return err
	}

	// Capture token for later requests if present.
	var resp map[string]interface{}
	if err := json.Unmarshal(s.lastResponseBody, &resp); err == nil {
		if token, ok := resp["token"].(string); ok && token != "" {
			s.token = token
		}
	}

	return nil
}

func (s *scenarioState) iPostSessionsWithHeaders(table *godog.Table) error {
	// Store headers for the next request
	s.pendingHeaders = make(map[string]string)
	for i := 1; i < len(table.Rows); i++ {
		key := table.Rows[i].Cells[0].Value
		value := table.Rows[i].Cells[1].Value
		// Replace <token> placeholder with actual token
		if strings.Contains(value, "<token>") {
			value = strings.Replace(value, "<token>", s.token, -1)
		}
		s.pendingHeaders[key] = value
	}
	return nil
}

func (s *scenarioState) iPostSessionsWithoutAuthHeader() error {
	return s.doPostRequest("/sessions", "", "")
}

func (s *scenarioState) iPostSessionsWithBody(body *godog.DocString) error {
	// Extract token from pending headers
	token := ""
	if authHeader, ok := s.pendingHeaders["Authorization"]; ok {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}

	if err := s.doPostRequest("/sessions", body.Content, token); err != nil {
		return err
	}

	// If successful, try to capture the session ID from response
	if s.lastResponse != nil && s.lastResponse.StatusCode == http.StatusCreated {
		var result map[string]interface{}
		if err := json.Unmarshal(s.lastResponseBody, &result); err == nil {
			if id, ok := result["id"].(string); ok {
				s.createdSessionID = id
			}
		}
	}

	return nil
}

func (s *scenarioState) iPostSessionSetsWithHeaders(sessionID string, table *godog.Table) error {
	// Store headers for the next request
	s.pendingHeaders = make(map[string]string)
	for i := 1; i < len(table.Rows); i++ {
		key := table.Rows[i].Cells[0].Value
		value := table.Rows[i].Cells[1].Value
		// Replace <token> placeholder with actual token
		if strings.Contains(value, "<token>") {
			value = strings.Replace(value, "<token>", s.token, -1)
		}
		s.pendingHeaders[key] = value
	}

	// Replace <session_id> with actual session ID
	if sessionID == "<session_id>" {
		sessionID = s.createdSessionID
	}

	// Extract token from pending headers
	token := ""
	if authHeader, ok := s.pendingHeaders["Authorization"]; ok {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Wait for body to be provided by next step - store the path for later
	s.pendingHeaders["_path"] = "/sessions/" + sessionID + "/sets"
	s.pendingHeaders["_token"] = token
	return nil
}

func (s *scenarioState) iPostInvalidSessionSetsWithHeaders(table *godog.Table) error {
	// Store headers for the next request
	s.pendingHeaders = make(map[string]string)
	for i := 1; i < len(table.Rows); i++ {
		key := table.Rows[i].Cells[0].Value
		value := table.Rows[i].Cells[1].Value
		// Replace <token> placeholder with actual token
		if strings.Contains(value, "<token>") {
			value = strings.Replace(value, "<token>", s.token, -1)
		}
		s.pendingHeaders[key] = value
	}

	// Extract token from pending headers
	token := ""
	if authHeader, ok := s.pendingHeaders["Authorization"]; ok {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Store the path for invalid session ID
	s.pendingHeaders["_path"] = "/sessions/invalid-session-id/sets"
	s.pendingHeaders["_token"] = token
	return nil
}

func (s *scenarioState) iGetSessionsWithHeaders(table *godog.Table) error {
	// Extract token from table
	token := ""
	for i := 1; i < len(table.Rows); i++ {
		if table.Rows[i].Cells[0].Value == "Authorization" {
			authValue := table.Rows[i].Cells[1].Value
			if strings.HasPrefix(authValue, "Bearer ") {
				token = strings.TrimPrefix(authValue, "Bearer ")
				if token == "<token>" {
					token = s.token
				}
			} else if authValue == "Bearer invalid.token" {
				token = "invalid.token"
			}
		}
	}

	return s.doGetRequest("/sessions", token)
}

func (s *scenarioState) andBody(body *godog.DocString) error {
	// Check if we have a pending path from previous step
	if path, ok := s.pendingHeaders["_path"]; ok {
		token := s.pendingHeaders["_token"]
		return s.doPostRequest(path, body.Content, token)
	}
	return fmt.Errorf("no pending request to attach body to")
}

func (s *scenarioState) iGetPlanNextWithHeaders(table *godog.Table) error {
	// Extract token from table
	token := ""
	for i := 1; i < len(table.Rows); i++ {
		if table.Rows[i].Cells[0].Value == "Authorization" {
			authValue := table.Rows[i].Cells[1].Value
			if strings.HasPrefix(authValue, "Bearer ") {
				token = strings.TrimPrefix(authValue, "Bearer ")
				if token == "<token>" {
					token = s.token
				}
			}
		}
	}

	return s.doGetRequest("/plan/next", token)
}

func (s *scenarioState) iGetPlanNextWithoutAuthHeader() error {
	return s.doGetRequest("/plan/next", "")
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
	var result map[string]interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check each expected value from the table
	for i := 1; i < len(table.Rows); i++ {
		field := table.Rows[i].Cells[0].Value
		expectedValue := table.Rows[i].Cells[1].Value

		actualValue, exists := result[field]
		if !exists {
			return fmt.Errorf("field %q not found in response", field)
		}

		// Handle "contains" expectations
		if strings.HasPrefix(expectedValue, "contains ") {
			expectedSubstring := strings.Trim(strings.TrimPrefix(expectedValue, "contains "), `"`)
			actualStr := fmt.Sprintf("%v", actualValue)
			if !strings.Contains(strings.ToLower(actualStr), strings.ToLower(expectedSubstring)) {
				return fmt.Errorf("field %q value %q does not contain %q", field, actualStr, expectedSubstring)
			}
			continue
		}

		// Direct value comparison - handle numeric types
		actualStr := fmt.Sprintf("%v", actualValue)
		if actualStr != expectedValue {
			return fmt.Errorf("field %q has value %q, expected %q", field, actualStr, expectedValue)
		}
	}

	return nil
}

func (s *scenarioState) theResponseJSONShouldIncludeList(table *godog.Table) error {
	var result interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check each expectation from the table
	for i := 1; i < len(table.Rows); i++ {
		field := table.Rows[i].Cells[0].Value
		expectation := table.Rows[i].Cells[1].Value

		// Parse the expectation (e.g., "equals 2", "equals \"deadlift-uuid\"")
		parts := strings.SplitN(expectation, " ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid expectation format: %q", expectation)
		}

		operator := parts[0]
		expectedValue := strings.Trim(parts[1], `"`)

		// Navigate to the field using path notation (e.g., "[0].id", "[0].sets.length")
		actualValue, err := navigateJSONPath(result, field)
		if err != nil {
			return fmt.Errorf("failed to navigate to field %q: %w", field, err)
		}

		// Apply the operator
		switch operator {
		case "equals":
			actualStr := fmt.Sprintf("%v", actualValue)
			if actualStr != expectedValue {
				return fmt.Errorf("field %q has value %q, expected %q", field, actualStr, expectedValue)
			}
		default:
			return fmt.Errorf("unsupported operator: %q", operator)
		}
	}

	return nil
}

// ========== Data setup steps ==========

func (s *scenarioState) myLastRecordedSetHas(table *godog.Table) error {
	ctx := context.Background()

	// Parse table data into a map
	data := make(map[string]string)
	for i := 1; i < len(table.Rows); i++ {
		key := table.Rows[i].Cells[0].Value
		value := table.Rows[i].Cells[1].Value
		data[key] = value
	}

	// Get user ID for the current token
	// First, get the user from the database
	var userID string
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", "user@example.com").Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	// Create a session
	sessionType := data["session_type"]
	if sessionType == "" {
		sessionType = "lower" // default
	}

	var sessionID string
	err = s.db.QueryRowContext(ctx,
		"INSERT INTO sessions (user_id, performed_at, session_type, notes) VALUES ($1, NOW(), $2, 'test') RETURNING id",
		userID, sessionType,
	).Scan(&sessionID)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Insert the set
	exerciseID := data["exercise_id"]
	reps := data["reps"]
	weightKg := data["weight_kg"]

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO sets (session_id, exercise_id, set_index, reps, weight_kg, rpe) VALUES ($1, $2, 1, $3, $4, 8)",
		sessionID, exerciseID, reps, weightKg,
	)
	if err != nil {
		return fmt.Errorf("failed to insert set: %w", err)
	}

	return nil
}

func (s *scenarioState) iHaveAddedTwoSetsToSession(sessionID string) error {
	ctx := context.Background()

	// Replace <session_id> with actual session ID
	if sessionID == "<session_id>" {
		sessionID = s.createdSessionID
	}

	// Insert two sets
	sets := []struct {
		exerciseID string
		setIndex   int
		reps       int
		weightKg   float64
	}{
		{"deadlift-uuid", 1, 8, 100.0},
		{"deadlift-uuid", 2, 7, 100.0},
	}

	for _, set := range sets {
		_, err := s.db.ExecContext(ctx,
			"INSERT INTO sets (session_id, exercise_id, set_index, reps, weight_kg, rpe) VALUES ($1, $2, $3, $4, $5, 8)",
			sessionID, set.exerciseID, set.setIndex, set.reps, set.weightKg,
		)
		if err != nil {
			return fmt.Errorf("failed to insert set: %w", err)
		}
	}

	return nil
}

func (s *scenarioState) iHaveCreatedASessionWithID(sessionID string) error {
	ctx := context.Background()

	// Get user ID for the current token
	var userID string
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", "user@example.com").Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	// Create a session
	var createdSessionID string
	err = s.db.QueryRowContext(ctx,
		"INSERT INTO sessions (user_id, performed_at, session_type, notes) VALUES ($1, NOW(), 'upper', 'test session') RETURNING id",
		userID,
	).Scan(&createdSessionID)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Store the session ID
	s.createdSessionID = createdSessionID

	return nil
}

func (s *scenarioState) iHaveNoRecordedSessionsOrSets() error {
	// This is essentially the same as having an empty database for the current user
	// The database is already cleaned before each scenario, and the user is created fresh
	// So we don't need to do anything here
	return nil
}

func (s *scenarioState) theResponseJSONShouldIncludeTable(table *godog.Table) error {
	var result map[string]interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check each expected field from the table
	for i := 1; i < len(table.Rows); i++ {
		field := table.Rows[i].Cells[0].Value
		expectedValue := table.Rows[i].Cells[1].Value

		actualValue, exists := result[field]
		if !exists {
			return fmt.Errorf("field %q not found in response", field)
		}

		// Handle "contains" expectations
		if strings.HasPrefix(expectedValue, "contains ") {
			expectedSubstring := strings.Trim(strings.TrimPrefix(expectedValue, "contains "), `"`)
			actualStr := fmt.Sprintf("%v", actualValue)
			if !strings.Contains(strings.ToLower(actualStr), strings.ToLower(expectedSubstring)) {
				return fmt.Errorf("field %q value %q does not contain %q", field, actualStr, expectedSubstring)
			}
			continue
		}

		// Handle "or less" expectations
		if strings.HasSuffix(expectedValue, " or less") {
			expectedMax := strings.TrimSuffix(expectedValue, " or less")
			var maxVal, actualVal float64
			fmt.Sscanf(expectedMax, "%f", &maxVal)

			switch v := actualValue.(type) {
			case float64:
				actualVal = v
			case int:
				actualVal = float64(v)
			default:
				fmt.Sscanf(fmt.Sprintf("%v", actualValue), "%f", &actualVal)
			}

			if actualVal > maxVal {
				return fmt.Errorf("field %q has value %v, expected %v or less", field, actualVal, maxVal)
			}
			continue
		}

		// Direct value comparison
		actualStr := fmt.Sprintf("%v", actualValue)
		if actualStr != expectedValue {
			return fmt.Errorf("field %q has value %q, expected %q", field, actualStr, expectedValue)
		}
	}

	return nil
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

// navigateJSONPath navigates to a field in a JSON structure using path notation
// Supports: "[0].id", "[0].sets.length", "[0].sets[0].exercise_id"
func navigateJSONPath(data interface{}, path string) (interface{}, error) {
	current := data

	// Parse path into parts
	parts := []string{}
	currentPart := ""
	inBracket := false

	for _, char := range path {
		switch char {
		case '[':
			if currentPart != "" {
				parts = append(parts, currentPart)
				currentPart = ""
			}
			inBracket = true
			currentPart += string(char)
		case ']':
			currentPart += string(char)
			inBracket = false
			parts = append(parts, currentPart)
			currentPart = ""
		case '.':
			if inBracket {
				currentPart += string(char)
			} else {
				if currentPart != "" {
					parts = append(parts, currentPart)
					currentPart = ""
				}
			}
		default:
			currentPart += string(char)
		}
	}
	if currentPart != "" {
		parts = append(parts, currentPart)
	}

	// Navigate through parts
	for _, part := range parts {
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			// Array index
			indexStr := strings.TrimPrefix(strings.TrimSuffix(part, "]"), "[")
			index := 0
			fmt.Sscanf(indexStr, "%d", &index)

			arr, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array at path part %q, got %T", part, current)
			}
			if index >= len(arr) {
				return nil, fmt.Errorf("index %d out of range (length %d)", index, len(arr))
			}
			current = arr[index]
		} else if part == "length" {
			// Array or map length
			switch v := current.(type) {
			case []interface{}:
				return len(v), nil
			case map[string]interface{}:
				return len(v), nil
			default:
				return nil, fmt.Errorf("cannot get length of %T", current)
			}
		} else {
			// Object field
			obj, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected object at path part %q, got %T", part, current)
			}
			value, exists := obj[part]
			if !exists {
				return nil, fmt.Errorf("field %q not found", part)
			}
			current = value
		}
	}

	return current, nil
}
