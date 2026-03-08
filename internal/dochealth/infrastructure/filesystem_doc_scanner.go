// Package infrastructure provides adapters for the DocHealth bounded context.
package infrastructure

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	dochealthdomain "github.com/alty-cli/alty/internal/dochealth/domain"
)

var (
	frontmatterRE  = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---`)
	lastReviewedRE = regexp.MustCompile(`(?m)^last_reviewed:\s*(.+)$`)
	placeholderRE  = regexp.MustCompile(`^[A-Z]{4}-[A-Z]{2}-[A-Z]{2}$`)
	markdownLinkRE = regexp.MustCompile(`(?:^|[^!])\[([^\]]*)\]\(([^)]*)\)`)
)

// FilesystemDocScanner scans the filesystem for document health status.
type FilesystemDocScanner struct{}

// NewFilesystemDocScanner creates a new FilesystemDocScanner.
func NewFilesystemDocScanner() *FilesystemDocScanner {
	return &FilesystemDocScanner{}
}

// LoadRegistry loads document registry entries from a TOML file.
func (s *FilesystemDocScanner) LoadRegistry(registryPath string) ([]dochealthdomain.DocRegistryEntry, error) {
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		return nil, nil
	}

	content, err := os.ReadFile(registryPath)
	if err != nil {
		return nil, nil
	}

	rawEntries := parseTOMLDocs(string(content))
	var entries []dochealthdomain.DocRegistryEntry
	for _, raw := range rawEntries {
		path, ok := raw["path"].(string)
		if !ok {
			continue
		}
		owner := ""
		if o, ok := raw["owner"].(string); ok {
			owner = o
		}
		interval := 30
		if i, ok := raw["review_interval_days"].(int); ok {
			interval = i
		}
		entry, err := dochealthdomain.NewDocRegistryEntry(path, owner, interval)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// ScanRegistered scans registered documents for health status.
func (s *FilesystemDocScanner) ScanRegistered(entries []dochealthdomain.DocRegistryEntry, projectDir string) ([]dochealthdomain.DocStatus, error) {
	var statuses []dochealthdomain.DocStatus
	now := time.Now().Truncate(24 * time.Hour)

	for _, entry := range entries {
		filePath := filepath.Join(projectDir, entry.Path())
		info, err := os.Stat(filePath)
		exists := err == nil && !info.IsDir()

		var lastReviewed *time.Time
		var brokenLinks []dochealthdomain.BrokenLink
		if exists {
			content, readErr := os.ReadFile(filePath)
			if readErr == nil {
				lastReviewed = parseLastReviewed(string(content))
				links := extractMarkdownLinks(string(content))
				brokenLinks = checkBrokenLinks(links, filePath)
			}
		}

		status := dochealthdomain.CreateDocStatus(
			entry.Path(),
			exists,
			lastReviewed,
			entry.ReviewIntervalDays(),
			entry.Owner(),
			&now,
			brokenLinks,
		)
		statuses = append(statuses, status)
	}
	return statuses, nil
}

// ScanUnregistered scans for markdown files not in the registry.
func (s *FilesystemDocScanner) ScanUnregistered(
	docsDir string,
	registeredPaths map[string]bool,
	excludeDirs []string,
) ([]dochealthdomain.DocStatus, error) {
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		return nil, nil
	}

	now := time.Now().Truncate(24 * time.Hour)
	excludeSet := make(map[string]bool)
	for _, d := range excludeDirs {
		excludeSet[d] = true
	}

	var mdFiles []string
	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".md" {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking docs directory: %w", err)
	}
	sort.Strings(mdFiles)

	var statuses []dochealthdomain.DocStatus
	parentDir := filepath.Dir(docsDir)

	for _, mdFile := range mdFiles {
		// Build relative path from docs_dir's parent (project root)
		relPath, err := filepath.Rel(parentDir, mdFile)
		if err != nil {
			relPath, _ = filepath.Rel(docsDir, mdFile)
		}

		// Skip registered paths
		if registeredPaths != nil && registeredPaths[relPath] {
			continue
		}

		// Skip excluded directories
		relFromDocs, _ := filepath.Rel(docsDir, mdFile)
		parts := strings.Split(relFromDocs, string(filepath.Separator))
		excluded := false
		for _, part := range parts {
			if excludeSet[part] {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Check frontmatter and links
		content, readErr := os.ReadFile(mdFile)
		var lastReviewed *time.Time
		var brokenLinks []dochealthdomain.BrokenLink
		if readErr == nil {
			lastReviewed = parseLastReviewed(string(content))
			links := extractMarkdownLinks(string(content))
			brokenLinks = checkBrokenLinks(links, mdFile)
		}

		status := dochealthdomain.CreateDocStatus(
			relPath,
			true,
			lastReviewed,
			30,
			"",
			&now,
			brokenLinks,
		)
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func parseLastReviewed(content string) *time.Time {
	fmMatch := frontmatterRE.FindStringSubmatch(content)
	if fmMatch == nil {
		return nil
	}

	lrMatch := lastReviewedRE.FindStringSubmatch(fmMatch[1])
	if lrMatch == nil {
		return nil
	}

	raw := strings.TrimSpace(lrMatch[1])
	raw = strings.Trim(raw, "\"'")

	if placeholderRE.MatchString(raw) {
		return nil
	}

	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil
	}
	return &t
}

type linkInfo struct {
	lineNumber int
	linkText   string
	target     string
}

func extractMarkdownLinks(content string) []linkInfo {
	var results []linkInfo
	inFence := false
	for lineno, line := range strings.Split(content, "\n") {
		stripped := strings.TrimSpace(line)
		if strings.HasPrefix(stripped, "```") || strings.HasPrefix(stripped, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		matches := markdownLinkRE.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			results = append(results, linkInfo{
				lineNumber: lineno + 1,
				linkText:   m[1],
				target:     m[2],
			})
		}
	}
	return results
}

