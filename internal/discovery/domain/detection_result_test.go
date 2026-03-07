package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// -- ConflictSeverity tests --

func TestConflictSeverityValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "compatible", string(SeverityCompatible))
	assert.Equal(t, "warning", string(SeverityWarning))
	assert.Equal(t, "conflict", string(SeverityConflict))
}

func TestAllConflictSeverities(t *testing.T) {
	t.Parallel()
	assert.Len(t, AllConflictSeverities(), 3)
}

// -- DetectionResult creation tests --

func TestDetectionResultCreationWithToolsAndConflicts(t *testing.T) {
	t.Parallel()
	tools := []DetectedTool{
		NewDetectedTool("claude-code", "/home/user/.claude", ""),
		NewDetectedTool("cursor", "", ""),
	}
	conflicts := []string{"Global cursor setting overrides local"}
	severityMap := map[string]ConflictSeverity{
		"Global cursor setting overrides local": SeverityWarning,
	}
	result := NewDetectionResult(tools, conflicts, severityMap)

	assert.Len(t, result.DetectedTools(), 2)
	assert.Equal(t, "claude-code", result.DetectedTools()[0].Name())
	assert.Len(t, result.Conflicts(), 1)
	assert.Equal(t, SeverityWarning, result.SeverityMap()["Global cursor setting overrides local"])
}

func TestDetectionResultEmpty(t *testing.T) {
	t.Parallel()
	result := NewDetectionResult(nil, nil, nil)
	assert.Empty(t, result.DetectedTools())
	assert.Empty(t, result.Conflicts())
	assert.Empty(t, result.SeverityMap())
}

func TestDetectionResultMultipleSeverities(t *testing.T) {
	t.Parallel()
	severityMap := map[string]ConflictSeverity{
		"Same value in both":     SeverityCompatible,
		"Cursor SQLite detected": SeverityWarning,
		"Contradicting settings": SeverityConflict,
	}
	result := NewDetectionResult(nil,
		[]string{"Same value in both", "Cursor SQLite detected", "Contradicting settings"},
		severityMap)
	assert.Equal(t, SeverityCompatible, result.SeverityMap()["Same value in both"])
	assert.Equal(t, SeverityWarning, result.SeverityMap()["Cursor SQLite detected"])
	assert.Equal(t, SeverityConflict, result.SeverityMap()["Contradicting settings"])
}

// -- DetectionResult defensive copy tests --

func TestDetectionResultDetectedToolsDefensiveCopy(t *testing.T) {
	t.Parallel()
	tools := []DetectedTool{NewDetectedTool("claude-code", "", "")}
	result := NewDetectionResult(tools, nil, nil)
	// Mutating input should not affect result
	tools[0] = NewDetectedTool("mutated", "", "")
	assert.Equal(t, "claude-code", result.DetectedTools()[0].Name())
}

func TestDetectionResultConflictsDefensiveCopy(t *testing.T) {
	t.Parallel()
	conflicts := []string{"a conflict"}
	result := NewDetectionResult(nil, conflicts, nil)
	conflicts[0] = "mutated"
	assert.Equal(t, "a conflict", result.Conflicts()[0])
}

func TestDetectionResultSeverityMapDefensiveCopy(t *testing.T) {
	t.Parallel()
	sMap := map[string]ConflictSeverity{"a": SeverityWarning}
	result := NewDetectionResult(nil, nil, sMap)
	// Mutating returned map should not affect internal state
	returned := result.SeverityMap()
	returned["new_key"] = SeverityConflict
	assert.NotContains(t, result.SeverityMap(), "new_key")
}
