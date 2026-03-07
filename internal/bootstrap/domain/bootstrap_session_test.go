package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

func makePreview() vo.Preview {
	return vo.NewPreview(
		[]vo.FileAction{vo.NewFileAction("docs/PRD.md", vo.FileActionCreate, "", "")},
		nil, nil,
	)
}

// -- Creation --

func TestNewBootstrapSessionStartsInCreatedState(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	assert.Equal(t, SessionStatusCreated, session.Status())
}

func TestNewBootstrapSessionHasUniqueID(t *testing.T) {
	t.Parallel()
	s1 := NewBootstrapSession("/tmp/a")
	s2 := NewBootstrapSession("/tmp/b")
	assert.NotEqual(t, s1.SessionID(), s2.SessionID())
}

func TestNewBootstrapSessionHasProjectDir(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	assert.Equal(t, "/tmp/proj", session.ProjectDir())
}

// -- State Transitions --

func TestSetPreviewTransitionsToPreviewed(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	assert.Equal(t, SessionStatusPreviewed, session.Status())
}

func TestConfirmRequiresPreviewFirst(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	err := session.Confirm()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "cannot confirm without preview")
}

func TestConfirmTransitionsToConfirmed(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	require.NoError(t, session.Confirm())
	assert.Equal(t, SessionStatusConfirmed, session.Status())
}

func TestCancelTransitionsToCancelled(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	require.NoError(t, session.Cancel())
	assert.Equal(t, SessionStatusCancelled, session.Status())
}

func TestCancelFromCreatedReturnsError(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	err := session.Cancel()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "can only cancel from previewed state")
}

func TestBeginExecutionRequiresConfirmation(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	err := session.BeginExecution()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "cannot execute without confirmation")
}

func TestBeginExecutionTransitionsToExecuting(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	require.NoError(t, session.Confirm())
	require.NoError(t, session.BeginExecution())
	assert.Equal(t, SessionStatusExecuting, session.Status())
}

func TestCompleteTransitionsToCompleted(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	require.NoError(t, session.Confirm())
	require.NoError(t, session.BeginExecution())
	require.NoError(t, session.Complete())
	assert.Equal(t, SessionStatusCompleted, session.Status())
}

func TestCompleteProducesBootstrapCompletedEvent(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	require.NoError(t, session.Confirm())
	require.NoError(t, session.BeginExecution())
	require.NoError(t, session.Complete())
	assert.Len(t, session.Events(), 1)
	event := session.Events()[0]
	assert.Equal(t, session.SessionID(), event.SessionID())
	assert.Equal(t, "/tmp/proj", event.ProjectDir())
}

// -- Invariants --

func TestPreviewTwiceReplacesPreview(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p1 := makePreview()
	p2 := vo.NewPreview(
		[]vo.FileAction{vo.NewFileAction("docs/DDD.md", vo.FileActionCreate, "", "")},
		nil, nil,
	)
	require.NoError(t, session.SetPreview(&p1))
	require.NoError(t, session.SetPreview(&p2))
	assert.Equal(t, SessionStatusPreviewed, session.Status())
}

func TestCannotConfirmCancelledSession(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	require.NoError(t, session.Cancel())
	err := session.Confirm()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot confirm without preview")
}

func TestCannotExecuteCancelledSession(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	require.NoError(t, session.Cancel())
	err := session.BeginExecution()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot execute without confirmation")
}

func TestCannotPreviewFromConfirmedState(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	require.NoError(t, session.Confirm())
	err := session.SetPreview(&p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot preview in confirmed state")
}

func TestCannotPreviewFromExecutingState(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	require.NoError(t, session.Confirm())
	require.NoError(t, session.BeginExecution())
	err := session.SetPreview(&p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot preview in executing state")
}

func TestCannotPreviewFromCompletedState(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	require.NoError(t, session.Confirm())
	require.NoError(t, session.BeginExecution())
	require.NoError(t, session.Complete())
	err := session.SetPreview(&p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot preview in completed state")
}

func TestDoubleCompleteReturnsError(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	require.NoError(t, session.Confirm())
	require.NoError(t, session.BeginExecution())
	require.NoError(t, session.Complete())
	err := session.Complete()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot complete unless executing")
}

func TestEventsReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	p := makePreview()
	require.NoError(t, session.SetPreview(&p))
	require.NoError(t, session.Confirm())
	require.NoError(t, session.BeginExecution())
	require.NoError(t, session.Complete())
	events := session.Events()
	events = events[:0]
	_ = events
	assert.Len(t, session.Events(), 1)
}

// -- Detected Tools --

func TestNewSessionHasEmptyDetectedTools(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	assert.Empty(t, session.DetectedTools())
}

func TestSetDetectedToolsStoresTools(t *testing.T) {
	t.Parallel()
	session := NewBootstrapSession("/tmp/proj")
	session.SetDetectedTools([]string{"claude", "cursor"})
	assert.Equal(t, []string{"claude", "cursor"}, session.DetectedTools())
}
