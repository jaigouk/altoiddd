// Package errors defines sentinel errors for domain invariant violations.
package errors

import "errors"

var (
	// ErrInvariantViolation indicates a domain rule was violated.
	ErrInvariantViolation = errors.New("invariant violation")

	// ErrNotFound indicates a requested entity does not exist.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists indicates a duplicate entity was rejected.
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidTransition indicates an invalid state transition was attempted.
	ErrInvalidTransition = errors.New("invalid state transition")
)
