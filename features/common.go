package features

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type scenarioState struct {
	db                 *sql.DB
	client             *http.Client
	baseURL            string
	lastResponse       *http.Response
	lastResponseBody   []byte
	token              string
	sessionID          string
	lastRequestMethod  string
	lastRequestPath    string
	lastRequestHeaders map[string]string
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

// replacePlaceholders replaces <token> and <session_id> placeholders with actual values
func (s *scenarioState) replacePlaceholders(text string) string {
	result := text
	result = strings.ReplaceAll(result, "<token>", s.token)
	result = strings.ReplaceAll(result, "<session_id>", s.sessionID)
	return result
}
