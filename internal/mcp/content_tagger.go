package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TagContent wraps text with boundary markers indicating its source.
// This helps AI clients distinguish tool output from other content.
func TagContent(text, method string) string {
	label := methodLabel(method)
	return fmt.Sprintf("[%s START]\n%s\n[%s END]", label, text, label)
}

// methodLabel returns a human-readable label for an MCP method.
func methodLabel(method string) string {
	switch method {
	case "tools/call":
		return "TOOL OUTPUT"
	default:
		return "MCP OUTPUT"
	}
}

// ContentTagMiddleware returns MCP middleware that adds boundary markers
// to tool and resource response text content.
func ContentTagMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			result, err := next(ctx, method, req)
			if err != nil || result == nil {
				return result, err
			}

			if toolResult, ok := result.(*mcp.CallToolResult); ok {
				tagToolResult(toolResult, method)
			}
			return result, nil
		}
	}
}

// tagToolResult wraps text content in a CallToolResult with boundary markers.
func tagToolResult(result *mcp.CallToolResult, method string) {
	for i, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			result.Content[i] = &mcp.TextContent{
				Text: TagContent(tc.Text, method),
			}
		}
	}
}
