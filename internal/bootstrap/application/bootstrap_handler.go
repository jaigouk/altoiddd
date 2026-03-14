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
	ContentFor(path string, projectName string) string
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

// BootstrapHandler orchestrates the preview -> confirm -> execute bootstrap flow.
type BootstrapHandler struct {
	toolDetection   ToolDetector
	fileChecker     FileChecker
	publisher       sharedapp.EventPublisher
	fileWriter      sharedapp.FileWriter
	contentProvider ContentProvider
	mu              sync.Mutex
	sessions        map[string]*domain.BootstrapSession
}

// NewBootstrapHandler creates a new BootstrapHandler with injected dependencies.
func NewBootstrapHandler(toolDetection ToolDetector, fileChecker FileChecker, publisher sharedapp.EventPublisher, fileWriter sharedapp.FileWriter, contentProvider ContentProvider) *BootstrapHandler {
	return &BootstrapHandler{
		toolDetection:   toolDetection,
		fileChecker:     fileChecker,
		publisher:       publisher,
		fileWriter:      fileWriter,
		contentProvider: contentProvider,
		sessions:        make(map[string]*domain.BootstrapSession),
	}
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

	preview := session.Preview()
	if preview != nil {
		for _, action := range preview.FileActions() {
			if action.ActionType() != vo.FileActionCreate {
				continue
			}
			content := h.contentProvider.ContentFor(action.Path(), filepath.Base(session.ProjectDir()))
			target := filepath.Join(session.ProjectDir(), action.Path())
			if err := h.fileWriter.WriteFile(context.Background(), target, content); err != nil {
				return nil, fmt.Errorf("writing %s: %w", action.Path(), err)
			}
		}
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
