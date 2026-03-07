// Package domain provides the DocHealth bounded context's core domain model.
// It contains value objects for document health checking: statuses, registry entries,
// broken links, doc statuses, health reports, and review results.
package domain

import (
	"fmt"
	"time"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// DocHealthStatus enumerates possible document health states.
type DocHealthStatus string

// Document health status constants.
const (
	DocHealthOK            DocHealthStatus = "ok"
	DocHealthStale         DocHealthStatus = "stale"
	DocHealthMissing       DocHealthStatus = "missing"
	DocHealthNoFrontmatter DocHealthStatus = "no_frontmatter"
)

// BrokenLink is a broken internal markdown link found in a document.
type BrokenLink struct {
	linkText   string
	target     string
	reason     string
	lineNumber int
}

// NewBrokenLink creates a validated BrokenLink value object.
func NewBrokenLink(lineNumber int, linkText, target, reason string) (BrokenLink, error) {
	if lineNumber < 1 {
		return BrokenLink{}, fmt.Errorf("line_number must be >= 1, got %d: %w",
			lineNumber, domainerrors.ErrInvariantViolation)
	}
	return BrokenLink{
		lineNumber: lineNumber,
		linkText:   linkText,
		target:     target,
		reason:     reason,
	}, nil
}

// LineNumber returns the line number where the broken link was found.
func (bl BrokenLink) LineNumber() int { return bl.lineNumber }

// LinkText returns the link text.
func (bl BrokenLink) LinkText() string { return bl.linkText }

// Target returns the link target.
func (bl BrokenLink) Target() string { return bl.target }

// Reason returns why the link is broken.
func (bl BrokenLink) Reason() string { return bl.reason }

// DocRegistryEntry is a document registered for health tracking.
type DocRegistryEntry struct {
	path               string
	owner              string
	reviewIntervalDays int
}

// NewDocRegistryEntry creates a validated DocRegistryEntry.
func NewDocRegistryEntry(path, owner string, reviewIntervalDays int) (DocRegistryEntry, error) {
	if reviewIntervalDays <= 0 {
		return DocRegistryEntry{}, fmt.Errorf(
			"review_interval_days must be positive, got %d: %w",
			reviewIntervalDays, domainerrors.ErrInvariantViolation)
	}
	return DocRegistryEntry{
		path:               path,
		owner:              owner,
		reviewIntervalDays: reviewIntervalDays,
	}, nil
}

// Path returns the document path.
func (e DocRegistryEntry) Path() string { return e.path }

// Owner returns the document owner.
func (e DocRegistryEntry) Owner() string { return e.owner }

// ReviewIntervalDays returns the review interval in days.
func (e DocRegistryEntry) ReviewIntervalDays() int { return e.reviewIntervalDays }

// DocStatus captures the health of a single document.
type DocStatus struct {
	path               string
	status             DocHealthStatus
	lastReviewed       *time.Time
	daysSince          *int
	owner              string
	brokenLinks        []BrokenLink
	reviewIntervalDays int
}

// NewDocStatus creates a DocStatus value object directly.
func NewDocStatus(
	path string,
	status DocHealthStatus,
	lastReviewed *time.Time,
	daysSince *int,
	reviewIntervalDays int,
	owner string,
	brokenLinks []BrokenLink,
) DocStatus {
	bl := make([]BrokenLink, len(brokenLinks))
	copy(bl, brokenLinks)
	return DocStatus{
		path:               path,
		status:             status,
		lastReviewed:       lastReviewed,
		daysSince:          daysSince,
		reviewIntervalDays: reviewIntervalDays,
		owner:              owner,
		brokenLinks:        bl,
	}
}

// Path returns the document path.
func (s DocStatus) Path() string { return s.path }

// Status returns the document health status.
func (s DocStatus) Status() DocHealthStatus { return s.status }

// LastReviewed returns the last review timestamp.
func (s DocStatus) LastReviewed() *time.Time { return s.lastReviewed }

// DaysSince returns the number of days since last review.
func (s DocStatus) DaysSince() *int { return s.daysSince }

// ReviewIntervalDays returns the review interval in days.
func (s DocStatus) ReviewIntervalDays() int { return s.reviewIntervalDays }

// Owner returns the document owner.
func (s DocStatus) Owner() string { return s.owner }

// BrokenLinks returns a defensive copy.
func (s DocStatus) BrokenLinks() []BrokenLink {
	out := make([]BrokenLink, len(s.brokenLinks))
	copy(out, s.brokenLinks)
	return out
}

// CreateDocStatus is a factory that auto-calculates status and days_since.
func CreateDocStatus(
	path string,
	exists bool,
	lastReviewed *time.Time,
	reviewIntervalDays int,
	owner string,
	today *time.Time,
	brokenLinks []BrokenLink,
) DocStatus {
	bl := make([]BrokenLink, len(brokenLinks))
	copy(bl, brokenLinks)

	if !exists {
		return DocStatus{
			path:               path,
			status:             DocHealthMissing,
			reviewIntervalDays: reviewIntervalDays,
			owner:              owner,
			brokenLinks:        bl,
		}
	}
	if lastReviewed == nil {
		return DocStatus{
			path:               path,
			status:             DocHealthNoFrontmatter,
			reviewIntervalDays: reviewIntervalDays,
			owner:              owner,
			brokenLinks:        bl,
		}
	}

	effectiveToday := time.Now().Truncate(24 * time.Hour)
	if today != nil {
		effectiveToday = *today
	}
	days := int(effectiveToday.Sub(*lastReviewed).Hours() / 24)

	status := DocHealthOK
	if days > reviewIntervalDays {
		status = DocHealthStale
	}

	lr := *lastReviewed
	return DocStatus{
		path:               path,
		status:             status,
		lastReviewed:       &lr,
		daysSince:          &days,
		reviewIntervalDays: reviewIntervalDays,
		owner:              owner,
		brokenLinks:        bl,
	}
}

// DocHealthReport aggregates document health statuses into a summary.
type DocHealthReport struct {
	statuses []DocStatus
}

// NewDocHealthReport creates a DocHealthReport.
func NewDocHealthReport(statuses []DocStatus) DocHealthReport {
	s := make([]DocStatus, len(statuses))
	copy(s, statuses)
	return DocHealthReport{statuses: s}
}

// Statuses returns a defensive copy of all statuses.
func (r DocHealthReport) Statuses() []DocStatus {
	out := make([]DocStatus, len(r.statuses))
	copy(out, r.statuses)
	return out
}

// IssueCount returns the count of documents with non-OK status or broken links.
func (r DocHealthReport) IssueCount() int {
	count := 0
	for _, s := range r.statuses {
		if s.status != DocHealthOK || len(s.brokenLinks) > 0 {
			count++
		}
	}
	return count
}

// TotalChecked returns the total number of documents checked.
func (r DocHealthReport) TotalChecked() int {
	return len(r.statuses)
}

// HasIssues returns whether any document has a non-OK status or broken links.
func (r DocHealthReport) HasIssues() bool {
	return r.IssueCount() > 0
}

// DocReviewResult is the result of marking a document as reviewed.
type DocReviewResult struct {
	newDate time.Time
	path    string
}

// NewDocReviewResult creates a DocReviewResult.
func NewDocReviewResult(path string, newDate time.Time) DocReviewResult {
	return DocReviewResult{path: path, newDate: newDate}
}

// Path returns the document path.
func (r DocReviewResult) Path() string { return r.path }

// NewDate returns the new review date.
func (r DocReviewResult) NewDate() time.Time { return r.newDate }
