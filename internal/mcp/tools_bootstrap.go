package mcp

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/alto-cli/alto/internal/composition"
	shareddomain "github.com/alto-cli/alto/internal/shared/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// --- Input structs ---

// InitProjectInput is the typed input for init_project.
type InitProjectInput struct {
	ProjectDir string `json:"project_dir" jsonschema:"the project directory to bootstrap"`
}

// RescueProjectInput is the typed input for rescue_project.
type RescueProjectInput struct {
	ProjectDir string `json:"project_dir" jsonschema:"the existing project directory to rescue"`
}

// GenerateArtifactsInput is the typed input for generate_artifacts.
type GenerateArtifactsInput struct {
	SessionID  string `json:"session_id" jsonschema:"the discovery session ID from guide_complete"`
	ProjectDir string `json:"project_dir" jsonschema:"the project directory to write artifacts to"`
}

// GenerateFitnessInput is the typed input for generate_fitness.
type GenerateFitnessInput struct {
	SessionID  string `json:"session_id" jsonschema:"the session ID from generate_artifacts"`
	ProjectDir string `json:"project_dir" jsonschema:"the project directory"`
}

// GenerateTicketsInput is the typed input for generate_tickets.
type GenerateTicketsInput struct {
	SessionID string `json:"session_id" jsonschema:"the session ID from generate_artifacts"`
}

// GenerateConfigsInput is the typed input for generate_configs.
type GenerateConfigsInput struct {
	SessionID string   `json:"session_id" jsonschema:"the session ID from generate_artifacts"`
	Tools     []string `json:"tools" jsonschema:"tool names to generate configs for (claude-code, cursor, roo-code, opencode)"`
}

// DetectToolsInput is the typed input for detect_tools.
type DetectToolsInput struct {
	ProjectDir string `json:"project_dir" jsonschema:"the project directory to scan for AI coding tools"`
}

// CheckQualityInput is the typed input for check_quality.
type CheckQualityInput struct {
	ProjectDir string   `json:"project_dir" jsonschema:"the project directory"`
	Gates      []string `json:"gates" jsonschema:"quality gates to run (lint, types, tests, fitness)"`
}

// DocHealthInput is the typed input for doc_health.
type DocHealthInput struct {
	ProjectDir string `json:"project_dir" jsonschema:"the project directory to check documentation health"`
}

// DocReviewInput is the typed input for doc_review.
type DocReviewInput struct {
	ProjectDir string `json:"project_dir" jsonschema:"the project directory"`
	DocPath    string `json:"doc_path" jsonschema:"the document path to mark as reviewed"`
}

// TicketHealthInput is the typed input for ticket_health.
type TicketHealthInput struct{}

// TicketVerifyInput is the typed input for ticket_verify.
type TicketVerifyInput struct {
	TicketID string `json:"ticket_id" jsonschema:"the ticket ID to verify claims for"`
}

// SpikeFollowUpAuditInput is the typed input for spike_follow_up_audit.
type SpikeFollowUpAuditInput struct {
	SpikeID    string `json:"spike_id" jsonschema:"the spike ticket ID to audit"`
	ProjectDir string `json:"project_dir" jsonschema:"the project directory"`
}

// KBLookupInput is the typed input for kb_lookup.
type KBLookupInput struct {
	Topic   string `json:"topic" jsonschema:"knowledge topic path (e.g. ddd/aggregate, tools/claude-code/agents)"`
	Version string `json:"version,omitempty" jsonschema:"optional version (defaults to latest)"`
}

// --- Validation helpers ---

