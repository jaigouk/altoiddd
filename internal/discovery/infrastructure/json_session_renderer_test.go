package infrastructure_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
)

func newAnsweringSession(t *testing.T) *domain.DiscoverySession {
	t.Helper()
	session := domain.NewDiscoverySession("# Test Project")
	require.NoError(t, session.DetectPersona("1")) // developer
	return session
}

// --- RenderQuestion ---

func TestJSONSessionRenderer_RenderQuestion(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()
	session := newAnsweringSession(t)
	catalog := domain.QuestionCatalog()
	question := catalog[0] // Q1

	data, err := renderer.RenderQuestion(session, question)
	require.NoError(t, err)

	var qo domain.QuestionOutput
	err = json.Unmarshal(data, &qo)
	require.NoError(t, err)

	assert.Equal(t, session.SessionID(), qo.SessionID)
	assert.Equal(t, "Q1", qo.QuestionID)
	assert.Equal(t, question.TechnicalText(), qo.QuestionText)
	assert.Equal(t, "open", qo.QuestionType)
	assert.Equal(t, "actors", qo.Phase)
	assert.True(t, qo.Required) // Q1 is MVP
	assert.True(t, qo.SkipAllowed)
	assert.Empty(t, qo.Choices)
}

func TestJSONSessionRenderer_RenderQuestion_NonTechnicalRegister(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()
	session := domain.NewDiscoverySession("# Test")
	require.NoError(t, session.DetectPersona("2")) // product_owner -> non_technical

	catalog := domain.QuestionCatalog()
	data, err := renderer.RenderQuestion(session, catalog[0])
	require.NoError(t, err)

	var qo domain.QuestionOutput
	require.NoError(t, json.Unmarshal(data, &qo))
	assert.Equal(t, catalog[0].NonTechnicalText(), qo.QuestionText)
}

// --- RenderPersonaPrompt ---

func TestJSONSessionRenderer_RenderPersonaPrompt(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()
	session := domain.NewDiscoverySession("# Test Project")

	data, err := renderer.RenderPersonaPrompt(session)
	require.NoError(t, err)

	var pp domain.PersonaPrompt
	err = json.Unmarshal(data, &pp)
	require.NoError(t, err)

	assert.Equal(t, session.SessionID(), pp.SessionID)
	assert.Equal(t, "Which best describes you?", pp.Prompt)
	assert.Len(t, pp.Choices, 4)
	assert.Equal(t, "Developer (technical background)", pp.Choices[0])
	assert.Equal(t, "Product Owner (defines what to build)", pp.Choices[1])
	assert.Equal(t, "Domain Expert (business knowledge)", pp.Choices[2])
	assert.Equal(t, "Mixed / Other", pp.Choices[3])
}

// --- RenderPlaybackPrompt ---

func TestJSONSessionRenderer_RenderPlaybackPrompt(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()
	session := newAnsweringSession(t)
	summary := "Q: Who are the actors?\nA: developers"

	data, err := renderer.RenderPlaybackPrompt(session, summary)
	require.NoError(t, err)

	var pp domain.PlaybackPrompt
	err = json.Unmarshal(data, &pp)
	require.NoError(t, err)

	assert.Equal(t, session.SessionID(), pp.SessionID)
	assert.Equal(t, summary, pp.Summary)
}

// --- RenderSessionStatus ---

func TestJSONSessionRenderer_RenderSessionStatus_Created(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()
	session := domain.NewDiscoverySession("# Test")

	data, err := renderer.RenderSessionStatus(session)
	require.NoError(t, err)

	var sso domain.SessionStatusOutput
	require.NoError(t, json.Unmarshal(data, &sso))

	assert.Equal(t, session.SessionID(), sso.SessionID)
	assert.Equal(t, "created", sso.Status)
	assert.Empty(t, sso.Persona)
	assert.Equal(t, 10, sso.QuestionsTotal)
	assert.Equal(t, 0, sso.Answered)
	assert.Equal(t, 0, sso.Skipped)
	assert.Equal(t, "persona_detection", sso.NextAction)
}

