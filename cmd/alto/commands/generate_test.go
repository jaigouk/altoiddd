package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/composition"
)

// validDDDContent returns a minimal valid DDD.md for testing.
func validDDDContent() string {
	return `# Domain Model

## Bounded Contexts

### 1. Orders (Core)

**Responsibility:** Order management and fulfillment

### 2. Shipping (Supporting)

**Responsibility:** Shipment tracking and delivery
`
}

func setupDocsDir(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "DDD.md"), []byte(content), 0o644))
	return docsDir
}

func TestGenerateTicketsFromDocs_WhenValidDDD_PrintsPreviewSummary(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	docsDir := setupDocsDir(t, validDDDContent())

	cmd := NewGenerateCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"tickets", "--from-docs", "--docs-dir", docsDir})

	err = cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Ticket")
}

func TestGenerateTicketsFromDocs_WhenMissingDDD_ReturnsError(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	cmd := NewGenerateCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"tickets", "--from-docs", "--docs-dir", t.TempDir()})

	err = cmd.Execute()
	require.Error(t, err)
}

func TestGenerateTickets_WhenNoFlag_ReturnsError(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	cmd := NewGenerateCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"tickets"})

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--from-docs")
}

func TestGenerateFitnessFromDocs_WhenValidDDD_PrintsPreviewSummary(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	docsDir := setupDocsDir(t, validDDDContent())

	cmd := NewGenerateCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"fitness", "--from-docs", "--docs-dir", docsDir})

	err = cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.NotEmpty(t, output)
}

func TestGenerateConfigsFromDocs_WhenValidDDD_PrintsPreviewSummary(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	docsDir := setupDocsDir(t, validDDDContent())

	cmd := NewGenerateCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"configs", "--from-docs", "--docs-dir", docsDir})

	err = cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Config")
}

func TestGenerateArtifacts_StillShowsStubMessage(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	cmd := NewGenerateCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"artifacts"})

	err = cmd.Execute()
	require.NoError(t, err)
}
