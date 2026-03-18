package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/alto-cli/alto/internal/composition"
	shareddomain "github.com/alto-cli/alto/internal/shared/domain"
)

// --- URI parsing helpers ---

// extractURIPath parses an alto:// URI and returns the path (without leading slash).
func extractURIPath(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("invalid URI: %w", err)
	}
	path := u.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	return path, nil
}

// resourceError returns a resource error result.
func resourceError(uri, msg string) (*mcp.ReadResourceResult, error) {
	return nil, fmt.Errorf("%s: %s", uri, msg)
}

// resourceText returns a text resource result.
func resourceText(uri, mimeType, text string) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: uri, MIMEType: mimeType, Text: text},
		},
	}, nil
}

// --- Knowledge resource handlers ---

func knowledgeDDDHandler(app *composition.App) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// alto://knowledge/ddd/{topic} → host=knowledge, path=/ddd/{topic}
		path, err := extractURIPath(req.Params.URI)
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}
		// path = "ddd/{topic}"
		topic := strings.TrimPrefix(path, "ddd/")
		if compErr := SafeComponent(topic); compErr != nil {
			return resourceError(req.Params.URI, compErr.Error())
		}
		entry, err := app.KnowledgeLookupHandler.Lookup(ctx, "ddd/"+topic, "")
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}
		return resourceText(req.Params.URI, "text/markdown", entry.Content())
	}
}

func knowledgeToolsHandler(app *composition.App) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// alto://knowledge/tools/{tool}/{subtopic}
		path, err := extractURIPath(req.Params.URI)
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}
		// path = "tools/{tool}/{subtopic}"
		rest := strings.TrimPrefix(path, "tools/")
		parts := strings.SplitN(rest, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return resourceError(req.Params.URI, "expected tools/{tool}/{subtopic}")
		}
		tool, subtopic := parts[0], parts[1]
		if compErr := SafeComponent(tool); compErr != nil {
			return resourceError(req.Params.URI, compErr.Error())
		}
		if compErr := SafeComponent(subtopic); compErr != nil {
			return resourceError(req.Params.URI, compErr.Error())
		}
		entry, err := app.KnowledgeLookupHandler.Lookup(ctx, "tools/"+tool+"/"+subtopic, "")
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}
		return resourceText(req.Params.URI, "text/markdown", entry.Content())
	}
}

func knowledgeConventionsHandler(app *composition.App) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		path, err := extractURIPath(req.Params.URI)
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}
		topic := strings.TrimPrefix(path, "conventions/")
		if compErr := SafeComponent(topic); compErr != nil {
			return resourceError(req.Params.URI, compErr.Error())
		}
		entry, err := app.KnowledgeLookupHandler.Lookup(ctx, "conventions/"+topic, "")
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}
		return resourceText(req.Params.URI, "text/markdown", entry.Content())
	}
}

func knowledgeCrossToolHandler(app *composition.App) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		path, err := extractURIPath(req.Params.URI)
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}
		topic := strings.TrimPrefix(path, "cross-tool/")
		if compErr := SafeComponent(topic); compErr != nil {
			return resourceError(req.Params.URI, compErr.Error())
		}
		entry, err := app.KnowledgeLookupHandler.Lookup(ctx, "cross-tool/"+topic, "")
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}
		return resourceText(req.Params.URI, "text/markdown", entry.Content())
	}
}

// --- Project document resource handlers ---

