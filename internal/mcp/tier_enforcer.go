package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tier represents a tool access tier. Higher tiers include lower tier permissions.
type Tier int

const (
	// TierRead allows read-only tools (detect_tools, doc_health, ticket_health).
	TierRead Tier = iota
	// TierWrite allows tools that modify project files (init, generate_*, doc_review).
	TierWrite
	// TierExec allows tools that execute subprocesses (check_quality).
	TierExec
)

// TierPolicy maps tool names to their required tier.
type TierPolicy struct {
	maxTier  Tier
	toolTier map[string]Tier
}

// NewTierPolicy creates a policy with the given maximum allowed tier
// and tool-to-tier mappings.
func NewTierPolicy(maxTier Tier, toolTier map[string]Tier) *TierPolicy {
	return &TierPolicy{
		maxTier:  maxTier,
		toolTier: toolTier,
	}
}

// IsAllowed returns true if the given tool is allowed under the policy.
// Tools not in the mapping are allowed by default.
func (tp *TierPolicy) IsAllowed(tool string) bool {
	requiredTier, exists := tp.toolTier[tool]
	if !exists {
		return true // unlisted tools are allowed
	}
	return requiredTier <= tp.maxTier
}

// extractToolName extracts the tool name from an MCP request when the
// method is "tools/call". Returns empty string for other methods.
func extractToolName(method string, req mcp.Request) string {
	if method != "tools/call" {
		return ""
	}
	params := req.GetParams()
	if ctp, ok := params.(*mcp.CallToolParamsRaw); ok {
		return ctp.Name
	}
	return ""
}

// TierEnforceMiddleware returns MCP middleware that restricts tool access
// based on tier policy. Non-tool methods pass through unconditionally.
func TierEnforceMiddleware(policy *TierPolicy) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			toolName := extractToolName(method, req)
			if toolName != "" && !policy.IsAllowed(toolName) {
				return nil, fmt.Errorf("tool %q requires higher access tier", toolName)
			}
			return next(ctx, method, req)
		}
	}
}
