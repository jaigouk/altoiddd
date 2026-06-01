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
	sharedapp "github.com/alto-cli/alto/internal/shared/application"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// --- minimal stubs (handler dependencies the test path does not exercise) ---

type stubToolDetector struct{}

func (stubToolDetector) Detect(string) ([]string, error)        { return nil, nil }
func (stubToolDetector) ScanConflicts(string) ([]string, error) { return nil, nil }

type stubFileChecker struct{}

func (stubFileChecker) Exists(string) bool { return false }

type stubFileWriter struct{}

func (stubFileWriter) WriteFile(context.Context, string, string) error { return nil }

type stubContentProvider struct{}

func (stubContentProvider) ContentFor(string, domain.ProjectConfig) string { return "" }

type stubPublisher struct{}

func (stubPublisher) Publish(context.Context, any) error { return nil }

// --- ScaffoldWriter spy ---

type spyScaffoldWriter struct {
	mu        sync.Mutex
	called    int
	gotTarget string
	gotParams domain.ScaffoldParams
	gotForce  bool
	err       error
}

func (s *spyScaffoldWriter) WriteScaffold(_ context.Context, targetDir string, params domain.ScaffoldParams, force bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.called++
	s.gotTarget = targetDir
	s.gotParams = params
	s.gotForce = force
	return s.err
}

func newHandler(opts ...bootstrapapp.BootstrapOption) *bootstrapapp.BootstrapHandler {
	return bootstrapapp.NewBootstrapHandler(
		stubToolDetector{},
		stubFileChecker{},
		stubPublisher{},
		sharedapp.FileWriter(stubFileWriter{}),
		stubContentProvider{},
		opts...,
	)
}

func newParams(t *testing.T) domain.ScaffoldParams {
	t.Helper()
	p, err := domain.NewScaffoldParams("demo", "demo-", "beads", []string{"Orders"}, "claude")
	require.NoError(t, err)
	return p
}

func TestBootstrapHandler_WriteScaffold_CallsScaffoldWriterOnce(t *testing.T) {
	t.Parallel()
	sw := &spyScaffoldWriter{}
	h := newHandler(bootstrapapp.WithScaffoldWriter(sw))

	err := h.WriteScaffold(context.Background(), "/tmp/target", newParams(t), false)
	require.NoError(t, err)
	assert.Equal(t, 1, sw.called)
	assert.Equal(t, "/tmp/target", sw.gotTarget)
	assert.Equal(t, "demo", sw.gotParams.ProjectName)
	assert.False(t, sw.gotForce)
}

func TestBootstrapHandler_WriteScaffold_ForceTrue_Propagates(t *testing.T) {
	t.Parallel()
	sw := &spyScaffoldWriter{}
	h := newHandler(bootstrapapp.WithScaffoldWriter(sw))

	err := h.WriteScaffold(context.Background(), "/tmp/target", newParams(t), true)
	require.NoError(t, err)
	assert.True(t, sw.gotForce)
}

func TestBootstrapHandler_WriteScaffold_ScaffoldWriterError_Propagates(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("disk failure")
	sw := &spyScaffoldWriter{err: sentinel}
	h := newHandler(bootstrapapp.WithScaffoldWriter(sw))

	err := h.WriteScaffold(context.Background(), "/tmp/target", newParams(t), false)
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
}

func TestBootstrapHandler_WriteScaffold_WithoutScaffoldWriter_ReturnsErrInvariantViolation(t *testing.T) {
	t.Parallel()
	h := newHandler()
	err := h.WriteScaffold(context.Background(), "/tmp/target", newParams(t), false)
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestBootstrapHandler_WriteScaffold_OpencodeWithoutGenerator_ReturnsErrInvariantViolation(t *testing.T) {
	// Contract: --primary-tool=opencode without WAG wired → wrapped ErrInvariantViolation.
	t.Parallel()
	sw := &spyScaffoldWriter{}
	h := newHandler(bootstrapapp.WithScaffoldWriter(sw))

	p, err := domain.NewScaffoldParams("demo", "demo-", "beads", nil, "opencode")
	require.NoError(t, err)

	err = h.WriteScaffold(context.Background(), "/tmp/target", p, false)
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}
