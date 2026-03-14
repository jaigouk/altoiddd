package application

import (
	"context"
	"fmt"
	"strings"

	rescuedomain "github.com/alty-cli/alty/internal/rescue/domain"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// StackProfileDetector detects the stack profile for a project directory.
type StackProfileDetector interface {
	DetectProfile(projectDir string) vo.StackProfile
}

// GapQueryHandler is a CQRS query handler that analyzes a project for
// structural gaps without performing any mutations (no git ops, no file writes).
type GapQueryHandler struct {
	projectScan     ProjectScan
	profileDetector StackProfileDetector
}

// NewGapQueryHandler creates a GapQueryHandler with injected dependencies.
func NewGapQueryHandler(projectScan ProjectScan, profileDetector StackProfileDetector) *GapQueryHandler {
	return &GapQueryHandler{
		projectScan:     projectScan,
		profileDetector: profileDetector,
	}
}

// GapReportEntry is a single gap finding for presentation.
type GapReportEntry struct {
	Path     string
	GapType  string
	Severity string
}

// GapReport contains the result of a gap analysis, ready for presentation.
type GapReport struct {
	Entries     []GapReportEntry
	HasRequired bool
}

// FormatReport returns a formatted string representation of the gap report.
func (r GapReport) FormatReport() string {
	if len(r.Entries) == 0 {
		return "No gaps found, project is compliant."
	}

	var sb strings.Builder
	sb.WriteString("Gap Analysis Report\n")
	sb.WriteString("----------------------------------------\n")

	// Group by severity in order: required → recommended → optional
	for _, sev := range rescuedomain.AllGapSeverities() {
		var sevEntries []GapReportEntry
		for _, e := range r.Entries {
			if e.Severity == string(sev) {
				sevEntries = append(sevEntries, e)
			}
		}
		if len(sevEntries) == 0 {
			continue
		}
		fmt.Fprintf(&sb, "\n[%s]\n", sev)
		for _, e := range sevEntries {
			fmt.Fprintf(&sb, "  %-40s %s\n", e.Path, e.GapType)
		}
	}

	if r.HasRequired {
		sb.WriteString("\nRequired gaps found. Run 'alty init --existing' to fix.\n")
	}

	return sb.String()
}

// AnalyzeGaps scans a project directory and returns a presentation-ready gap report.
func (h *GapQueryHandler) AnalyzeGaps(
	ctx context.Context,
	projectDir string,
) (*GapReport, error) {
	profile := h.profileDetector.DetectProfile(projectDir)

	scan, err := h.projectScan.Scan(ctx, projectDir, profile)
	if err != nil {
		return nil, fmt.Errorf("scanning project: %w", err)
	}

	gaps := IdentifyGaps(scan, profile)

	report := &GapReport{}
	for _, g := range gaps {
		report.Entries = append(report.Entries, GapReportEntry{
			Path:     g.Path(),
			GapType:  string(g.GapType()),
			Severity: string(g.Severity()),
		})
		if g.Severity() == rescuedomain.GapSeverityRequired {
			report.HasRequired = true
		}
	}

	return report, nil
}
