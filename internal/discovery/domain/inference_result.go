package domain

import (
	"fmt"
	"strings"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// InferenceResult is a value object wrapping a domain model inferred from project
// documentation, along with a confidence level and the list of source docs used.
type InferenceResult struct {
	model      *ddd.DomainModel
	confidence string
	sourceDocs []string
}

// NewInferenceResult creates an InferenceResult. Model and confidence are required.
func NewInferenceResult(model *ddd.DomainModel, confidence string, sourceDocs []string) (*InferenceResult, error) {
	if model == nil {
		return nil, fmt.Errorf("inference result model cannot be nil: %w", domainerrors.ErrInvariantViolation)
	}
	if strings.TrimSpace(confidence) == "" {
		return nil, fmt.Errorf("inference result confidence cannot be empty: %w", domainerrors.ErrInvariantViolation)
	}
	docs := make([]string, len(sourceDocs))
	copy(docs, sourceDocs)
	return &InferenceResult{model: model, confidence: confidence, sourceDocs: docs}, nil
}

// Model returns the inferred domain model.
func (r *InferenceResult) Model() *ddd.DomainModel { return r.model }

// Confidence returns the confidence level of the inference (e.g. "high", "medium", "low").
func (r *InferenceResult) Confidence() string { return r.confidence }

// SourceDocs returns a defensive copy of the source document names used for inference.
func (r *InferenceResult) SourceDocs() []string {
	out := make([]string, len(r.sourceDocs))
	copy(out, r.sourceDocs)
	return out
}