// sanitizeProjectDir validates and cleans a user-supplied project directory.
// Rejects traversal sequences, null bytes, and empty input.
// Returns the cleaned path or a tool error.
func sanitizeProjectDir(dir string) (string, *mcp.CallToolResult, any, error) {
	if strings.TrimSpace(dir) == "" {
		r, m, e := toolError("project_dir is required")
		return "", r, m, e
	}

	// Reject null bytes before any path processing.
	if strings.ContainsRune(dir, 0) {
		r, m, e := toolError("project_dir must not contain null bytes")
		return "", r, m, e
	}

	// Reject path traversal components before Clean normalizes them away.
	for _, component := range strings.Split(filepath.ToSlash(dir), "/") {
		if component == ".." {
			r, m, e := toolError("project_dir must not contain path traversal sequences")
			return "", r, m, e
		}
	}

	return filepath.Clean(dir), nil, nil, nil
}

// --- Tool handlers ---

func initProjectHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, InitProjectInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input InitProjectInput) (*mcp.CallToolResult, any, error) {
		projectDir, r, m, e := sanitizeProjectDir(input.ProjectDir)
		if r != nil {
			return r, m, e
		}

		if app.BootstrapHandler == nil {
			return toolError("bootstrap handler not available")
		}

		// MVP: combine Preview + Confirm + Execute into single call.
		session, err := app.BootstrapHandler.Preview(projectDir)
		if err != nil {
			return toolError(fmt.Sprintf("bootstrap preview: %s", err))
		}

		session, err = app.BootstrapHandler.Confirm(session.SessionID())
		if err != nil {
			return toolError(fmt.Sprintf("bootstrap confirm: %s", err))
		}

		session, err = app.BootstrapHandler.Execute(session.SessionID())
		if err != nil {
			return toolError(fmt.Sprintf("bootstrap execute: %s", err))
		}

		return textResult(fmt.Sprintf("Project bootstrapped.\nsession_id: %s\nproject_dir: %s\nstatus: %s",
			session.SessionID(), session.ProjectDir(), session.Status()))
	}
}

func rescueProjectHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, RescueProjectInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input RescueProjectInput) (*mcp.CallToolResult, any, error) {
		projectDir, r, m, e := sanitizeProjectDir(input.ProjectDir)
		if r != nil {
			return r, m, e
		}

		if app.RescueHandler == nil {
			return toolError("rescue handler not available")
		}

		// Detect stack profile for the project.
		profile := detectStackProfile(app, projectDir)

		analysis, err := app.RescueHandler.Rescue(ctx, projectDir, profile, false, false)
		if err != nil {
			return toolError(fmt.Sprintf("rescue: %s", err))
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "Rescue analysis complete.\n")
		fmt.Fprintf(&sb, "analysis_id: %s\n", analysis.AnalysisID())
		fmt.Fprintf(&sb, "project_dir: %s\n", analysis.ProjectDir())
		fmt.Fprintf(&sb, "status: %s\n", analysis.Status())
		fmt.Fprintf(&sb, "gaps: %d\n", len(analysis.Gaps()))

		return textResult(sb.String())
	}
}

func generateArtifactsHandler(app *composition.App, store *ModelStore) func(context.Context, *mcp.CallToolRequest, GenerateArtifactsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input GenerateArtifactsInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}
		projectDir, r, m, e := sanitizeProjectDir(input.ProjectDir)
		if r != nil {
			return r, m, e
		}

		if app.DiscoveryHandler == nil || app.ArtifactGenerationHandler == nil {
			return toolError("discovery or artifact generation handler not available")
		}

		// Get the completed session's event.
		session, err := app.DiscoveryHandler.GetSession(input.SessionID)
		if err != nil {
			return toolError(fmt.Sprintf("get session: %s", err))
		}

		events := session.Events()
		if len(events) == 0 {
			return toolError("session has no completed events — run guide_complete first")
		}

		event := events[0]
		preview, err := app.ArtifactGenerationHandler.BuildPreview(ctx, event)
		if err != nil {
			return toolError(fmt.Sprintf("generate artifacts: %s", err))
		}

		// Write artifacts to project directory.
		docsDir := filepath.Join(projectDir, "docs")
		if err := app.ArtifactGenerationHandler.WriteArtifacts(ctx, preview, docsDir, projectDir); err != nil {
			return toolError(fmt.Sprintf("write artifacts: %s", err))
		}

		// Cache model + detected profile in store.
		profile := detectStackProfile(app, projectDir)
		store.Put(input.SessionID, preview.Model, profile)

		return textResult(fmt.Sprintf("Artifacts generated.\nsession_id: %s\nproject_dir: %s\n"+
			"Files written: docs/PRD.md, docs/DDD.md, docs/ARCHITECTURE.md, .alto/bounded_context_map.yaml\n"+
			"Domain model cached for generate_fitness, generate_tickets, generate_configs.",
			input.SessionID, projectDir))
	}
}

