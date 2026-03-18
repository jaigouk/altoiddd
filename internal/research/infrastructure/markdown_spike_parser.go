// Package infrastructure provides adapters for the Research bounded context.
package infrastructure

import (
	"context"
	"os"
	"regexp"
	"strings"

	researchapp "github.com/alto-cli/alto/internal/research/application"
	researchdomain "github.com/alto-cli/alto/internal/research/domain"
)

// Compile-time interface check.
var _ researchapp.SpikeReportParser = (*MarkdownSpikeParser)(nil)

var (
	// Matches ## or ### with optional "N." prefix
	followupHeadingRE = regexp.MustCompile(`(?i)^(#{2,3})\s+(?:\d+\.\s+)?(?:Step\s+\d+:\s+)?(.+)$`)
	followupKeywords  = regexp.MustCompile(`(?i)follow.?up`)
	ticketKeywords    = regexp.MustCompile(`(?i)ticket|implementation|work`)

	// ### Ticket N: Title
	ticketHeadingRE = regexp.MustCompile(`(?i)^#{3,4}\s+(?:Ticket\s+\d+:\s*)(.+)$`)

	// ### N. Title
	numberedHeadingRE = regexp.MustCompile(`^#{3,4}\s+\d+\.\s+(.+)$`)

	// - **Title**: Description
	boldListRE = regexp.MustCompile(`^[-*]\s+\*\*(.+?)\*\*(?:\s*[:\x{2014}\x{2013}-]\s*(.*))?$`)

	// - Plain title text
	plainListRE = regexp.MustCompile(`^[-*]\s+(.+)$`)

	// Any heading at level 1+
	anyHeadingRE = regexp.MustCompile(`^(#{1,6})\s+`)
)

// MarkdownSpikeParser parses Markdown spike reports to extract follow-up intents.
type MarkdownSpikeParser struct{}

// NewMarkdownSpikeParser creates a new MarkdownSpikeParser.
func NewMarkdownSpikeParser() *MarkdownSpikeParser {
	return &MarkdownSpikeParser{}
}

// Parse extracts follow-up intents from a Markdown spike report.
func (p *MarkdownSpikeParser) Parse(_ context.Context, reportPath string) ([]researchdomain.FollowUpIntent, error) {
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, nil
	}

	lines := strings.Split(string(data), "\n")

	sectionStart, sectionLevel := p.findFollowUpSection(lines)
	if sectionStart < 0 {
		return nil, nil
	}

	sectionLines := p.extractSection(lines, sectionStart, sectionLevel)
	return p.parseItems(sectionLines), nil
}

func (p *MarkdownSpikeParser) findFollowUpSection(lines []string) (int, int) {
	for i, line := range lines {
		m := followupHeadingRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		level := len(m[1])
		headingText := m[2]
		if followupKeywords.MatchString(headingText) && ticketKeywords.MatchString(headingText) {
			return i, level
		}
	}
	return -1, 0
}

func (p *MarkdownSpikeParser) extractSection(lines []string, start, level int) []string {
	var result []string
	for _, line := range lines[start+1:] {
		hm := anyHeadingRE.FindStringSubmatch(line)
		if hm != nil {
			currentLevel := len(hm[1])
			if currentLevel <= level {
				break
			}
		}
		result = append(result, line)
	}
	return result
}

func (p *MarkdownSpikeParser) parseItems(lines []string) []researchdomain.FollowUpIntent {
	var intents []researchdomain.FollowUpIntent
	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Try ### Ticket N: Title
		if m := ticketHeadingRE.FindStringSubmatch(line); m != nil {
			title := strings.TrimSpace(m[1])
			desc := p.collectDescription(lines, i+1)
			intent, err := researchdomain.NewFollowUpIntent(title, desc)
			if err == nil {
				intents = append(intents, intent)
			}
			i++
			continue
		}

		// Try ### N. Title
		if m := numberedHeadingRE.FindStringSubmatch(line); m != nil {
			title := strings.TrimSpace(m[1])
			desc := p.collectDescription(lines, i+1)
			intent, err := researchdomain.NewFollowUpIntent(title, desc)
			if err == nil {
				intents = append(intents, intent)
			}
			i++
			continue
		}

		// Try - **Title**: Description
		if m := boldListRE.FindStringSubmatch(trimmed); m != nil {
			title := strings.TrimSpace(m[1])
			desc := ""
			if len(m) > 2 {
				desc = strings.TrimSpace(m[2])
			}
			intent, err := researchdomain.NewFollowUpIntent(title, desc)
			if err == nil {
				intents = append(intents, intent)
			}
			i++
			continue
		}

		// Try - Plain title
		if m := plainListRE.FindStringSubmatch(trimmed); m != nil {
			if !strings.HasPrefix(trimmed, "- http") && !strings.HasPrefix(trimmed, "- [") {
				title := strings.TrimSpace(m[1])
				if !strings.HasPrefix(title, "**") && len(title) > 3 {
					intent, err := researchdomain.NewFollowUpIntent(title, "")
					if err == nil {
						intents = append(intents, intent)
					}
					i++
					continue
				}
			}
		}

		i++
	}
	return intents
}

func (p *MarkdownSpikeParser) collectDescription(lines []string, start int) string {
	var descLines []string
	metaPrefixes := []string{"**Type:", "**Priority:", "**Depends on:", "**Steps:", "**Bounded Context:"}
	for _, line := range lines[start:] {
		stripped := strings.TrimSpace(line)
		if strings.HasPrefix(stripped, "#") || boldListRE.MatchString(stripped) {
			break
		}
		isMeta := false
		for _, prefix := range metaPrefixes {
			if strings.HasPrefix(stripped, prefix) {
				isMeta = true
				break
			}
		}
		if isMeta {
			break
		}
		if stripped != "" {
			descLines = append(descLines, stripped)
		}
	}
	return strings.Join(descLines, " ")
}
