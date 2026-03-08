package application_test

import (
	"context"
	"time"

	"github.com/alty-cli/alty/internal/dochealth/application"
	dochealthdomain "github.com/alty-cli/alty/internal/dochealth/domain"
)

// Compile-time interface satisfaction checks.
var (
	_ application.DocHealth = (*mockDocHealth)(nil)
	_ application.DocReview = (*mockDocReview)(nil)
)

// --- mockDocHealth ---

type mockDocHealth struct{}

func (m *mockDocHealth) Check(_ context.Context, _ string) (dochealthdomain.DocHealthReport, error) {
	return dochealthdomain.DocHealthReport{}, nil
}

func (m *mockDocHealth) CheckKnowledge(_ context.Context, _ string) (dochealthdomain.DocHealthReport, error) {
	return dochealthdomain.DocHealthReport{}, nil
}

// --- mockDocReview ---

type mockDocReview struct{}

func (m *mockDocReview) ReviewableDocs(_ context.Context, _ string) ([]dochealthdomain.DocStatus, error) {
	return nil, nil
}

func (m *mockDocReview) MarkReviewed(_ context.Context, _ string, _ string, _ *time.Time) (dochealthdomain.DocReviewResult, error) {
	return dochealthdomain.DocReviewResult{}, nil
}

func (m *mockDocReview) MarkAllReviewed(_ context.Context, _ string, _ *time.Time) ([]dochealthdomain.DocReviewResult, error) {
	return nil, nil
}
