package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/dochealth/application"
	"github.com/alto-cli/alto/internal/dochealth/domain"
)

// ---------------------------------------------------------------------------
// Mock DocReview port
// ---------------------------------------------------------------------------

type stubDocReviewer struct {
	reviewableDocs    []domain.DocStatus
	reviewableDocsErr error
	markReviewedRes   domain.DocReviewResult
	markReviewedErr   error
	markAllRes        []domain.DocReviewResult
	markAllErr        error
}

func (m *stubDocReviewer) ReviewableDocs(_ context.Context, _ string) ([]domain.DocStatus, error) {
	return m.reviewableDocs, m.reviewableDocsErr
}

func (m *stubDocReviewer) MarkReviewed(_ context.Context, _ string, _ string, _ *time.Time) (domain.DocReviewResult, error) {
	return m.markReviewedRes, m.markReviewedErr
}

func (m *stubDocReviewer) MarkAllReviewed(_ context.Context, _ string, _ *time.Time) ([]domain.DocReviewResult, error) {
	return m.markAllRes, m.markAllErr
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestDocReviewHandler_ReviewableDocs_HappyPath(t *testing.T) {
	t.Parallel()
	statuses := []domain.DocStatus{
		domain.NewDocStatus("docs/PRD.md", domain.DocHealthStale, nil, nil, 30, "", nil),
	}
	mock := &stubDocReviewer{reviewableDocs: statuses}
	handler := application.NewDocReviewHandler(mock)

	result, err := handler.ReviewableDocs(context.Background(), "/project")

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "docs/PRD.md", result[0].Path())
}

func TestDocReviewHandler_ReviewableDocs_Empty(t *testing.T) {
	t.Parallel()
	mock := &stubDocReviewer{reviewableDocs: nil}
	handler := application.NewDocReviewHandler(mock)

	result, err := handler.ReviewableDocs(context.Background(), "/project")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestDocReviewHandler_ReviewableDocs_Error(t *testing.T) {
	t.Parallel()
	mock := &stubDocReviewer{reviewableDocsErr: errors.New("scan failed")}
	handler := application.NewDocReviewHandler(mock)

	_, err := handler.ReviewableDocs(context.Background(), "/project")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "scan failed")
}

func TestDocReviewHandler_MarkReviewed_HappyPath(t *testing.T) {
	t.Parallel()
	now := time.Now().Truncate(24 * time.Hour)
	mock := &stubDocReviewer{
		markReviewedRes: domain.NewDocReviewResult("docs/PRD.md", now),
	}
	handler := application.NewDocReviewHandler(mock)

	result, err := handler.MarkReviewed(context.Background(), "docs/PRD.md", "/project", nil)

	require.NoError(t, err)
	assert.Equal(t, "docs/PRD.md", result.Path())
	assert.Equal(t, now, result.NewDate())
}

func TestDocReviewHandler_MarkReviewed_DocNotFound(t *testing.T) {
	t.Parallel()
	mock := &stubDocReviewer{markReviewedErr: errors.New("file not found: nonexistent.md")}
	handler := application.NewDocReviewHandler(mock)

	_, err := handler.MarkReviewed(context.Background(), "nonexistent.md", "/project", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDocReviewHandler_MarkAllReviewed_HappyPath(t *testing.T) {
	t.Parallel()
	now := time.Now().Truncate(24 * time.Hour)
	mock := &stubDocReviewer{
		markAllRes: []domain.DocReviewResult{
			domain.NewDocReviewResult("docs/PRD.md", now),
			domain.NewDocReviewResult("docs/DDD.md", now),
		},
	}
	handler := application.NewDocReviewHandler(mock)

	results, err := handler.MarkAllReviewed(context.Background(), "/project", nil)

	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestDocReviewHandler_MarkAllReviewed_Error(t *testing.T) {
	t.Parallel()
	mock := &stubDocReviewer{markAllErr: errors.New("permission denied")}
	handler := application.NewDocReviewHandler(mock)

	_, err := handler.MarkAllReviewed(context.Background(), "/project", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}
