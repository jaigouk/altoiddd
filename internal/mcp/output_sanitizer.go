package mcp

import (
	"context"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// absolutePathPattern matches Unix and Windows absolute paths.
var absolutePathPattern = regexp.MustCompile(`(?:/(?:home|Users|tmp|var|etc|opt|usr|root)/[^\s"',;:)}\]]+|[A-Z]:\\[^\s"',;:)}\]]+)`)

// secretPatterns matches common secret formats that should be redacted.
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|api[_-]?secret|access[_-]?token|auth[_-]?token|bearer)\s*[=:]\s*\S+`),
	regexp.MustCompile(`sk-[a-zA-Z0-9]{20,}`),                                      // OpenAI-style keys
	regexp.MustCompile(`(?i)(password|passwd|secret)\s*[=:]\s*\S+`),                // password fields
	regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),                                      // GitHub PAT
	regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`),                                     // AWS access key
	regexp.MustCompile(`-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----[\s\S]*?-----END`), // PEM keys
}

// SanitizeOutput strips absolute paths and secret patterns from text.
// Absolute paths are replaced with just the filename component.
// Secrets are replaced with [REDACTED].
func SanitizeOutput(text string) string {
	// Redact secrets first (more specific patterns).
	for _, pat := range secretPatterns {
		text = pat.ReplaceAllString(text, "[REDACTED]")
	}

	// Replace absolute paths with just the filename.
	text = absolutePathPattern.ReplaceAllStringFunc(text, func(match string) string {
		// Determine separator: Windows paths use backslash, Unix uses forward slash.
		sep := "/"
		if strings.ContainsRune(match, '\\') {
			sep = `\`
		}
		parts := strings.Split(match, sep)
		if len(parts) > 1 {
			return parts[len(parts)-1]
		}
		return "[REDACTED-PATH]"
	})

	return text
}

// OutputSanitizeMiddleware returns MCP middleware that strips absolute paths
// and secret patterns from tool result text content.
func OutputSanitizeMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			result, err := next(ctx, method, req)
			if err != nil || result == nil {
				return result, err
			}

			// Sanitize CallToolResult text content.
			if toolResult, ok := result.(*mcp.CallToolResult); ok {
				sanitizeToolResult(toolResult)
			}
			return result, nil
		}
	}
}

// sanitizeToolResult sanitizes text content within a CallToolResult.
func sanitizeToolResult(result *mcp.CallToolResult) {
	for i, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			sanitized := SanitizeOutput(tc.Text)
			if sanitized != tc.Text {
				result.Content[i] = &mcp.TextContent{Text: sanitized}
			}
		}
	}
}
