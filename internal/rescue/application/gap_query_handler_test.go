package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/rescue/application"
	rescuedomain "github.com/alty-cli/alty/internal/rescue/domain"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Mock StackProfileDetector
// ---------------------------------------------------------------------------

type fakeProfileDetector struct {
	profile vo.StackProfile
}

func (f *fakeProfileDetector) DetectProfile(_ string) vo.StackProfile {
	return f.profile
}

// ---------------------------------------------------------------------------
// GapQueryHandler
// ---------------------------------------------------------------------------

func TestGapQueryHandler_AnalyzeGaps_WhenMissingDocs_ExpectGapsReturned(t *testing.T) {
	t.Parallel()

	// Given: a project with no docs, no configs, no structure
	scanner := newFakeScanner(nil) // defaults: no docs, no configs, no knowledge
	handler := application.NewGapQueryHandler(scanner, &fakeProfileDetector{})

	// When
	report, err := handler.AnalyzeGaps(context.Background(), "/tmp/proj")

	// Then
	require.NoError(t, err)
	assert.NotEmpty(t, report.Entries)

	// Should have missing docs (PRD.md, DDD.md, ARCHITECTURE.md)
	var docEntries []application.GapReportEntry
	for _, e := range report.Entries {
		if e.GapType == string(rescuedomain.GapTypeMissingDoc) {
			docEntries = append(docEntries, e)
		}
	}
	assert.NotEmpty(t, docEntries, "should detect missing documentation gaps")
}

func TestGapQueryHandler_AnalyzeGaps_WhenMissingAltyConfig_ExpectRecommendedGap(t *testing.T) {
	t.Parallel()

	// Given: a project with all docs but no .alty/config.toml
	scan := rescuedomain.NewProjectScan(
		"/tmp/proj",
		[]string{"docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md"},
		[]string{".claude/CLAUDE.md"},
		nil,
		true, true, true, false, true, // hasAltyConfig = false
	)
	scanner := newFakeScanner(&scan)
	handler := application.NewGapQueryHandler(scanner, &fakeProfileDetector{})

	// When
	report, err := handler.AnalyzeGaps(context.Background(), "/tmp/proj")

	// Then
	require.NoError(t, err)

	var configEntry *application.GapReportEntry
	for _, e := range report.Entries {
		if e.Path == ".alty/config.toml" {
			e := e
			configEntry = &e
			break
		}
	}
	require.NotNil(t, configEntry, "should detect missing .alty/config.toml")
	assert.Equal(t, string(rescuedomain.GapSeverityRecommended), configEntry.Severity)
}

func TestGapQueryHandler_AnalyzeGaps_WhenCleanProject_ExpectEmptyGaps(t *testing.T) {
	t.Parallel()

	// Given: a project with everything in place
	scan := rescuedomain.NewProjectScan(
		"/tmp/proj",
		[]string{"docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md", "AGENTS.md"},
		[]string{".claude/CLAUDE.md", ".alty/config.toml"},
		nil,
		true, true, true, true, true,
	)
	scanner := newFakeScanner(&scan)
	handler := application.NewGapQueryHandler(scanner, &fakeProfileDetector{})

	// When
	report, err := handler.AnalyzeGaps(context.Background(), "/tmp/proj")

	// Then
	require.NoError(t, err)
	assert.Empty(t, report.Entries)
}

func TestGapQueryHandler_AnalyzeGaps_WhenStackProfile_ExpectProfileGapsIncluded(t *testing.T) {
	t.Parallel()

	// Given: a Go project missing go.mod
	scan := rescuedomain.NewProjectScan(
		"/tmp/proj",
		[]string{"docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md", "AGENTS.md"},
		[]string{".claude/CLAUDE.md", ".alty/config.toml"},
		nil,
		true, true, true, true, true,
	)
	scanner := newFakeScanner(&scan)
	handler := application.NewGapQueryHandler(scanner, &fakeProfileDetector{profile: vo.GoModProfile{}})

	// When
	report, err := handler.AnalyzeGaps(context.Background(), "/tmp/proj")

	// Then
	require.NoError(t, err)

	// GoModProfile has ProjectManifest() = "go.mod" and SourceLayout() entries
	// Since go.mod is not in existingConfigs, it should show as a gap
	var manifestEntry *application.GapReportEntry
	for _, e := range report.Entries {
		if e.Path == "go.mod" {
			e := e
			manifestEntry = &e
			break
		}
	}
	assert.NotNil(t, manifestEntry, "should detect missing go.mod from profile")
}

func TestGapReport_FormatReport_WhenNoGaps_ExpectCompliantMessage(t *testing.T) {
	t.Parallel()

	report := &application.GapReport{}
	output := report.FormatReport()
	assert.Contains(t, output, "No gaps found")
}

func TestGapReport_FormatReport_WhenRequired_ExpectRequiredMessage(t *testing.T) {
	t.Parallel()

	report := &application.GapReport{
		Entries: []application.GapReportEntry{
			{Path: "docs/PRD.md", GapType: "missing_doc", Severity: "required"},
		},
		HasRequired: true,
	}
	output := report.FormatReport()
	assert.Contains(t, output, "Gap Analysis Report")
	assert.Contains(t, output, "docs/PRD.md")
	assert.Contains(t, output, "Required gaps found")
}
