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
	ctx.Step(`^the response JSON should include "([^"]*)"$`, state.theResponseJSONShouldIncludeFields)
	ctx.Step(`^the response JSON should include "([^"]*)" and "([^"]*)"$`, state.theResponseJSONShouldIncludeTwoFields)
	ctx.Step(`^the response JSON should include a non-empty "([^"]*)"$`, state.theResponseJSONShouldIncludeNonEmptyField)
	ctx.Step(`^the response JSON field "([^"]*)" should be "([^"]*)"$`, state.theResponseJSONFieldShouldBe)
	ctx.Step(`^the response JSON field "([^"]*)" should contain "([^"]*)"$`, state.theResponseJSONFieldShouldContain)
	ctx.Step(`^the response JSON should include:$`, state.theResponseJSONShouldIncludeTable)
	ctx.Step(`^the response JSON should include default values:$`, state.theResponseJSONShouldIncludeTable)
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

func (s *scenarioState) theResponseJSONShouldIncludeFields(fields string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Split by comma to handle multiple fields
	fieldList := strings.Split(fields, ",")
	for _, field := range fieldList {
		field = strings.TrimSpace(field)
		if !hasField(result, field) {
			return fmt.Errorf("field %q not found in response: %s", field, string(s.lastResponseBody))
		}
	}

	return nil
}

func (s *scenarioState) theResponseJSONShouldIncludeTwoFields(field1, field2 string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if !hasField(result, field1) {
		return fmt.Errorf("field %q not found in response: %s", field1, string(s.lastResponseBody))
	}

	if !hasField(result, field2) {
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

	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("field %q is not a string", field)
	}

	if strValue == "" {
		return fmt.Errorf("field %q is empty", field)
	}

	return nil
}

func (s *scenarioState) theResponseJSONFieldShouldBe(field, expectedValue string) error {
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

func (s *scenarioState) theResponseJSONFieldShouldContain(field, substring string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	actualValue, exists := result[field]
	if !exists {
		return fmt.Errorf("field %q not found in response", field)
	}

	actualStr := fmt.Sprintf("%v", actualValue)
	if !strings.Contains(strings.ToLower(actualStr), strings.ToLower(substring)) {
		return fmt.Errorf("field %q value %q does not contain %q", field, actualStr, substring)
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

		actualValue, exists := result[field]
		if !exists {
			return fmt.Errorf("field %q not found in response", field)
		}

		// Handle "contains X" expectation
		if strings.HasPrefix(expectation, "contains ") {
			substring := strings.TrimPrefix(expectation, "contains ")
			substring = strings.Trim(substring, `"`)
			actualStr := fmt.Sprintf("%v", actualValue)
			if !strings.Contains(strings.ToLower(actualStr), strings.ToLower(substring)) {
				return fmt.Errorf("field %q value %q does not contain %q", field, actualStr, substring)
			}
			continue
		}

		// Handle "X or less" / "X or more" expectations
		if strings.Contains(expectation, " or ") {
			actualFloat, ok := actualValue.(float64)
			if !ok {
				return fmt.Errorf("field %q is not a number", field)
			}

			parts := strings.Split(expectation, " or ")
			threshold, _ := strconv.ParseFloat(parts[0], 64)

			if strings.Contains(expectation, " or less") {
				if actualFloat > threshold {
					return fmt.Errorf("field %q value %.0f is not %s", field, actualFloat, expectation)
				}
			} else if strings.Contains(expectation, " or more") {
				if actualFloat < threshold {
					return fmt.Errorf("field %q value %.0f is not %s", field, actualFloat, expectation)
				}
			}
			continue
		}

		// Exact match
		actualStr := fmt.Sprintf("%v", actualValue)
		if actualStr != expectation {
			return fmt.Errorf("field %q has value %q, expected %q", field, actualStr, expectation)
		}
	}

	return nil
}

func (s *scenarioState) theResponseJSONShouldIncludeList(table *godog.Table) error {
	var result []map[string]interface{}
	if err := json.Unmarshal(s.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse JSON as array: %w", err)
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
		expectation := row.Cells[1].Value

		// Parse expectation (e.g., "equals 2")
		parts := strings.SplitN(expectation, " ", 2)
		if len(parts) < 2 {
			return fmt.Errorf("invalid expectation format: %q", expectation)
		}

		operator := parts[0]
		expectedValue := strings.Trim(parts[1], `"`)

		// Navigate to the value
		actualValue, err := navigateToField(result, fieldPath)
		if err != nil {
			return fmt.Errorf("failed to get field %q: %w", fieldPath, err)
		}

		// Check the expectation
		if operator == "equals" {
			actualStr := fmt.Sprintf("%v", actualValue)
			if actualStr != expectedValue {
				return fmt.Errorf("field %q has value %q, expected %q", fieldPath, actualStr, expectedValue)
			}
		} else {
			return fmt.Errorf("unsupported operator: %q", operator)
		}
	}

	return nil
}

// hasField checks if a field exists, supporting dot notation for nested fields
func hasField(data map[string]interface{}, field string) bool {
	if !strings.Contains(field, ".") {
		_, exists := data[field]
		return exists
	}

	// Handle nested fields
	parts := strings.Split(field, ".")
	current := data

	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return false
		}

		if i == len(parts)-1 {
			return true
		}

		nested, ok := value.(map[string]interface{})
		if !ok {
			return false
		}
		current = nested
	}

	return false
}