func generateFitnessHandler(app *composition.App, store *ModelStore) func(context.Context, *mcp.CallToolRequest, GenerateFitnessInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input GenerateFitnessInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}
		projectDir, r, m, e := sanitizeProjectDir(input.ProjectDir)
		if r != nil {
			return r, m, e
		}

		model, profile, err := store.Get(input.SessionID)
		if err != nil {
			return toolError(err.Error())
		}

		if app.FitnessGenerationHandler == nil {
			return toolError("fitness generation handler not available")
		}

		projectName := filepath.Base(projectDir)
		preview, err := app.FitnessGenerationHandler.BuildPreview(model, projectName, profile, nil)
		if err != nil {
			return toolError(fmt.Sprintf("generate fitness: %s", err))
		}

		if preview == nil {
			return textResult("No fitness tests generated — stack profile does not support fitness tests.")
		}

		return textResult(fmt.Sprintf("Fitness tests generated.\n%s", preview.Summary))
	}
}

func generateTicketsHandler(app *composition.App, store *ModelStore) func(context.Context, *mcp.CallToolRequest, GenerateTicketsInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input GenerateTicketsInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}

		model, profile, err := store.Get(input.SessionID)
		if err != nil {
			return toolError(err.Error())
		}

		if app.TicketGenerationHandler == nil {
			return toolError("ticket generation handler not available")
		}

		preview, err := app.TicketGenerationHandler.BuildPreview(model, profile)
		if err != nil {
			return toolError(fmt.Sprintf("generate tickets: %s", err))
		}

		return textResult(fmt.Sprintf("Tickets generated.\n%s", preview.Summary))
	}
}

func generateConfigsHandler(app *composition.App, store *ModelStore) func(context.Context, *mcp.CallToolRequest, GenerateConfigsInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input GenerateConfigsInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}

		model, profile, err := store.Get(input.SessionID)
		if err != nil {
			return toolError(err.Error())
		}

		tools, err := ParseSupportedTools(input.Tools)
		if err != nil {
			return toolError(err.Error())
		}

		if app.ConfigGenerationHandler == nil {
			return toolError("config generation handler not available")
		}

		preview, err := app.ConfigGenerationHandler.BuildPreview(model, tools, profile)
		if err != nil {
			return toolError(fmt.Sprintf("generate configs: %s", err))
		}

		return textResult(fmt.Sprintf("Configs generated.\n%s", preview.Summary))
	}
}

func detectToolsHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, DetectToolsInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input DetectToolsInput) (*mcp.CallToolResult, any, error) {
		projectDir, r, m, e := sanitizeProjectDir(input.ProjectDir)
		if r != nil {
			return r, m, e
		}

		if app.DetectionHandler == nil {
			return toolError("detection handler not available")
		}

		result, err := app.DetectionHandler.Detect(projectDir)
		if err != nil {
			return toolError(fmt.Sprintf("detect tools: %s", err))
		}

		detected := result.DetectedTools()
		if len(detected) == 0 {
			return textResult("No AI coding tools detected in project.")
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "Detected %d tool(s):\n", len(detected))
		for _, tool := range detected {
			fmt.Fprintf(&sb, "- %s\n", tool.Name())
		}

		conflicts := result.Conflicts()
		if len(conflicts) > 0 {
			fmt.Fprintf(&sb, "\nConflicts:\n")
			for _, c := range conflicts {
				fmt.Fprintf(&sb, "- %s\n", c)
			}
		}

		return textResult(sb.String())
	}
}

func checkQualityHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, CheckQualityInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input CheckQualityInput) (*mcp.CallToolResult, any, error) {
		if _, r, m, e := sanitizeProjectDir(input.ProjectDir); r != nil {
			return r, m, e
		}

		gates, err := ParseQualityGates(input.Gates)
		if err != nil {
			return toolError(err.Error())
		}

		if app.QualityGateHandler == nil {
			return toolError("quality gate handler not available")
		}

		report, err := app.QualityGateHandler.Check(ctx, gates)
		if err != nil {
			return toolError(fmt.Sprintf("check quality: %s", err))
		}

		var sb strings.Builder
		for _, r := range report.Results() {
			status := "PASS"
			if !r.Passed() {
				status = "FAIL"
			}
			fmt.Fprintf(&sb, "[%s] %s\n", status, r.Gate())
			if r.Output() != "" {
				fmt.Fprintf(&sb, "  %s\n", r.Output())
			}
		}
		fmt.Fprintf(&sb, "\nOverall: ")
		if report.Passed() {
			fmt.Fprint(&sb, "PASSED")
		} else {
			fmt.Fprint(&sb, "FAILED")
		}

		return textResult(sb.String())
	}
}

func docHealthHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, DocHealthInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input DocHealthInput) (*mcp.CallToolResult, any, error) {
		projectDir, r, m, e := sanitizeProjectDir(input.ProjectDir)
		if r != nil {
			return r, m, e
		}

		if app.DocHealthHandler == nil {
			return toolError("doc health handler not available")
		}

		report, err := app.DocHealthHandler.Handle(ctx, projectDir)
		if err != nil {
			return toolError(fmt.Sprintf("doc health: %s", err))
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "Documentation health: %d checked, %d issues\n",
			report.TotalChecked(), report.IssueCount())
		for _, s := range report.Statuses() {
			fmt.Fprintf(&sb, "- %s: %s\n", s.Path(), s.Status())
		}

		return textResult(sb.String())
	}
}

func docReviewHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, DocReviewInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input DocReviewInput) (*mcp.CallToolResult, any, error) {
		projectDir, r, m, e := sanitizeProjectDir(input.ProjectDir)
		if r != nil {
			return r, m, e
		}
		if strings.TrimSpace(input.DocPath) == "" {
			return toolError("doc_path is required")
		}

		if app.DocReviewHandler == nil {
			return toolError("doc review handler not available")
		}

		result, err := app.DocReviewHandler.MarkReviewed(ctx, input.DocPath, projectDir, nil)
		if err != nil {
			return toolError(fmt.Sprintf("doc review: %s", err))
		}

		return textResult(fmt.Sprintf("Document reviewed.\npath: %s\nnew_date: %s",
			result.Path(), result.NewDate().Format("2006-01-02")))
	}
}

func ticketHealthHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, TicketHealthInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ TicketHealthInput) (*mcp.CallToolResult, any, error) {
		if app.TicketHealthHandler == nil {
			return toolError("ticket health handler not available")
		}

		report, err := app.TicketHealthHandler.Report(ctx)
		if err != nil {
			return toolError(fmt.Sprintf("ticket health: %s", err))
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "Ticket health: %d open, %d need review\n",
			report.TotalOpen(), report.ReviewNeededCount())
		fmt.Fprintf(&sb, "Freshness: %.1f%% (%s)\n",
			report.FreshnessPct(), report.FreshnessLabel())
		for _, t := range report.FlaggedTickets() {
			fmt.Fprintf(&sb, "- %s (%s): %d flags\n", t.TicketID(), t.Title(), t.FlagCount())
		}

		return textResult(sb.String())
	}
}

func ticketVerifyHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, TicketVerifyInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input TicketVerifyInput) (*mcp.CallToolResult, any, error) {
		if strings.TrimSpace(input.TicketID) == "" {
			return toolError("ticket_id is required")
		}

		if app.TicketVerifyHandler == nil {
			return toolError("ticket verify handler not available")
		}

		results, err := app.TicketVerifyHandler.Verify(ctx, input.TicketID)
		if err != nil {
			return toolError(fmt.Sprintf("verify ticket: %s", err))
		}

		// Tally results
		claimCount := len(results)
		verified := 0
		mismatches := 0
		for _, r := range results {
			if r.Match() {
				verified++
			} else {
				mismatches++
			}
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "Ticket verification: %s\n", input.TicketID)
		fmt.Fprintf(&sb, "Claims found: %d\n", claimCount)
		fmt.Fprintf(&sb, "Verified: %d\n", verified)
		fmt.Fprintf(&sb, "Mismatches: %d\n", mismatches)

		for _, r := range results {
			if r.Match() {
				fmt.Fprintf(&sb, "\n  [PASS] %s = %s", r.Claim().ClaimText(), r.ActualValue())
			} else {
				fmt.Fprintf(&sb, "\n  [FAIL] %s — %s", r.Claim().ClaimText(), r.Discrepancy())
			}
		}

		return textResult(sb.String())
	}
}

func spikeFollowUpAuditHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, SpikeFollowUpAuditInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input SpikeFollowUpAuditInput) (*mcp.CallToolResult, any, error) {
		if strings.TrimSpace(input.SpikeID) == "" {
			return toolError("spike_id is required")
		}
		projectDir, r, m, e := sanitizeProjectDir(input.ProjectDir)
		if r != nil {
			return r, m, e
		}

		if app.SpikeFollowUpHandler == nil {
			return toolError("spike follow-up handler not available")
		}

		result, err := app.SpikeFollowUpHandler.Audit(ctx, input.SpikeID, projectDir)
		if err != nil {
			return toolError(fmt.Sprintf("spike follow-up audit: %s", err))
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "Spike follow-up audit: %s\n", result.SpikeID())
		fmt.Fprintf(&sb, "Defined intents: %d\n", result.DefinedCount())
		fmt.Fprintf(&sb, "Orphaned intents: %d\n", result.OrphanedCount())
		if result.HasOrphans() {
			fmt.Fprintf(&sb, "\nOrphaned follow-ups (no ticket created):\n")
			for _, intent := range result.OrphanedIntents() {
				fmt.Fprintf(&sb, "- %s\n", intent.Description())
			}
		}

		return textResult(sb.String())
	}
}

func kbLookupHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, KBLookupInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input KBLookupInput) (*mcp.CallToolResult, any, error) {
		if strings.TrimSpace(input.Topic) == "" {
			return toolError("topic is required")
		}

		if app.KnowledgeLookupHandler == nil {
			return toolError("knowledge lookup handler not available")
		}

		entry, err := app.KnowledgeLookupHandler.Lookup(ctx, input.Topic, input.Version)
		if err != nil {
			return toolError(fmt.Sprintf("lookup: %s", err))
		}

		return textResult(entry.Content())
	}
}

// --- Coordinator-aware handlers ---

