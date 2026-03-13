package domain

import (
	"time"
)

// DDDVersion represents the version metadata of a DDD.md document.
// It is a value object - immutable after creation.
//
// Note: YAML parsing/serialization is handled by infrastructure adapters
// (ParseDDDVersionFromContent, ApplyVersionToContent) to keep the domain
// layer free of external dependencies.
type DDDVersion struct {
	version          int
	round            string
	updated          string
	convergenceDelta int
}

// NewDDDVersion creates a new DDDVersion with the given values.
func NewDDDVersion(version int, round, updated string, convergenceDelta int) DDDVersion {
	return DDDVersion{
		version:          version,
		round:            round,
		updated:          updated,
		convergenceDelta: convergenceDelta,
	}
}

// Version returns the version number.
func (v DDDVersion) Version() int { return v.version }

// Round returns the round name (e.g., "express", "challenge", "simulate").
func (v DDDVersion) Round() string { return v.round }

// Updated returns the update date string.
func (v DDDVersion) Updated() string { return v.updated }

// ConvergenceDelta returns the convergence delta from the last round.
func (v DDDVersion) ConvergenceDelta() int { return v.convergenceDelta }

// Increment creates a new DDDVersion with incremented version number
// and updated metadata. The original DDDVersion is unchanged (immutability).
func (v DDDVersion) Increment(round string, convergenceDelta int, updatedAt time.Time) DDDVersion {
	return DDDVersion{
		version:          v.version + 1,
		round:            round,
		updated:          updatedAt.Format("2006-01-02"),
		convergenceDelta: convergenceDelta,
	}
}
