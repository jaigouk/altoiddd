// Package domain provides the Challenge bounded context's core domain model.
package domain

import "errors"

// Domain errors for the Challenge bounded context.
var (
	// ErrChallengeNotFound is returned when a challenge ID doesn't exist in the session.
	ErrChallengeNotFound = errors.New("challenge not found")
	// ErrChallengeAlreadyAnswered is returned when attempting to respond to a challenge twice.
	ErrChallengeAlreadyAnswered = errors.New("challenge already answered")
	// ErrSessionNotFound is returned when a session ID doesn't exist.
	ErrSessionNotFound = errors.New("session not found")
)