func generateArtifactsHandlerWithCoordinator(app *composition.App, coord *shareddomain.WorkflowCoordinator) func(context.Context, *mcp.CallToolRequest, GenerateArtifactsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input GenerateArtifactsInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}
		projectDir, r, m, e := sanitizeProjectDir(input.ProjectDir)
		if r != nil {
			return r, m, e
		}

		// Check precondition
		if !coord.CanExecute(input.SessionID, shareddomain.StepArtifactGeneration) {
			return toolError("precondition not met: discovery must complete first")
		}

		// Begin step
		if err := coord.BeginStep(input.SessionID, shareddomain.StepArtifactGeneration); err != nil {
			return toolError(fmt.Sprintf("begin step: %s", err))
		}

		if app.DiscoveryHandler == nil || app.ArtifactGenerationHandler == nil {
			return toolError("discovery or artifact generation handler not available")
		}

		// Get the completed session's event.
		session, err := app.DiscoveryHandler.GetSession(input.SessionID)
		if err != nil {
			return toolError(fmt.Sprintf("get session: %s", err))
		}

		events := session.Events()
		if len(events) == 0 {
			return toolError("session has no completed events — run guide_complete first")
		}

		event := events[0]
		preview, err := app.ArtifactGenerationHandler.BuildPreview(ctx, event)
		if err != nil {
			return toolError(fmt.Sprintf("generate artifacts: %s", err))
		}

		// Write artifacts to project directory.
		docsDir := filepath.Join(projectDir, "docs")
		if err := app.ArtifactGenerationHandler.WriteArtifacts(ctx, preview, docsDir, projectDir); err != nil {
			return toolError(fmt.Sprintf("write artifacts: %s", err))
		}

		// Store session context in coordinator
		profile := detectStackProfile(app, projectDir)
		sessionCtx := &shareddomain.SessionContext{
			SessionID:    input.SessionID,
			DomainModel:  preview.Model,
			StackProfile: profile,
			ProjectDir:   projectDir,
		}
		if err := coord.SetSessionContext(input.SessionID, sessionCtx); err != nil {
			return toolError(fmt.Sprintf("set session context: %s", err))
		}

		// Complete step
		if err := coord.CompleteStep(input.SessionID, shareddomain.StepArtifactGeneration); err != nil {
			return toolError(fmt.Sprintf("complete step: %s", err))
		}

		return textResult(fmt.Sprintf("Artifacts generated.\nsession_id: %s\nproject_dir: %s\n"+
			"Files written: docs/PRD.md, docs/DDD.md, docs/ARCHITECTURE.md, .alto/bounded_context_map.yaml\n"+
			"Session context stored for generate_fitness, generate_tickets, generate_configs.",
			input.SessionID, projectDir))
	}
}

func generateFitnessHandlerWithCoordinator(app *composition.App, coord *shareddomain.WorkflowCoordinator) func(context.Context, *mcp.CallToolRequest, GenerateFitnessInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input GenerateFitnessInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}
		projectDir, r, m, e := sanitizeProjectDir(input.ProjectDir)
		if r != nil {
			return r, m, e
		}

		// Check precondition
		if !coord.CanExecute(input.SessionID, shareddomain.StepFitness) {
			return toolError("precondition not met: artifact generation must complete first")
		}

		// Begin step
		if err := coord.BeginStep(input.SessionID, shareddomain.StepFitness); err != nil {
			return toolError(fmt.Sprintf("begin step: %s", err))
		}

		// Get session context
		sessionCtx, err := coord.SessionContext(input.SessionID)
		if err != nil {
			return toolError(fmt.Sprintf("session context: %s", err))
		}

		if app.FitnessGenerationHandler == nil {
			return toolError("fitness generation handler not available")
		}

		projectName := filepath.Base(projectDir)
		preview, err := app.FitnessGenerationHandler.BuildPreview(sessionCtx.DomainModel, projectName, sessionCtx.StackProfile, nil)
		if err != nil {
			return toolError(fmt.Sprintf("generate fitness: %s", err))
		}

		// Complete step
		if err := coord.CompleteStep(input.SessionID, shareddomain.StepFitness); err != nil {
			return toolError(fmt.Sprintf("complete step: %s", err))
		}

		if preview == nil {
			return textResult("No fitness tests generated — stack profile does not support fitness tests.")
		}

		return textResult(fmt.Sprintf("Fitness tests generated.\n%s", preview.Summary))
	}
}

