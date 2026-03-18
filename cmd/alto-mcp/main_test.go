package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/composition"
	discoveryapp "github.com/alto-cli/alto/internal/discovery/application"
	knowledgeapp "github.com/alto-cli/alto/internal/knowledge/application"
	knowledgedomain "github.com/alto-cli/alto/internal/knowledge/domain"
	ticketapp "github.com/alto-cli/alto/internal/ticket/application"
	ticketdomain "github.com/alto-cli/alto/internal/ticket/domain"
	ttapp "github.com/alto-cli/alto/internal/tooltranslation/application"
)

// stubKnowledgeReader implements knowledgeapp.KnowledgeReader for testing.
type stubKnowledgeReader struct {
	entries map[string]knowledgedomain.KnowledgeEntry
}

func (s *stubKnowledgeReader) ReadEntry(_ context.Context, path knowledgedomain.KnowledgePath, _ string) (knowledgedomain.KnowledgeEntry, error) {
	entry, ok := s.entries[path.Raw()]
	if !ok {
		return knowledgedomain.KnowledgeEntry{}, fmt.Errorf("entry not found: %s", path.Raw())
	}
	return entry, nil
}

func (s *stubKnowledgeReader) ListTopics(_ context.Context, _ knowledgedomain.KnowledgeCategory, _ *string) ([]string, error) {
	return nil, nil
}

// stubTicketReader implements ticketapp.TicketReader for testing.
type stubTicketReader struct{}

func (s *stubTicketReader) ReadOpenTickets(_ context.Context) ([]ticketdomain.OpenTicketData, error) {
	return nil, nil
}

func (s *stubTicketReader) ReadFlags(_ context.Context, _ string) ([]ticketdomain.FreshnessFlag, error) {
	return nil, nil
}

// stubPublisher implements sharedapp.EventPublisher for testing.
type stubPublisher struct{}

func (s *stubPublisher) Publish(_ context.Context, _ any) error { return nil }

// stubFileWriter implements sharedapp.FileWriter for testing.
type stubFileWriter struct{}

func (s *stubFileWriter) WriteFile(_ context.Context, _ string, _ string) error {
	return nil
}

// testApp creates a minimal App for tests with all required handlers.
func testApp() *composition.App {
	return &composition.App{
		DiscoveryHandler:       discoveryapp.NewDiscoveryHandler(&stubPublisher{}),
		KnowledgeLookupHandler: knowledgeapp.NewKnowledgeLookupHandler(&stubKnowledgeReader{entries: map[string]knowledgedomain.KnowledgeEntry{}}),
		TicketHealthHandler:    ticketapp.NewTicketHealthHandler(&stubTicketReader{}),
		PersonaHandler:         ttapp.NewPersonaHandler(&stubFileWriter{}),
	}
}

// --- In-Memory Transport Tests ---

func TestEchoTool_InMemory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	session := connectInMemory(t, ctx)
	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "echo",
		Arguments: map[string]any{"message": "hello from test"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError, "tool should not return error")
	require.Len(t, result.Content, 1)

	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "content should be TextContent")
	assert.Contains(t, text.Text, "Echo: hello from test")
}

func TestEchoTool_EmptyMessage_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	session := connectInMemory(t, ctx)
	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "echo",
		Arguments: map[string]any{"message": ""},
	})
	require.NoError(t, err, "protocol error should not occur")
	assert.True(t, result.IsError, "tool should return isError=true")
}

func TestListTools_InMemory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	session := connectInMemory(t, ctx)
	defer session.Close()

	result, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, result.Tools, 27) // echo + 8 guide_* + 4 challenge_* + 14 bootstrap tools
	// Find echo tool in the list
	var foundEcho bool
	for _, tool := range result.Tools {
		if tool.Name == "echo" {
			foundEcho = true
			break
		}
	}
	assert.True(t, foundEcho, "echo tool should be registered")
	assert.NotEmpty(t, result.Tools[0].Description)
	assert.NotNil(t, result.Tools[0].InputSchema, "auto-schema should be populated")
}

func TestStaticResource_InMemory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	session := connectInMemory(t, ctx)
	defer session.Close()

	result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "alto://tickets/ready"})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "Ticket Health Report")
	assert.Equal(t, "text/plain", result.Contents[0].MIMEType)
}

func TestListResources_InMemory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	session := connectInMemory(t, ctx)
	defer session.Close()

	result, err := session.ListResources(ctx, nil)
	require.NoError(t, err)
	require.Len(t, result.Resources, 1)
	assert.Equal(t, "alto://tickets/ready", result.Resources[0].URI)
}

func TestListResourceTemplates_InMemory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	session := connectInMemory(t, ctx)
	defer session.Close()

	result, err := session.ListResourceTemplates(ctx, nil)
	require.NoError(t, err)
	// 4 knowledge + 3 project + 1 ticket by ID + 1 persona + 1 session status = 10 templates
	assert.Len(t, result.ResourceTemplates, 10)
}

// --- Streamable HTTP Transport Tests ---

func TestEchoTool_StreamableHTTP(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	session := connectStreamableHTTP(t, ctx)
	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "echo",
		Arguments: map[string]any{"message": "hello via HTTP"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 1)

	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, text.Text, "Echo: hello via HTTP")
}

