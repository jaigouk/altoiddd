package infrastructure_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/ticket/infrastructure"
)

var ctx = context.Background()

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func writeIssuesJSONL(t *testing.T, beadsDir string, issues []map[string]any) {
	t.Helper()
	require.NoError(t, os.MkdirAll(beadsDir, 0o755))
	f, err := os.Create(filepath.Join(beadsDir, "issues.jsonl"))
	require.NoError(t, err)
	defer f.Close()
	for _, issue := range issues {
		data, _ := json.Marshal(issue)
		_, err := f.Write(append(data, '\n'))
		require.NoError(t, err)
	}
}

func openIssue(id, title string) map[string]any {
	return map[string]any{
		"id":         id,
		"title":      title,
		"status":     "open",
		"priority":   "P1",
		"issue_type": "task",
	}
}

func closedIssue(id, title string) map[string]any {
	return map[string]any{
		"id":     id,
		"title":  title,
		"status": "closed",
	}
}

// ---------------------------------------------------------------------------
// ReadOpenTickets
// ---------------------------------------------------------------------------

func TestReaderReadsOpenTickets(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	writeIssuesJSONL(t, beadsDir, []map[string]any{
		openIssue("k7m.25", "Ticket Health"),
		openIssue("k7m.20", "Ticket Gen"),
	})
	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	tickets := reader.ReadOpenTickets(ctx)

	assert.Len(t, tickets, 2)
	ids := make(map[string]bool)
	for _, tk := range tickets {
		ids[tk.TicketID()] = true
	}
	assert.True(t, ids["k7m.25"])
	assert.True(t, ids["k7m.20"])
}

func TestReaderFiltersClosedTickets(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	writeIssuesJSONL(t, beadsDir, []map[string]any{
		openIssue("k7m.25", "Open ticket"),
		closedIssue("k7m.19", "Closed ticket"),
	})
	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	tickets := reader.ReadOpenTickets(ctx)

	assert.Len(t, tickets, 1)
	assert.Equal(t, "k7m.25", tickets[0].TicketID())
}

func TestReaderHandlesMissingDir(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads") // does not exist
	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	tickets := reader.ReadOpenTickets(ctx)

	assert.Empty(t, tickets)
}

func TestReaderHandlesCorruptedLines(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	require.NoError(t, os.MkdirAll(beadsDir, 0o755))
	f, err := os.Create(filepath.Join(beadsDir, "issues.jsonl"))
	require.NoError(t, err)
	data1, _ := json.Marshal(openIssue("k7m.25", "Good"))
	f.Write(append(data1, '\n'))
	f.WriteString("this is not valid json\n")
	data2, _ := json.Marshal(openIssue("k7m.20", "Also good"))
	f.Write(append(data2, '\n'))
	f.Close()

	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	tickets := reader.ReadOpenTickets(ctx)
	assert.Len(t, tickets, 2)
}

func TestReaderExtractsTitle(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	writeIssuesJSONL(t, beadsDir, []map[string]any{
		openIssue("k7m.25", "My Title"),
	})
	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	tickets := reader.ReadOpenTickets(ctx)
	assert.Equal(t, "My Title", tickets[0].Title())
}

func TestReaderSkipsBlankLines(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	require.NoError(t, os.MkdirAll(beadsDir, 0o755))
	f, err := os.Create(filepath.Join(beadsDir, "issues.jsonl"))
	require.NoError(t, err)
	data1, _ := json.Marshal(openIssue("k7m.25", "A"))
	f.Write(append(data1, '\n'))
	f.WriteString("\n")
	f.WriteString("   \n")
	data2, _ := json.Marshal(openIssue("k7m.26", "B"))
	f.Write(append(data2, '\n'))
	f.Close()

	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	tickets := reader.ReadOpenTickets(ctx)
	assert.Len(t, tickets, 2)
}

func TestReaderHandlesEmptyFile(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	require.NoError(t, os.MkdirAll(beadsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(beadsDir, "issues.jsonl"), []byte(""), 0o644))

	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	tickets := reader.ReadOpenTickets(ctx)
	assert.Empty(t, tickets)
}

// ---------------------------------------------------------------------------
// ReadFlags
// ---------------------------------------------------------------------------

func TestReaderReadsFlagsEmpty(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	require.NoError(t, os.MkdirAll(beadsDir, 0o755))

	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	flags := reader.ReadFlags(ctx, "k7m.25")
	assert.Empty(t, flags)
}

