package domain

import vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"

// DiscoveryCompletedEvent is emitted when a discovery session completes successfully.
type DiscoveryCompletedEvent struct {
	techStack             *vo.TechStack
	sessionID             string
	persona               DiscoveryPersona
	register              DiscoveryRegister
	answers               []Answer
	playbackConfirmations []Playback
}

// NewDiscoveryCompletedEvent creates a DiscoveryCompletedEvent with defensive copies.
func NewDiscoveryCompletedEvent(
	sessionID string,
	persona DiscoveryPersona,
	register DiscoveryRegister,
	answers []Answer,
	playbackConfirmations []Playback,
	techStack *vo.TechStack,
) DiscoveryCompletedEvent {
	a := make([]Answer, len(answers))
	copy(a, answers)
	p := make([]Playback, len(playbackConfirmations))
	copy(p, playbackConfirmations)
	var ts *vo.TechStack
	if techStack != nil {
		cp := *techStack
		ts = &cp
	}
	return DiscoveryCompletedEvent{
		sessionID:             sessionID,
		persona:               persona,
		register:              register,
		answers:               a,
		playbackConfirmations: p,
		techStack:             ts,
	}
}

// SessionID returns the session identifier.
func (e DiscoveryCompletedEvent) SessionID() string { return e.sessionID }

// Persona returns the detected persona.
func (e DiscoveryCompletedEvent) Persona() DiscoveryPersona { return e.persona }

// Register returns the language register.
func (e DiscoveryCompletedEvent) Register() DiscoveryRegister { return e.register }

// Answers returns a defensive copy of the answers.
func (e DiscoveryCompletedEvent) Answers() []Answer {
	out := make([]Answer, len(e.answers))
	copy(out, e.answers)
	return out
}

// PlaybackConfirmations returns a defensive copy of the playback confirmations.
func (e DiscoveryCompletedEvent) PlaybackConfirmations() []Playback {
	out := make([]Playback, len(e.playbackConfirmations))
	copy(out, e.playbackConfirmations)
	return out
}

// TechStack returns the tech stack, or nil if not set.
func (e DiscoveryCompletedEvent) TechStack() *vo.TechStack { return e.techStack }

// Equal returns true if two events have the same values.
func (e DiscoveryCompletedEvent) Equal(other DiscoveryCompletedEvent) bool {
	if e.sessionID != other.sessionID || e.persona != other.persona || e.register != other.register {
		return false
	}
	if len(e.answers) != len(other.answers) {
		return false
	}
	for i := range e.answers {
		if !e.answers[i].Equal(other.answers[i]) {
			return false
		}
	}
	if len(e.playbackConfirmations) != len(other.playbackConfirmations) {
		return false
	}
	for i := range e.playbackConfirmations {
		if !e.playbackConfirmations[i].Equal(other.playbackConfirmations[i]) {
			return false
		}
	}
	if (e.techStack == nil) != (other.techStack == nil) {
		return false
	}
	if e.techStack != nil && !e.techStack.Equal(*other.techStack) {
		return false
	}
	return true
}