func checkBrokenLinks(links []linkInfo, docPath string) []dochealthdomain.BrokenLink {
	var broken []dochealthdomain.BrokenLink
	for _, link := range links {
		if link.target == "" {
			bl, err := dochealthdomain.NewBrokenLink(link.lineNumber, link.linkText, link.target, "empty target")
			if err == nil {
				broken = append(broken, bl)
			}
			continue
		}
		if strings.HasPrefix(link.target, "http://") || strings.HasPrefix(link.target, "https://") ||
			strings.HasPrefix(link.target, "mailto:") {
			continue
		}
		if strings.HasPrefix(link.target, "#") {
			continue
		}
		pathPart := strings.SplitN(link.target, "#", 2)[0]
		if pathPart == "" {
			continue
		}
		resolved := filepath.Join(filepath.Dir(docPath), pathPart)
		if _, err := os.Stat(resolved); os.IsNotExist(err) {
			bl, err := dochealthdomain.NewBrokenLink(link.lineNumber, link.linkText, link.target, "target not found")
			if err == nil {
				broken = append(broken, bl)
			}
		}
	}
	return broken
}

// parseTOMLDocs parses [[docs]] entries from a simple TOML file.
func parseTOMLDocs(content string) []map[string]any {
	var entries []map[string]any
	var current map[string]any

	for _, line := range strings.Split(content, "\n") {
		stripped := strings.TrimSpace(line)
		if stripped == "[[docs]]" {
			if current != nil {
				entries = append(entries, current)
			}
			current = make(map[string]any)
			continue
		}
		if current == nil || stripped == "" || strings.HasPrefix(stripped, "#") || !strings.Contains(stripped, "=") {
			continue
		}
		key, value, found := strings.Cut(stripped, "=")
		if !found {
			continue
		}
		k := strings.TrimSpace(key)
		v := strings.TrimSpace(value)
		current[k] = parseTOMLValue(v)
	}
	if current != nil {
		entries = append(entries, current)
	}
	return entries
}

func parseTOMLValue(value string) any {
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return value[1 : len(value)-1]
	}
	var n int
	if _, err := fmt.Sscanf(value, "%d", &n); err == nil {
		return n
	}
	return value
}
