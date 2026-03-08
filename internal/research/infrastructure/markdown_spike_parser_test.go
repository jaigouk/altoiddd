package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/research/application"
	"github.com/alty-cli/alty/internal/research/infrastructure"
)

// Compile-time interface check.
var _ application.SpikeReportParser = (*infrastructure.MarkdownSpikeParser)(nil)

func writeReport(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "report.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

// -- Heading detection --

func TestSpikeParser_DetectsH2FollowUpTickets(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeReport(t, dir,
		"# Spike Report\n\n## Follow-Up Tickets\n\n### Ticket 1: Implement feature A\n\nDescription of A.\n")

	parser := infrastructure.NewMarkdownSpikeParser()
	intents, err := parser.Parse(context.Background(), path)
	require.NoError(t, err)
	require.Len(t, intents, 1)
	assert.Equal(t, "Implement feature A", intents[0].Title())
}

func TestSpikeParser_DetectsH2FollowUpImplementationTickets(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeReport(t, dir,
		"# Report\n\n## 8. Follow-Up Implementation Tickets\n\n### Ticket 1: Build the parser\n\nDetails.\n")

	parser := infrastructure.NewMarkdownSpikeParser()
	intents, err := parser.Parse(context.Background(), path)
	require.NoError(t, err)
	require.Len(t, intents, 1)
	assert.Equal(t, "Build the parser", intents[0].Title())
}

func TestSpikeParser_DetectsH2FollowUpTicketsNeeded(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeReport(t, dir,
		"# Report\n\n## Follow-up Tickets Needed\n\n### 1. Create domain model\n\nSteps here.\n")

	parser := infrastructure.NewMarkdownSpikeParser()
	intents, err := parser.Parse(context.Background(), path)
	require.NoError(t, err)
	require.Len(t, intents, 1)
	assert.Equal(t, "Create domain model", intents[0].Title())
}

func TestSpikeParser_DetectsH3FollowUpHeading(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeReport(t, dir,
		"# Report\n\n### Follow-up Tickets Needed\n\n- **Task A**: Build stuff\n- **Task B**: Fix stuff\n")

	parser := infrastructure.NewMarkdownSpikeParser()
	intents, err := parser.Parse(context.Background(), path)
	require.NoError(t, err)
	assert.Len(t, intents, 2)
}

func TestSpikeParser_DetectsNumberedHeadingPrefix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeReport(t, dir,
		"## 13. Follow-Up Implementation Tickets\n\n### Ticket 1: Gap analysis tool\n\nBuild it.\n")

	parser := infrastructure.NewMarkdownSpikeParser()
	intents, err := parser.Parse(context.Background(), path)
	require.NoError(t, err)
	require.Len(t, intents, 1)
	assert.Equal(t, "Gap analysis tool", intents[0].Title())
}

// -- Ticket extraction formats --

func TestSpikeParser_ExtractsBoldListFormat(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeReport(t, dir,
		"## Follow-Up Tickets\n\n- **Build CLI command**: Wire up the subcommand tree\n- **Add test coverage**: Cover edge cases\n")

	parser := infrastructure.NewMarkdownSpikeParser()
	intents, err := parser.Parse(context.Background(), path)
	require.NoError(t, err)
	require.Len(t, intents, 2)
	assert.Equal(t, "Build CLI command", intents[0].Title())
	assert.Equal(t, "Wire up the subcommand tree", intents[0].Description())
}

func TestSpikeParser_ExtractsPlainListFormat(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeReport(t, dir,
		"## Follow-Up Tickets\n\n- Create fitness functions\n- Add import linter\n")

	parser := infrastructure.NewMarkdownSpikeParser()
	intents, err := parser.Parse(context.Background(), path)
	require.NoError(t, err)
	require.Len(t, intents, 2)
	assert.Equal(t, "Create fitness functions", intents[0].Title())
}

func TestSpikeParser_ExtractsDescriptionFromBody(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeReport(t, dir,
		"## Follow-Up Tickets\n\n### Ticket 1: Build parser\n\n"+
			"Create a Markdown parser that extracts follow-up intents.\n\n**Type:** Task\n")

	parser := infrastructure.NewMarkdownSpikeParser()
	intents, err := parser.Parse(context.Background(), path)
	require.NoError(t, err)
	require.Len(t, intents, 1)
	assert.Contains(t, intents[0].Description(), "Markdown parser")
}

// -- Edge cases --

func TestSpikeParser_NoFollowUpSectionReturnsEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeReport(t, dir,
		"# Spike Report\n\n## Findings\n\nSome research results.\n\n## References\n\n- Link 1\n")

	parser := infrastructure.NewMarkdownSpikeParser()
	intents, err := parser.Parse(context.Background(), path)
	require.NoError(t, err)
	assert.Empty(t, intents)
}

func TestSpikeParser_EmptyFollowUpSectionReturnsEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeReport(t, dir,
		"# Report\n\n## Follow-Up Tickets\n\n## References\n\n- Link\n")

	parser := infrastructure.NewMarkdownSpikeParser()
	intents, err := parser.Parse(context.Background(), path)
	require.NoError(t, err)
	assert.Empty(t, intents)
}

func TestSpikeParser_NonexistentFileReturnsEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	parser := infrastructure.NewMarkdownSpikeParser()
	intents, err := parser.Parse(context.Background(), filepath.Join(dir, "does_not_exist.md"))
	require.NoError(t, err)
	assert.Empty(t, intents)
}

func TestSpikeParser_StopsAtNextSameLevelHeading(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeReport(t, dir,
		"## Follow-Up Tickets\n\n### Ticket 1: Real ticket\n\nDetails.\n\n"+
			"## References\n\n### Not a ticket: This is a reference\n\nLink.\n")

	parser := infrastructure.NewMarkdownSpikeParser()
	intents, err := parser.Parse(context.Background(), path)
	require.NoError(t, err)
	require.Len(t, intents, 1)
	assert.Equal(t, "Real ticket", intents[0].Title())
}