func TestStaticResource_StreamableHTTP(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	session := connectStreamableHTTP(t, ctx)
	defer session.Close()

	result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "alto://tickets/ready"})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "Ticket Health Report")
}

// --- SSE Transport Tests ---

func TestEchoTool_SSE(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	session := connectSSE(t, ctx)
	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "echo",
		Arguments: map[string]any{"message": "hello via SSE"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 1)

	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, text.Text, "Echo: hello via SSE")
}

func TestStaticResource_SSE(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	session := connectSSE(t, ctx)
	defer session.Close()

	result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "alto://tickets/ready"})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "Ticket Health Report")
}

// --- Error handling tests ---

func TestEchoTool_ErrorIsToolError_NotProtocolError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	session := connectInMemory(t, ctx)
	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "echo",
		Arguments: map[string]any{"message": ""},
	})
	// Tool errors are NOT protocol errors — they come back in the result.
	require.NoError(t, err, "should not be a protocol error")
	assert.True(t, result.IsError, "should be a tool error")
	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, text.Text, "message is required")
}

func TestUnknownTool_ReturnsProtocolError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	session := connectInMemory(t, ctx)
	defer session.Close()

	_, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "nonexistent",
		Arguments: map[string]any{},
	})
	// Unknown tool is a protocol error (JSON-RPC error).
	require.Error(t, err)
}

// --- Concurrent requests test ---

func TestConcurrentToolCalls_NoRace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	session := connectInMemory(t, ctx)
	defer session.Close()

	const n = 20
	errCh := make(chan error, n)
	for i := range n {
		go func(i int) {
			result, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name:      "echo",
				Arguments: map[string]any{"message": fmt.Sprintf("concurrent-%d", i)},
			})
			if err != nil {
				errCh <- err
				return
			}
			if result.IsError {
				errCh <- fmt.Errorf("tool error for concurrent-%d", i)
				return
			}
			errCh <- nil
		}(i)
	}

	for range n {
		err := <-errCh
		assert.NoError(t, err)
	}
}

// --- Auth Middleware Tests ---

func TestAuthMiddleware_RejectsUnauthenticated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	server := newServer(testApp())
	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, nil)

	// Wrap with auth middleware using a simple token verifier.
	authMiddleware := auth.RequireBearerToken(
		func(_ context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
			if token == "valid-test-token" {
				return &auth.TokenInfo{
					UserID:     "test-user",
					Scopes:     []string{"read", "write"},
					Expiration: time.Now().Add(1 * time.Hour),
				}, nil
			}
			return nil, auth.ErrInvalidToken
		},
		&auth.RequireBearerTokenOptions{
			Scopes: []string{"read"},
		},
	)

	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()

	httpServer := &http.Server{
		Handler:           authMiddleware(handler),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() { _ = httpServer.Serve(listener) }()
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	})

	// Unauthenticated POST should get 401.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s", addr), nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_AcceptsValidToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	server := newServer(testApp())
	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, nil)

	authMiddleware := auth.RequireBearerToken(
		func(_ context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
			if token == "valid-test-token" {
				return &auth.TokenInfo{
					UserID:     "test-user",
					Scopes:     []string{"read", "write"},
					Expiration: time.Now().Add(1 * time.Hour),
				}, nil
			}
			return nil, auth.ErrInvalidToken
		},
		&auth.RequireBearerTokenOptions{
			Scopes: []string{"read"},
		},
	)

	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()

	httpServer := &http.Server{
		Handler:           authMiddleware(handler),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() { _ = httpServer.Serve(listener) }()
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	})

	// Authenticated client should connect and call tools.
	client := mcp.NewClient(&mcp.Implementation{Name: "auth-client", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint: fmt.Sprintf("http://%s", addr),
		HTTPClient: &http.Client{
			Transport: &bearerTokenTransport{token: "valid-test-token"},
		},
	}, nil)
	require.NoError(t, err)
	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "echo",
		Arguments: map[string]any{"message": "authenticated"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)
	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, text.Text, "Echo: authenticated")
}

// bearerTokenTransport injects a Bearer token into every HTTP request.
type bearerTokenTransport struct {
	token string
}

func (t *bearerTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return http.DefaultTransport.RoundTrip(req)
}

// --- Helpers ---

func connectInMemory(t *testing.T, ctx context.Context) *mcp.ClientSession {
	t.Helper()
	server := newServer(testApp())
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	ct, st := mcp.NewInMemoryTransports()
	go func() {
		_ = server.Run(ctx, st)
	}()
	session, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })
	return session
}

func connectStreamableHTTP(t *testing.T, ctx context.Context) *mcp.ClientSession {
	t.Helper()
	server := newServer(testApp())
	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, nil)

	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()

	httpServer := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() { _ = httpServer.Serve(listener) }()
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	})

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint: fmt.Sprintf("http://%s", addr),
	}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })
	return session
}

func connectSSE(t *testing.T, ctx context.Context) *mcp.ClientSession {
	t.Helper()
	server := newServer(testApp())
	handler := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, nil)

	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()

	httpServer := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() { _ = httpServer.Serve(listener) }()
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	})

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, &mcp.SSEClientTransport{
		Endpoint: fmt.Sprintf("http://%s", addr),
	}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })
	return session
}
