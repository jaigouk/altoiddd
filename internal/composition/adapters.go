// Package composition contains adapter bridges that reconcile interface
// mismatches between handler-local interfaces and infrastructure adapters.
// These wrappers live here because the mismatch is a wiring concern.
package composition

import (
	"context"
	"fmt"

	bootstrapapp "github.com/alty-cli/alty/internal/bootstrap/application"
	discoveryinfra "github.com/alty-cli/alty/internal/discovery/infrastructure"
	dochealthapp "github.com/alty-cli/alty/internal/dochealth/application"
	dochealthdomain "github.com/alty-cli/alty/internal/dochealth/domain"
	dochealthinfra "github.com/alty-cli/alty/internal/dochealth/infrastructure"
	ticketapp "github.com/alty-cli/alty/internal/ticket/application"
	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
	ticketinfra "github.com/alty-cli/alty/internal/ticket/infrastructure"
)

// ---------------------------------------------------------------------------
// Bootstrap ToolDetector adapter
// ---------------------------------------------------------------------------

// Compile-time interface check.
var _ bootstrapapp.ToolDetector = (*bootstrapToolDetectorAdapter)(nil)

// bootstrapToolDetectorAdapter bridges FilesystemToolScanner (ctx params) to
// the bootstrap ToolDetector interface (no ctx params).
type bootstrapToolDetectorAdapter struct {
	scanner *discoveryinfra.FilesystemToolScanner
}

// Detect implements ToolDetector.
func (a *bootstrapToolDetectorAdapter) Detect(projectDir string) ([]string, error) {
	tools, err := a.scanner.Detect(context.Background(), projectDir)
	if err != nil {
		return nil, fmt.Errorf("detecting tools: %w", err)
	}
	return tools, nil
}

// ScanConflicts implements ToolDetector.
func (a *bootstrapToolDetectorAdapter) ScanConflicts(projectDir string) ([]string, error) {
	conflicts, err := a.scanner.ScanConflicts(context.Background(), projectDir)
	if err != nil {
		return nil, fmt.Errorf("scanning conflicts: %w", err)
	}
	return conflicts, nil
}

// Note: bootstrapToolDetectorAdapter also satisfies discoveryapp.ToolDetector
// via Go structural typing (same method set). Used for both bootstrap and
// discovery DetectionHandler wiring.

// ---------------------------------------------------------------------------
// DocHealth DocScanner adapter
// ---------------------------------------------------------------------------

// Compile-time interface check.
var _ dochealthapp.DocScanner = (*docScannerAdapter)(nil)

// docScannerAdapter bridges FilesystemDocScanner (no ctx, map params) to
// the DocScanner interface (with ctx, slice params).
type docScannerAdapter struct {
	scanner *dochealthinfra.FilesystemDocScanner
}

// LoadRegistry implements DocScanner.
func (a *docScannerAdapter) LoadRegistry(
	_ context.Context,
	registryPath string,
) ([]dochealthdomain.DocRegistryEntry, error) {
	entries, err := a.scanner.LoadRegistry(registryPath)
	if err != nil {
		return nil, fmt.Errorf("loading registry: %w", err)
	}
	return entries, nil
}

// ScanRegistered implements DocScanner.
func (a *docScannerAdapter) ScanRegistered(
	_ context.Context,
	entries []dochealthdomain.DocRegistryEntry,
	projectDir string,
) ([]dochealthdomain.DocStatus, error) {
	statuses, err := a.scanner.ScanRegistered(entries, projectDir)
	if err != nil {
		return nil, fmt.Errorf("scanning registered docs: %w", err)
	}
	return statuses, nil
}

// ScanUnregistered implements DocScanner.
func (a *docScannerAdapter) ScanUnregistered(
	_ context.Context,
	docsDir string,
	registeredPaths []string,
	excludeDirs []string,
) ([]dochealthdomain.DocStatus, error) {
	// Convert []string to map[string]bool for infrastructure adapter.
	pathMap := make(map[string]bool, len(registeredPaths))
	for _, p := range registeredPaths {
		pathMap[p] = true
	}
	statuses, err := a.scanner.ScanUnregistered(docsDir, pathMap, excludeDirs)
	if err != nil {
		return nil, fmt.Errorf("scanning unregistered docs: %w", err)
	}
	return statuses, nil
}

// ---------------------------------------------------------------------------
// Ticket TicketReader adapter
// ---------------------------------------------------------------------------

// Compile-time interface check.
var _ ticketapp.TicketReader = (*ticketReaderAdapter)(nil)

// ticketReaderAdapter bridges BeadsTicketReader (no ctx, no error) to the
// TicketReader interface (with ctx, with error).
type ticketReaderAdapter struct {
	reader *ticketinfra.BeadsTicketReader
}

// ReadOpenTickets implements TicketReader.
func (a *ticketReaderAdapter) ReadOpenTickets(
	ctx context.Context,
) ([]ticketdomain.OpenTicketData, error) {
	return a.reader.ReadOpenTickets(ctx), nil
}

// ReadFlags implements TicketReader.
func (a *ticketReaderAdapter) ReadFlags(
	ctx context.Context,
	ticketID string,
) ([]ticketdomain.FreshnessFlag, error) {
	return a.reader.ReadFlags(ctx, ticketID), nil
}
