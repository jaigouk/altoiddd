package mcp

import (
	"fmt"

	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	ttdomain "github.com/alty-cli/alty/internal/tooltranslation/domain"
)

var qualityGateMap = map[string]vo.QualityGate{
	"lint":    vo.QualityGateLint,
	"types":   vo.QualityGateTypes,
	"tests":   vo.QualityGateTests,
	"fitness": vo.QualityGateFitness,
}

var supportedToolMap = map[string]ttdomain.SupportedTool{
	"claude-code": ttdomain.ToolClaudeCode,
	"cursor":      ttdomain.ToolCursor,
	"roo-code":    ttdomain.ToolRooCode,
	"opencode":    ttdomain.ToolOpenCode,
}

// ParseQualityGates maps string gate names to typed QualityGate values.
// Returns an error if any name is unrecognised.
func ParseQualityGates(names []string) ([]vo.QualityGate, error) {
	if len(names) == 0 {
		return nil, nil
	}
	out := make([]vo.QualityGate, 0, len(names))
	for _, n := range names {
		g, ok := qualityGateMap[n]
		if !ok {
			return nil, fmt.Errorf("unknown quality gate: %q", n)
		}
		out = append(out, g)
	}
	return out, nil
}

// ParseSupportedTools maps string tool names to typed SupportedTool values.
// Returns an error if any name is unrecognised.
func ParseSupportedTools(names []string) ([]ttdomain.SupportedTool, error) {
	if len(names) == 0 {
		return nil, nil
	}
	out := make([]ttdomain.SupportedTool, 0, len(names))
	for _, n := range names {
		t, ok := supportedToolMap[n]
		if !ok {
			return nil, fmt.Errorf("unknown tool: %q", n)
		}
		out = append(out, t)
	}
	return out, nil
}
