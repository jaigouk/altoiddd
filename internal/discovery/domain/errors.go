package domain

import (
	"errors"
	"fmt"
)

// ErrInferenceDismissed indicates the user declined an inference result during
// the --existing flow. The caller catches this with errors.Is and falls through
// to the normal storytelling discovery path.
var ErrInferenceDismissed = errors.New("user dismissed inference result")

// ErrInferenceFailed is the sentinel matched by InferenceFailedError.Is.
// Callers use errors.Is(err, ErrInferenceFailed) to detect that doc inference
// was attempted against discovered docs but the underlying LLM or regex step
// failed — distinct from "no docs were found in the first place".
var ErrInferenceFailed = errors.New("inference failed")

// ErrNoDocsFound is returned when a docs directory is reachable but contained
// no inference-eligible documentation. Callers may use this signal to attempt
// a fallback search location (e.g., "." after "docs/").
var ErrNoDocsFound = errors.New("no docs found")

// InferenceFailedError carries the list of source docs that were discovered
// alongside the underlying failure reason, so the CLI layer can present an
// actionable message instead of a generic "no docs found" fallthrough.
//
// Docs MUST be sorted by the producer for deterministic output.
type InferenceFailedError struct {
	Docs   []string
	Reason error
}

// Error renders a single-line summary including the doc count and the
// underlying reason. The exact format is part of the contract surfaced to
// the CLI layer.
func (e *InferenceFailedError) Error() string {
	return fmt.Sprintf("inference failed (found %d doc(s)): %v", len(e.Docs), e.Reason)
}

// Unwrap exposes the underlying reason so errors.Is/As can walk past this
// wrapper to the original failure (e.g., an LLM client error).
func (e *InferenceFailedError) Unwrap() error { return e.Reason }

// Is matches the ErrInferenceFailed sentinel so callers can classify any
// InferenceFailedError as "inference failed" without unwrapping by type.
func (e *InferenceFailedError) Is(target error) bool { return target == ErrInferenceFailed }