func TestJSONSessionRenderer_RenderSessionStatus_Answering(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()
	session := newAnsweringSession(t)
	require.NoError(t, session.AnswerQuestion("Q1", "developers"))
	require.NoError(t, session.AnswerQuestion("Q2", "orders, products"))

	data, err := renderer.RenderSessionStatus(session)
	require.NoError(t, err)

	var sso domain.SessionStatusOutput
	require.NoError(t, json.Unmarshal(data, &sso))

	assert.Equal(t, "answering", sso.Status)
	assert.Equal(t, "developer", sso.Persona)
	assert.Equal(t, 2, sso.Answered)
	assert.Equal(t, 0, sso.Skipped)
	assert.Equal(t, "answer_question", sso.NextAction)
}

func TestJSONSessionRenderer_RenderSessionStatus_PlaybackPending(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()
	session := newAnsweringSession(t)
	require.NoError(t, session.AnswerQuestion("Q1", "developers"))
	require.NoError(t, session.AnswerQuestion("Q2", "orders"))
	require.NoError(t, session.AnswerQuestion("Q3", "user places order"))

	assert.Equal(t, domain.StatusPlaybackPending, session.Status())

	data, err := renderer.RenderSessionStatus(session)
	require.NoError(t, err)

	var sso domain.SessionStatusOutput
	require.NoError(t, json.Unmarshal(data, &sso))

	assert.Equal(t, "playback_pending", sso.Status)
	assert.Equal(t, 3, sso.Answered)
	assert.Equal(t, "confirm_playback", sso.NextAction)
}

func TestJSONSessionRenderer_RenderSessionStatus_WithSkipped(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()
	session := newAnsweringSession(t)
	require.NoError(t, session.AnswerQuestion("Q1", "developers"))
	require.NoError(t, session.SkipQuestion("Q2", "not sure yet"))

	data, err := renderer.RenderSessionStatus(session)
	require.NoError(t, err)

	var sso domain.SessionStatusOutput
	require.NoError(t, json.Unmarshal(data, &sso))

	assert.Equal(t, 1, sso.Answered)
	assert.Equal(t, 1, sso.Skipped)
}

// --- ParseAnswerInput ---

func TestJSONSessionRenderer_ParseAnswerInput(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()

	tests := []struct {
		name string
		ai   domain.AnswerInput
	}{
		{
			name: "with answer",
			ai: domain.AnswerInput{
				SessionID:  "sess-1",
				QuestionID: "Q1",
				Answer:     "developers and API clients",
			},
		},
		{
			name: "skipped",
			ai: domain.AnswerInput{
				SessionID:  "sess-1",
				QuestionID: "Q5",
				Skipped:    true,
				SkipReason: "not applicable",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.ai)
			require.NoError(t, err)

			got, err := renderer.ParseAnswerInput(data)
			require.NoError(t, err)
			assert.Equal(t, tt.ai, *got)
		})
	}
}

func TestJSONSessionRenderer_ParseAnswerInput_InvalidJSON(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()
	_, err := renderer.ParseAnswerInput([]byte(`{invalid`))
	require.Error(t, err)
}

// --- ParsePersonaResponse ---

func TestJSONSessionRenderer_ParsePersonaResponse(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()
	pr := domain.PersonaResponse{SessionID: "sess-1", Choice: "2"}
	data, err := json.Marshal(pr)
	require.NoError(t, err)

	got, err := renderer.ParsePersonaResponse(data)
	require.NoError(t, err)
	assert.Equal(t, pr, *got)
}

func TestJSONSessionRenderer_ParsePersonaResponse_InvalidJSON(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()
	_, err := renderer.ParsePersonaResponse([]byte(`not json`))
	require.Error(t, err)
}

// --- ParsePlaybackResponse ---

func TestJSONSessionRenderer_ParsePlaybackResponse(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()

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
			pr := domain.PlaybackResponse{SessionID: "sess-1", Confirmed: tt.confirmed}
			data, err := json.Marshal(pr)
			require.NoError(t, err)

			got, err := renderer.ParsePlaybackResponse(data)
			require.NoError(t, err)
			assert.Equal(t, pr, *got)
		})
	}
}

func TestJSONSessionRenderer_ParsePlaybackResponse_InvalidJSON(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewJSONSessionRenderer()
	_, err := renderer.ParsePlaybackResponse([]byte(`{`))
	require.Error(t, err)
}
