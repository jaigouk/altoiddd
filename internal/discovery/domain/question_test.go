package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuestionHasRequiredFields(t *testing.T) {
	t.Parallel()
	q := NewQuestion("Q1", PhaseActors, "Who are the actors?", "Who will use this?", []string{"actors"})
	assert.Equal(t, "Q1", q.ID())
	assert.Equal(t, PhaseActors, q.Phase())
	assert.Equal(t, "Who are the actors?", q.TechnicalText())
	assert.Equal(t, "Who will use this?", q.NonTechnicalText())
	assert.Equal(t, []string{"actors"}, q.Produces())
}

func TestCatalogHasTenQuestions(t *testing.T) {
	t.Parallel()
	assert.Len(t, QuestionCatalog(), 10)
}

func TestCatalogIDsAreQ1ThroughQ10(t *testing.T) {
	t.Parallel()
	catalog := QuestionCatalog()
	ids := make([]string, len(catalog))
	for i, q := range catalog {
		ids[i] = q.ID()
	}
	expected := []string{"Q1", "Q2", "Q3", "Q4", "Q5", "Q6", "Q7", "Q8", "Q9", "Q10"}
	assert.Equal(t, expected, ids)
}

func TestCatalogPhaseAssignments(t *testing.T) {
	t.Parallel()
	catalog := QuestionCatalog()
	phaseMap := make(map[string]QuestionPhase)
	for _, q := range catalog {
		phaseMap[q.ID()] = q.Phase()
	}
	assert.Equal(t, PhaseActors, phaseMap["Q1"])
	assert.Equal(t, PhaseActors, phaseMap["Q2"])
	assert.Equal(t, PhaseStory, phaseMap["Q3"])
	assert.Equal(t, PhaseStory, phaseMap["Q4"])
	assert.Equal(t, PhaseStory, phaseMap["Q5"])
	assert.Equal(t, PhaseEvents, phaseMap["Q6"])
	assert.Equal(t, PhaseEvents, phaseMap["Q7"])
	assert.Equal(t, PhaseEvents, phaseMap["Q8"])
	assert.Equal(t, PhaseBoundaries, phaseMap["Q9"])
	assert.Equal(t, PhaseBoundaries, phaseMap["Q10"])
}

func TestEachQuestionHasBothRegisterTexts(t *testing.T) {
	t.Parallel()
	for _, q := range QuestionCatalog() {
		assert.NotEmpty(t, q.TechnicalText(), "%s missing technical_text", q.ID())
		assert.NotEmpty(t, q.NonTechnicalText(), "%s missing non_technical_text", q.ID())
	}
}

func TestEachQuestionHasProduces(t *testing.T) {
	t.Parallel()
	for _, q := range QuestionCatalog() {
		assert.NotEmpty(t, q.Produces(), "%s has empty produces", q.ID())
	}
}

func TestDualRegisterTextsDiffer(t *testing.T) {
	t.Parallel()
	for _, q := range QuestionCatalog() {
		assert.NotEqual(t, q.TechnicalText(), q.NonTechnicalText(), "%s has identical register texts", q.ID())
	}
}

func TestMVPQuestionIDsContainsFive(t *testing.T) {
	t.Parallel()
	assert.Len(t, MVPQuestionIDs(), 5)
}

func TestMVPQuestionIDsAreCorrect(t *testing.T) {
	t.Parallel()
	expected := map[string]bool{"Q1": true, "Q3": true, "Q4": true, "Q9": true, "Q10": true}
	assert.Equal(t, expected, MVPQuestionIDs())
}

func TestMVPQuestionIDsAreSubsetOfCatalog(t *testing.T) {
	t.Parallel()
	catalogIDs := make(map[string]bool)
	for _, q := range QuestionCatalog() {
		catalogIDs[q.ID()] = true
	}
	for id := range MVPQuestionIDs() {
		assert.True(t, catalogIDs[id], "MVP ID %s not in catalog", id)
	}
}

func TestProducesReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	q := QuestionCatalog()[0]
	p1 := q.Produces()
	p1[0] = "mutated"
	assert.NotEqual(t, "mutated", q.Produces()[0])
}
