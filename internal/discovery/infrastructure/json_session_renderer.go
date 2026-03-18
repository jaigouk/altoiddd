package infrastructure

import (
	"encoding/json"
	"fmt"

	"github.com/alto-cli/alto/internal/discovery/domain"
)

// JSONSessionRenderer renders discovery session state as JSON for AI agent consumption
// and parses JSON responses from agents.
type JSONSessionRenderer struct{}

// NewJSONSessionRenderer creates a new JSONSessionRenderer.
func NewJSONSessionRenderer() *JSONSessionRenderer {
	return &JSONSessionRenderer{}
}

// personaChoiceLabels are the human-readable persona choice labels.
var personaChoiceLabels = []string{
	"Developer (technical background)",
	"Product Owner (defines what to build)",
	"Domain Expert (business knowledge)",
	"Mixed / Other",
}

// RenderQuestion serializes a discovery question as JSON for an agent.
func (r *JSONSessionRenderer) RenderQuestion(session *domain.DiscoverySession, question domain.Question) ([]byte, error) {
	register, _ := session.Register()

	var text string
	if register == domain.RegisterTechnical {
		text = question.TechnicalText()
	} else {
		text = question.NonTechnicalText()
	}

	mvp := domain.MVPQuestionIDs()

	qo := domain.QuestionOutput{
		SessionID:    session.SessionID(),
		QuestionID:   question.ID(),
		QuestionText: text,
		QuestionType: "open",
		Phase:        string(question.Phase()),
		Required:     mvp[question.ID()],
		SkipAllowed:  true,
	}

	data, err := json.Marshal(qo)
	if err != nil {
		return nil, fmt.Errorf("marshaling question output: %w", err)
	}
	return data, nil
}

// RenderPersonaPrompt serializes the persona detection prompt as JSON.
func (r *JSONSessionRenderer) RenderPersonaPrompt(session *domain.DiscoverySession) ([]byte, error) {
	pp := domain.PersonaPrompt{
		SessionID: session.SessionID(),
		Prompt:    "Which best describes you?",
		Choices:   personaChoiceLabels,
	}

	data, err := json.Marshal(pp)
	if err != nil {
		return nil, fmt.Errorf("marshaling persona prompt: %w", err)
	}
	return data, nil
}

// RenderPlaybackPrompt serializes a playback confirmation prompt as JSON.
func (r *JSONSessionRenderer) RenderPlaybackPrompt(session *domain.DiscoverySession, summary string) ([]byte, error) {
	pp := domain.PlaybackPrompt{
		SessionID: session.SessionID(),
		Summary:   summary,
	}

	data, err := json.Marshal(pp)
	if err != nil {
		return nil, fmt.Errorf("marshaling playback prompt: %w", err)
	}
	return data, nil
}

// RenderSessionStatus serializes the current session status as JSON.
func (r *JSONSessionRenderer) RenderSessionStatus(session *domain.DiscoverySession) ([]byte, error) {
	sso := domain.SessionStatusOutput{
		SessionID:      session.SessionID(),
		Status:         string(session.Status()),
		QuestionsTotal: len(domain.QuestionCatalog()),
		Answered:       len(session.Answers()),
		NextAction:     nextAction(session.Status()),
	}

	if p, ok := session.Persona(); ok {
		sso.Persona = string(p)
	}

	// Count skipped questions from snapshot approach — we need to count skipped entries.
	// The session exposes SkipReason per question, so iterate the catalog.
	skippedCount := 0
	for _, q := range domain.QuestionCatalog() {
		if session.SkipReason(q.ID()) != "" {
			skippedCount++
		}
	}
	sso.Skipped = skippedCount

	data, err := json.Marshal(sso)
	if err != nil {
		return nil, fmt.Errorf("marshaling session status: %w", err)
	}
	return data, nil
}

// ParseAnswerInput deserializes an agent's answer from JSON.
func (r *JSONSessionRenderer) ParseAnswerInput(data []byte) (*domain.AnswerInput, error) {
	var ai domain.AnswerInput
	if err := json.Unmarshal(data, &ai); err != nil {
		return nil, fmt.Errorf("parsing answer input: %w", err)
	}
	return &ai, nil
}

// ParsePersonaResponse deserializes an agent's persona selection from JSON.
func (r *JSONSessionRenderer) ParsePersonaResponse(data []byte) (*domain.PersonaResponse, error) {
	var pr domain.PersonaResponse
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, fmt.Errorf("parsing persona response: %w", err)
	}
	return &pr, nil
}

// ParsePlaybackResponse deserializes an agent's playback confirmation from JSON.
func (r *JSONSessionRenderer) ParsePlaybackResponse(data []byte) (*domain.PlaybackResponse, error) {
	var pr domain.PlaybackResponse
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, fmt.Errorf("parsing playback response: %w", err)
	}
	return &pr, nil
}

// nextAction maps session status to the next action the agent should take.
func nextAction(status domain.DiscoveryStatus) string {
	switch status {
	case domain.StatusCreated:
		return "persona_detection"
	case domain.StatusPersonaDetected, domain.StatusAnswering:
		return "answer_question"
	case domain.StatusPlaybackPending:
		return "confirm_playback"
	case domain.StatusCompleted, domain.StatusCancelled,
		domain.StatusRound1Complete, domain.StatusRound2Complete:
		return "complete"
	case domain.StatusChallenging:
		return "challenge"
	case domain.StatusSimulating:
		return "simulate"
	default:
		return "unknown"
	}
}
