package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// VerifiableClaim represents a quantitative claim in a ticket that can be verified.
type VerifiableClaim struct {
	ticketID     string
	claimText    string
	command      string
	claimedValue string
	location     string
}

// NewVerifiableClaim creates a VerifiableClaim value object.
// Returns error if ticketID or claimText is empty.
func NewVerifiableClaim(ticketID, claimText, command, claimedValue, location string) (VerifiableClaim, error) {
	if ticketID == "" {
		return VerifiableClaim{}, fmt.Errorf("ticket ID cannot be empty")
	}
	if claimText == "" {
		return VerifiableClaim{}, fmt.Errorf("claim text cannot be empty")
	}
	return VerifiableClaim{
		ticketID:     ticketID,
		claimText:    claimText,
		command:      command,
		claimedValue: claimedValue,
		location:     location,
	}, nil
}

// TicketID returns the ticket identifier.
func (c VerifiableClaim) TicketID() string { return c.ticketID }

// ClaimText returns the original claim text.
func (c VerifiableClaim) ClaimText() string { return c.claimText }

// Command returns the verification command (may be empty).
func (c VerifiableClaim) Command() string { return c.command }

// ClaimedValue returns the claimed numeric value as string.
func (c VerifiableClaim) ClaimedValue() string { return c.claimedValue }

// Location returns where the claim appears in the ticket.
func (c VerifiableClaim) Location() string { return c.location }

// IsVerifiable returns true if the claim has an associated command.
func (c VerifiableClaim) IsVerifiable() bool { return c.command != "" }

// VerificationResult represents the outcome of verifying a claim.
type VerificationResult struct {
	claim       VerifiableClaim
	actualValue string
	match       bool
	discrepancy string
}

// NewVerificationResult creates a VerificationResult by comparing claimed vs actual.
func NewVerificationResult(claim VerifiableClaim, actualValue string, err error) VerificationResult {
	if err != nil {
		return VerificationResult{
			claim:       claim,
			actualValue: "",
			match:       false,
			discrepancy: fmt.Sprintf("command error: %v", err),
		}
	}

	if !claim.IsVerifiable() {
		return VerificationResult{
			claim:       claim,
			actualValue: "",
			match:       false,
			discrepancy: "UNVERIFIED: no command to verify claim",
		}
	}

	actual := strings.TrimSpace(actualValue)
	claimed := strings.TrimSpace(claim.claimedValue)

	if actual == claimed {
		return VerificationResult{
			claim:       claim,
			actualValue: actual,
			match:       true,
			discrepancy: "",
		}
	}

	discrepancy := fmt.Sprintf("claimed %s, actual %s", claimed, actual)

	// Calculate magnitude if both are numeric
	claimedNum, errC := strconv.ParseFloat(claimed, 64)
	actualNum, errA := strconv.ParseFloat(actual, 64)
	if errC == nil && errA == nil && claimedNum > 0 {
		ratio := actualNum / claimedNum
		if ratio >= 2 {
			discrepancy += fmt.Sprintf(" (%.0fx difference)", ratio)
		} else if ratio <= 0.5 {
			discrepancy += fmt.Sprintf(" (%.1fx difference)", ratio)
		}
	}

	return VerificationResult{
		claim:       claim,
		actualValue: actual,
		match:       false,
		discrepancy: discrepancy,
	}
}

// Claim returns the original claim.
func (r VerificationResult) Claim() VerifiableClaim { return r.claim }

// ActualValue returns the actual value from verification.
func (r VerificationResult) ActualValue() string { return r.actualValue }

// Match returns true if claimed equals actual.
func (r VerificationResult) Match() bool { return r.match }

// Discrepancy returns a description of the mismatch (empty if match).
func (r VerificationResult) Discrepancy() string { return r.discrepancy }

// ClaimVerifier parses ticket markdown to find verifiable claims.
type ClaimVerifier struct {
	// Patterns for finding quantitative claims
	boldNumberPattern *regexp.Regexp
	codeBlockPattern  *regexp.Regexp
}

// NewClaimVerifier creates a ClaimVerifier with default patterns.
func NewClaimVerifier() *ClaimVerifier {
	return &ClaimVerifier{
		// Matches **N things** or **N-word things**
		boldNumberPattern: regexp.MustCompile(`\*\*(\d+)\s+([a-zA-Z][a-zA-Z\s-]*[a-zA-Z])\*\*`),
		// Matches ```bash or ``` code blocks
		codeBlockPattern: regexp.MustCompile("(?s)```(?:bash|sh)?\\s*\\n([^`]+)```"),
	}
}

// ParseClaims extracts verifiable claims from ticket markdown.
func (v *ClaimVerifier) ParseClaims(ticketID, markdown string) []VerifiableClaim {
	var claims []VerifiableClaim

	// Find all code blocks first (to associate commands with claims)
	codeBlocks := v.codeBlockPattern.FindAllStringSubmatch(markdown, -1)
	var lastCommand string
	if len(codeBlocks) > 0 {
		// Use the last code block as the potential command
		lastCommand = strings.TrimSpace(codeBlocks[len(codeBlocks)-1][1])
	}

	// Find all bold number patterns
	matches := v.boldNumberPattern.FindAllStringSubmatchIndex(markdown, -1)

	for _, match := range matches {
		if len(match) >= 6 {
			fullMatch := markdown[match[0]:match[1]]
			number := markdown[match[2]:match[3]]
			word := markdown[match[4]:match[5]]

			// Filter out non-quantitative phrases
			wordLower := strings.ToLower(word)
			if isQuantitativeWord(wordLower) {
				claim, err := NewVerifiableClaim(
					ticketID,
					fullMatch,
					lastCommand,
					number,
					fmt.Sprintf("position %d", match[0]),
				)
				if err == nil {
					claims = append(claims, claim)
				}
			}
		}
	}

	return claims
}

// isQuantitativeWord returns true if the word suggests a countable thing.
func isQuantitativeWord(word string) bool {
	quantitativeTerms := []string{
		"finding", "findings",
		"issue", "issues",
		"error", "errors",
		"warning", "warnings",
		"function", "functions",
		"file", "files",
		"line", "lines",
		"test", "tests",
		"failure", "failures",
		"critical", "production",
	}

	for _, term := range quantitativeTerms {
		if strings.Contains(word, term) {
			return true
		}
	}
	return false
}
