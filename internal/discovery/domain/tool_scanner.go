package domain

import "strings"

// knownTools maps tool names to their relative config paths (from home dir).
var knownTools = map[string]string{
	"claude-code": ".claude",
	"cursor":      ".cursor",
	"roo-code":    ".roo",
	"opencode":    ".config/opencode",
}

// ToolScanner classifies tools and conflicts into a structured DetectionResult.
// Pure domain service with no external dependencies.
type ToolScanner struct{}

// NewToolScanner creates a new ToolScanner.
func NewToolScanner() *ToolScanner {
	return &ToolScanner{}
}

// KnownTools returns a copy of the known tool registry.
func (ts *ToolScanner) KnownTools() map[string]string {
	out := make(map[string]string, len(knownTools))
	for k, v := range knownTools {
		out[k] = v
	}
	return out
}

// BuildResult builds a DetectionResult from raw detection data.
func (ts *ToolScanner) BuildResult(toolNames []string, conflicts []string) DetectionResult {
	tools := make([]DetectedTool, len(toolNames))
	for i, name := range toolNames {
		tools[i] = ts.buildTool(name)
	}
	severityMap := make(map[string]ConflictSeverity, len(conflicts))
	for _, c := range conflicts {
		severityMap[c] = ts.classifyConflict(c)
	}
	return NewDetectionResult(tools, conflicts, severityMap)
}

func (ts *ToolScanner) buildTool(name string) DetectedTool {
	configPath := knownTools[name]
	return NewDetectedTool(name, configPath, "")
}

func (ts *ToolScanner) classifyConflict(description string) ConflictSeverity {
	lower := strings.ToLower(description)
	if strings.Contains(lower, "compatible") {
		return SeverityCompatible
	}
	if strings.Contains(lower, "contradict") {
		return SeverityConflict
	}
	return SeverityWarning
}
