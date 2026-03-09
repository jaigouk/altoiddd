package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// WorkflowStep Value Object
// ---------------------------------------------------------------------------

func TestWorkflowStep_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		step     WorkflowStep
		expected string
	}{
		{StepArtifactGeneration, "artifact_generation"},
		{StepFitness, "fitness"},
		{StepTickets, "tickets"},
		{StepConfigs, "configs"},
		{StepRippleReview, "ripple_review"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.step.String())
		})
	}
}

func TestAllWorkflowSteps_ReturnsAllSteps(t *testing.T) {
	t.Parallel()

	steps := AllWorkflowSteps()
	assert.Len(t, steps, 5)
	assert.Contains(t, steps, StepArtifactGeneration)
	assert.Contains(t, steps, StepFitness)
	assert.Contains(t, steps, StepTickets)
	assert.Contains(t, steps, StepConfigs)
	assert.Contains(t, steps, StepRippleReview)
}

// ---------------------------------------------------------------------------
// SessionTracker Aggregate
// ---------------------------------------------------------------------------

func TestNewSessionTracker_CreatesEmptyTracker(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	require.NotNil(t, tracker)
}

func TestSessionTracker_MarkReady_MakesActionAvailable(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	tracker.MarkReady("session-1", StepArtifactGeneration)

	actions := tracker.ReadyActions("session-1")
	assert.Len(t, actions, 1)
	assert.Equal(t, StepArtifactGeneration.String(), actions[0].Name())
}

func TestSessionTracker_MarkReady_MultipleSteps(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	tracker.MarkReady("session-1", StepFitness, StepTickets, StepConfigs)

	actions := tracker.ReadyActions("session-1")
	assert.Len(t, actions, 3)

	names := make([]string, 0, len(actions))
	for _, a := range actions {
		names = append(names, a.Name())
	}
	assert.Contains(t, names, "fitness")
	assert.Contains(t, names, "tickets")
	assert.Contains(t, names, "configs")
}

func TestSessionTracker_MarkCompleted_RemovesFromReady(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	tracker.MarkReady("session-1", StepFitness)
	tracker.MarkCompleted("session-1", StepFitness)

	actions := tracker.ReadyActions("session-1")
	assert.Empty(t, actions)
}

func TestSessionTracker_ReadyActions_UnknownSession_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	actions := tracker.ReadyActions("nonexistent")
	assert.Empty(t, actions)
}

func TestSessionTracker_MarkReady_Idempotent(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	tracker.MarkReady("session-1", StepFitness)
	tracker.MarkReady("session-1", StepFitness) // duplicate

	actions := tracker.ReadyActions("session-1")
	assert.Len(t, actions, 1) // still just one
}

func TestSessionTracker_MarkCompleted_BeforeReady_NoOp(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	// Completing a step that was never marked ready should not panic or error
	tracker.MarkCompleted("session-1", StepFitness)

	actions := tracker.ReadyActions("session-1")
	assert.Empty(t, actions)
}

func TestSessionTracker_MultipleSessions_Independent(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	tracker.MarkReady("session-1", StepFitness)
	tracker.MarkReady("session-2", StepTickets)

	actions1 := tracker.ReadyActions("session-1")
	actions2 := tracker.ReadyActions("session-2")

	assert.Len(t, actions1, 1)
	assert.Len(t, actions2, 1)
	assert.Equal(t, "fitness", actions1[0].Name())
	assert.Equal(t, "tickets", actions2[0].Name())
}

func TestSessionTracker_AllStepsCompleted_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	tracker.MarkReady("session-1", StepArtifactGeneration, StepFitness)
	tracker.MarkCompleted("session-1", StepArtifactGeneration)
	tracker.MarkCompleted("session-1", StepFitness)

	actions := tracker.ReadyActions("session-1")
	assert.Empty(t, actions)
}

// ---------------------------------------------------------------------------
// ReadyAction Value Object
// ---------------------------------------------------------------------------

func TestReadyAction_Name(t *testing.T) {
	t.Parallel()

	action := NewReadyAction("test_action")
	assert.Equal(t, "test_action", action.Name())
}

func TestReadyAction_Equality(t *testing.T) {
	t.Parallel()

	a1 := NewReadyAction("fitness")
	a2 := NewReadyAction("fitness")
	a3 := NewReadyAction("tickets")

	assert.Equal(t, a1, a2)
	assert.NotEqual(t, a1, a3)
}

// ---------------------------------------------------------------------------
// Thread Safety (basic check)
// ---------------------------------------------------------------------------

func TestSessionTracker_ConcurrentAccess_NoRace(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	done := make(chan struct{})

	// Launch multiple goroutines doing concurrent operations
	for i := 0; i < 10; i++ {
		go func(id int) {
			sessionID := "session-concurrent"
			tracker.MarkReady(sessionID, StepFitness)
			_ = tracker.ReadyActions(sessionID)
			tracker.MarkCompleted(sessionID, StepFitness)
			done <- struct{}{}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	// If we get here without race detector complaining, test passes
}
