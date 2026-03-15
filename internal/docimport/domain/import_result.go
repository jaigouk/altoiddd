// Package domain contains the core domain types for the DocImport bounded context.
package domain

import (
	"fmt"
	"strings"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// ParseWarning represents a section of a document that could not be fully parsed.
type ParseWarning struct {
	section string
	reason  string
}

// NewParseWarning creates a ParseWarning value object.
func NewParseWarning(section, reason string) (ParseWarning, error) {
	if strings.TrimSpace(section) == "" {
		return ParseWarning{}, fmt.Errorf("parse warning section cannot be empty: %w", domainerrors.ErrInvariantViolation)
	}
	if strings.TrimSpace(reason) == "" {
		return ParseWarning{}, fmt.Errorf("parse warning reason cannot be empty: %w", domainerrors.ErrInvariantViolation)
	}
	return ParseWarning{section: section, reason: reason}, nil
}

// Section returns the document section that produced the warning.
func (w ParseWarning) Section() string { return w.section }

// Reason returns why the section could not be parsed.
func (w ParseWarning) Reason() string { return w.reason }

// ImportResult wraps a parsed DomainModel with any warnings encountered during import.
type ImportResult struct {
	model    *ddd.DomainModel
	warnings []ParseWarning
}

// NewImportResult creates an ImportResult value object.
func NewImportResult(model *ddd.DomainModel, warnings []ParseWarning) (*ImportResult, error) {
	if model == nil {
		return nil, fmt.Errorf("import result model cannot be nil: %w", domainerrors.ErrInvariantViolation)
	}
	w := make([]ParseWarning, len(warnings))
	copy(w, warnings)
	return &ImportResult{model: model, warnings: w}, nil
}

// Model returns the parsed domain model.
func (r *ImportResult) Model() *ddd.DomainModel { return r.model }

// Warnings returns a defensive copy of parse warnings.
func (r *ImportResult) Warnings() []ParseWarning {
	out := make([]ParseWarning, len(r.warnings))
	copy(out, r.warnings)
	return out
}
