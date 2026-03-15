package domain

import (
	"fmt"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// FixedQuestionFlow implements the existing 10-question sequential flow.
// It enforces phase ordering, playback every 3 answers, and MVP question completeness.
type FixedQuestionFlow struct{}

// Compile-time check that FixedQuestionFlow satisfies DiscoveryFlow.
var _ DiscoveryFlow = (*FixedQuestionFlow)(nil)

// NewFixedQuestionFlow creates a FixedQuestionFlow.
func NewFixedQuestionFlow() *FixedQuestionFlow { return &FixedQuestionFlow{} }

// ValidateQuestionOrder enforces phase ordering: Actors -> Story -> Events -> Boundaries.
// SEED phase is always allowed.
func (f *FixedQuestionFlow) ValidateQuestionOrder(ref QuestionRef, answered []Answer, skipped map[string]bool) error {
	targetIdx := -1
	for i, p := range questionPhases {
		if p == ref.Phase() {
			targetIdx = i
			break
		}
	}
	if targetIdx < 0 {
		return nil // SEED phase always allowed
	}

	allHandled := make(map[string]bool)
	for _, a := range answered {
		allHandled[a.QuestionID()] = true
	}
	for id := range skipped {
		allHandled[id] = true
	}

	catalog := QuestionCatalog()
	for i := 0; i < targetIdx; i++ {
		earlierPhase := questionPhases[i]
		for _, q := range catalog {
			if q.Phase() == earlierPhase && !allHandled[q.ID()] {
				return fmt.Errorf("cannot answer %s (%s phase) before completing %s phase (question %s not answered or skipped): %w",
					ref.ID(), ref.Phase(), earlierPhase, q.ID(), domainerrors.ErrInvariantViolation)
			}
		}
	}
	return nil
}

// IsPlaybackDue returns true when answer count reaches the fixed interval of 3.
func (f *FixedQuestionFlow) IsPlaybackDue(answersSinceLastPlayback int) bool {
	return answersSinceLastPlayback >= fixedPlaybackInterval
}

// PlaybackInterval returns the fixed playback interval of 3.
func (f *FixedQuestionFlow) PlaybackInterval() int { return fixedPlaybackInterval }

// CheckCompleteness checks that all MVP questions have been answered.
func (f *FixedQuestionFlow) CheckCompleteness(answers []Answer, _ map[string]bool) error {
	answeredIDs := make(map[string]bool)
	for _, a := range answers {
		answeredIDs[a.QuestionID()] = true
	}
	mvpIDs := MVPQuestionIDs()
	var missing []string
	for id := range mvpIDs {
		if !answeredIDs[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("cannot complete: MVP questions not answered: %v: %w",
			missing, domainerrors.ErrInvariantViolation)
	}
	return nil
}

// fixedPlaybackInterval is the playback cadence for fixed-question flow.
const fixedPlaybackInterval = 3
