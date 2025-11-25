package features

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
)

// registerSessionsSteps registers session-related step definitions.
func registerSessionsSteps(ctx *godog.ScenarioContext, state *scenarioState) {
	ctx.Step(`^I have created a session with id "([^"]*)"$`, state.iHaveCreatedASessionWithID)
	ctx.Step(`^I POST /sessions with headers:$`, state.iPostSessionsWithHeaders)
	ctx.Step(`^I POST /sessions without an Authorization header$`, state.iPostSessionsWithoutAuthHeader)
	ctx.Step(`^And body:$`, state.andBody)
	ctx.Step(`^I POST /sessions/([^/]+)/sets with headers:$`, state.iPostSessionSetsWithHeaders)
	ctx.Step(`^I POST /sessions/invalid-session-id/sets with headers:$`, state.iPostInvalidSessionSetsWithHeaders)
	ctx.Step(`^I GET /sessions with headers:$`, state.iGetSessionsWithHeaders)
}

// ========== Sessions HTTP request steps ==========

// iHaveCreatedASessionWithID creates a session and stores its ID in state
func (s *scenarioState) iHaveCreatedASessionWithID(placeholderID string) error {
	// Create a session via API
	sessionBody := map[string]interface{}{
		"performed_at":  "2024-01-01T10:00:00Z",
		"session_type":  "upper",
		"notes":         "Test session",
	}
	bodyBytes, _ := json.Marshal(sessionBody)

	req, err := http.NewRequest("POST", s.baseURL+"/sessions", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create session request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create session: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode session response: %w", err)
	}

	sessionID, ok := result["id"].(string)
	if !ok || sessionID == "" {
		return fmt.Errorf("no session id in response")
	}

	s.sessionID = sessionID
	return nil
}

// iPostSessionsWithHeaders handles POST /sessions with headers table
func (s *scenarioState) iPostSessionsWithHeaders(table *godog.Table) error {
	// Extract headers from table
	headers := s.extractHeaders(table)

	// Store last request method for later body attachment
	s.lastRequestMethod = "POST"
	s.lastRequestPath = "/sessions"
	s.lastRequestHeaders = headers

	return nil
}

func (s *scenarioState) iPostSessionsWithoutAuthHeader() error {
	return s.doPostRequest("/sessions", "{}", "")
}

// andBody attaches body to the last request and executes it
func (s *scenarioState) andBody(body *godog.DocString) error {
	bodyContent := s.replacePlaceholders(body.Content)

	// Execute the last request with this body
	if s.lastRequestMethod == "POST" {
		return s.doPostRequestWithHeaders(s.lastRequestPath, bodyContent, s.lastRequestHeaders)
	} else if s.lastRequestMethod == "GET" {
		return s.doGetRequestWithHeaders(s.lastRequestPath, s.lastRequestHeaders)
	}

	return fmt.Errorf("no pending request to attach body to")
}

func (s *scenarioState) iPostSessionSetsWithHeaders(sessionID string, table *godog.Table) error {
	// Replace placeholder sessionID
	sessionID = s.replacePlaceholders(sessionID)

	headers := s.extractHeaders(table)

	// Store last request for later body attachment
	s.lastRequestMethod = "POST"
	s.lastRequestPath = "/sessions/" + sessionID + "/sets"
	s.lastRequestHeaders = headers

	return nil
}

func (s *scenarioState) iPostInvalidSessionSetsWithHeaders(table *godog.Table) error {
	headers := s.extractHeaders(table)

	// Store last request for later body attachment
	s.lastRequestMethod = "POST"
	s.lastRequestPath = "/sessions/invalid-session-id/sets"
	s.lastRequestHeaders = headers

	return nil
}

func (s *scenarioState) iGetSessionsWithHeaders(table *godog.Table) error {
	headers := s.extractHeaders(table)
	return s.doGetRequestWithHeaders("/sessions", headers)
}

// extractHeaders parses a godog table and returns a map of headers
func (s *scenarioState) extractHeaders(table *godog.Table) map[string]string {
	headers := make(map[string]string)

	for i, row := range table.Rows {
		if i == 0 {
			continue // Skip header row
		}
		if len(row.Cells) >= 2 {
			key := row.Cells[0].Value
			value := s.replacePlaceholders(row.Cells[1].Value)
			headers[key] = value
		}
	}

	return headers
}

// doPostRequestWithHeaders performs a POST request with custom headers
func (s *scenarioState) doPostRequestWithHeaders(path, body string, headers map[string]string) error {
	req, err := http.NewRequest("POST", s.baseURL+path, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
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

// doGetRequestWithHeaders performs a GET request with custom headers
func (s *scenarioState) doGetRequestWithHeaders(path string, headers map[string]string) error {
	req, err := http.NewRequest("GET", s.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
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
