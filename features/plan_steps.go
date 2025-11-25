package features

import (
	"github.com/cucumber/godog"
)

// registerPlanSteps registers plan-related step definitions.
func registerPlanSteps(ctx *godog.ScenarioContext, state *scenarioState) {
	ctx.Step(`^I GET /plan/next with headers:$`, state.iGetPlanNextWithHeaders)
	ctx.Step(`^I GET /plan/next without an Authorization header$`, state.iGetPlanNextWithoutAuthHeader)
}

// ========== Plan HTTP request steps ==========

func (s *scenarioState) iGetPlanNextWithHeaders(table *godog.Table) error {
	return godog.ErrPending
}

func (s *scenarioState) iGetPlanNextWithoutAuthHeader() error {
	return godog.ErrPending
}
