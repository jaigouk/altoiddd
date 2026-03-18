package infrastructure

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	researchapp "github.com/alto-cli/alto/internal/research/application"
	researchdomain "github.com/alto-cli/alto/internal/research/domain"
)

// Compile-time interface check.
var _ researchapp.SpikeFollowUp = (*SpikeFollowUpAdapter)(nil)

// Prefixes stripped before comparison (case-insensitive).
var stripPrefixes = []string{
	"task:",
	"spike:",
	"bug:",
	"feature:",
	"(optional)",
}

// Short stop-words excluded from keyword overlap scoring.
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true,
	"of": true, "to": true, "in": true, "for": true, "on": true,
	"with": true, "is": true, "it": true, "be": true, "as": true,
	"at": true, "by": true, "from": true, "that": true, "this": true,
}

// Minimum Jaccard similarity for keyword overlap match.
const keywordOverlapThreshold = 0.4

// parenRE extracts content from parentheses.
var parenRE = regexp.MustCompile(`\(([^)]+)\)`)

// tokenSplitRE splits on whitespace/punctuation for tokenization.
var tokenSplitRE = regexp.MustCompile(`[\s\-_/,()]+`)

// parenTokenSplitRE splits parenthetical content.
var parenTokenSplitRE = regexp.MustCompile(`[\s,+/]+`)

// SpikeFollowUpAdapter audits spike follow-ups against created tickets.
// Implements SpikeFollowUp port by scanning research reports for follow-up
// intents and comparing them against beads tickets using fuzzy title matching.
type SpikeFollowUpAdapter struct {
	parser *MarkdownSpikeParser
}

// NewSpikeFollowUpAdapter creates a new SpikeFollowUpAdapter.
func NewSpikeFollowUpAdapter() *SpikeFollowUpAdapter {
	return &SpikeFollowUpAdapter{
		parser: NewMarkdownSpikeParser(),
	}
}

// Audit audits a spike's follow-up intents against created tickets.
func (a *SpikeFollowUpAdapter) Audit(ctx context.Context, spikeID string, projectDir string) (researchdomain.FollowUpAuditResult, error) {
	researchDir := filepath.Join(projectDir, "docs", "research")
	if _, err := os.Stat(researchDir); os.IsNotExist(err) {
		return researchdomain.NewFollowUpAuditResult(spikeID, "", nil, nil, nil), nil
	}

	// Scan all Markdown reports in docs/research/
	entries, err := os.ReadDir(researchDir)
	if err != nil {
		return researchdomain.NewFollowUpAuditResult(spikeID, "", nil, nil, nil), nil
	}

	// Sort entries by name for deterministic order
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var allIntents []researchdomain.FollowUpIntent
	reportPath := ""
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		fullPath := filepath.Join(researchDir, entry.Name())
		intents, parseErr := a.parser.Parse(ctx, fullPath)
		if parseErr != nil {
			continue
		}
		if len(intents) > 0 {
			allIntents = append(allIntents, intents...)
			if reportPath == "" {
				reportPath = fullPath
			}
		}
	}

	if len(allIntents) == 0 {
		return researchdomain.NewFollowUpAuditResult(spikeID, reportPath, nil, nil, nil), nil
	}

	// Load existing beads tickets
	existingTitles := loadTicketTitles(projectDir)

	// Match intents against tickets
	var matchedIDs []string
	var orphaned []researchdomain.FollowUpIntent

	for _, intent := range allIntents {
		ticketID := FuzzyMatch(intent.Title(), existingTitles)
		if ticketID != "" {
			matchedIDs = append(matchedIDs, ticketID)
		} else {
			orphaned = append(orphaned, intent)
		}
	}

	return researchdomain.NewFollowUpAuditResult(spikeID, reportPath, allIntents, matchedIDs, orphaned), nil
}

