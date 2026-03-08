package mcp

import (
	"context"
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AuditMiddleware returns MCP middleware that logs every method invocation
// with session ID, method name, duration, and success/failure status.
func AuditMiddleware(logger *slog.Logger) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()
			sessionID := req.GetSession().ID()

			result, err := next(ctx, method, req)

			duration := time.Since(start)
			if err != nil {
				logger.ErrorContext(ctx, "MCP request failed",
					"session", sessionID,
					"method", method,
					"duration", duration,
					"error", err,
				)
			} else {
				logger.InfoContext(ctx, "MCP request",
					"session", sessionID,
					"method", method,
					"duration", duration,
				)
			}
			return result, err
		}
	}
}
