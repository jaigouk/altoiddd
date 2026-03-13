// Package domain contains the Discovery bounded context domain model.
package domain

import (
	"encoding/json"
	"fmt"
)

// DiscoveryPersona is the user's self-identified role during discovery.
// NOTE: Named DiscoveryPersona to avoid collision with shared kernel PersonaType.
type DiscoveryPersona string

// DiscoveryPersona constants.
const (
	PersonaDeveloper    DiscoveryPersona = "developer"
	PersonaProductOwner DiscoveryPersona = "product_owner"
	PersonaDomainExpert DiscoveryPersona = "domain_expert"
	PersonaMixed        DiscoveryPersona = "mixed"
)

// AllDiscoveryPersonas returns all valid persona values.
func AllDiscoveryPersonas() []DiscoveryPersona {
	return []DiscoveryPersona{PersonaDeveloper, PersonaProductOwner, PersonaDomainExpert, PersonaMixed}
}

// ParseDiscoveryPersona parses a string into a DiscoveryPersona.
func ParseDiscoveryPersona(s string) (DiscoveryPersona, error) {
	switch DiscoveryPersona(s) {
	case PersonaDeveloper, PersonaProductOwner, PersonaDomainExpert, PersonaMixed:
		return DiscoveryPersona(s), nil
	default:
		return "", fmt.Errorf("invalid persona: %q", s)
	}
}

// DiscoveryRegister is the language register for question phrasing.
type DiscoveryRegister string

// DiscoveryRegister constants.
const (
	RegisterTechnical    DiscoveryRegister = "technical"
	RegisterNonTechnical DiscoveryRegister = "non_technical"
)

// AllRegisters returns all valid register values.
func AllRegisters() []DiscoveryRegister {
	return []DiscoveryRegister{RegisterTechnical, RegisterNonTechnical}
}

// ParseDiscoveryRegister parses a string into a DiscoveryRegister.
func ParseDiscoveryRegister(s string) (DiscoveryRegister, error) {
	switch DiscoveryRegister(s) {
	case RegisterTechnical, RegisterNonTechnical:
		return DiscoveryRegister(s), nil
	default:
		return "", fmt.Errorf("invalid register: %q", s)
	}
}

// QuestionPhase represents phases of the 10-question DDD discovery flow.
type QuestionPhase string

// QuestionPhase constants.
const (
	PhaseSeed           QuestionPhase = "seed"
	PhaseActors         QuestionPhase = "actors"
	PhaseStory          QuestionPhase = "story"
	PhaseEvents         QuestionPhase = "events"
	PhaseBoundaries     QuestionPhase = "boundaries"
	PhaseClassification QuestionPhase = "classification"
)

// AllQuestionPhases returns all question phases in order.
func AllQuestionPhases() []QuestionPhase {
	return []QuestionPhase{PhaseSeed, PhaseActors, PhaseStory, PhaseEvents, PhaseBoundaries, PhaseClassification}
}

// DiscoveryMode represents the discovery thoroughness level.
type DiscoveryMode string

// DiscoveryMode constants.
const (
	ModeExpress DiscoveryMode = "express"
	ModeDeep    DiscoveryMode = "deep"
)

// AllDiscoveryModes returns all valid discovery modes.
func AllDiscoveryModes() []DiscoveryMode {
	return []DiscoveryMode{ModeExpress, ModeDeep}
}

// ParseDiscoveryMode parses a string into a DiscoveryMode.
func ParseDiscoveryMode(s string) (DiscoveryMode, error) {
	switch DiscoveryMode(s) {
	case ModeExpress, ModeDeep:
		return DiscoveryMode(s), nil
	default:
		return "", fmt.Errorf("invalid discovery mode: %q", s)
	}
}

// DiscoveryRound indicates which round of discovery is active.
type DiscoveryRound string

// DiscoveryRound constants.
const (
	RoundDiscovery DiscoveryRound = "discovery"
	RoundChallenge DiscoveryRound = "challenge"
	RoundSimulate  DiscoveryRound = "simulate"
)

