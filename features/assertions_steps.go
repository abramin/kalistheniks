package features

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// registerAssertionSteps registers assertion-related step definitions.
func registerAssertionSteps(ctx *godog.ScenarioContext, state *scenarioState) {
	ctx.Step(`^the response status should be (\d+)$`, state.theResponseStatusShouldBe)
	ctx.Step(`^the response JSON should include "([^"]*)" and "([^"]*)"$`, state.theResponseJSONShouldIncludeFields)
	ctx.Step(`^the response JSON should include a non-empty "([^"]*)"$`, state.theResponseJSONShouldIncludeNonEmptyField)
	ctx.Step(`^the response JSON should include an "([^"]*)" explaining the email is taken$`, state.theResponseJSONShouldIncludeErrorAboutEmailTaken)
	ctx.Step(`^the response JSON should include "([^"]*)"$`, state.theResponseJSONShouldIncludeField)
	ctx.Step(`^the response JSON should include an "([^"]*)" about invalid request body$`, state.theResponseJSONShouldIncludeErrorAboutInvalidBody)
	ctx.Step(`^the response JSON should include default values:$`, state.theResponseJSONShouldIncludeDefaultValues)
	ctx.Step(`^the response JSON should include a list where:$`, state.theResponseJSONShouldIncludeList)
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
