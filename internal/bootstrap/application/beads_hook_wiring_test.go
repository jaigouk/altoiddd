package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bootstrapapp "github.com/alto-cli/alto/internal/bootstrap/application"
	"github.com/alto-cli/alto/internal/bootstrap/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// spyBeadsHookWriter captures invocations for assertions.
type spyBeadsHookWriter struct {
	mu          sync.Mutex
	calls       int
	gotTarget   string
	gotPrimary  string
	gotForce    bool
	returnedErr error
}

func (s *spyBeadsHookWriter) WriteBeadsPostCloseHook(_ context.Context, targetDir, primaryTool string, force bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.gotTarget = targetDir
	s.gotPrimary = primaryTool
	s.gotForce = force
	return s.returnedErr
}

func newParamsWithHooks(t *testing.T) domain.ScaffoldParams {
	t.Helper()
	p, err := domain.NewScaffoldParams("demo", "demo-", "beads", []string{"Orders"}, "claude")
	require.NoError(t, err)
	return p // IncludeHooks defaults to true
}

func TestBootstrapHandler_WriteScaffold_WhenIncludeHooks_CallsHookWriter(t *testing.T) {
	t.Parallel()

	sw := &spyScaffoldWriter{}
	hw := &spyBeadsHookWriter{}
	h := newHandler(
		bootstrapapp.WithScaffoldWriter(sw),
		bootstrapapp.WithBeadsHookWriter(hw),
	)

	err := h.WriteScaffold(context.Background(), "/tmp/target", newParamsWithHooks(t), false)

	require.NoError(t, err)
	assert.Equal(t, 1, hw.calls)
	assert.Equal(t, "/tmp/target", hw.gotTarget)
	assert.Equal(t, "claude", hw.gotPrimary)
	assert.False(t, hw.gotForce, "default forceHooks is false")
}

func TestBootstrapHandler_WriteScaffold_WhenNoHooks_SkipsHookWriter(t *testing.T) {
	t.Parallel()

	sw := &spyScaffoldWriter{}
	hw := &spyBeadsHookWriter{}
	h := newHandler(
		bootstrapapp.WithScaffoldWriter(sw),
		bootstrapapp.WithBeadsHookWriter(hw),
	)
	params := newParamsWithHooks(t).WithIncludeHooks(false)

	err := h.WriteScaffold(context.Background(), "/tmp/target", params, false)

	require.NoError(t, err)
	assert.Zero(t, hw.calls, "hook writer must not be invoked when IncludeHooks=false")
}

func TestBootstrapHandler_WriteScaffold_WhenHookWriterMissing_ReturnsInvariantViolation(t *testing.T) {
	t.Parallel()

	sw := &spyScaffoldWriter{}
	h := newHandler(bootstrapapp.WithScaffoldWriter(sw))

	err := h.WriteScaffold(context.Background(), "/tmp/target", newParamsWithHooks(t), false)

	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestBootstrapHandler_WriteScaffold_WhenHookWriterErrors_WrapsError(t *testing.T) {
	t.Parallel()

	sw := &spyScaffoldWriter{}
	sentinel := errors.New("disk full")
	hw := &spyBeadsHookWriter{returnedErr: sentinel}
	h := newHandler(
		bootstrapapp.WithScaffoldWriter(sw),
		bootstrapapp.WithBeadsHookWriter(hw),
	)

	err := h.WriteScaffold(context.Background(), "/tmp/target", newParamsWithHooks(t), false)

	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	assert.Contains(t, err.Error(), "writing beads post-close hook")
}

func TestBootstrapHandler_SetForceHooks_ThreadsThroughToWriter(t *testing.T) {
	t.Parallel()

	sw := &spyScaffoldWriter{}
	hw := &spyBeadsHookWriter{}
	h := newHandler(
		bootstrapapp.WithScaffoldWriter(sw),
		bootstrapapp.WithBeadsHookWriter(hw),
	)
	h.SetForceHooks(true)

	err := h.WriteScaffold(context.Background(), "/tmp/target", newParamsWithHooks(t), false)

	require.NoError(t, err)
	assert.True(t, hw.gotForce)
}
