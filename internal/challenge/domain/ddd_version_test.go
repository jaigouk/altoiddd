package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	challengedomain "github.com/alto-cli/alto/internal/challenge/domain"
)

// --- NewDDDVersion constructor tests ---

func TestNewDDDVersion(t *testing.T) {
	t.Parallel()

	version := challengedomain.NewDDDVersion(3, "simulate", "2026-03-15", 7)

	assert.Equal(t, 3, version.Version())
	assert.Equal(t, "simulate", version.Round())
	assert.Equal(t, "2026-03-15", version.Updated())
	assert.Equal(t, 7, version.ConvergenceDelta())
}

func TestNewDDDVersion_ZeroValues(t *testing.T) {
	t.Parallel()

	version := challengedomain.NewDDDVersion(0, "", "", 0)

	assert.Equal(t, 0, version.Version())
	assert.Empty(t, version.Round())
	assert.Empty(t, version.Updated())
	assert.Equal(t, 0, version.ConvergenceDelta())
}

// --- DDDVersion.Increment tests ---

func TestDDDVersion_Increment(t *testing.T) {
	t.Parallel()

	original := challengedomain.NewDDDVersion(1, "express", "2026-03-01", 0)

	updated := original.Increment("challenge", 5, time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC))

	assert.Equal(t, 2, updated.Version())
	assert.Equal(t, "challenge", updated.Round())
	assert.Equal(t, 5, updated.ConvergenceDelta())
	assert.Equal(t, "2026-03-12", updated.Updated())
}

func TestDDDVersion_IncrementFromZero(t *testing.T) {
	t.Parallel()

	// Starting from unversioned document (version 0)
	original := challengedomain.NewDDDVersion(0, "", "", 0)

	updated := original.Increment("express", 0, time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC))

	assert.Equal(t, 1, updated.Version())
	assert.Equal(t, "express", updated.Round())
}

func TestDDDVersion_IncrementPreservesImmutability(t *testing.T) {
	t.Parallel()

	original := challengedomain.NewDDDVersion(1, "express", "2026-03-01", 0)

	_ = original.Increment("challenge", 3, time.Now())

	// Original should be unchanged
	assert.Equal(t, 1, original.Version())
	assert.Equal(t, "express", original.Round())
}

func TestDDDVersion_IncrementMultipleTimes(t *testing.T) {
	t.Parallel()

	v1 := challengedomain.NewDDDVersion(1, "express", "2026-03-01", 0)
	v2 := v1.Increment("challenge", 3, time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC))
	v3 := v2.Increment("simulate", 1, time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC))

	assert.Equal(t, 1, v1.Version())
	assert.Equal(t, 2, v2.Version())
	assert.Equal(t, 3, v3.Version())
	assert.Equal(t, "simulate", v3.Round())
	assert.Equal(t, 1, v3.ConvergenceDelta())
}
