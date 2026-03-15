package domain

// QuestionOutput is the JSON-serializable representation of a discovery question
// emitted for an AI agent to consume.
type QuestionOutput struct {
	SessionID    string   `json:"session_id"`
	QuestionID   string   `json:"question_id"`
	QuestionText string   `json:"question_text"`
	QuestionType string   `json:"question_type"`
	Phase        string   `json:"phase"`
	Required     bool     `json:"required"`
	SkipAllowed  bool     `json:"skip_allowed"`
	Choices      []string `json:"choices,omitempty"`
}

// AnswerInput is the JSON-serializable representation of an agent's answer
// to a discovery question.
type AnswerInput struct {
	SessionID  string `json:"session_id"`
	QuestionID string `json:"question_id"`
	Answer     string `json:"answer,omitempty"`
	Skipped    bool   `json:"skipped"`
	SkipReason string `json:"skip_reason,omitempty"`
}

// PersonaPrompt is the JSON-serializable persona detection prompt
// emitted at the start of a discovery session.
type PersonaPrompt struct {
	SessionID string   `json:"session_id"`
	Prompt    string   `json:"prompt"`
	Choices   []string `json:"choices"`
}

// PersonaResponse is the JSON-serializable agent response to a persona prompt.
type PersonaResponse struct {
	SessionID string `json:"session_id"`
	Choice    string `json:"choice"`
}

// PlaybackPrompt is the JSON-serializable playback confirmation prompt
// emitted after every 3 answered questions.
type PlaybackPrompt struct {
	SessionID string `json:"session_id"`
	Summary   string `json:"summary"`
}

// PlaybackResponse is the JSON-serializable agent response to a playback prompt.
type PlaybackResponse struct {
	SessionID string `json:"session_id"`
	Confirmed bool   `json:"confirmed"`
}

// SessionStatusOutput is the JSON-serializable session status
// providing progress and next-action guidance to an AI agent.
type SessionStatusOutput struct {
	SessionID      string `json:"session_id"`
	Status         string `json:"status"`
	Persona        string `json:"persona,omitempty"`
	QuestionsTotal int    `json:"questions_total"`
	Answered       int    `json:"answered"`
	Skipped        int    `json:"skipped"`
	NextAction     string `json:"next_action"`
}
