// Package application defines ports for the DocHealth bounded context.
package application

import (
	"context"
	"time"

	dochealthdomain "github.com/alty-cli/alty/internal/dochealth/domain"
)

// DocHealth checks the health and consistency of project documentation
// and knowledge base entries.
type DocHealth interface {
	// Check checks the health of project documentation.
	Check(ctx context.Context, projectDir string) (dochealthdomain.DocHealthReport, error)

	// CheckKnowledge checks the health of knowledge base entries.
	CheckKnowledge(ctx context.Context, knowledgeDir string) (dochealthdomain.DocHealthReport, error)
}

// DocReview manages document review operations: marking docs as reviewed
// and querying which docs are due for review.
type DocReview interface {
	// ReviewableDocs returns docs that are due for review.
	ReviewableDocs(ctx context.Context, projectDir string) ([]dochealthdomain.DocStatus, error)

	// MarkReviewed marks a document as reviewed.
	MarkReviewed(ctx context.Context, docPath string, projectDir string, reviewDate *time.Time) (dochealthdomain.DocReviewResult, error)

	// MarkAllReviewed marks all stale docs as reviewed.
	MarkAllReviewed(ctx context.Context, projectDir string, reviewDate *time.Time) ([]dochealthdomain.DocReviewResult, error)
}
