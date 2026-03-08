// cmd/alty-mcp is the MCP server binary for alty.
//
// It exposes alty's capabilities (bootstrap, discovery, knowledge base) as MCP
// tools and resources over three transports: stdio, SSE, and Streamable HTTP.
//
// Usage:
//
//	alty-mcp                          # stdio (default, for Claude Code)
//	alty-mcp --transport=sse          # SSE on localhost:8080
//	alty-mcp --transport=http         # Streamable HTTP on localhost:8080
//	alty-mcp --transport=all          # All 3 simultaneously
//	alty-mcp --transport=http --addr=0.0.0.0:9090  # Custom bind address
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const version = "0.1.0"

func main() {
	transport := flag.String("transport", "stdio", "Transport mode: stdio, sse, http, all")
	addr := flag.String("addr", "127.0.0.1:8080", "HTTP listen address (for sse/http/all modes)")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	slog.Info("alty-mcp starting",
		"version", version,
		"transport", *transport,
		"implementation", "alty-mcp",
	)

	server := newServer()

	validTransports := map[string]bool{"stdio": true, "sse": true, "http": true, "all": true}
	if !validTransports[*transport] {
		log.Fatalf("unknown transport: %s (valid: stdio, sse, http, all)", *transport)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	switch *transport {
	case "stdio":
		runStdio(ctx, server)
	case "sse":
		runHTTP(ctx, server, *addr, "sse")
	case "http":
		runHTTP(ctx, server, *addr, "streamable")
	case "all":
		runAll(ctx, server, *addr)
	}
}

// newServer creates the MCP server with all tools and resources registered.
func newServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "alty-mcp",
		Version: version,
	}, nil)

	// Add MCP-level logging middleware.
	server.AddReceivingMiddleware(loggingMiddleware())

	registerTools(server)
	registerResources(server)

	return server
}

// --- Tools ---

// EchoInput is the typed input for the echo tool.
type EchoInput struct {
	Message string `json:"message" jsonschema:"the message to echo back"`
}

func registerTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "echo",
		Description: "Echo back the input message (spike PoC tool)",
	}, echoHandler)
}

func echoHandler(_ context.Context, _ *mcp.CallToolRequest, input EchoInput) (*mcp.CallToolResult, any, error) {
	if input.Message == "" {
		return nil, nil, fmt.Errorf("message is required")
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Echo: " + input.Message},
		},
	}, nil, nil
}

// --- Resources ---

func registerResources(server *mcp.Server) {
	// Static resource.
	server.AddResource(&mcp.Resource{
		Name:     "Hello Resource",
		MIMEType: "text/plain",
		URI:      "alty://test/hello",
	}, helloResourceHandler)

	// Resource template with dynamic URI parameter.
	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "Test Resource Template",
		Description: "A test resource template with dynamic name parameter",
		MIMEType:    "text/plain",
		URITemplate: "alty://test/{name}",
	}, testTemplateHandler)
}

func helloResourceHandler(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "text/plain", Text: "Hello from alty-mcp!"},
		},
	}, nil
}

func testTemplateHandler(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	u, err := url.Parse(req.Params.URI)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}
	// alty://test/{name} parses as: scheme=alty, host=test, path=/{name}
	name := u.Path
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	if name == "" {
		return nil, fmt.Errorf("name parameter is required")
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "text/plain", Text: fmt.Sprintf("Hello, %s! (from template)", name)},
		},
	}, nil
}

// --- Transport runners ---

func runStdio(ctx context.Context, server *mcp.Server) {
	slog.Info("serving over stdio")
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("stdio server failed: %v", err)
	}
}

func runHTTP(ctx context.Context, server *mcp.Server, addr, mode string) {
	mux := http.NewServeMux()

	switch mode {
	case "sse":
		handler := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server {
			return server
		}, nil)
		mux.Handle("/sse", handler)
		slog.Info("SSE handler registered", "path", "/sse")
	case "streamable":
		handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
			return server
		}, nil)
		mux.Handle("/mcp", handler)
		slog.Info("Streamable HTTP handler registered", "path", "/mcp")
	}

	serveHTTP(ctx, mux, addr)
}

func runAll(ctx context.Context, server *mcp.Server, addr string) {
	mux := http.NewServeMux()

	// SSE on /sse
	sseHandler := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, nil)
	mux.Handle("/sse", sseHandler)
	slog.Info("SSE handler registered", "path", "/sse")

	// Streamable HTTP on /mcp
	streamableHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, nil)
	mux.Handle("/mcp", streamableHandler)
	slog.Info("Streamable HTTP handler registered", "path", "/mcp")

	// stdio in background goroutine
	go func() {
		slog.Info("serving stdio in background")
		if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
			slog.Error("stdio server failed", "error", err)
		}
	}()

	serveHTTP(ctx, mux, addr)
}

func serveHTTP(ctx context.Context, handler http.Handler, addr string) {
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Graceful shutdown on context cancellation.
	go func() {
		<-ctx.Done()
		slog.Info("shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("HTTP shutdown error", "error", err)
		}
	}()

	slog.Info("HTTP server listening", "addr", addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

// --- Middleware ---

func loggingMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()
			sessionID := req.GetSession().ID()

			result, err := next(ctx, method, req)

			duration := time.Since(start)
			if err != nil {
				slog.Error("MCP request failed",
					"session", sessionID,
					"method", method,
					"duration", duration,
					"error", err,
				)
			} else {
				slog.Info("MCP request",
					"session", sessionID,
					"method", method,
					"duration", duration,
				)
			}
			return result, err
		}
	}
}
