package domain

// requiredPhases are the phases that must be covered for a complete domain model.
var requiredPhases = []QuestionPhase{PhaseActors, PhaseStory, PhaseEvents, PhaseBoundaries}

// ModelCompleteness represents the degree to which the domain model covers
// all required aspects (actors, stories, events, boundaries).
type ModelCompleteness struct {
	gaps []ModelGap
}

// NewModelCompleteness creates a ModelCompleteness value object.
func NewModelCompleteness(gaps []ModelGap) ModelCompleteness {
	g := make([]ModelGap, len(gaps))
	copy(g, gaps)
	return ModelCompleteness{gaps: g}
}

// IsComplete returns true if there are no gaps in the domain model.
func (mc ModelCompleteness) IsComplete() bool {
	return len(mc.gaps) == 0
}

// Gaps returns a defensive copy of the model gaps.
func (mc ModelCompleteness) Gaps() []ModelGap {
	out := make([]ModelGap, len(mc.gaps))
	copy(out, mc.gaps)
	return out
}

// CoveredPhases returns the required phases that have no gaps.
func (mc ModelCompleteness) CoveredPhases() []QuestionPhase {
	gapped := make(map[QuestionPhase]bool, len(mc.gaps))
	for _, g := range mc.gaps {
		gapped[g.Phase()] = true
	}
	var covered []QuestionPhase
	for _, p := range requiredPhases {
		if !gapped[p] {
			covered = append(covered, p)
		}
	}
	return covered
}