func projectDocHandler(_ *composition.App, docSubpath, docName string) mcp.ResourceHandler {
	return func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// alto://project/{dir}/{docSubpath} → host=project, path=/{dir}/{docSubpath}
		u, err := url.Parse(req.Params.URI)
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}
		// host = "project", path = "/{dir}/domain-model"
		path := u.Path
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}
		// Strip the trailing doc suffix to get the dir
		dir := strings.TrimSuffix(path, "/"+docSubpath)
		if dir == "" || dir == path {
			return resourceError(req.Params.URI, "project directory is required")
		}

		// Validate the project directory path. Use CWD as allowed root.
		cwd, err := os.Getwd()
		if err != nil {
			return resourceError(req.Params.URI, "cannot determine working directory")
		}
		resolvedDir, err := SafeProjectPath(dir, []string{cwd})
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}

		filePath := filepath.Join(resolvedDir, "docs", docName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				return resourceError(req.Params.URI, fmt.Sprintf("%s not found in project", docName))
			}
			return resourceError(req.Params.URI, fmt.Sprintf("reading %s: %v", docName, err))
		}
		return resourceText(req.Params.URI, "text/markdown", string(content))
	}
}

// --- Ticket resource handlers ---

func ticketsReadyHandler(app *composition.App) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		report, err := app.TicketHealthHandler.Report(ctx)
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "Ticket Health Report\n")
		fmt.Fprintf(&sb, "Total open: %d\n", report.TotalOpen())
		fmt.Fprintf(&sb, "Freshness: %.0f%% (%s)\n", report.FreshnessPct(), report.FreshnessLabel())
		fmt.Fprintf(&sb, "Review needed: %d\n", report.ReviewNeededCount())
		if oldest := report.OldestLastReviewed(); oldest != nil {
			fmt.Fprintf(&sb, "Oldest last reviewed: %s\n", *oldest)
		}
		for _, ft := range report.FlaggedTickets() {
			fmt.Fprintf(&sb, "- %s: %s (%d flags)\n", ft.TicketID(), ft.Title(), ft.FlagCount())
		}
		return resourceText(req.Params.URI, "text/plain", sb.String())
	}
}

func ticketByIDHandler(_ *composition.App) mcp.ResourceHandler {
	return func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// alto://tickets/{id} → host=tickets, path=/{id}
		u, err := url.Parse(req.Params.URI)
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}
		id := u.Path
		if len(id) > 0 && id[0] == '/' {
			id = id[1:]
		}
		if err := SafeTicketID(id); err != nil {
			return resourceError(req.Params.URI, err.Error())
		}

		// Use bd show subprocess to get ticket details.
		// NOTE: In production, bd must be in PATH. For now, return a placeholder
		// since the subprocess approach is deferred to integration testing.
		return resourceError(req.Params.URI, fmt.Sprintf("ticket lookup for %q requires bd binary (not available in this context)", id))
	}
}

// --- Persona resource handler ---

func personaHandler(app *composition.App) mcp.ResourceHandler {
	return func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// alto://personas/{type} → host=personas, path=/{type}
		u, err := url.Parse(req.Params.URI)
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}
		pType := u.Path
		if len(pType) > 0 && pType[0] == '/' {
			pType = pType[1:]
		}
		if compErr := SafeComponent(pType); compErr != nil {
			return resourceError(req.Params.URI, compErr.Error())
		}

		personas := app.PersonaHandler.ListPersonas()
		for _, p := range personas {
			if strings.EqualFold(string(p.PersonaType()), pType) {
				text := fmt.Sprintf("Name: %s\nType: %s\nRegister: %s\nDescription: %s\n\n%s",
					p.Name(), p.PersonaType(), p.Register(), p.Description(), p.InstructionsTemplate())
				return resourceText(req.Params.URI, "text/plain", text)
			}
		}
		return resourceError(req.Params.URI, fmt.Sprintf("persona %q not found", pType))
	}
}

// --- Registration ---

