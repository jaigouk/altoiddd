package mcp

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// internalErrorPattern matches common internal error indicators that
// should not be exposed to MCP clients.
var internalErrorPattern = regexp.MustCompile(`(?i)(stack trace|goroutine|panic:|runtime error:|internal server error)`)

// SanitizeError converts internal errors to safe user-facing messages.
// Errors containing stack traces, panics, or internal details are replaced
// with a generic message. Absolute paths in error messages are stripped.
func SanitizeError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()

	// Check for internal error indicators.
	if internalErrorPattern.MatchString(msg) {
		return fmt.Errorf("internal error occurred")
	}

	// Strip absolute paths from the error message.
	msg = absolutePathPattern.ReplaceAllStringFunc(msg, func(match string) string {
		parts := strings.Split(match, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
		return "<path>"
	})

	// Strip secret patterns.
	for _, pat := range secretPatterns {
		msg = pat.ReplaceAllString(msg, "[REDACTED]")
	}

	return fmt.Errorf("%s", msg)
}

// ErrorSanitizeMiddleware returns MCP middleware that sanitizes errors
// returned by handlers, removing internal details and absolute paths.
func ErrorSanitizeMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			result, err := next(ctx, method, req)
			if err != nil {
				return result, SanitizeError(err)
			}
			return result, nil
		}
	}
}
