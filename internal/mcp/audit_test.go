package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AuditMiddleware tests use full server/client roundtrips since mcp.Request
// has unexported interface methods and cannot be stubbed outside the SDK.

func setupAuditServer(t *testing.T, logger *slog.Logger) *gomcp.Server {
	t.Helper()
	server := gomcp.NewServer(&gomcp.Implementation{Name: "test", Version: "0.1"}, nil)
	server.AddReceivingMiddleware(AuditMiddleware(logger))
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "echo",
		Description: "echo tool",
	}, func(_ context.Context, _ *gomcp.CallToolRequest, input struct {
		Msg string `json:"msg"`
	},
	) (*gomcp.CallToolResult, any, error) {
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{&gomcp.TextContent{Text: input.Msg}},
		}, nil, nil
	})
	return server
}

func connectAudit(t *testing.T, ctx context.Context, server *gomcp.Server) *gomcp.ClientSession {
	t.Helper()
	ct, st := gomcp.NewInMemoryTransports()
	go func() { _ = server.Run(ctx, st) }()
	client := gomcp.NewClient(&gomcp.Implementation{Name: "test-client", Version: "0.1"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	return session
}

func findLogEntry(buf *bytes.Buffer, msgField string) (map[string]any, bool) {
	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	for _, line := range lines {
		var entry map[string]any
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		if entry["msg"] == msgField {
			return entry, true
		}
	}
	return nil, false
}

func findLogEntryWithMethod(buf *bytes.Buffer, msgField, method string) (map[string]any, bool) {
	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	for _, line := range lines {
		var entry map[string]any
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		if entry["msg"] == msgField && entry["method"] == method {
			return entry, true
		}
	}
	return nil, false
}

func TestAuditMiddleware_LogsToolCall(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	server := setupAuditServer(t, logger)
	ctx := context.Background()
	session := connectAudit(t, ctx, server)

	_, err := session.CallTool(ctx, &gomcp.CallToolParams{
		Name:      "echo",
		Arguments: map[string]any{"msg": "hello"},
	})
	require.NoError(t, err)

	entry, found := findLogEntryWithMethod(&buf, "MCP request", "tools/call")
	require.True(t, found, "should have logged a tools/call entry")
	_, hasSession := entry["session"]
	assert.True(t, hasSession, "should include session field")
	assert.NotNil(t, entry["duration"], "should include duration")
	assert.Equal(t, "tools/call", entry["method"])
}

func TestAuditMiddleware_LogsError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	server := gomcp.NewServer(&gomcp.Implementation{Name: "test", Version: "0.1"}, nil)
	server.AddReceivingMiddleware(AuditMiddleware(logger))
	// No tools registered — calling an unknown tool should produce an error.

	ctx := context.Background()
	session := connectAudit(t, ctx, server)

	_, err := session.CallTool(ctx, &gomcp.CallToolParams{Name: "nonexistent"})
	require.Error(t, err)

	entry, found := findLogEntry(&buf, "MCP request failed")
	require.True(t, found, "should have logged an error entry")
	assert.Equal(t, "ERROR", entry["level"])
}

func TestAuditMiddleware_LogsResourceRead(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	server := gomcp.NewServer(&gomcp.Implementation{Name: "test", Version: "0.1"}, nil)
	server.AddReceivingMiddleware(AuditMiddleware(logger))
	server.AddResource(&gomcp.Resource{
		Name:     "Test Resource",
		URI:      "test://hello",
		MIMEType: "text/plain",
	}, func(_ context.Context, _ *gomcp.ReadResourceRequest) (*gomcp.ReadResourceResult, error) {
		return &gomcp.ReadResourceResult{
			Contents: []*gomcp.ResourceContents{
				{URI: "test://hello", MIMEType: "text/plain", Text: "hello"},
			},
		}, nil
	})

	ctx := context.Background()
	session := connectAudit(t, ctx, server)

	_, err := session.ReadResource(ctx, &gomcp.ReadResourceParams{URI: "test://hello"})
	require.NoError(t, err)

	_, found := findLogEntryWithMethod(&buf, "MCP request", "resources/read")
	assert.True(t, found, "should have logged a resources/read entry")
}

func TestAuditMiddleware_IncludesDuration(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	server := setupAuditServer(t, logger)
	ctx := context.Background()
	session := connectAudit(t, ctx, server)

	_, err := session.CallTool(ctx, &gomcp.CallToolParams{
		Name:      "echo",
		Arguments: map[string]any{"msg": "test"},
	})
	require.NoError(t, err)

	entry, found := findLogEntryWithMethod(&buf, "MCP request", "tools/call")
	require.True(t, found)
	_, hasDuration := entry["duration"]
	assert.True(t, hasDuration, "log entry should include duration")
}

func TestAuditMiddleware_IncludesSessionID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	server := setupAuditServer(t, logger)
	ctx := context.Background()
	session := connectAudit(t, ctx, server)

	_, err := session.CallTool(ctx, &gomcp.CallToolParams{
		Name:      "echo",
		Arguments: map[string]any{"msg": "test"},
	})
	require.NoError(t, err)

	entry, found := findLogEntryWithMethod(&buf, "MCP request", "tools/call")
	require.True(t, found)
	_, hasSession := entry["session"]
	assert.True(t, hasSession, "should include session field")
}
