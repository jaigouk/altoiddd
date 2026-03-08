// Package application provides command handlers for the Bootstrap bounded context.
package application

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/alty-cli/alty/internal/bootstrap/domain"
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
	toolDetection ToolDetector
	fileChecker   FileChecker
	mu            sync.Mutex
	sessions      map[string]*domain.BootstrapSession
}

// NewBootstrapHandler creates a new BootstrapHandler with injected dependencies.
func NewBootstrapHandler(toolDetection ToolDetector, fileChecker FileChecker) *BootstrapHandler {
	return &BootstrapHandler{
		toolDetection: toolDetection,
		fileChecker:   fileChecker,
		sessions:      make(map[string]*domain.BootstrapSession),
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
	if err := session.Complete(); err != nil {
		return nil, fmt.Errorf("complete session: %w", err)
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
