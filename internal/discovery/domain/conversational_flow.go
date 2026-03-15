package domain

import (
	"fmt"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

const defaultConversationalPlaybackInterval = 5

// ConversationalFlow implements adaptive LLM-driven discovery.
// It allows any question order, uses a configurable playback interval,
// and checks semantic phase coverage for completeness.
type ConversationalFlow struct {
	playbackInterval    int
	registeredQuestions map[string]QuestionRef
}

// Compile-time check that ConversationalFlow satisfies DiscoveryFlow.
var _ DiscoveryFlow = (*ConversationalFlow)(nil)

// NewConversationalFlow creates a ConversationalFlow with configurable playback interval.
// If playbackInterval is <= 0, defaults to 5.
func NewConversationalFlow(playbackInterval int) *ConversationalFlow {
	if playbackInterval <= 0 {
		playbackInterval = defaultConversationalPlaybackInterval
	}
	return &ConversationalFlow{
		playbackInterval:    playbackInterval,
		registeredQuestions: make(map[string]QuestionRef),
	}
}

// RegisterQuestion registers a dynamically generated question with phase metadata.
// This is needed for semantic completeness checking.
func (f *ConversationalFlow) RegisterQuestion(ref QuestionRef) {
	f.registeredQuestions[ref.ID()] = ref
}

// ValidateQuestionOrder always returns nil — conversational mode allows any order.
func (f *ConversationalFlow) ValidateQuestionOrder(_ QuestionRef, _ []Answer, _ map[string]bool) error {
	return nil
}

// IsPlaybackDue returns true when the answer count reaches the configured interval.
func (f *ConversationalFlow) IsPlaybackDue(answersSinceLastPlayback int) bool {
	return answersSinceLastPlayback >= f.playbackInterval
}

// PlaybackInterval returns the configured playback interval.
func (f *ConversationalFlow) PlaybackInterval() int { return f.playbackInterval }

// CheckCompleteness verifies that all required phases have at least one answer.
// Uses registered question metadata to determine which phases are covered.
func (f *ConversationalFlow) CheckCompleteness(answers []Answer, _ map[string]bool) error {
	completeness := f.Completeness(answers)
	if completeness.IsComplete() {
		return nil
	}
	gaps := completeness.Gaps()
	var missingPhases []string
	for _, g := range gaps {
		missingPhases = append(missingPhases, string(g.Phase()))
	}
	return fmt.Errorf("cannot complete: domain model gaps in phases %v: %w",
		missingPhases, domainerrors.ErrInvariantViolation)
}

// Completeness computes the current model completeness based on answered questions.
func (f *ConversationalFlow) Completeness(answers []Answer) ModelCompleteness {
	// Build set of phases that have at least one answer
	coveredPhases := make(map[QuestionPhase]bool)
	for _, a := range answers {
		if ref, ok := f.registeredQuestions[a.QuestionID()]; ok {
			coveredPhases[ref.Phase()] = true
		} else {
			// For catalog questions (Q1-Q10), look up phase from catalog
			if q, exists := QuestionByID()[a.QuestionID()]; exists {
				coveredPhases[q.Phase()] = true
			}
		}
	}

	// Identify gaps
	var gaps []ModelGap
	for _, phase := range requiredPhases {
		if !coveredPhases[phase] {
			gap, _ := NewModelGap(phase, fmt.Sprintf("no answers covering %s phase", phase))
			gaps = append(gaps, gap)
		}
	}
	return NewModelCompleteness(gaps)
}
