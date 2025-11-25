package features

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	ctx.Step(`^the response JSON should include:$`, state.theResponseJSONShouldIncludeTable)
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

func (s *scenarioState) theResponseJSONShouldIncludeTable(table *godog.Table) error {
	var result map[string]interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Process each row in the table
	for i, row := range table.Rows {
		if i == 0 {
			continue // Skip header row
		}

		if len(row.Cells) < 2 {
			continue
		}

		field := row.Cells[0].Value
		expectation := row.Cells[1].Value

		// Check if expectation starts with "contains"
		if strings.HasPrefix(expectation, "contains ") {
			// Extract the substring to look for
			substring := strings.TrimPrefix(expectation, "contains ")
			substring = strings.Trim(substring, `"`)

			// Get the actual value
			actualValue, exists := result[field]
			if !exists {
				return fmt.Errorf("field %q not found in response", field)
			}

			actualStr := fmt.Sprintf("%v", actualValue)
			if !strings.Contains(actualStr, substring) {
				return fmt.Errorf("field %q value %q does not contain %q", field, actualStr, substring)
			}
		} else if strings.Contains(expectation, " or ") {
			// Handle "or" conditions (e.g., "5 or less")
			parts := strings.Split(expectation, " or ")
			actualValue, exists := result[field]
			if !exists {
				return fmt.Errorf("field %q not found in response", field)
			}

			actualFloat, ok := actualValue.(float64)
			if !ok {
				return fmt.Errorf("field %q is not a number", field)
			}

			// For "5 or less", check if actual is <= 5
			if strings.Contains(expectation, " or less") {
				maxValue, _ := strconv.ParseFloat(parts[0], 64)
				if actualFloat > maxValue {
					return fmt.Errorf("field %q value %.0f is not %s", field, actualFloat, expectation)
				}
			} else if strings.Contains(expectation, " or more") {
				minValue, _ := strconv.ParseFloat(parts[0], 64)
				if actualFloat < minValue {
					return fmt.Errorf("field %q value %.0f is not %s", field, actualFloat, expectation)
				}
			}
		} else {
			// Exact match
			actualValue, exists := result[field]
			if !exists {
				return fmt.Errorf("field %q not found in response", field)
			}

			actualStr := fmt.Sprintf("%v", actualValue)
			if actualStr != expectation {
				return fmt.Errorf("field %q has value %q, expected %q", field, actualStr, expectation)
			}
		}
	}

	return nil
}

func (s *scenarioState) theResponseJSONShouldIncludeDefaultValues(table *godog.Table) error {
	// This is the same as theResponseJSONShouldIncludeTable
	return s.theResponseJSONShouldIncludeTable(table)
}

func (s *scenarioState) theResponseJSONShouldIncludeList(table *godog.Table) error {
	var result interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Process each row in the table
	for i, row := range table.Rows {
		if i == 0 {
			continue // Skip header row
		}

		if len(row.Cells) < 2 {
			continue
		}

		fieldPath := row.Cells[0].Value
		expectationType := row.Cells[1].Value

		// Parse the expectation (e.g., "equals 2", "equals \"deadlift-uuid\"")
		parts := strings.SplitN(expectationType, " ", 2)
		if len(parts) < 2 {
			return fmt.Errorf("invalid expectation format: %q", expectationType)
		}

		operator := parts[0]
		expectedValue := strings.Trim(parts[1], `"`)

		// Navigate to the field using the path (e.g., "[0].sets.length")
		actualValue, err := getValueByPath(result, fieldPath)
		if err != nil {
			return fmt.Errorf("failed to get field %q: %w", fieldPath, err)
		}

		// Apply the operator
		switch operator {
		case "equals":
			actualStr := fmt.Sprintf("%v", actualValue)
			if actualStr != expectedValue {
				return fmt.Errorf("field %q has value %q, expected %q", fieldPath, actualStr, expectedValue)
			}
		default:
			return fmt.Errorf("unsupported operator: %q", operator)
		}
	}

	return nil
}

// getValueByPath navigates through a nested JSON structure using a path like "[0].sets.length"
func getValueByPath(data interface{}, path string) (interface{}, error) {
	parts := parseFieldPath(path)
	current := data

	for _, part := range parts {
		switch {
		case strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]"):
			// Array index
			indexStr := strings.Trim(part, "[]")
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %q", part)
			}

			arr, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array, got %T", current)
			}

			if index < 0 || index >= len(arr) {
				return nil, fmt.Errorf("array index %d out of bounds (length %d)", index, len(arr))
			}

			current = arr[index]

		case part == "length":
			// Special case for array length
			arr, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array for .length, got %T", current)
			}
			return len(arr), nil

		default:
			// Object field
			obj, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected object, got %T", current)
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

// parseFieldPath splits a path like "[0].sets[1].reps" into ["[0]", "sets", "[1]", "reps"]
func parseFieldPath(path string) []string {
	var parts []string
	var current strings.Builder

	inBracket := false
	for _, ch := range path {
		switch ch {
		case '[':
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			inBracket = true
			current.WriteRune(ch)
		case ']':
			current.WriteRune(ch)
			parts = append(parts, current.String())
			current.Reset()
			inBracket = false
		case '.':
			if inBracket {
				current.WriteRune(ch)
			} else {
				if current.Len() > 0 {
					parts = append(parts, current.String())
					current.Reset()
				}
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