// navigateToField navigates through a list response using simple path notation
// Supports: [0].field, [0].nested.field, [0].sets.length, [0].sets[0].field
func navigateToField(data []map[string]interface{}, path string) (interface{}, error) {
	// Start with the array index
	if !strings.HasPrefix(path, "[") {
		return nil, fmt.Errorf("path must start with array index: %q", path)
	}

	// Find the first closing bracket
	closeBracket := strings.Index(path, "]")
	if closeBracket == -1 {
		return nil, fmt.Errorf("invalid array index in path: %q", path)
	}

	// Extract index
	indexStr := path[1:closeBracket]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid array index: %q", indexStr)
	}

	if index < 0 || index >= len(data) {
		return nil, fmt.Errorf("array index %d out of bounds (length %d)", index, len(data))
	}

	current := data[index]
	remaining := path[closeBracket+1:]

	// If nothing remains, return the object
	if remaining == "" {
		return current, nil
	}

	// Skip the dot
	if strings.HasPrefix(remaining, ".") {
		remaining = remaining[1:]
	}

	// Navigate through the remaining path
	return navigateObjectField(current, remaining)
}

// navigateObjectField navigates through nested object fields
// Supports: field, nested.field, sets.length, sets[0].field
func navigateObjectField(obj map[string]interface{}, path string) (interface{}, error) {
	// Split by dot, but be careful with array indices
	parts := strings.Split(path, ".")
	current := interface{}(obj)

	for _, part := range parts {
		// Check if this part has an array index
		if strings.Contains(part, "[") {
			// Split field name and index
			openBracket := strings.Index(part, "[")
			fieldName := part[:openBracket]
			indexPart := part[openBracket:]

			// Get the field
			currentMap, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected object, got %T", current)
			}

			value, exists := currentMap[fieldName]
			if !exists {
				return nil, fmt.Errorf("field %q not found", fieldName)
			}

			// Parse array index
			closeBracket := strings.Index(indexPart, "]")
			indexStr := indexPart[1:closeBracket]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %q", indexStr)
			}

			arr, ok := value.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array, got %T", value)
			}

			if index < 0 || index >= len(arr) {
				return nil, fmt.Errorf("array index %d out of bounds (length %d)", index, len(arr))
			}

			current = arr[index]
			continue
		}

		// Handle .length special case
		if part == "length" {
			arr, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array for .length, got %T", current)
			}
			return len(arr), nil
		}

		// Regular field access
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected object, got %T", current)
		}

		value, exists := currentMap[part]
		if !exists {
			return nil, fmt.Errorf("field %q not found", part)
		}

		current = value
	}

	return current, nil
}