// loadTicketTitles loads ticket ID -> title mapping from .beads/issues.jsonl.
func loadTicketTitles(projectDir string) map[string]string {
	issuesPath := filepath.Join(projectDir, ".beads", "issues.jsonl")
	data, err := os.ReadFile(issuesPath)
	if err != nil {
		return map[string]string{}
	}

	titles := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var issue map[string]string
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			continue
		}
		ticketID := issue["id"]
		title := issue["title"]
		if ticketID != "" && title != "" {
			titles[ticketID] = title
		}
	}
	return titles
}

// FuzzyMatch finds a beads ticket that fuzzy-matches the intent title.
// Exported for testing. Returns the matching ticket ID, or empty string.
//
// Matching strategy (ordered by strictness):
//  1. Case-insensitive exact match (after prefix stripping)
//  2. Case-insensitive substring (intent in ticket or ticket in intent)
//  3. Keyword overlap (Jaccard similarity >= threshold)
func FuzzyMatch(intentTitle string, existing map[string]string) string {
	strippedIntent := stripPrefixesFromTitle(intentTitle)
	if strippedIntent == "" {
		return ""
	}
	intentLower := strings.ToLower(strippedIntent)

	for ticketID, title := range existing {
		strippedTicket := stripPrefixesFromTitle(title)
		if strippedTicket == "" {
			continue
		}
		ticketLower := strings.ToLower(strippedTicket)

		// Tier 1: Exact match
		if intentLower == ticketLower {
			return ticketID
		}

		// Tier 2: Substring
		if strings.Contains(ticketLower, intentLower) || strings.Contains(intentLower, ticketLower) {
			return ticketID
		}

		// Tier 3: Keyword overlap
		if keywordOverlap(intentLower, ticketLower) >= keywordOverlapThreshold {
			return ticketID
		}
	}

	return ""
}

// stripPrefixesFromTitle removes common prefixes from a title for comparison.
func stripPrefixesFromTitle(title string) string {
	result := strings.TrimSpace(title)
	lower := strings.ToLower(result)
	for _, prefix := range stripPrefixes {
		if strings.HasPrefix(lower, prefix) {
			result = strings.TrimSpace(result[len(prefix):])
			lower = strings.ToLower(result)
		}
	}
	return result
}

// tokenize splits a title into meaningful keywords.
// Splits on whitespace/punctuation, lowercases, removes stop-words,
// and extracts parenthetical content as additional tokens.
func tokenize(text string) map[string]bool {
	// Extract parenthetical content as extra tokens
	parenMatches := parenRE.FindAllStringSubmatch(text, -1)
	parenTokens := make(map[string]bool)
	for _, m := range parenMatches {
		for _, token := range parenTokenSplitRE.Split(strings.ToLower(m[1]), -1) {
			token = strings.TrimSpace(token)
			if token != "" && !stopWords[token] && len(token) > 1 {
				parenTokens[token] = true
			}
		}
	}

	// Main tokenization
	words := tokenSplitRE.Split(strings.ToLower(text), -1)
	tokens := make(map[string]bool)
	for _, w := range words {
		if w != "" && !stopWords[w] && len(w) > 1 {
			tokens[w] = true
		}
	}

	// Merge
	for k := range parenTokens {
		tokens[k] = true
	}
	return tokens
}

// keywordOverlap computes fuzzy keyword overlap between two titles.
// Uses prefix matching (first 5 chars) for each token pair to handle
// morphological variants like "generate"/"generation".
// Returns a score from 0.0 to 1.0.
func keywordOverlap(a, b string) float64 {
	tokensA := tokenize(a)
	tokensB := tokenize(b)
	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0.0
	}

	matchedA := 0
	matchedBTokens := make(map[string]bool)
	for ta := range tokensA {
		for tb := range tokensB {
			if matchedBTokens[tb] {
				continue
			}
			if ta == tb || (len(ta) >= 5 && len(tb) >= 5 && ta[:5] == tb[:5]) {
				matchedA++
				matchedBTokens[tb] = true
				break
			}
		}
	}

	totalUnique := len(tokensA) + len(tokensB) - matchedA
	if totalUnique == 0 {
		return 0.0
	}
	return float64(matchedA) / float64(totalUnique)
}
