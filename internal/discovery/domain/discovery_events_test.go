package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
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
