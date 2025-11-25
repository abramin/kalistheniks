package features

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cucumber/godog"
	"golang.org/x/crypto/bcrypt"
)

// registerAuthSteps registers authentication-related step definitions.
func registerAuthSteps(ctx *godog.ScenarioContext, state *scenarioState) {
	ctx.Step(`^the database is empty$`, state.theDatabaseIsEmpty)
	ctx.Step(`^a user already exists with email "([^"]*)"$`, state.aUserAlreadyExistsWithEmail)
	ctx.Step(`^a user exists with email "([^"]*)" and password "([^"]*)"$`, state.aUserExistsWithEmailAndPassword)
	ctx.Step(`^I have a valid token from logging in as "([^"]*)"$`, state.iHaveAValidTokenFromLoggingInAs)
	ctx.Step(`^I POST /signup with body:$`, state.iPostSignupWithBody)
	ctx.Step(`^I POST /login with body:$`, state.iPostLoginWithBody)
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
