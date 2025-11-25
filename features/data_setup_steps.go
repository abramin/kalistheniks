package features

import (
	"github.com/cucumber/godog"
)

// registerDataSetupSteps registers data setup-related step definitions.
func registerDataSetupSteps(ctx *godog.ScenarioContext, state *scenarioState) {
	ctx.Step(`^my last recorded set has:$`, state.myLastRecordedSetHas)
	ctx.Step(`^I have added two sets to session "([^"]*)"$`, state.iHaveAddedTwoSetsToSession)
}

// ========== Data setup steps ==========

func (s *scenarioState) myLastRecordedSetHas(table *godog.Table) error {
	return godog.ErrPending
}

func (s *scenarioState) iHaveAddedTwoSetsToSession(sessionID string) error {
	return godog.ErrPending
}
