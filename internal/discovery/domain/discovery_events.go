package domain

import (
	"encoding/json"
	"fmt"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

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

// MarshalJSON implements json.Marshaler for event bus serialization.
func (e DiscoveryCompletedEvent) MarshalJSON() ([]byte, error) {
	type proxy struct {
		TechStack             *vo.TechStack     `json:"tech_stack"`
		SessionID             string            `json:"session_id"`
		Persona               DiscoveryPersona  `json:"persona"`
		Register              DiscoveryRegister `json:"register"`
		Answers               []Answer          `json:"answers"`
		PlaybackConfirmations []Playback        `json:"playback_confirmations"`
	}
	data, err := json.Marshal(proxy{
		TechStack:             e.techStack,
		SessionID:             e.sessionID,
		Persona:               e.persona,
		Register:              e.register,
		Answers:               e.answers,
		PlaybackConfirmations: e.playbackConfirmations,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling DiscoveryCompletedEvent: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (e *DiscoveryCompletedEvent) UnmarshalJSON(data []byte) error {
	type proxy struct {
		TechStack             *vo.TechStack     `json:"tech_stack"`
		SessionID             string            `json:"session_id"`
		Persona               DiscoveryPersona  `json:"persona"`
		Register              DiscoveryRegister `json:"register"`
		Answers               []Answer          `json:"answers"`
		PlaybackConfirmations []Playback        `json:"playback_confirmations"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling DiscoveryCompletedEvent: %w", err)
	}
	e.techStack = p.TechStack
	e.sessionID = p.SessionID
	e.persona = p.Persona
	e.register = p.Register
	e.answers = p.Answers
	e.playbackConfirmations = p.PlaybackConfirmations
	return nil
}

// BoundedContextClassifiedEvent is emitted when a bounded context is classified.
type BoundedContextClassifiedEvent struct {
	sessionID      string
	contextName    string
	classification vo.SubdomainClassification
	rationale      string
}

// NewBoundedContextClassifiedEvent creates a BoundedContextClassifiedEvent.
func NewBoundedContextClassifiedEvent(
	sessionID string,
	contextName string,
	classification vo.SubdomainClassification,
	rationale string,
) BoundedContextClassifiedEvent {
	return BoundedContextClassifiedEvent{
		sessionID:      sessionID,
		contextName:    contextName,
		classification: classification,
		rationale:      rationale,
	}
}

// SessionID returns the session identifier.
func (e BoundedContextClassifiedEvent) SessionID() string { return e.sessionID }

// ContextName returns the bounded context name.
func (e BoundedContextClassifiedEvent) ContextName() string { return e.contextName }

// Classification returns the subdomain classification.
func (e BoundedContextClassifiedEvent) Classification() vo.SubdomainClassification {
	return e.classification
}

// Rationale returns the classification rationale.
func (e BoundedContextClassifiedEvent) Rationale() string { return e.rationale }

// Equal returns true if two events have the same values.
func (e BoundedContextClassifiedEvent) Equal(other BoundedContextClassifiedEvent) bool {
	return e.sessionID == other.sessionID &&
		e.contextName == other.contextName &&
		e.classification == other.classification &&
		e.rationale == other.rationale
}

// MarshalJSON implements json.Marshaler for event bus serialization.
func (e BoundedContextClassifiedEvent) MarshalJSON() ([]byte, error) {
	type proxy struct {
		SessionID      string                     `json:"session_id"`
		ContextName    string                     `json:"context_name"`
		Classification vo.SubdomainClassification `json:"classification"`
		Rationale      string                     `json:"rationale"`
	}
	data, err := json.Marshal(proxy{
		SessionID:      e.sessionID,
		ContextName:    e.contextName,
		Classification: e.classification,
		Rationale:      e.rationale,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling BoundedContextClassifiedEvent: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (e *BoundedContextClassifiedEvent) UnmarshalJSON(data []byte) error {
	type proxy struct {
		SessionID      string                     `json:"session_id"`
		ContextName    string                     `json:"context_name"`
		Classification vo.SubdomainClassification `json:"classification"`
		Rationale      string                     `json:"rationale"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling BoundedContextClassifiedEvent: %w", err)
	}
	e.sessionID = p.SessionID
	e.contextName = p.ContextName
	e.classification = p.Classification
	e.rationale = p.Rationale
	return nil
}
