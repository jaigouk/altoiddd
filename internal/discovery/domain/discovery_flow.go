package domain

// DiscoveryFlow encapsulates mode-specific question flow behavior.
// Fixed mode enforces phase ordering and MVP questions.
// Conversational mode allows adaptive ordering and semantic completeness.
type DiscoveryFlow interface {
	// ValidateQuestionOrder checks if answering this question is allowed given current state.
	ValidateQuestionOrder(ref QuestionRef, answered []Answer, skipped map[string]bool) error

	// IsPlaybackDue returns true if a playback should be triggered.
	IsPlaybackDue(answersSinceLastPlayback int) bool

	// PlaybackInterval returns the number of answers between playbacks.
	PlaybackInterval() int

	// CheckCompleteness verifies the session has sufficient coverage to complete.
	CheckCompleteness(answers []Answer, skipped map[string]bool) error
}

// QuestionRef is a mode-agnostic reference to a question.
// Fixed mode: ID is "Q1"-"Q10". Conversational mode: generated IDs.
type QuestionRef struct {
	id    string
	phase QuestionPhase
}

// NewQuestionRef creates a QuestionRef value object.
func NewQuestionRef(id string, phase QuestionPhase) QuestionRef {
	return QuestionRef{id: id, phase: phase}
}

// ID returns the question identifier.
func (r QuestionRef) ID() string { return r.id }

// Phase returns the question phase.
func (r QuestionRef) Phase() QuestionPhase { return r.phase }

// Equal returns true if two QuestionRefs have the same values.
func (r QuestionRef) Equal(other QuestionRef) bool {
	return r.id == other.id && r.phase == other.phase
}
