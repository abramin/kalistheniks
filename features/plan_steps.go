package features

import (
	"github.com/cucumber/godog"
)

// registerPlanSteps registers plan-related step definitions.
func registerPlanSteps(ctx *godog.ScenarioContext, state *scenarioState) {
	ctx.Step(`^I GET /plan/next with headers:$`, state.iGetPlanNextWithHeaders)
	ctx.Step(`^I GET /plan/next without an Authorization header$`, state.iGetPlanNextWithoutAuthHeader)
	ctx.Step(`^I have no recorded sessions or sets$`, state.iHaveNoRecordedSessionsOrSets)
}

// ========== Plan HTTP request steps ==========

func (s *scenarioState) iGetPlanNextWithHeaders(table *godog.Table) error {
	headers := s.extractHeaders(table)
	return s.doGetRequestWithHeaders("/plan/next", headers)
}

func (s *scenarioState) iGetPlanNextWithoutAuthHeader() error {
	return s.doGetRequest("/plan/next", "")
}

func (s *scenarioState) iHaveNoRecordedSessionsOrSets() error {
	// This is already handled by the "the database is empty" step,
	// but we include it here for clarity in the feature file
	return nil
}
