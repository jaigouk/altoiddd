package domain

// ConflictSeverity classifies the severity of a configuration conflict.
type ConflictSeverity string

// ConflictSeverity constants.
const (
	SeverityCompatible ConflictSeverity = "compatible"
	SeverityWarning    ConflictSeverity = "warning"
	SeverityConflict   ConflictSeverity = "conflict"
)

// AllConflictSeverities returns all valid severity values.
func AllConflictSeverities() []ConflictSeverity {
	return []ConflictSeverity{SeverityCompatible, SeverityWarning, SeverityConflict}
}

// DetectionResult is an immutable result of scanning for AI coding tools and conflicts.
type DetectionResult struct {
	severityMap   map[string]ConflictSeverity
	detectedTools []DetectedTool
	conflicts     []string
}

// NewDetectionResult creates a DetectionResult with defensive copies.
func NewDetectionResult(
	detectedTools []DetectedTool,
	conflicts []string,
	severityMap map[string]ConflictSeverity,
) DetectionResult {
	dt := make([]DetectedTool, len(detectedTools))
	copy(dt, detectedTools)
	c := make([]string, len(conflicts))
	copy(c, conflicts)
	sm := make(map[string]ConflictSeverity, len(severityMap))
	for k, v := range severityMap {
		sm[k] = v
	}
	return DetectionResult{detectedTools: dt, conflicts: c, severityMap: sm}
}

// DetectedTools returns a defensive copy of detected tools.
func (dr DetectionResult) DetectedTools() []DetectedTool {
	out := make([]DetectedTool, len(dr.detectedTools))
	copy(out, dr.detectedTools)
	return out
}

// Conflicts returns a defensive copy of conflict descriptions.
func (dr DetectionResult) Conflicts() []string {
	out := make([]string, len(dr.conflicts))
	copy(out, dr.conflicts)
	return out
}

// SeverityMap returns a defensive copy of the severity mapping.
func (dr DetectionResult) SeverityMap() map[string]ConflictSeverity {
	out := make(map[string]ConflictSeverity, len(dr.severityMap))
	for k, v := range dr.severityMap {
		out[k] = v
	}
	return out
}
