package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BeadsTicketContentReader reads ticket content from the beads issue tracker.
type BeadsTicketContentReader struct {
	beadsDir string
}

// NewBeadsTicketContentReader creates a reader for the given beads directory.
func NewBeadsTicketContentReader(beadsDir string) *BeadsTicketContentReader {
	return &BeadsTicketContentReader{beadsDir: beadsDir}
}

// ReadTicketContent reads the full description/content of a ticket.
func (r *BeadsTicketContentReader) ReadTicketContent(_ context.Context, ticketID string) (string, error) {
	if ticketID == "" {
		return "", fmt.Errorf("ticket ID cannot be empty")
	}

	issuesPath := filepath.Join(r.beadsDir, "issues.jsonl")
	data, err := os.ReadFile(issuesPath)
	if err != nil {
		return "", fmt.Errorf("reading issues file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var issue map[string]any
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			continue // Skip malformed lines
		}

		id, ok := issue["id"].(string)
		if !ok || id != ticketID {
			continue
		}

		// Extract description (the main content)
		desc, _ := issue["description"].(string)
		notes, _ := issue["notes"].(string)

		// Combine description and notes for full content
		content := desc
		if notes != "" {
			content += "\n\n## Notes\n\n" + notes
		}

		return content, nil
	}

	return "", fmt.Errorf("ticket not found: %s", ticketID)
}
