// Package application provides query handlers for the DocHealth bounded context.
package application

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/alto-cli/alto/internal/dochealth/domain"
)

// DocScanner is a handler-local interface for scanning project documentation.
// It is consumed by DocHealthHandler and defined here per Go convention
// (define interfaces where consumed, not where implemented).
type DocScanner interface {
	// LoadRegistry loads document registry entries from a TOML file.
	LoadRegistry(ctx context.Context, registryPath string) ([]domain.DocRegistryEntry, error)

	// ScanRegistered scans registered documents for health status.
	ScanRegistered(ctx context.Context, entries []domain.DocRegistryEntry, projectDir string) ([]domain.DocStatus, error)

	// ScanUnregistered scans for markdown files not in the registry.
	ScanUnregistered(ctx context.Context, docsDir string, registeredPaths []string, excludeDirs []string) ([]domain.DocStatus, error)
}

// defaultEntries are used when no registry file is found.
func defaultEntries() []domain.DocRegistryEntry {
	// These are well-known defaults; review_interval_days=30 is a sensible default.
	prd, _ := domain.NewDocRegistryEntry("docs/PRD.md", "", 30)
	ddd, _ := domain.NewDocRegistryEntry("docs/DDD.md", "", 30)
	arch, _ := domain.NewDocRegistryEntry("docs/ARCHITECTURE.md", "", 30)
	return []domain.DocRegistryEntry{prd, ddd, arch}
}

// DocHealthHandler orchestrates document health checking.
// It loads a doc registry (or uses defaults), scans registered docs,
// discovers unregistered docs, and combines into a DocHealthReport.
type DocHealthHandler struct {
	scanner DocScanner
}

// NewDocHealthHandler creates a new DocHealthHandler.
func NewDocHealthHandler(scanner DocScanner) *DocHealthHandler {
	return &DocHealthHandler{scanner: scanner}
}

// Handle executes the doc health query.
func (h *DocHealthHandler) Handle(ctx context.Context, projectDir string) (domain.DocHealthReport, error) {
	registryPath := filepath.Join(projectDir, ".alto", "maintenance", "doc-registry.toml")
	entries, err := h.scanner.LoadRegistry(ctx, registryPath)
	if err != nil {
		return domain.DocHealthReport{}, fmt.Errorf("load registry: %w", err)
	}

	if len(entries) == 0 {
		entries = defaultEntries()
	}

	registeredStatuses, err := h.scanner.ScanRegistered(ctx, entries, projectDir)
	if err != nil {
		return domain.DocHealthReport{}, fmt.Errorf("scan registered docs: %w", err)
	}

	registeredPaths := make([]string, len(entries))
	for i, e := range entries {
		registeredPaths[i] = e.Path()
	}

	docsDir := filepath.Join(projectDir, "docs")
	excludeDirs := []string{"templates", "beads_templates"}
	unregisteredStatuses, err := h.scanner.ScanUnregistered(ctx, docsDir, registeredPaths, excludeDirs)
	if err != nil {
		return domain.DocHealthReport{}, fmt.Errorf("scan unregistered docs: %w", err)
	}

	allStatuses := make([]domain.DocStatus, 0, len(registeredStatuses)+len(unregisteredStatuses))
	allStatuses = append(allStatuses, registeredStatuses...)
	allStatuses = append(allStatuses, unregisteredStatuses...)

	return domain.NewDocHealthReport(allStatuses), nil
}
