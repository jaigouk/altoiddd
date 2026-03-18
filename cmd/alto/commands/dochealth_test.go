package commands_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/cmd/alto/commands"
	"github.com/alto-cli/alto/internal/composition"
	dochealthapp "github.com/alto-cli/alto/internal/dochealth/application"
	"github.com/alto-cli/alto/internal/dochealth/domain"
)

// mockDocReview implements dochealthapp.DocReview for testing.
type mockDocReview struct {
	reviewableDocs    []domain.DocStatus
	reviewableDocsErr error
	markResult        domain.DocReviewResult
	markErr           error
	markAllResults    []domain.DocReviewResult
	markAllErr        error
}

func (m *mockDocReview) ReviewableDocs(_ context.Context, _ string) ([]domain.DocStatus, error) {
	return m.reviewableDocs, m.reviewableDocsErr
}

func (m *mockDocReview) MarkReviewed(_ context.Context, docPath, _ string, _ *time.Time) (domain.DocReviewResult, error) {
	if m.markErr != nil {
		return domain.DocReviewResult{}, m.markErr
	}
	return m.markResult, nil
}

func (m *mockDocReview) MarkAllReviewed(_ context.Context, _ string, _ *time.Time) ([]domain.DocReviewResult, error) {
	return m.markAllResults, m.markAllErr
}

func TestDocReviewList_PrintsReviewableDocs(t *testing.T) {
	// Setup mock with reviewable docs.
	lastReviewed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	daysSince := 60
	mock := &mockDocReview{
		reviewableDocs: []domain.DocStatus{
			domain.NewDocStatus("docs/README.md", domain.DocHealthStale, &lastReviewed, &daysSince, 30, "team", nil),
			domain.NewDocStatus("docs/ARCHITECTURE.md", domain.DocHealthNoFrontmatter, nil, nil, 30, "team", nil),
		},
	}
	handler := dochealthapp.NewDocReviewHandler(mock)
	app := &composition.App{DocReviewHandler: handler}

	// Create command and capture output.
	cmd := commands.NewDocReviewCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "docs/README.md")
	assert.Contains(t, output, "docs/ARCHITECTURE.md")
	assert.Contains(t, output, "stale")
}

func TestDocReviewList_NoDocs_PrintsEmpty(t *testing.T) {
	mock := &mockDocReview{
		reviewableDocs: nil,
	}
	handler := dochealthapp.NewDocReviewHandler(mock)
	app := &composition.App{DocReviewHandler: handler}

	cmd := commands.NewDocReviewCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, strings.ToLower(output), "no docs due for review")
}

func TestDocReviewMark_UpdatesFrontmatter(t *testing.T) {
	newDate := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	mock := &mockDocReview{
		markResult: domain.NewDocReviewResult("docs/README.md", newDate),
	}
	handler := dochealthapp.NewDocReviewHandler(mock)
	app := &composition.App{DocReviewHandler: handler}

	cmd := commands.NewDocReviewCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"mark", "docs/README.md"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "docs/README.md")
	assert.Contains(t, output, "2026-03-09")
}

func TestDocReviewMark_MissingArg_ReturnsError(t *testing.T) {
	mock := &mockDocReview{}
	handler := dochealthapp.NewDocReviewHandler(mock)
	app := &composition.App{DocReviewHandler: handler}

	cmd := commands.NewDocReviewCmd(app)
	var buf bytes.Buffer
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"mark"}) // No doc-path argument.

	err := cmd.Execute()
	require.Error(t, err)
}

func TestDocReviewMarkAll_UpdatesAllStale(t *testing.T) {
	newDate := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	mock := &mockDocReview{
		markAllResults: []domain.DocReviewResult{
			domain.NewDocReviewResult("docs/README.md", newDate),
			domain.NewDocReviewResult("docs/ARCHITECTURE.md", newDate),
		},
	}
	handler := dochealthapp.NewDocReviewHandler(mock)
	app := &composition.App{DocReviewHandler: handler}

	cmd := commands.NewDocReviewCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"mark-all"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "2") // Count of updated docs.
	assert.Contains(t, output, "docs/README.md")
}

func TestDocReview_DefaultToList(t *testing.T) {
	// Running "alto doc-review" with no subcommand should default to list.
	mock := &mockDocReview{
		reviewableDocs: nil,
	}
	handler := dochealthapp.NewDocReviewHandler(mock)
	app := &composition.App{DocReviewHandler: handler}

	cmd := commands.NewDocReviewCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{}) // No subcommand.

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should behave like list - show "no docs due" message.
	assert.Contains(t, strings.ToLower(output), "no docs due for review")
}
