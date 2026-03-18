package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

func TestDiscoveryCompletedEventCreation(t *testing.T) {
	t.Parallel()
	event := NewDiscoveryCompletedEvent(
		"abc-123",
		PersonaDeveloper,
		RegisterTechnical,
		[]Answer{NewAnswer("Q1", "Users")},
		[]Playback{NewPlayback("Sum", true, "")},
		nil,
	)
	assert.Equal(t, "abc-123", event.SessionID())
	assert.Equal(t, PersonaDeveloper, event.Persona())
	assert.Equal(t, RegisterTechnical, event.Register())
	assert.Len(t, event.Answers(), 1)
	assert.Len(t, event.PlaybackConfirmations(), 1)
	assert.Nil(t, event.TechStack())
}

func TestDiscoveryCompletedEventWithTechStack(t *testing.T) {
	t.Parallel()
	ts := vo.NewTechStack("python", "uv")
	event := NewDiscoveryCompletedEvent(
		"abc-123",
		PersonaDeveloper,
		RegisterTechnical,
		nil,
		nil,
		&ts,
	)
	assert.NotNil(t, event.TechStack())
	assert.Equal(t, "python", event.TechStack().Language())
}

func TestDiscoveryCompletedEventEquality(t *testing.T) {
	t.Parallel()
	answers := []Answer{NewAnswer("Q1", "Users")}
	e1 := NewDiscoveryCompletedEvent("abc-123", PersonaDeveloper, RegisterTechnical, answers, nil, nil)
	e2 := NewDiscoveryCompletedEvent("abc-123", PersonaDeveloper, RegisterTechnical, answers, nil, nil)
	assert.True(t, e1.Equal(e2))
}

func TestDiscoveryCompletedEventAnswersDefensiveCopy(t *testing.T) {
	t.Parallel()
	answers := []Answer{NewAnswer("Q1", "Users")}
	event := NewDiscoveryCompletedEvent("abc-123", PersonaDeveloper, RegisterTechnical, answers, nil, nil)
	got := event.Answers()
	got[0] = NewAnswer("mutated", "mutated")
	assert.Equal(t, "Q1", event.Answers()[0].QuestionID())
}

func TestDiscoveryCompletedEvent_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	original := NewDiscoveryCompletedEvent(
		"sess-rt",
		PersonaDeveloper,
		RegisterTechnical,
		[]Answer{NewAnswer("Q1", "Users")},
		[]Playback{NewPlayback("Sum", true, "")},
		nil,
	)

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"session_id"`)
	assert.Contains(t, string(data), `"sess-rt"`)

	var restored DiscoveryCompletedEvent
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, original.SessionID(), restored.SessionID())
	assert.Equal(t, original.Persona(), restored.Persona())
	assert.Equal(t, original.Register(), restored.Register())
	require.Len(t, restored.Answers(), 1)
	assert.Equal(t, "Q1", restored.Answers()[0].QuestionID())
	assert.Equal(t, "Users", restored.Answers()[0].ResponseText())
	require.Len(t, restored.PlaybackConfirmations(), 1)
	assert.Equal(t, "Sum", restored.PlaybackConfirmations()[0].SummaryText())
	assert.True(t, restored.PlaybackConfirmations()[0].Confirmed())
}

func TestDiscoveryCompletedEvent_JSONRoundtrip_WithTechStack(t *testing.T) {
	t.Parallel()

	ts := vo.NewTechStack("Go", "go modules")
	original := NewDiscoveryCompletedEvent(
		"sess-ts", PersonaDeveloper, RegisterTechnical,
		nil, nil, &ts,
	)

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored DiscoveryCompletedEvent
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	require.NotNil(t, restored.TechStack())
	assert.Equal(t, "Go", restored.TechStack().Language())
	assert.Equal(t, "go modules", restored.TechStack().PackageManager())
}
