package application

import (
	"context"
	"fmt"
	"time"

	"github.com/alto-cli/alto/internal/dochealth/domain"
)

// DocReviewHandler orchestrates document review operations.
// It delegates to the DocReview port for actual file I/O.
type DocReviewHandler struct {
	reviewer DocReview
}

// NewDocReviewHandler creates a new DocReviewHandler.
func NewDocReviewHandler(reviewer DocReview) *DocReviewHandler {
	return &DocReviewHandler{reviewer: reviewer}
}

// ReviewableDocs returns docs that are due for review.
func (h *DocReviewHandler) ReviewableDocs(ctx context.Context, projectDir string) ([]domain.DocStatus, error) {
	statuses, err := h.reviewer.ReviewableDocs(ctx, projectDir)
	if err != nil {
		return nil, fmt.Errorf("reviewable docs: %w", err)
	}
	return statuses, nil
}

// MarkReviewed marks a document as reviewed. Pass nil for reviewDate to use current time.
func (h *DocReviewHandler) MarkReviewed(ctx context.Context, docPath, projectDir string, reviewDate *time.Time) (domain.DocReviewResult, error) {
	result, err := h.reviewer.MarkReviewed(ctx, docPath, projectDir, reviewDate)
	if err != nil {
		return domain.DocReviewResult{}, fmt.Errorf("mark reviewed: %w", err)
	}
	return result, nil
}

// MarkAllReviewed marks all stale docs as reviewed. Pass nil for reviewDate to use current time.
func (h *DocReviewHandler) MarkAllReviewed(ctx context.Context, projectDir string, reviewDate *time.Time) ([]domain.DocReviewResult, error) {
	results, err := h.reviewer.MarkAllReviewed(ctx, projectDir, reviewDate)
	if err != nil {
		return nil, fmt.Errorf("mark all reviewed: %w", err)
	}
	return results, nil
}
