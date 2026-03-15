// Package application provides command handlers for the Bootstrap bounded context.
package application

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/alty-cli/alty/internal/bootstrap/domain"
	sharedapp "github.com/alty-cli/alty/internal/shared/application"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ToolDetector detects installed AI coding tools and scans for config conflicts.
// Defined where consumed per Go convention (no ctx parameter needed for sync ops).
type ToolDetector interface {
	Detect(projectDir string) ([]string, error)
	ScanConflicts(projectDir string) ([]string, error)
}

// FileChecker checks whether a file exists. Extracted as a port so the
// application layer has no direct os.Stat dependency.
type FileChecker interface {
	Exists(path string) bool
}

// ContentProvider returns generated content for a planned file path.
type ContentProvider interface {
	ContentFor(path string, config domain.ProjectConfig) string
}

// plannedFiles lists files alty plans to create in a new project.
var plannedFiles = []string{
	"docs/PRD.md",
	"docs/DDD.md",
	"docs/ARCHITECTURE.md",
	"AGENTS.md",
	".alty/config.toml",
	".alty/knowledge/_index.toml",
	".alty/maintenance/doc-registry.toml",
}

// ScaffoldCommitMessage is the commit message used when auto-committing scaffold files.
const ScaffoldCommitMessage = "chore: initialize alty project structure"

// BootstrapHandler orchestrates the preview -> confirm -> execute bootstrap flow.
type BootstrapHandler struct {
	toolDetection   ToolDetector
	fileChecker     FileChecker
	publisher       sharedapp.EventPublisher
	fileWriter      sharedapp.FileWriter
	contentProvider ContentProvider
	gitCommitter    GitCommitter
	mu              sync.Mutex
	sessions        map[string]*domain.BootstrapSession
	configs         map[string]domain.ProjectConfig
}

// BootstrapOption configures optional dependencies for BootstrapHandler.
type BootstrapOption func(*BootstrapHandler)

// WithGitCommitter injects an optional GitCommitter for auto-committing scaffold files.
func WithGitCommitter(gc GitCommitter) BootstrapOption {
	return func(h *BootstrapHandler) {
		h.gitCommitter = gc
	}
}