func generateTicketsHandlerWithCoordinator(app *composition.App, coord *shareddomain.WorkflowCoordinator) func(context.Context, *mcp.CallToolRequest, GenerateTicketsInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input GenerateTicketsInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}

		// Check precondition
		if !coord.CanExecute(input.SessionID, shareddomain.StepTickets) {
			return toolError("precondition not met: artifact generation must complete first")
		}

		// Begin step
		if err := coord.BeginStep(input.SessionID, shareddomain.StepTickets); err != nil {
			return toolError(fmt.Sprintf("begin step: %s", err))
		}

		// Get session context
		sessionCtx, err := coord.SessionContext(input.SessionID)
		if err != nil {
			return toolError(fmt.Sprintf("session context: %s", err))
		}

		if app.TicketGenerationHandler == nil {
			return toolError("ticket generation handler not available")
		}

		preview, err := app.TicketGenerationHandler.BuildPreview(sessionCtx.DomainModel, sessionCtx.StackProfile)
		if err != nil {
			return toolError(fmt.Sprintf("generate tickets: %s", err))
		}

		// Complete step
		if err := coord.CompleteStep(input.SessionID, shareddomain.StepTickets); err != nil {
			return toolError(fmt.Sprintf("complete step: %s", err))
		}

		return textResult(fmt.Sprintf("Tickets generated.\n%s", preview.Summary))
	}
}

func generateConfigsHandlerWithCoordinator(app *composition.App, coord *shareddomain.WorkflowCoordinator) func(context.Context, *mcp.CallToolRequest, GenerateConfigsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input GenerateConfigsInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}

		// Check precondition
		if !coord.CanExecute(input.SessionID, shareddomain.StepConfigs) {
			return toolError("precondition not met: artifact generation must complete first")
		}

		// Begin step
		if err := coord.BeginStep(input.SessionID, shareddomain.StepConfigs); err != nil {
			return toolError(fmt.Sprintf("begin step: %s", err))
		}

		// Get session context
		sessionCtx, err := coord.SessionContext(input.SessionID)
		if err != nil {
			return toolError(fmt.Sprintf("session context: %s", err))
		}

		tools, err := ParseSupportedTools(input.Tools)
		if err != nil {
			return toolError(err.Error())
		}

		if app.ConfigGenerationHandler == nil {
			return toolError("config generation handler not available")
		}

		preview, err := app.ConfigGenerationHandler.BuildPreview(sessionCtx.DomainModel, tools, sessionCtx.StackProfile)
		if err != nil {
			return toolError(fmt.Sprintf("generate configs: %s", err))
		}

		// Write config files to project directory
		if err := app.ConfigGenerationHandler.ApproveAndWrite(ctx, preview, sessionCtx.ProjectDir); err != nil {
			return toolError(fmt.Sprintf("write configs: %s", err))
		}

		// Complete step
		if err := coord.CompleteStep(input.SessionID, shareddomain.StepConfigs); err != nil {
			return toolError(fmt.Sprintf("complete step: %s", err))
		}

		return textResult(fmt.Sprintf("Configs generated and written to %s.\n%s", sessionCtx.ProjectDir, preview.Summary))
	}
}

// detectStackProfile uses DetectionHandler to determine the project's stack profile.
func detectStackProfile(app *composition.App, projectDir string) vo.StackProfile {
	if app.DetectionHandler == nil {
		return vo.GenericProfile{}
	}

	result, err := app.DetectionHandler.Detect(projectDir)
	if err != nil {
		return vo.GenericProfile{}
	}

	for _, tool := range result.DetectedTools() {
		name := strings.ToLower(tool.Name())
		if strings.Contains(name, "python") || strings.Contains(name, "uv") {
			return vo.PythonUvProfile{}
		}
	}

	return vo.GenericProfile{}
}

// --- Registration ---

