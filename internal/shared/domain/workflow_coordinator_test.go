package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// StepStatus Value Object
// ---------------------------------------------------------------------------

func TestStepStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status   StepStatus
		expected string
	}{
		{StepPending, "pending"},
		{StepReady, "ready"},
		{StepInProgress, "in_progress"},
		{StepCompleted, "completed"},
		{StepSkipped, "skipped"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.status.String())
		})
	}
}

func TestAllStepStatuses_ReturnsAllStatuses(t *testing.T) {
	t.Parallel()

	statuses := AllStepStatuses()
	assert.Len(t, statuses, 5)
	assert.Contains(t, statuses, StepPending)
	assert.Contains(t, statuses, StepReady)
	assert.Contains(t, statuses, StepInProgress)
	assert.Contains(t, statuses, StepCompleted)
	assert.Contains(t, statuses, StepSkipped)
}

// ---------------------------------------------------------------------------
// WorkflowCoordinator — CanExecute
// ---------------------------------------------------------------------------

func TestWorkflowCoordinator_CanExecute_ReturnsFalseForPendingStep(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	// Step is pending (never marked ready)
	assert.False(t, coord.CanExecute("session-1", StepFitness))
}

func TestWorkflowCoordinator_CanExecute_ReturnsTrueForReadyStep(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	coord.MarkReady("session-1", StepFitness)

	assert.True(t, coord.CanExecute("session-1", StepFitness))
}

func TestWorkflowCoordinator_CanExecute_ReturnsFalseForInProgressStep(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	coord.MarkReady("session-1", StepFitness)
	err := coord.BeginStep("session-1", StepFitness)
	require.NoError(t, err)

	assert.False(t, coord.CanExecute("session-1", StepFitness))
}

func TestWorkflowCoordinator_CanExecute_ReturnsFalseForCompletedStep(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	coord.MarkReady("session-1", StepFitness)
	_ = coord.BeginStep("session-1", StepFitness)
	_ = coord.CompleteStep("session-1", StepFitness)

	assert.False(t, coord.CanExecute("session-1", StepFitness))
}

func TestWorkflowCoordinator_CanExecute_ReturnsFalseForSkippedStep(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	coord.MarkReady("session-1", StepFitness)
	_ = coord.SkipStep("session-1", StepFitness)

	assert.False(t, coord.CanExecute("session-1", StepFitness))
}

// ---------------------------------------------------------------------------
// WorkflowCoordinator — BeginStep
// ---------------------------------------------------------------------------

func TestWorkflowCoordinator_BeginStep_TransitionsReadyToInProgress(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	coord.MarkReady("session-1", StepFitness)

	err := coord.BeginStep("session-1", StepFitness)
	require.NoError(t, err)

	status := coord.StepStatus("session-1", StepFitness)
	assert.Equal(t, StepInProgress, status)
}

func TestWorkflowCoordinator_BeginStep_FailsIfNotReady(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	// Step is pending, not ready

	err := coord.BeginStep("session-1", StepFitness)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not ready")
}

func TestWorkflowCoordinator_BeginStep_FailsIfAlreadyInProgress(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	coord.MarkReady("session-1", StepFitness)
	_ = coord.BeginStep("session-1", StepFitness)

	err := coord.BeginStep("session-1", StepFitness)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already in progress")
}

// ---------------------------------------------------------------------------
// WorkflowCoordinator — CompleteStep
// ---------------------------------------------------------------------------

func TestWorkflowCoordinator_CompleteStep_TransitionsInProgressToCompleted(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	coord.MarkReady("session-1", StepFitness)
	_ = coord.BeginStep("session-1", StepFitness)

	err := coord.CompleteStep("session-1", StepFitness)
	require.NoError(t, err)

	status := coord.StepStatus("session-1", StepFitness)
	assert.Equal(t, StepCompleted, status)
}

func TestWorkflowCoordinator_CompleteStep_FailsIfNotInProgress(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	coord.MarkReady("session-1", StepFitness)
	// Not started yet

	err := coord.CompleteStep("session-1", StepFitness)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in progress")
}

// ---------------------------------------------------------------------------
// WorkflowCoordinator — SkipStep
// ---------------------------------------------------------------------------

func TestWorkflowCoordinator_SkipStep_TransitionsReadyToSkipped(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	coord.MarkReady("session-1", StepFitness)

	err := coord.SkipStep("session-1", StepFitness)
	require.NoError(t, err)

	status := coord.StepStatus("session-1", StepFitness)
	assert.Equal(t, StepSkipped, status)
}

func TestWorkflowCoordinator_SkipStep_FailsIfNotReady(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	// Step is pending

	err := coord.SkipStep("session-1", StepFitness)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not ready")
}

func TestWorkflowCoordinator_SkipStep_FailsIfInProgress(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	coord.MarkReady("session-1", StepFitness)
	_ = coord.BeginStep("session-1", StepFitness)

	err := coord.SkipStep("session-1", StepFitness)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "in progress")
}

// ---------------------------------------------------------------------------
// WorkflowCoordinator — SessionContext
// ---------------------------------------------------------------------------

func TestWorkflowCoordinator_SetSessionContext_StoresContext(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	ctx := &SessionContext{
		SessionID:  "session-1",
		ProjectDir: "/tmp/test-project",
	}

	err := coord.SetSessionContext("session-1", ctx)
	require.NoError(t, err)

	retrieved, err := coord.SessionContext("session-1")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/test-project", retrieved.ProjectDir)
}

func TestWorkflowCoordinator_SessionContext_ReturnsErrorIfNotFound(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()

	_, err := coord.SessionContext("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// WorkflowCoordinator — AvailableActions
// ---------------------------------------------------------------------------

func TestWorkflowCoordinator_AvailableActions_ReturnsOnlyReadySteps(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	coord.MarkReady("session-1", StepFitness, StepTickets)
	_ = coord.BeginStep("session-1", StepFitness) // Now in progress

	actions := coord.AvailableActions("session-1")
	assert.Len(t, actions, 1)
	assert.Equal(t, "tickets", actions[0].Name())
}

// ---------------------------------------------------------------------------
// WorkflowCoordinator — StepStatus query
// ---------------------------------------------------------------------------

func TestWorkflowCoordinator_StepStatus_ReturnsPendingForUnknownStep(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()

	status := coord.StepStatus("session-1", StepFitness)
	assert.Equal(t, StepPending, status)
}

// ---------------------------------------------------------------------------
// Backward Compatibility — existing methods still work
// ---------------------------------------------------------------------------

func TestWorkflowCoordinator_MarkReady_MakesStepReady(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	coord.MarkReady("session-1", StepFitness)

	status := coord.StepStatus("session-1", StepFitness)
	assert.Equal(t, StepReady, status)
}

func TestWorkflowCoordinator_IsReady_ReturnsCorrectValue(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	assert.False(t, coord.IsReady("session-1", StepFitness))

	coord.MarkReady("session-1", StepFitness)
	assert.True(t, coord.IsReady("session-1", StepFitness))
}

func TestWorkflowCoordinator_IsCompleted_ReturnsCorrectValue(t *testing.T) {
	t.Parallel()

	coord := NewWorkflowCoordinator()
	coord.MarkReady("session-1", StepFitness)
	_ = coord.BeginStep("session-1", StepFitness)

	assert.False(t, coord.IsCompleted("session-1", StepFitness))

	_ = coord.CompleteStep("session-1", StepFitness)
	assert.True(t, coord.IsCompleted("session-1", StepFitness))
}
