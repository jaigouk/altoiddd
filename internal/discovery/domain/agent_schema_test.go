package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/discovery/domain"
)

// --- QuestionOutput ---

func TestQuestionOutput_RoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		qo   domain.QuestionOutput
	}{
		{
			name: "open question without choices",
			qo: domain.QuestionOutput{
				SessionID:    "sess-1",
				QuestionID:   "Q1",
				QuestionText: "Who are the actors?",
				QuestionType: "open",
				Phase:        "actors",
				Required:     true,
				SkipAllowed:  false,
			},
		},
		{
			name: "choice question with choices",
			qo: domain.QuestionOutput{
				SessionID:    "sess-2",
				QuestionID:   "Q10",
				QuestionText: "Classify each context",
				QuestionType: "choice",
				Phase:        "boundaries",
				Required:     false,
				SkipAllowed:  true,
				Choices:      []string{"core", "supporting", "generic"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.qo)
			require.NoError(t, err)

			var got domain.QuestionOutput
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)
			assert.Equal(t, tt.qo, got)
		})
	}
}

func TestQuestionOutput_OmitsEmptyChoices(t *testing.T) {
	t.Parallel()
	qo := domain.QuestionOutput{
		SessionID:    "sess-1",
		QuestionID:   "Q1",
		QuestionText: "Who are the actors?",
		QuestionType: "open",
		Phase:        "actors",
		Required:     true,
		SkipAllowed:  false,
	}
	data, err := json.Marshal(qo)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "choices")
}

// --- AnswerInput ---

func TestAnswerInput_RoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ai   domain.AnswerInput
	}{
		{
			name: "answered question",
			ai: domain.AnswerInput{
				SessionID:  "sess-1",
				QuestionID: "Q1",
				Answer:     "developers and API clients",
			},
		},
		{
			name: "skipped question",
			ai: domain.AnswerInput{
				SessionID:  "sess-1",
				QuestionID: "Q5",
				Skipped:    true,
				SkipReason: "not applicable yet",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.ai)
			require.NoError(t, err)

			var got domain.AnswerInput
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)
			assert.Equal(t, tt.ai, got)
		})
	}
}

func TestAnswerInput_OmitsEmptyOptionals(t *testing.T) {
	t.Parallel()
	ai := domain.AnswerInput{
		SessionID:  "sess-1",
		QuestionID: "Q1",
		Answer:     "some answer",
	}
	data, err := json.Marshal(ai)
	require.NoError(t, err)
	raw := string(data)
	assert.NotContains(t, raw, "skip_reason")
	assert.Contains(t, raw, `"skipped":false`)
}

// --- PersonaPrompt ---

func TestPersonaPrompt_RoundTrip(t *testing.T) {
	t.Parallel()
	pp := domain.PersonaPrompt{
		SessionID: "sess-1",
		Prompt:    "Which best describes you?",
		Choices: []string{
			"Developer (technical background)",
			"Product Owner (defines what to build)",
			"Domain Expert (business knowledge)",
			"Mixed / Other",
		},
	}
	data, err := json.Marshal(pp)
	require.NoError(t, err)

	var got domain.PersonaPrompt
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)
	assert.Equal(t, pp, got)
}

// --- PersonaResponse ---

func TestPersonaResponse_RoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		choice string
	}{
		{"developer", "1"},
		{"product owner", "2"},
		{"domain expert", "3"},
		{"mixed", "4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pr := domain.PersonaResponse{
				SessionID: "sess-1",
				Choice:    tt.choice,
			}
			data, err := json.Marshal(pr)
			require.NoError(t, err)

			var got domain.PersonaResponse
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)
			assert.Equal(t, pr, got)
		})
	}
}

// --- PlaybackPrompt ---

func TestPlaybackPrompt_RoundTrip(t *testing.T) {
	t.Parallel()
	pp := domain.PlaybackPrompt{
		SessionID: "sess-1",
		Summary:   "Q: Who are the actors?\nA: developers\n",
	}
	data, err := json.Marshal(pp)
	require.NoError(t, err)

	var got domain.PlaybackPrompt
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)
	assert.Equal(t, pp, got)
}

// --- PlaybackResponse ---

func TestPlaybackResponse_RoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		confirmed bool
	}{
		{"confirmed", true},
		{"rejected", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pr := domain.PlaybackResponse{
				SessionID: "sess-1",
				Confirmed: tt.confirmed,
			}
			data, err := json.Marshal(pr)
			require.NoError(t, err)

			var got domain.PlaybackResponse
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)
			assert.Equal(t, pr, got)
		})
	}
}

// --- SessionStatusOutput ---

func TestSessionStatusOutput_RoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		sso  domain.SessionStatusOutput
	}{
		{
			name: "initial state",
			sso: domain.SessionStatusOutput{
				SessionID:      "sess-1",
				Status:         "created",
				QuestionsTotal: 10,
				Answered:       0,
				Skipped:        0,
				NextAction:     "persona_detection",
			},
		},
		{
			name: "mid-session with persona",
			sso: domain.SessionStatusOutput{
				SessionID:      "sess-2",
				Status:         "answering",
				Persona:        "developer",
				QuestionsTotal: 10,
				Answered:       3,
				Skipped:        1,
				NextAction:     "answer_question",
			},
		},
		{
			name: "playback pending",
			sso: domain.SessionStatusOutput{
				SessionID:      "sess-3",
				Status:         "playback_pending",
				Persona:        "product_owner",
				QuestionsTotal: 10,
				Answered:       3,
				Skipped:        0,
				NextAction:     "confirm_playback",
			},
		},
		{
			name: "completed",
			sso: domain.SessionStatusOutput{
				SessionID:      "sess-4",
				Status:         "completed",
				Persona:        "mixed",
				QuestionsTotal: 10,
				Answered:       8,
				Skipped:        2,
				NextAction:     "complete",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.sso)
			require.NoError(t, err)

			var got domain.SessionStatusOutput
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)
			assert.Equal(t, tt.sso, got)
		})
	}
}

func TestSessionStatusOutput_OmitsEmptyPersona(t *testing.T) {
	t.Parallel()
	sso := domain.SessionStatusOutput{
		SessionID:      "sess-1",
		Status:         "created",
		QuestionsTotal: 10,
		NextAction:     "persona_detection",
	}
	data, err := json.Marshal(sso)
	require.NoError(t, err)
	// The "persona" key should not appear when empty (omitempty).
	// Note: "persona_detection" in next_action is expected; we check the key pattern.
	assert.NotContains(t, string(data), `"persona":`)
}
