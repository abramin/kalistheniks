package features

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

type scenarioState struct {
	token string
}

func TestFeatures(t *testing.T) {
	opts := godog.Options{
		Output: colors.Colored(os.Stdout),
		Paths:  []string{"features"},
		Strict: true,
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

// InitializeTestSuite allows shared setup/teardown across scenarios.
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {})
	ctx.AfterSuite(func() {})
}

// InitializeScenario registers step definitions for all feature files.
func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &scenarioState{}

	ctx.Before(func(context.Context, *godog.Scenario) (context.Context, error) {
		state.token = ""
		return context.Background(), nil
	})
	// Common steps
	ctx.Step(`^the database is empty$`, func() error {
		return nil
	})
	ctx.Step(`^I have a valid token from logging in as "([^"]*)"$`, func(email string) error {
		state.token = email // placeholder; replace with real token acquisition
		return nil
	})
	ctx.Step(`^I POST /signup with body:$`, func(body *godog.DocString) error {
		return nil
	})
	ctx.Step(`^I POST /login with body:$`, func(body *godog.DocString) error {
		// TODO: call login endpoint with provided body.
		return godog.ErrPending
	})
	ctx.Step(`^I POST /sessions with headers:$`, func(table *godog.Table) error {
		// TODO: call create session endpoint with headers and stored token.
		return godog.ErrPending
	})
	ctx.Step(`^I POST /sessions without an Authorization header$`, func() error {
		// TODO: call endpoint without token.
		return godog.ErrPending
	})
	ctx.Step(`^I POST /sessions with body:$`, func(body *godog.DocString) error {
		// TODO: call create session with body.
		return godog.ErrPending
	})
	ctx.Step(`^I POST /sessions/([^/]+)/sets with headers:$`, func(sessionID string, table *godog.Table) error {
		// TODO: call add set endpoint.
		return godog.ErrPending
	})
	ctx.Step(`^I POST /sessions/invalid-session-id/sets with headers:$`, func(table *godog.Table) error {
		// TODO: call add set with invalid session.
		return godog.ErrPending
	})
	ctx.Step(`^I GET /sessions with headers:$`, func(table *godog.Table) error {
		// TODO: call list sessions.
		return godog.ErrPending
	})
	ctx.Step(`^I GET /plan/next with headers:$`, func(table *godog.Table) error {
		// TODO: call plan next endpoint.
		return godog.ErrPending
	})
	ctx.Step(`^I GET /plan/next without an Authorization header$`, func() error {
		// TODO: call plan next without token.
		return godog.ErrPending
	})
	ctx.Step(`^the response status should be (\d+)$`, func(code int) error {
		// TODO: assert HTTP status.
		return godog.ErrPending
	})
	ctx.Step(`^the response JSON should include "([^"]*)" and "([^"]*)"$`, func(field1, field2 string) error {
		// TODO: assert JSON fields exist.
		return godog.ErrPending
	})
	ctx.Step(`^the response JSON should include a non-empty "([^"]*)"$`, func(field string) error {
		// TODO: assert field is present and not empty.
		return godog.ErrPending
	})
	ctx.Step(`^the response JSON should include an "([^"]*)" explaining the email is taken$`, func(field string) error {
		// TODO: assert duplicate signup error.
		return godog.ErrPending
	})
	ctx.Step(`^the response JSON should include "([^"]*)"$`, func(field string) error {
		// TODO: generic field presence check.
		return godog.ErrPending
	})
	ctx.Step(`^the response JSON should include an "([^"]*)" about invalid request body$`, func(field string) error {
		// TODO: assert validation error.
		return godog.ErrPending
	})
	ctx.Step(`^the response JSON should include default values:$`, func(table *godog.Table) error {
		// TODO: assert default plan response.
		return godog.ErrPending
	})
	ctx.Step(`^the response JSON should include a list where:$`, func(table *godog.Table) error {
		// TODO: assert sessions list contents.
		return godog.ErrPending
	})
	ctx.Step(`^my last recorded set has:$`, func(table *godog.Table) error {
		// TODO: seed last set data.
		return godog.ErrPending
	})
	ctx.Step(`^I have added two sets to session "([^"]*)"$`, func(sessionID string) error {
		// TODO: seed two sets for given session.
		return godog.ErrPending
	})
}
