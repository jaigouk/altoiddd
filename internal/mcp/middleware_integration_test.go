package mcp

import (
	"context"
	"fmt"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for middleware wrappers using full server/client roundtrips.

func setupMiddlewareServer(t *testing.T, middlewares ...gomcp.Middleware) (*gomcp.Server, func(context.Context) *gomcp.ClientSession) {
	t.Helper()
	server := gomcp.NewServer(&gomcp.Implementation{Name: "mw-test", Version: "0.1"}, nil)
	server.AddReceivingMiddleware(middlewares...)

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "greet",
		Description: "returns greeting with path info",
	}, func(_ context.Context, _ *gomcp.CallToolRequest, input struct {
		Name string `json:"name"`
	},
	) (*gomcp.CallToolResult, any, error) {
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{
				&gomcp.TextContent{Text: fmt.Sprintf("Hello %s from /Users/test/project/main.go", input.Name)},
			},
		}, nil, nil
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "secret_tool",
		Description: "returns text with secrets",
	}, func(_ context.Context, _ *gomcp.CallToolRequest, _ struct{}) (*gomcp.CallToolResult, any, error) {
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{
				&gomcp.TextContent{Text: "api_key=sk_live_secret123456789012345"},
			},
		}, nil, nil
	})

	connect := func(ctx context.Context) *gomcp.ClientSession {
		ct, st := gomcp.NewInMemoryTransports()
		go func() { _ = server.Run(ctx, st) }()
		client := gomcp.NewClient(&gomcp.Implementation{Name: "mw-client", Version: "0.1"}, nil)
		session, err := client.Connect(ctx, ct, nil)
		require.NoError(t, err)
		return session
	}

	return server, connect
}

// --- OutputSanitizeMiddleware integration ---

func TestOutputSanitizeMiddleware_StripsPathsFromToolResult(t *testing.T) {
	_, connect := setupMiddlewareServer(t, OutputSanitizeMiddleware())
	ctx := context.Background()
	session := connect(ctx)
	defer session.Close()

	result, err := session.CallTool(ctx, &gomcp.CallToolParams{
		Name:      "greet",
		Arguments: map[string]any{"name": "world"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	text, ok := result.Content[0].(*gomcp.TextContent)
	require.True(t, ok)
	assert.NotContains(t, text.Text, "/Users/test")
	assert.Contains(t, text.Text, "main.go")
}

func TestOutputSanitizeMiddleware_RedactsSecretsFromToolResult(t *testing.T) {
	_, connect := setupMiddlewareServer(t, OutputSanitizeMiddleware())
	ctx := context.Background()
	session := connect(ctx)
	defer session.Close()

	result, err := session.CallTool(ctx, &gomcp.CallToolParams{
		Name: "secret_tool",
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	text, ok := result.Content[0].(*gomcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, text.Text, "[REDACTED]")
	assert.NotContains(t, text.Text, "sk_live_secret")
}

// --- ContentTagMiddleware integration ---

func TestContentTagMiddleware_WrapsToolOutput(t *testing.T) {
	_, connect := setupMiddlewareServer(t, ContentTagMiddleware())
	ctx := context.Background()
	session := connect(ctx)
	defer session.Close()

	result, err := session.CallTool(ctx, &gomcp.CallToolParams{
		Name:      "greet",
		Arguments: map[string]any{"name": "test"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	text, ok := result.Content[0].(*gomcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, text.Text, "[TOOL OUTPUT START]")
	assert.Contains(t, text.Text, "[TOOL OUTPUT END]")
	assert.Contains(t, text.Text, "Hello test")
}

// --- ErrorSanitizeMiddleware integration ---

func TestErrorSanitizeMiddleware_SanitizesProtocolError(t *testing.T) {
	_, connect := setupMiddlewareServer(t, ErrorSanitizeMiddleware())
	ctx := context.Background()
	session := connect(ctx)
	defer session.Close()

	// Call nonexistent tool — produces protocol error.
	_, err := session.CallTool(ctx, &gomcp.CallToolParams{Name: "nonexistent"})
	require.Error(t, err)
	// Error should not contain stack traces or internal details.
	assert.NotContains(t, err.Error(), "goroutine")
}

// --- Composed middleware stack ---

func TestMiddlewareStack_AllApplied(t *testing.T) {
	_, connect := setupMiddlewareServer(t,
		ErrorSanitizeMiddleware(),
		OutputSanitizeMiddleware(),
		ContentTagMiddleware(),
	)
	ctx := context.Background()
	session := connect(ctx)
	defer session.Close()

	result, err := session.CallTool(ctx, &gomcp.CallToolParams{
		Name:      "greet",
		Arguments: map[string]any{"name": "stack-test"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	text, ok := result.Content[0].(*gomcp.TextContent)
	require.True(t, ok)
	// ContentTag wraps output.
	assert.Contains(t, text.Text, "[TOOL OUTPUT START]")
	// OutputSanitizer stripped absolute path.
	assert.NotContains(t, text.Text, "/Users/test")
	assert.Contains(t, text.Text, "Hello stack-test")
}