// AllDiscoveryRounds returns all valid discovery rounds.
func AllDiscoveryRounds() []DiscoveryRound {
	return []DiscoveryRound{RoundDiscovery, RoundChallenge, RoundSimulate}
}

// ParseDiscoveryRound parses a string into a DiscoveryRound.
func ParseDiscoveryRound(s string) (DiscoveryRound, error) {
	switch DiscoveryRound(s) {
	case RoundDiscovery, RoundChallenge, RoundSimulate:
		return DiscoveryRound(s), nil
	default:
		return "", fmt.Errorf("invalid discovery round: %q", s)
	}
}

// Answer is a user's response to a single discovery question.
type Answer struct {
	questionID   string
	responseText string
}

// NewAnswer creates an Answer value object.
func NewAnswer(questionID, responseText string) Answer {
	return Answer{questionID: questionID, responseText: responseText}
}

// QuestionID returns the question identifier.
func (a Answer) QuestionID() string { return a.questionID }

// ResponseText returns the user's free-text answer.
func (a Answer) ResponseText() string { return a.responseText }

// Equal returns true if two answers have the same values.
func (a Answer) Equal(other Answer) bool {
	return a.questionID == other.questionID && a.responseText == other.responseText
}

// MarshalJSON implements json.Marshaler for event bus serialization.
func (a Answer) MarshalJSON() ([]byte, error) {
	type proxy struct {
		QuestionID   string `json:"question_id"`
		ResponseText string `json:"response_text"`
	}
	data, err := json.Marshal(proxy{
		QuestionID:   a.questionID,
		ResponseText: a.responseText,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling Answer: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (a *Answer) UnmarshalJSON(data []byte) error {
	type proxy struct {
		QuestionID   string `json:"question_id"`
		ResponseText string `json:"response_text"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling Answer: %w", err)
	}
	a.questionID = p.QuestionID
	a.responseText = p.ResponseText
	return nil
}

// Playback is a playback summary shown to the user for confirmation.
type Playback struct {
	summaryText string
	corrections string
	confirmed   bool
}

// NewPlayback creates a Playback value object.
func NewPlayback(summaryText string, confirmed bool, corrections string) Playback {
	return Playback{summaryText: summaryText, confirmed: confirmed, corrections: corrections}
}

// SummaryText returns the generated summary text.
func (p Playback) SummaryText() string { return p.summaryText }

// Confirmed returns whether the user confirmed the playback.
func (p Playback) Confirmed() bool { return p.confirmed }

// Corrections returns the user-provided corrections.
func (p Playback) Corrections() string { return p.corrections }

// Equal returns true if two playbacks have the same values.
func (p Playback) Equal(other Playback) bool {
	return p.summaryText == other.summaryText &&
		p.confirmed == other.confirmed &&
		p.corrections == other.corrections
}

// MarshalJSON implements json.Marshaler for event bus serialization.
func (p Playback) MarshalJSON() ([]byte, error) {
	type proxy struct {
		SummaryText string `json:"summary_text"`
		Confirmed   bool   `json:"confirmed"`
		Corrections string `json:"corrections"`
	}
	data, err := json.Marshal(proxy{
		SummaryText: p.summaryText,
		Confirmed:   p.confirmed,
		Corrections: p.corrections,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling Playback: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (p *Playback) UnmarshalJSON(data []byte) error {
	type proxy struct {
		SummaryText string `json:"summary_text"`
		Confirmed   bool   `json:"confirmed"`
		Corrections string `json:"corrections"`
	}
	var pp proxy
	if err := json.Unmarshal(data, &pp); err != nil {
		return fmt.Errorf("unmarshaling Playback: %w", err)
	}
	p.summaryText = pp.SummaryText
	p.confirmed = pp.Confirmed
	p.corrections = pp.Corrections
	return nil
}