// NewBootstrapHandler creates a new BootstrapHandler with injected dependencies.
func NewBootstrapHandler(toolDetection ToolDetector, fileChecker FileChecker, publisher sharedapp.EventPublisher, fileWriter sharedapp.FileWriter, contentProvider ContentProvider, opts ...BootstrapOption) *BootstrapHandler {
	h := &BootstrapHandler{
		toolDetection:   toolDetection,
		fileChecker:     fileChecker,
		publisher:       publisher,
		fileWriter:      fileWriter,
		contentProvider: contentProvider,
		sessions:        make(map[string]*domain.BootstrapSession),
		configs:         make(map[string]domain.ProjectConfig),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Preview creates a new session and generates a preview of planned actions.
func (h *BootstrapHandler) Preview(projectDir string) (*domain.BootstrapSession, error) {
	readme := filepath.Join(projectDir, "README.md")
	if !h.fileChecker.Exists(readme) {
		return nil, fmt.Errorf("create a README.md with your project idea first")
	}

	tools, err := h.toolDetection.Detect(projectDir)
	if err != nil {
		return nil, fmt.Errorf("detecting tools: %w", err)
	}

	conflicts, err := h.toolDetection.ScanConflicts(projectDir)
	if err != nil {
		return nil, fmt.Errorf("scanning conflicts: %w", err)
	}

	var fileActions []vo.FileAction
	for _, planned := range plannedFiles {
		fullPath := filepath.Join(projectDir, planned)
		if h.fileChecker.Exists(fullPath) {
			fileActions = append(fileActions,
				vo.NewFileAction(planned, vo.FileActionSkip, "already exists", ""))
		} else {
			fileActions = append(fileActions,
				vo.NewFileAction(planned, vo.FileActionCreate, "", ""))
		}
	}

	preview := vo.NewPreview(fileActions, nil, conflicts)
	session := domain.NewBootstrapSession(projectDir)
	session.SetDetectedTools(tools)
	if err := session.SetPreview(&preview); err != nil {
		return nil, fmt.Errorf("setting preview: %w", err)
	}

	h.mu.Lock()
	h.sessions[session.SessionID()] = session
	h.mu.Unlock()
	return session, nil
}

// SetGitCommitter sets or clears the GitCommitter at runtime.
// Used by the CLI to honor the --no-commit flag.
func (h *BootstrapHandler) SetGitCommitter(gc GitCommitter) {
	h.gitCommitter = gc
}

// WithProjectConfig associates a ProjectConfig with a session. Must be called
// before Execute so that content generation receives detected project settings.
func (h *BootstrapHandler) WithProjectConfig(sessionID string, config domain.ProjectConfig) {
	h.mu.Lock()
	h.configs[sessionID] = config
	h.mu.Unlock()
}

// Confirm confirms a previewed session, enabling execution.
func (h *BootstrapHandler) Confirm(sessionID string) (*domain.BootstrapSession, error) {
	session, err := h.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.Confirm(); err != nil {
		return nil, fmt.Errorf("confirm session: %w", err)
	}
	return session, nil
}

// Cancel cancels a previewed session.
func (h *BootstrapHandler) Cancel(sessionID string) (*domain.BootstrapSession, error) {
	session, err := h.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.Cancel(); err != nil {
		return nil, fmt.Errorf("cancel session: %w", err)
	}
	return session, nil
}

// Execute executes a confirmed session.
func (h *BootstrapHandler) Execute(sessionID string) (*domain.BootstrapSession, error) {
	session, err := h.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.BeginExecution(); err != nil {
		return nil, fmt.Errorf("begin execution: %w", err)
	}

	h.mu.Lock()
	config, hasConfig := h.configs[sessionID]
	h.mu.Unlock()

	if !hasConfig {
		// Fallback: build minimal config from session data.
		config = domain.NewProjectConfig(
			filepath.Base(session.ProjectDir()),
			"", "", session.DetectedTools(),
		)
	}

	var writtenPaths []string
	preview := session.Preview()
	if preview != nil {
		for _, action := range preview.FileActions() {
			if action.ActionType() != vo.FileActionCreate {
				continue
			}
			content := h.contentProvider.ContentFor(action.Path(), config)
			target := filepath.Join(session.ProjectDir(), action.Path())
			if err := h.fileWriter.WriteFile(context.Background(), target, content); err != nil {
				return nil, fmt.Errorf("writing %s: %w", action.Path(), err)
			}
			writtenPaths = append(writtenPaths, action.Path())
		}
	}

	if err := h.commitScaffold(context.Background(), session.ProjectDir(), writtenPaths); err != nil {
		return nil, fmt.Errorf("committing scaffold: %w", err)
	}

	if err := session.Complete(); err != nil {
		return nil, fmt.Errorf("complete session: %w", err)
	}
	for _, event := range session.Events() {
		_ = h.publisher.Publish(context.Background(), event)
	}
	return session, nil
}

func (h *BootstrapHandler) getSession(sessionID string) (*domain.BootstrapSession, error) {
	h.mu.Lock()
	session, ok := h.sessions[sessionID]
	h.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("no active session with id '%s'", sessionID)
	}
	return session, nil
}

// commitScaffold stages and commits written files if a GitCommitter is configured.
// Skips gracefully when: no GitCommitter, not a git repo, or no files to commit.
func (h *BootstrapHandler) commitScaffold(ctx context.Context, projectDir string, paths []string) error {
	if h.gitCommitter == nil || len(paths) == 0 {
		return nil
	}

	hasGit, err := h.gitCommitter.HasGit(ctx, projectDir)
	if err != nil {
		return fmt.Errorf("checking git repo: %w", err)
	}
	if !hasGit {
		return nil
	}

	if err := h.gitCommitter.StageFiles(ctx, projectDir, paths); err != nil {
		return fmt.Errorf("staging files: %w", err)
	}

	if err := h.gitCommitter.Commit(ctx, projectDir, ScaffoldCommitMessage); err != nil {
		return fmt.Errorf("creating commit: %w", err)
	}

	return nil
}
