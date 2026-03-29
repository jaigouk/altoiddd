package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// -- Helpers --

// sessionWithModeAndPersona creates a session in StatusPersonaDetected with the given mode.
// Order: NewDiscoverySession → SetMode → DetectPersona.
func sessionWithModeAndPersona(mode DiscoveryMode) *DiscoverySession {
	session := NewDiscoverySession("A test project idea.")
	if err := session.SetMode(mode); err != nil {
		panic("SetMode failed: " + err.Error())
	}
	if err := session.DetectPersona("1"); err != nil {
		panic("DetectPersona failed: " + err.Error())
	}
	return session
}

// -- ComputeResumeCheckpoint Tests --

func TestComputeResumeCheckpoint_FreshSession_Index1(t *testing.T) {
	// Given a session with ModeRapid, 0 stories done, no boundaries
	session := sessionWithModeAndPersona(ModeRapid)

	// When we compute the resume checkpoint
	checkpoint, err := session.ComputeResumeCheckpoint()

	// Then StoryIndex=1 (next story to narrate), BoundariesDone=false
	require.NoError(t, err)
	assert.Equal(t, 1, checkpoint.StoryIndex)
	assert.False(t, checkpoint.BoundariesDone)
}

func TestComputeResumeCheckpoint_TwoStories_Index3(t *testing.T) {
	// Given a session with ModeRapid and 2 stories added
	session := sessionWithModeAndPersona(ModeRapid)
	require.NoError(t, session.AddStoryRef("story1.yaml"))
	require.NoError(t, session.AddStoryRef("story2.yaml"))

	// When we compute the resume checkpoint
	checkpoint, err := session.ComputeResumeCheckpoint()

	// Then StoryIndex=3 (StoryCount+1), BoundariesDone=false
	require.NoError(t, err)
	assert.Equal(t, 3, checkpoint.StoryIndex)
	assert.False(t, checkpoint.BoundariesDone)
}

func TestComputeResumeCheckpoint_BoundariesDone_True(t *testing.T) {
	// Given a session with ModeRapid and boundaries confirmed
	session := sessionWithModeAndPersona(ModeRapid)
	require.NoError(t, session.AddStoryRef("story1.yaml"))
	require.NoError(t, session.ConfirmBoundaries(nil))

	// When we compute the resume checkpoint
	checkpoint, err := session.ComputeResumeCheckpoint()

	// Then BoundariesDone=true
	require.NoError(t, err)
	assert.True(t, checkpoint.BoundariesDone)
}

func TestComputeResumeCheckpoint_LegacyModeExpress_Error(t *testing.T) {
	// Given a session with ModeExpress (legacy, not resumable)
	session := sessionWithModeAndPersona(ModeExpress)

	// When we compute the resume checkpoint
	_, err := session.ComputeResumeCheckpoint()

	// Then it returns ErrInvariantViolation
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestComputeResumeCheckpoint_LegacyModeDeep_Error(t *testing.T) {
	// Given a session with ModeDeep (legacy, not resumable)
	session := sessionWithModeAndPersona(ModeDeep)

	// When we compute the resume checkpoint
	_, err := session.ComputeResumeCheckpoint()

	// Then it returns ErrInvariantViolation
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestComputeResumeCheckpoint_NilMode_Error(t *testing.T) {
	// Given a session loaded from snapshot with nil mode (pre-mode legacy session).
	// FromSnapshot with mode=nil produces s.mode == nil.
	// We need StatusPersonaDetected so persona is set.
	snapshot := map[string]interface{}{
		"session_id":                  "test-nil-mode-session",
		"readme_content":              "A test project idea.",
		"status":                      "persona_detected",
		"persona":                     "developer",
		"register":                    "technical",
		"answers":                     []interface{}{},
		"skipped":                     []interface{}{},
		"playback_confirmations":      []interface{}{},
		"answers_since_last_playback": 0,
		// mode intentionally absent → s.mode == nil
	}

	session, err := FromSnapshot(snapshot)
	require.NoError(t, err)

	// When we compute the resume checkpoint
	_, err = session.ComputeResumeCheckpoint()

	// Then it returns ErrInvariantViolation (nil mode is not resumable)
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestComputeResumeCheckpoint_ModeConversational_Error(t *testing.T) {
	// Given a session with ModeConversational (not resumable)
	session := sessionWithModeAndPersona(ModeConversational)

	// When we compute the resume checkpoint
	_, err := session.ComputeResumeCheckpoint()

	// Then it returns ErrInvariantViolation
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}
