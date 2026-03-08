// Package infrastructure provides adapters for the Ticket bounded context.
package infrastructure

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
)

const bdTimeoutSeconds = 10

var (
	// Current format: **Ripple review needed** -- `<id>`
	rippleTriggerRE = regexp.MustCompile(`\*\*Ripple review needed\*\*\s*--\s*` + "`" + `([^` + "`" + `]+)` + "`")

	// Extracts "What changed:" summary text
	whatChangedRE = regexp.MustCompile(`\*\*What changed:\*\*\s*(.*?)(?:\n\n|\z)`)

	// Old format: **Ripple context diff from `<id>`:**
	ripplePatternOld = regexp.MustCompile(`\*\*Ripple context diff from ` + "`" + `([^` + "`" + `]+)` + "`" + `:\*\*\s*(.*)`)

	// Parses bd query output lines
	bdQueryLineRE = regexp.MustCompile(`^[○◐●✓❄]\s+(\S+)`)

	// Parses comment blocks
	commentBlockRE = regexp.MustCompile(`\n\[.+?\] at (\d{4}-\d{2}-\d{2}[\sT]?\d{0,8})`)
)

// BeadsTicketReader reads beads data and translates it into domain value objects.
// This is an Anti-Corruption Layer (ACL) that shields the domain from beads data format details.
type BeadsTicketReader struct {
	beadsDir string
}

// NewBeadsTicketReader creates a BeadsTicketReader.
func NewBeadsTicketReader(beadsDir string) *BeadsTicketReader {
	return &BeadsTicketReader{beadsDir: beadsDir}
}

// ReadOpenTickets reads all open tickets from issues.jsonl, enriched with labels.
func (r *BeadsTicketReader) ReadOpenTickets(ctx context.Context) []ticketdomain.OpenTicketData {
	tickets := r.readTicketsFromJSONL()
	flaggedIDs := r.getFlaggedIDs(ctx)
	return r.enrichLabels(tickets, flaggedIDs)
}

// ReadFlags reads freshness flags from ripple review comments.
func (r *BeadsTicketReader) ReadFlags(ctx context.Context, ticketID string) []ticketdomain.FreshnessFlag {
	flags := r.readFlagsFromBDComments(ctx, ticketID)
	if len(flags) > 0 {
		return flags
	}
	return r.readFlagsFromJSONL(ticketID)
}

// ------------------------------------------------------------------
// JSONL reading
// ------------------------------------------------------------------

type issueJSON struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

func (r *BeadsTicketReader) readTicketsFromJSONL() []ticketdomain.OpenTicketData {
	jsonlPath := filepath.Join(r.beadsDir, "issues.jsonl")
	f, err := os.Open(jsonlPath)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	var tickets []ticketdomain.OpenTicketData
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var issue issueJSON
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			continue
		}
		if issue.Status != "open" {
			continue
		}
		tickets = append(tickets, ticketdomain.NewOpenTicketData(issue.ID, issue.Title, nil, nil))
	}
	return tickets
}

type interactionJSON struct {
	IssueID   string `json:"issue_id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

func (r *BeadsTicketReader) readFlagsFromJSONL(ticketID string) []ticketdomain.FreshnessFlag {
	interactionsPath := filepath.Join(r.beadsDir, "interactions.jsonl")
	f, err := os.Open(interactionsPath)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	var flags []ticketdomain.FreshnessFlag
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var interaction interactionJSON
		if err := json.Unmarshal([]byte(line), &interaction); err != nil {
			continue
		}
		if interaction.IssueID != ticketID {
			continue
		}

		match := ripplePatternOld.FindStringSubmatch(interaction.Body)
		if match == nil {
			continue
		}
		triggeringID := match[1]
		summary := strings.TrimSpace(match[2])
		if summary == "" {
			continue
		}

		diff, err := ticketdomain.NewContextDiff(summary, triggeringID, interaction.CreatedAt)
		if err != nil {
			continue
		}
		flags = append(flags, ticketdomain.NewFreshnessFlag(diff, interaction.CreatedAt))
	}
	return flags
}

// ------------------------------------------------------------------
// bd CLI integration
// ------------------------------------------------------------------

func (r *BeadsTicketReader) getFlaggedIDs(ctx context.Context) map[string]struct{} {
	ctx, cancel := context.WithTimeout(ctx, bdTimeoutSeconds*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bd", "query", "label=review_needed")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	ids := make(map[string]struct{})
	for _, line := range strings.Split(string(output), "\n") {
		match := bdQueryLineRE.FindStringSubmatch(strings.TrimSpace(line))
		if match != nil {
			ids[match[1]] = struct{}{}
		}
	}
	return ids
}

func (r *BeadsTicketReader) enrichLabels(tickets []ticketdomain.OpenTicketData, flaggedIDs map[string]struct{}) []ticketdomain.OpenTicketData {
	enriched := make([]ticketdomain.OpenTicketData, len(tickets))
	for i, t := range tickets {
		var labels []string
		if _, ok := flaggedIDs[t.TicketID()]; ok {
			labels = []string{"review_needed"}
		}
		enriched[i] = ticketdomain.NewOpenTicketData(t.TicketID(), t.Title(), labels, t.LastReviewed())
	}
	return enriched
}

func (r *BeadsTicketReader) readFlagsFromBDComments(ctx context.Context, ticketID string) []ticketdomain.FreshnessFlag {
	ctx, cancel := context.WithTimeout(ctx, bdTimeoutSeconds*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bd", "comments", ticketID)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	return parseBDComments(string(output))
}

func parseBDComments(output string) []ticketdomain.FreshnessFlag {
	var flags []ticketdomain.FreshnessFlag

	blocks := commentBlockRE.Split(output, -1)
	dates := commentBlockRE.FindAllStringSubmatch(output, -1)

	for i, dateMatch := range dates {
		if i+1 >= len(blocks) {
			break
		}
		dateStr := strings.TrimSpace(dateMatch[1])
		body := blocks[i+1]

		// Try current format
		triggerMatch := rippleTriggerRE.FindStringSubmatch(body)
		changedMatch := whatChangedRE.FindStringSubmatch(body)
		if triggerMatch != nil && changedMatch != nil {
			summary := strings.TrimSpace(changedMatch[1])
			if summary != "" {
				diff, err := ticketdomain.NewContextDiff(summary, triggerMatch[1], dateStr)
				if err == nil {
					flags = append(flags, ticketdomain.NewFreshnessFlag(diff, dateStr))
				}
			}
			continue
		}

		// Try old format
		oldMatch := ripplePatternOld.FindStringSubmatch(body)
		if oldMatch != nil {
			summary := strings.TrimSpace(oldMatch[2])
			if summary != "" {
				diff, err := ticketdomain.NewContextDiff(summary, oldMatch[1], dateStr)
				if err == nil {
					flags = append(flags, ticketdomain.NewFreshnessFlag(diff, dateStr))
				}
			}
		}
	}
	return flags
}
