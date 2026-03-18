package infrastructure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	dochealthapp "github.com/alto-cli/alto/internal/dochealth/application"
	"github.com/alto-cli/alto/internal/dochealth/domain"
)

// Compile-time interface check.
var _ dochealthapp.DocReview = (*DocReviewAdapter)(nil)

// DocReviewAdapter implements the DocReview port by reading and updating
// YAML frontmatter in markdown files.
type DocReviewAdapter struct {
	scanner *FilesystemDocScanner
}

// NewDocReviewAdapter creates a new DocReviewAdapter.
func NewDocReviewAdapter(scanner *FilesystemDocScanner) *DocReviewAdapter {
	return &DocReviewAdapter{scanner: scanner}
}

// ReviewableDocs returns docs that are due for review by scanning the registry
// and checking frontmatter dates.
func (a *DocReviewAdapter) ReviewableDocs(ctx context.Context, projectDir string) ([]domain.DocStatus, error) {
	registryPath := filepath.Join(projectDir, ".alto", "maintenance", "doc-registry.toml")

	entries, err := a.scanner.LoadRegistry(registryPath)
	if err != nil {
		return nil, fmt.Errorf("loading doc registry: %w", err)
	}
	if len(entries) == 0 {
		return nil, nil
	}

	statuses, err := a.scanner.ScanRegistered(entries, projectDir)
	if err != nil {
		return nil, fmt.Errorf("scanning registered docs: %w", err)
	}

	// Filter to only stale docs.
	var reviewable []domain.DocStatus
	for _, s := range statuses {
		if s.Status() == domain.DocHealthStale || s.Status() == domain.DocHealthNoFrontmatter {
			reviewable = append(reviewable, s)
		}
	}
	return reviewable, nil
}

// MarkReviewed marks a document as reviewed by updating its YAML frontmatter.
// Pass nil for reviewDate to use current time.
func (a *DocReviewAdapter) MarkReviewed(_ context.Context, docPath, projectDir string, reviewDate *time.Time) (domain.DocReviewResult, error) {
	fullPath := filepath.Join(projectDir, docPath)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return domain.DocReviewResult{}, fmt.Errorf("reading %s: %w", docPath, err)
	}

	effectiveDate := time.Now().Truncate(24 * time.Hour)
	if reviewDate != nil {
		effectiveDate = *reviewDate
	}

	dateStr := effectiveDate.Format("2006-01-02")
	updated, err := updateFrontmatter(string(data), dateStr)
	if err != nil {
		return domain.DocReviewResult{}, fmt.Errorf("updating frontmatter in %s: %w", docPath, err)
	}

	if err := os.WriteFile(fullPath, []byte(updated), 0o644); err != nil {
		return domain.DocReviewResult{}, fmt.Errorf("writing %s: %w", docPath, err)
	}

	return domain.NewDocReviewResult(docPath, effectiveDate), nil
}

// MarkAllReviewed marks all stale docs as reviewed.
func (a *DocReviewAdapter) MarkAllReviewed(ctx context.Context, projectDir string, reviewDate *time.Time) ([]domain.DocReviewResult, error) {
	stale, err := a.ReviewableDocs(ctx, projectDir)
	if err != nil {
		return nil, fmt.Errorf("finding reviewable docs: %w", err)
	}

	var results []domain.DocReviewResult
	for _, doc := range stale {
		result, markErr := a.MarkReviewed(ctx, doc.Path(), projectDir, reviewDate)
		if markErr != nil {
			return nil, fmt.Errorf("marking %s: %w", doc.Path(), markErr)
		}
		results = append(results, result)
	}
	return results, nil
}

// updateFrontmatter updates or inserts the last_reviewed field in YAML frontmatter.
// It preserves the document body byte-exact outside the frontmatter block.
func updateFrontmatter(content, dateStr string) (string, error) {
	const delimiter = "---"

	// Check if file starts with frontmatter delimiter.
	if !strings.HasPrefix(content, delimiter+"\n") {
		// No frontmatter — insert it at the top.
		fm := delimiter + "\nlast_reviewed: \"" + dateStr + "\"\n" + delimiter + "\n"
		return fm + content, nil
	}

	// Find closing delimiter: look for "\n---\n" or "\n---" at EOF after the opening.
	rest := content[len(delimiter)+1:] // skip "---\n"
	closingIdx := strings.Index(rest, "\n"+delimiter+"\n")
	var fmYAML, body string
	if closingIdx == -1 {
		// Try "\n---" at end of string (no trailing newline after closing delimiter).
		closingIdx = strings.Index(rest, "\n"+delimiter)
		if closingIdx == -1 || closingIdx+len("\n"+delimiter) != len(rest) {
			// Malformed frontmatter — insert fresh.
			fm := delimiter + "\nlast_reviewed: \"" + dateStr + "\"\n" + delimiter + "\n"
			return fm + content, nil
		}
		fmYAML = rest[:closingIdx]
		body = ""
	} else {
		fmYAML = rest[:closingIdx]
		body = rest[closingIdx+len("\n"+delimiter+"\n"):]
	}

	// Parse frontmatter as ordered map.
	var fmMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(fmYAML), &fmMap); err != nil {
		return "", fmt.Errorf("parsing frontmatter YAML: %w", err)
	}
	if fmMap == nil {
		fmMap = make(map[string]interface{})
	}

	// Update the last_reviewed field.
	fmMap["last_reviewed"] = dateStr

	// Serialize back.
	fmBytes, err := yaml.Marshal(fmMap)
	if err != nil {
		return "", fmt.Errorf("serializing frontmatter YAML: %w", err)
	}
	// yaml.Marshal adds trailing newline; we want it before the closing delimiter.
	fmStr := strings.TrimRight(string(fmBytes), "\n")

	// Reassemble.
	var result strings.Builder
	result.WriteString(delimiter + "\n")
	result.WriteString(fmStr + "\n")
	result.WriteString(delimiter + "\n")
	result.WriteString(body)

	return result.String(), nil
}
