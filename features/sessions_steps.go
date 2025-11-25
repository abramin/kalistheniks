package features

import (
	"github.com/cucumber/godog"
)

// registerSessionsSteps registers session-related step definitions.
func registerSessionsSteps(ctx *godog.ScenarioContext, state *scenarioState) {
	ctx.Step(`^I POST /sessions with headers:$`, state.iPostSessionsWithHeaders)
	ctx.Step(`^I POST /sessions without an Authorization header$`, state.iPostSessionsWithoutAuthHeader)
	ctx.Step(`^I POST /sessions with body:$`, state.iPostSessionsWithBody)
	ctx.Step(`^I POST /sessions/([^/]+)/sets with headers:$`, state.iPostSessionSetsWithHeaders)
	ctx.Step(`^I POST /sessions/invalid-session-id/sets with headers:$`, state.iPostInvalidSessionSetsWithHeaders)
	ctx.Step(`^I GET /sessions with headers:$`, state.iGetSessionsWithHeaders)
}

// ========== Sessions HTTP request steps ==========

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