// RegisterResources registers all MCP resources for the alto server.
func RegisterResources(server *mcp.Server, app *composition.App) {
	// Knowledge base resources (4 templates)
	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "DDD Knowledge",
		Description: "DDD knowledge base entries by topic",
		MIMEType:    "text/markdown",
		URITemplate: "alto://knowledge/ddd/{topic}",
	}, knowledgeDDDHandler(app))

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "Tool Knowledge",
		Description: "Tool-specific knowledge base entries",
		MIMEType:    "text/markdown",
		URITemplate: "alto://knowledge/tools/{tool}/{subtopic}",
	}, knowledgeToolsHandler(app))

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "Conventions Knowledge",
		Description: "Project convention knowledge base entries",
		MIMEType:    "text/markdown",
		URITemplate: "alto://knowledge/conventions/{topic}",
	}, knowledgeConventionsHandler(app))

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "Cross-Tool Knowledge",
		Description: "Cross-tool knowledge base entries",
		MIMEType:    "text/markdown",
		URITemplate: "alto://knowledge/cross-tool/{topic}",
	}, knowledgeCrossToolHandler(app))

	// Project document resources (3 templates)
	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "Project Domain Model",
		Description: "Project DDD domain model (docs/DDD.md)",
		MIMEType:    "text/markdown",
		URITemplate: "alto://project/{dir}/domain-model",
	}, projectDocHandler(app, "domain-model", "DDD.md"))

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "Project Architecture",
		Description: "Project architecture document (docs/ARCHITECTURE.md)",
		MIMEType:    "text/markdown",
		URITemplate: "alto://project/{dir}/architecture",
	}, projectDocHandler(app, "architecture", "ARCHITECTURE.md"))

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "Project PRD",
		Description: "Project product requirements document (docs/PRD.md)",
		MIMEType:    "text/markdown",
		URITemplate: "alto://project/{dir}/prd",
	}, projectDocHandler(app, "prd", "PRD.md"))

	// Ticket resources (1 static + 1 template)
	server.AddResource(&mcp.Resource{
		Name:     "Ready Tickets",
		MIMEType: "text/plain",
		URI:      "alto://tickets/ready",
	}, ticketsReadyHandler(app))

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "Ticket by ID",
		Description: "Ticket details by ID",
		MIMEType:    "text/plain",
		URITemplate: "alto://tickets/{id}",
	}, ticketByIDHandler(app))

	// Persona resource (1 template)
	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "Persona",
		Description: "Persona definition by type (e.g. solo_developer, team_lead)",
		MIMEType:    "text/plain",
		URITemplate: "alto://personas/{type}",
	}, personaHandler(app))
}

// --- Session Status resource handler ---

func sessionStatusHandler(coord *shareddomain.WorkflowCoordinator) mcp.ResourceHandler {
	return func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// alto://session/{session_id}/status
		u, err := url.Parse(req.Params.URI)
		if err != nil {
			return resourceError(req.Params.URI, err.Error())
		}
		path := u.Path
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}
		// path = "{session_id}/status"
		sessionID := strings.TrimSuffix(path, "/status")
		if sessionID == "" || sessionID == path {
			return resourceError(req.Params.URI, "session_id is required")
		}

		// Build status response
		actions := coord.AvailableActions(sessionID)
		actionNames := make([]string, 0, len(actions))
		for _, a := range actions {
			actionNames = append(actionNames, a.Name())
		}

		// Build steps status map
		steps := make(map[string]string)
		for _, step := range shareddomain.AllWorkflowSteps() {
			status := coord.StepStatus(sessionID, step)
			steps[step.String()] = status.String()
		}

		response := struct {
			SessionID        string            `json:"session_id"`
			Steps            map[string]string `json:"steps"`
			AvailableActions []string          `json:"available_actions"`
		}{
			SessionID:        sessionID,
			Steps:            steps,
			AvailableActions: actionNames,
		}

		jsonBytes, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return resourceError(req.Params.URI, fmt.Sprintf("marshal response: %v", err))
		}

		return resourceText(req.Params.URI, "application/json", string(jsonBytes))
	}
}

// RegisterResourcesWithCoordinator registers all MCP resources including session_status.
func RegisterResourcesWithCoordinator(server *mcp.Server, app *composition.App, coord *shareddomain.WorkflowCoordinator) {
	// Register all standard resources
	RegisterResources(server, app)

	// Add session status resource
	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "Session Status",
		Description: "Workflow session status showing step states and available actions",
		MIMEType:    "application/json",
		URITemplate: "alto://session/{session_id}/status",
	}, sessionStatusHandler(coord))
}