// RegisterBootstrapToolsWithCoordinator registers all 14 bootstrap and generation MCP tools
// using WorkflowCoordinator for precondition checking and lifecycle tracking.
func RegisterBootstrapToolsWithCoordinator(server *mcp.Server, app *composition.App, coord *shareddomain.WorkflowCoordinator) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "init_project",
		Description: "Bootstrap a new project with DDD structure",
	}, initProjectHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "rescue_project",
		Description: "Rescue an existing project — analyse gaps and create migration plan",
	}, rescueProjectHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_artifacts",
		Description: "Generate DDD artifacts (PRD, DDD.md, Architecture) from a completed discovery session",
	}, generateArtifactsHandlerWithCoordinator(app, coord))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_fitness",
		Description: "Generate architecture fitness tests from domain model (requires generate_artifacts first)",
	}, generateFitnessHandlerWithCoordinator(app, coord))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_tickets",
		Description: "Generate implementation tickets from domain model (requires generate_artifacts first)",
	}, generateTicketsHandlerWithCoordinator(app, coord))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_configs",
		Description: "Generate tool-native configs for AI coding tools (requires generate_artifacts first)",
	}, generateConfigsHandlerWithCoordinator(app, coord))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "detect_tools",
		Description: "Detect AI coding tools installed in a project",
	}, detectToolsHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_quality",
		Description: "Run quality gates (lint, types, tests, fitness) on a project",
	}, checkQualityHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "doc_health",
		Description: "Check documentation health — find stale or missing docs",
	}, docHealthHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "doc_review",
		Description: "Mark a document as reviewed with current date",
	}, docReviewHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ticket_health",
		Description: "Show ticket health report — open tickets, freshness, flagged items",
	}, ticketHealthHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ticket_verify",
		Description: "Verify quantitative claims in a ticket against actual command output",
	}, ticketVerifyHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "spike_follow_up_audit",
		Description: "Audit a spike's follow-up intents — find orphaned research that never became tickets",
	}, spikeFollowUpAuditHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "kb_lookup",
		Description: "Look up a knowledge base entry by topic",
	}, kbLookupHandler(app))
}

// RegisterBootstrapTools registers all 14 bootstrap and generation MCP tools.
//
// Deprecated: Use RegisterBootstrapToolsWithCoordinator for precondition checking.
func RegisterBootstrapTools(server *mcp.Server, app *composition.App, store *ModelStore) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "init_project",
		Description: "Bootstrap a new project with DDD structure",
	}, initProjectHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "rescue_project",
		Description: "Rescue an existing project — analyse gaps and create migration plan",
	}, rescueProjectHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_artifacts",
		Description: "Generate DDD artifacts (PRD, DDD.md, Architecture) from a completed discovery session",
	}, generateArtifactsHandler(app, store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_fitness",
		Description: "Generate architecture fitness tests from domain model (requires generate_artifacts first)",
	}, generateFitnessHandler(app, store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_tickets",
		Description: "Generate implementation tickets from domain model (requires generate_artifacts first)",
	}, generateTicketsHandler(app, store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_configs",
		Description: "Generate tool-native configs for AI coding tools (requires generate_artifacts first)",
	}, generateConfigsHandler(app, store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "detect_tools",
		Description: "Detect AI coding tools installed in a project",
	}, detectToolsHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_quality",
		Description: "Run quality gates (lint, types, tests, fitness) on a project",
	}, checkQualityHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "doc_health",
		Description: "Check documentation health — find stale or missing docs",
	}, docHealthHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "doc_review",
		Description: "Mark a document as reviewed with current date",
	}, docReviewHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ticket_health",
		Description: "Show ticket health report — open tickets, freshness, flagged items",
	}, ticketHealthHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ticket_verify",
		Description: "Verify quantitative claims in a ticket against actual command output",
	}, ticketVerifyHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "spike_follow_up_audit",
		Description: "Audit a spike's follow-up intents — find orphaned research that never became tickets",
	}, spikeFollowUpAuditHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "kb_lookup",
		Description: "Look up a knowledge base entry by topic",
	}, kbLookupHandler(app))
}
