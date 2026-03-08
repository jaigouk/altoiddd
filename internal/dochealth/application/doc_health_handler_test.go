package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/dochealth/application"
	"github.com/alty-cli/alty/internal/dochealth/domain"
)

// ---------------------------------------------------------------------------
// Mock scanner
// ---------------------------------------------------------------------------

type mockDocScanner struct {
	registryEntries      []domain.DocRegistryEntry
	registeredStatuses   []domain.DocStatus
	unregisteredStatuses []domain.DocStatus

	loadRegistryCalls      []string
	scanRegisteredCalls    []scanRegisteredCall
	scanUnregisteredCalls  []scanUnregisteredCall
}

type scanRegisteredCall struct {
	entries    []domain.DocRegistryEntry
	projectDir string
}

type scanUnregisteredCall struct {
	docsDir         string
	registeredPaths []string
	excludeDirs     []string
}

func (m *mockDocScanner) LoadRegistry(_ context.Context, registryPath string) ([]domain.DocRegistryEntry, error) {
	m.loadRegistryCalls = append(m.loadRegistryCalls, registryPath)
	return m.registryEntries, nil
}

func (m *mockDocScanner) ScanRegistered(_ context.Context, entries []domain.DocRegistryEntry, projectDir string) ([]domain.DocStatus, error) {
	m.scanRegisteredCalls = append(m.scanRegisteredCalls, scanRegisteredCall{entries: entries, projectDir: projectDir})
	return m.registeredStatuses, nil
}

