package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/research/application"
	"github.com/alto-cli/alto/internal/research/domain"
)

// ---------------------------------------------------------------------------
// Mock SpikeFollowUp port
// ---------------------------------------------------------------------------

type stubSpikeFollowUp struct {
	auditResult domain.FollowUpAuditResult
	auditErr    error
}

func (m *stubSpikeFollowUp) Audit(_ context.Context, _ string, _ string) (domain.FollowUpAuditResult, error) {
	return m.auditResult, m.auditErr
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestSpikeFollowUpHandler_Audit_HappyPath(t *testing.T) {
	t.Parallel()
	intent, err := domain.NewFollowUpIntent("Create login feature", "User auth")
	require.NoError(t, err)
	result := domain.NewFollowUpAuditResult(
		"spike-1", "docs/research/spike-1.md",
		[]domain.FollowUpIntent{intent},
		[]string{"ticket-1"},
		nil,
	)
	mock := &stubSpikeFollowUp{auditResult: result}
	handler := application.NewSpikeFollowUpHandler(mock)

	auditResult, err := handler.Audit(context.Background(), "spike-1", "/project")

	require.NoError(t, err)
	assert.Equal(t, "spike-1", auditResult.SpikeID())
	assert.Equal(t, 1, auditResult.DefinedCount())
	assert.False(t, auditResult.HasOrphans())
}

func TestSpikeFollowUpHandler_Audit_WithOrphans(t *testing.T) {
	t.Parallel()
	intent1, _ := domain.NewFollowUpIntent("Create login feature", "")
	intent2, _ := domain.NewFollowUpIntent("Add rate limiting", "")
	result := domain.NewFollowUpAuditResult(
		"spike-2", "docs/research/spike-2.md",
		[]domain.FollowUpIntent{intent1, intent2},
		[]string{"ticket-1"},
		[]domain.FollowUpIntent{intent2},
	)
	mock := &stubSpikeFollowUp{auditResult: result}
	handler := application.NewSpikeFollowUpHandler(mock)

	auditResult, err := handler.Audit(context.Background(), "spike-2", "/project")

	require.NoError(t, err)
	assert.True(t, auditResult.HasOrphans())
	assert.Equal(t, 1, auditResult.OrphanedCount())
}

func TestSpikeFollowUpHandler_Audit_SpikeNotFound(t *testing.T) {
	t.Parallel()
	// When spike has no research dir, adapter returns empty result (not error).
	result := domain.NewFollowUpAuditResult("bad-id", "", nil, nil, nil)
	mock := &stubSpikeFollowUp{auditResult: result}
	handler := application.NewSpikeFollowUpHandler(mock)

	auditResult, err := handler.Audit(context.Background(), "bad-id", "/project")

	require.NoError(t, err)
	assert.Equal(t, 0, auditResult.DefinedCount())
}

func TestSpikeFollowUpHandler_Audit_Error(t *testing.T) {
	t.Parallel()
	mock := &stubSpikeFollowUp{auditErr: errors.New("filesystem error")}
	handler := application.NewSpikeFollowUpHandler(mock)

	_, err := handler.Audit(context.Background(), "spike-1", "/project")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "filesystem error")
}
