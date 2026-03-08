// Package composition contains adapter bridges that reconcile interface
// mismatches between handler-local interfaces and infrastructure adapters.
// These wrappers live here because the mismatch is a wiring concern.
package composition

import (
	"context"

	bootstrapapp "github.com/alty-cli/alty/internal/bootstrap/application"
	dochealthapp "github.com/alty-cli/alty/internal/dochealth/application"
	dochealthdomain "github.com/alty-cli/alty/internal/dochealth/domain"
	dochealthinfra "github.com/alty-cli/alty/internal/dochealth/infrastructure"
	discoveryinfra "github.com/alty-cli/alty/internal/discovery/infrastructure"
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

func (a *bootstrapToolDetectorAdapter) Detect(projectDir string) ([]string, error) {
	return a.scanner.Detect(context.Background(), projectDir)
}

func (a *bootstrapToolDetectorAdapter) ScanConflicts(projectDir string) ([]string, error) {
	return a.scanner.ScanConflicts(context.Background(), projectDir)
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

func (a *docScannerAdapter) LoadRegistry(
	_ context.Context,
	registryPath string,
) ([]dochealthdomain.DocRegistryEntry, error) {
	return a.scanner.LoadRegistry(registryPath)
}

func (a *docScannerAdapter) ScanRegistered(
	_ context.Context,
	entries []dochealthdomain.DocRegistryEntry,
	projectDir string,
) ([]dochealthdomain.DocStatus, error) {
	return a.scanner.ScanRegistered(entries, projectDir)
}

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
	return a.scanner.ScanUnregistered(docsDir, pathMap, excludeDirs)
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

func (a *ticketReaderAdapter) ReadOpenTickets(
	_ context.Context,
) ([]ticketdomain.OpenTicketData, error) {
	return a.reader.ReadOpenTickets(), nil
}

func (a *ticketReaderAdapter) ReadFlags(
	_ context.Context,
	ticketID string,
) ([]ticketdomain.FreshnessFlag, error) {
	return a.reader.ReadFlags(ticketID), nil
}