func (m *mockDocScanner) ScanUnregistered(_ context.Context, docsDir string, registeredPaths []string, excludeDirs []string) ([]domain.DocStatus, error) {
	m.scanUnregisteredCalls = append(m.scanUnregisteredCalls, scanUnregisteredCall{
		docsDir:         docsDir,
		registeredPaths: registeredPaths,
		excludeDirs:     excludeDirs,
	})
	return m.unregisteredStatuses, nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestDocHealthHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("checks registered docs", func(t *testing.T) {
		t.Parallel()
		entry, err := domain.NewDocRegistryEntry("docs/PRD.md", "", 30)
		require.NoError(t, err)
		registeredStatuses := []domain.DocStatus{
			domain.NewDocStatus("docs/PRD.md", domain.DocHealthOK, nil, nil, 30, "", nil),
		}
		scanner := &mockDocScanner{
			registryEntries:    []domain.DocRegistryEntry{entry},
			registeredStatuses: registeredStatuses,
		}
		handler := application.NewDocHealthHandler(scanner)

		report, err := handler.Handle(context.Background(), "/project")

		require.NoError(t, err)
		assert.Equal(t, 1, len(scanner.scanRegisteredCalls))
		assert.NotNil(t, report)
	})

	t.Run("uses defaults when registry missing", func(t *testing.T) {
		t.Parallel()
		defaultStatuses := []domain.DocStatus{
			domain.NewDocStatus("docs/PRD.md", domain.DocHealthOK, nil, nil, 30, "", nil),
			domain.NewDocStatus("docs/DDD.md", domain.DocHealthMissing, nil, nil, 30, "", nil),
			domain.NewDocStatus("docs/ARCHITECTURE.md", domain.DocHealthNoFrontmatter, nil, nil, 30, "", nil),
		}
		scanner := &mockDocScanner{
			registryEntries:    nil, // empty -> triggers defaults
			registeredStatuses: defaultStatuses,
		}
		handler := application.NewDocHealthHandler(scanner)

		_, err := handler.Handle(context.Background(), "/project")

		require.NoError(t, err)
		require.Equal(t, 1, len(scanner.scanRegisteredCalls))
		entriesUsed := scanner.scanRegisteredCalls[0].entries
		assert.Equal(t, 3, len(entriesUsed))
		paths := make(map[string]bool)
		for _, e := range entriesUsed {
			paths[e.Path()] = true
		}
		assert.True(t, paths["docs/PRD.md"])
		assert.True(t, paths["docs/DDD.md"])
		assert.True(t, paths["docs/ARCHITECTURE.md"])
	})

	t.Run("skips template dirs", func(t *testing.T) {
		t.Parallel()
		entry, err := domain.NewDocRegistryEntry("docs/PRD.md", "", 30)
		require.NoError(t, err)
		scanner := &mockDocScanner{
			registryEntries:    []domain.DocRegistryEntry{entry},
			registeredStatuses: []domain.DocStatus{domain.NewDocStatus("docs/PRD.md", domain.DocHealthOK, nil, nil, 30, "", nil)},
		}
		handler := application.NewDocHealthHandler(scanner)

		_, err = handler.Handle(context.Background(), "/project")

		require.NoError(t, err)
		require.Equal(t, 1, len(scanner.scanUnregisteredCalls))
		excludeDirs := scanner.scanUnregisteredCalls[0].excludeDirs
		assert.Contains(t, excludeDirs, "templates")
		assert.Contains(t, excludeDirs, "beads_templates")
	})

	t.Run("combines registered and unregistered", func(t *testing.T) {
		t.Parallel()
		entry, err := domain.NewDocRegistryEntry("docs/PRD.md", "", 30)
		require.NoError(t, err)
		registered := []domain.DocStatus{
			domain.NewDocStatus("docs/PRD.md", domain.DocHealthOK, nil, nil, 30, "", nil),
		}
		unregistered := []domain.DocStatus{
			domain.NewDocStatus("docs/notes.md", domain.DocHealthNoFrontmatter, nil, nil, 0, "", nil),
		}
		scanner := &mockDocScanner{
			registryEntries:      []domain.DocRegistryEntry{entry},
			registeredStatuses:   registered,
			unregisteredStatuses: unregistered,
		}
		handler := application.NewDocHealthHandler(scanner)

		report, err := handler.Handle(context.Background(), "/project")

		require.NoError(t, err)
		assert.Equal(t, 2, report.TotalChecked())
		paths := make(map[string]bool)
		for _, s := range report.Statuses() {
			paths[s.Path()] = true
		}
		assert.True(t, paths["docs/PRD.md"])
		assert.True(t, paths["docs/notes.md"])
	})

	t.Run("returns complete report with issues", func(t *testing.T) {
		t.Parallel()
		entry1, err := domain.NewDocRegistryEntry("docs/PRD.md", "", 30)
		require.NoError(t, err)
		entry2, err := domain.NewDocRegistryEntry("docs/DDD.md", "", 30)
		require.NoError(t, err)

		days := 45
		registered := []domain.DocStatus{
			domain.NewDocStatus("docs/PRD.md", domain.DocHealthOK, nil, nil, 30, "", nil),
			domain.NewDocStatus("docs/DDD.md", domain.DocHealthStale, nil, &days, 30, "", nil),
		}
		unregistered := []domain.DocStatus{
			domain.NewDocStatus("docs/extra.md", domain.DocHealthNoFrontmatter, nil, nil, 0, "", nil),
		}
		scanner := &mockDocScanner{
			registryEntries:      []domain.DocRegistryEntry{entry1, entry2},
			registeredStatuses:   registered,
			unregisteredStatuses: unregistered,
		}
		handler := application.NewDocHealthHandler(scanner)

		report, err := handler.Handle(context.Background(), "/project")

		require.NoError(t, err)
		assert.Equal(t, 3, report.TotalChecked())
		assert.Equal(t, 2, report.IssueCount())
		assert.True(t, report.HasIssues())
	})

	t.Run("passes registered paths to unregistered scan", func(t *testing.T) {
		t.Parallel()
		entry1, err := domain.NewDocRegistryEntry("docs/PRD.md", "", 30)
		require.NoError(t, err)
		entry2, err := domain.NewDocRegistryEntry("docs/DDD.md", "", 30)
		require.NoError(t, err)

		scanner := &mockDocScanner{
			registryEntries: []domain.DocRegistryEntry{entry1, entry2},
			registeredStatuses: []domain.DocStatus{
				domain.NewDocStatus("docs/PRD.md", domain.DocHealthOK, nil, nil, 30, "", nil),
				domain.NewDocStatus("docs/DDD.md", domain.DocHealthOK, nil, nil, 30, "", nil),
			},
		}
		handler := application.NewDocHealthHandler(scanner)

		_, err = handler.Handle(context.Background(), "/project")

		require.NoError(t, err)
		require.Equal(t, 1, len(scanner.scanUnregisteredCalls))
		registeredPaths := scanner.scanUnregisteredCalls[0].registeredPaths
		assert.Contains(t, registeredPaths, "docs/PRD.md")
		assert.Contains(t, registeredPaths, "docs/DDD.md")
		assert.Equal(t, 2, len(registeredPaths))
	})
}
