package features

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/cucumber/godog"
)

// registerDataSetupSteps registers data setup-related step definitions.
func registerDataSetupSteps(ctx *godog.ScenarioContext, state *scenarioState) {
	ctx.Step(`^my last recorded set has:$`, state.myLastRecordedSetHas)
	ctx.Step(`^I have added two sets to session "([^"]*)"$`, state.iHaveAddedTwoSetsToSession)
}

// ========== Data setup steps ==========

func (s *scenarioState) myLastRecordedSetHas(table *godog.Table) error {
	ctx := context.Background()

	// Parse the table to extract set data
	setData := make(map[string]string)
	for i, row := range table.Rows {
		if i == 0 {
			continue // Skip header row
		}
		if len(row.Cells) >= 2 {
			setData[row.Cells[0].Value] = row.Cells[1].Value
		}
	}

	// First, create a session
	sessionType := setData["session_type"]
	if sessionType == "" {
		sessionType = "upper"
	}

	const insertSessionSQL = `
		INSERT INTO sessions (user_id, performed_at, session_type, notes)
		SELECT id, '2024-01-01T10:00:00Z', $1, 'Test session'
		FROM users
		WHERE email = 'user@example.com'
		RETURNING id
	`

	var sessionID string
	err := s.db.QueryRowContext(ctx, insertSessionSQL, sessionType).Scan(&sessionID)
	if err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}

	// Now insert the set
	exerciseID := setData["exercise_id"]
	reps, _ := strconv.Atoi(setData["reps"])
	weightKg, _ := strconv.ParseFloat(setData["weight_kg"], 64)

	const insertSetSQL = `
		INSERT INTO sets (session_id, exercise_id, set_index, reps, weight_kg, rpe)
		VALUES ($1, $2, 1, $3, $4, 7)
	`

	_, err = s.db.ExecContext(ctx, insertSetSQL, sessionID, exerciseID, reps, weightKg)
	if err != nil {
		return fmt.Errorf("failed to insert set: %w", err)
	}

	return nil
}

func (s *scenarioState) iHaveAddedTwoSetsToSession(sessionID string) error {
	// Replace placeholder sessionID
	sessionID = s.replacePlaceholders(sessionID)

	// Add first set
	set1Body := map[string]interface{}{
		"exercise_id": "deadlift-uuid",
		"set_index":   1,
		"reps":        8,
		"weight_kg":   100.0,
		"rpe":         7,
	}
	bodyBytes, _ := json.Marshal(set1Body)

	req, err := http.NewRequest("POST", s.baseURL+"/sessions/"+sessionID+"/sets", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create set 1 request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add set 1: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add set 1: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Add second set
	set2Body := map[string]interface{}{
		"exercise_id": "deadlift-uuid",
		"set_index":   2,
		"reps":        10,
		"weight_kg":   105.0,
		"rpe":         8,
	}
	bodyBytes, _ = json.Marshal(set2Body)

	req, err = http.NewRequest("POST", s.baseURL+"/sessions/"+sessionID+"/sets", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create set 2 request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err = s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add set 2: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add set 2: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