func TestReaderReadsFlagsFromInteractions(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	require.NoError(t, os.MkdirAll(beadsDir, 0o755))

	interaction := map[string]any{
		"issue_id":   "k7m.25",
		"type":       "comment",
		"body":       "**Ripple context diff from `k7m.19`:**\nImplemented fitness test generation aggregate",
		"created_at": "2026-03-01T10:00:00",
		"created_by": "agent",
	}
	data, _ := json.Marshal(interaction)
	require.NoError(t, os.WriteFile(filepath.Join(beadsDir, "interactions.jsonl"), append(data, '\n'), 0o644))

	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	flags := reader.ReadFlags(ctx, "k7m.25")

	assert.Len(t, flags, 1)
	assert.Contains(t, flags[0].ContextDiff().Summary(), "fitness")
}

func TestReaderReadsFlagsFiltersByTicket(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	require.NoError(t, os.MkdirAll(beadsDir, 0o755))

	f, err := os.Create(filepath.Join(beadsDir, "interactions.jsonl"))
	require.NoError(t, err)
	i1 := map[string]any{
		"issue_id":   "k7m.25",
		"body":       "**Ripple context diff from `k7m.19`:**\nChange for 25",
		"created_at": "2026-03-01T10:00:00",
	}
	i2 := map[string]any{
		"issue_id":   "k7m.20",
		"body":       "**Ripple context diff from `k7m.18`:**\nChange for 20",
		"created_at": "2026-03-01T11:00:00",
	}
	d1, _ := json.Marshal(i1)
	d2, _ := json.Marshal(i2)
	f.Write(append(d1, '\n'))
	f.Write(append(d2, '\n'))
	f.Close()

	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	flags := reader.ReadFlags(ctx, "k7m.25")
	assert.Len(t, flags, 1)
	assert.Contains(t, flags[0].ContextDiff().Summary(), "25")
}

func TestReaderHandlesMissingInteractionsFile(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	require.NoError(t, os.MkdirAll(beadsDir, 0o755))

	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	flags := reader.ReadFlags(ctx, "k7m.25")
	assert.Empty(t, flags)
}

func TestReaderHandlesEmptyInteractionsFile(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	require.NoError(t, os.MkdirAll(beadsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(beadsDir, "interactions.jsonl"), []byte(""), 0o644))

	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	flags := reader.ReadFlags(ctx, "k7m.25")
	assert.Empty(t, flags)
}

func TestReaderHandlesCorruptedInteractionLines(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	require.NoError(t, os.MkdirAll(beadsDir, 0o755))

	f, err := os.Create(filepath.Join(beadsDir, "interactions.jsonl"))
	require.NoError(t, err)
	f.WriteString("not valid json at all\n")
	good := map[string]any{
		"issue_id":   "k7m.25",
		"body":       "**Ripple context diff from `k7m.19`:**\nAdded fitness functions",
		"created_at": "2026-03-01T10:00:00",
	}
	data, _ := json.Marshal(good)
	f.Write(append(data, '\n'))
	f.Close()

	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	flags := reader.ReadFlags(ctx, "k7m.25")
	assert.Len(t, flags, 1)
	assert.Contains(t, flags[0].ContextDiff().Summary(), "fitness")
}

func TestReaderMultipleRippleComments(t *testing.T) {
	t.Parallel()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	require.NoError(t, os.MkdirAll(beadsDir, 0o755))

	f, err := os.Create(filepath.Join(beadsDir, "interactions.jsonl"))
	require.NoError(t, err)
	i1 := map[string]any{
		"issue_id":   "k7m.25",
		"body":       "**Ripple context diff from `k7m.19`:**\nFirst change",
		"created_at": "2026-03-01T10:00:00",
	}
	i2 := map[string]any{
		"issue_id":   "k7m.25",
		"body":       "**Ripple context diff from `k7m.20`:**\nSecond change",
		"created_at": "2026-03-02T10:00:00",
	}
	d1, _ := json.Marshal(i1)
	d2, _ := json.Marshal(i2)
	f.Write(append(d1, '\n'))
	f.Write(append(d2, '\n'))
	f.Close()

	reader := infrastructure.NewBeadsTicketReader(beadsDir)
	flags := reader.ReadFlags(ctx, "k7m.25")
	assert.Len(t, flags, 2)
	summaries := make(map[string]bool)
	for _, fl := range flags {
		summaries[fl.ContextDiff().Summary()] = true
	}
	assert.True(t, summaries["First change"])
	assert.True(t, summaries["Second change"])
}
