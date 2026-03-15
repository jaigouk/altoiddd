// Package infrastructure contains adapters for the DocImport bounded context.
package infrastructure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	importdomain "github.com/alty-cli/alty/internal/docimport/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// MarkdownDocParser parses DDD.md files into a DomainModel aggregate.
type MarkdownDocParser struct{}

// NewMarkdownDocParser creates a new MarkdownDocParser.
func NewMarkdownDocParser() *MarkdownDocParser {
	return &MarkdownDocParser{}
}

// Import reads a DDD.md file from docDir and returns an ImportResult with a populated DomainModel.
func (p *MarkdownDocParser) Import(_ context.Context, docDir string) (*importdomain.ImportResult, error) {
	path := filepath.Join(docDir, "DDD.md")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading DDD.md: %w", err)
	}

	content := string(data)
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("no bounded contexts found in DDD.md: file is empty")
	}

	model := ddd.NewDomainModel("imported")

	// Parse classification table (from ## 3. Subdomain Classification)
	classMap := p.parseClassificationTable(content)

	// Parse bounded contexts (### Context: Name or ### N. Name (Classification))
	contexts := p.parseBoundedContexts(content, classMap)
	if len(contexts) == 0 {
		return nil, fmt.Errorf("no bounded contexts found in DDD.md")
	}

	for _, bc := range contexts {
		if addErr := model.AddBoundedContext(bc); addErr != nil {
			return nil, fmt.Errorf("adding bounded context %q: %w", bc.Name(), addErr)
		}

		// Classify if we have classification info
		if bc.Classification() != nil {
			if classErr := model.ClassifySubdomain(bc.Name(), *bc.Classification(), bc.ClassificationRationale()); classErr != nil {
				return nil, fmt.Errorf("classifying %q: %w", bc.Name(), classErr)
			}
		}
	}

	// Parse context map relationships
	rels := p.parseContextRelationships(content)
	for _, rel := range rels {
		if addErr := model.AddContextRelationship(rel); addErr != nil {
			return nil, fmt.Errorf("adding relationship: %w", addErr)
		}
	}

	result, resultErr := importdomain.NewImportResult(model, nil)
	if resultErr != nil {
		return nil, fmt.Errorf("creating import result: %w", resultErr)
	}
	return result, nil
}

// contextHeadingRe matches "### Context: Name" style headings.
var contextHeadingRe = regexp.MustCompile(`^###\s+Context:\s+(.+)$`)

// numberedContextHeadingRe matches "### N. Name (Classification)" style headings.
var numberedContextHeadingRe = regexp.MustCompile(`^###\s+\d+\.\s+(.+?)(?:\s+\((\w+)\))?$`)

// responsibilityRe matches "**Responsibility:** text" lines.
var responsibilityRe = regexp.MustCompile(`^\*\*Responsibility:\*\*\s+(.+)$`)

// classificationTableRowRe matches "| Name | **Type** | Rationale |" rows.
var classificationTableRowRe = regexp.MustCompile(`^\|\s*(.+?)\s*\|\s*\*\*(\w+)\*\*\s*\|\s*(.+?)\s*\|`)

// contextMapRowRe matches "| Upstream | Downstream | Pattern |" rows.
var contextMapRowRe = regexp.MustCompile(`^\|\s*(.+?)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*\|$`)

func (p *MarkdownDocParser) parseClassificationTable(content string) map[string]classInfo {
	result := make(map[string]classInfo)
	lines := strings.Split(content, "\n")

	inClassSection := false
	inTable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## ") && strings.Contains(trimmed, "Subdomain Classification") {
			inClassSection = true
			continue
		}

		if inClassSection && strings.HasPrefix(trimmed, "## ") {
			break
		}

		if inClassSection && strings.HasPrefix(trimmed, "### Summary") {
			inTable = true
			continue
		}

		if !inTable {
			continue
		}

		// Skip table header and separator
		if strings.HasPrefix(trimmed, "| Subdomain") || strings.HasPrefix(trimmed, "|---") {
			continue
		}

		// End of table
		if trimmed == "" && inTable {
			// Tables can have blank lines between them; stop at next heading
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			break
		}

		matches := classificationTableRowRe.FindStringSubmatch(trimmed)
		if matches == nil {
			continue
		}

		name := strings.TrimSpace(matches[1])
		classStr := strings.ToLower(strings.TrimSpace(matches[2]))
		rationale := strings.TrimSpace(matches[3])

		classification, ok := parseClassification(classStr)
		if !ok {
			continue
		}

		result[name] = classInfo{
			classification: classification,
			rationale:      rationale,
		}
	}

	return result
}

type classInfo struct {
	classification vo.SubdomainClassification
	rationale      string
}

func (p *MarkdownDocParser) parseBoundedContexts(content string, classMap map[string]classInfo) []vo.DomainBoundedContext {
	lines := strings.Split(content, "\n")
	var contexts []vo.DomainBoundedContext

	var currentName string
	var currentResp string
	var currentClass *vo.SubdomainClassification
	var currentRationale string

	flushContext := func() {
		if currentName == "" {
			return
		}

		// Try to get classification from class table if not inline
		if currentClass == nil {
			if info, ok := classMap[currentName]; ok {
				currentClass = &info.classification
				currentRationale = info.rationale
			}
		}

		bc := vo.NewDomainBoundedContext(
			currentName,
			currentResp,
			nil, // key domain objects extracted separately if needed
			currentClass,
			currentRationale,
		)
		contexts = append(contexts, bc)

		currentName = ""
		currentResp = ""
		currentClass = nil
		currentRationale = ""
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for "### Context: Name" style
		if matches := contextHeadingRe.FindStringSubmatch(trimmed); matches != nil {
			flushContext()
			currentName = strings.TrimSpace(matches[1])
			continue
		}

		// Check for "### N. Name (Classification)" style
		if matches := numberedContextHeadingRe.FindStringSubmatch(trimmed); matches != nil {
			// Skip non-context headings like "### Summary", "### Classification Flow"
			name := strings.TrimSpace(matches[1])
			if isNonContextHeading(name) {
				continue
			}

			flushContext()
			currentName = name

			if len(matches) > 2 && matches[2] != "" {
				classStr := strings.ToLower(matches[2])
				if c, ok := parseClassification(classStr); ok {
					currentClass = &c
				}
			}
			continue
		}

		// Check for "### Context Map" — stop parsing contexts
		if strings.HasPrefix(trimmed, "### Context Map") {
			flushContext()
			break
		}

		// Stop at next ## heading (different section)
		if strings.HasPrefix(trimmed, "## ") && currentName != "" {
			flushContext()
			continue
		}

		// Extract responsibility
		if currentName != "" {
			if matches := responsibilityRe.FindStringSubmatch(trimmed); matches != nil {
				currentResp = strings.TrimSpace(matches[1])
			}
		}
	}

	flushContext()
	return contexts
}

func (p *MarkdownDocParser) parseContextRelationships(content string) []vo.ContextRelationship {
	lines := strings.Split(content, "\n")
	var rels []vo.ContextRelationship

	inContextMap := false
	inTable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "### Context Map") {
			inContextMap = true
			continue
		}

		if inContextMap && strings.HasPrefix(trimmed, "## ") {
			break
		}

		if !inContextMap {
			// Also check for context map table outside a ### heading
			if strings.HasPrefix(trimmed, "| Upstream Context") {
				inContextMap = true
				inTable = true
				continue
			}
			continue
		}

		if strings.HasPrefix(trimmed, "| Upstream") {
			inTable = true
			continue
		}

		if strings.HasPrefix(trimmed, "|---") {
			continue
		}

		if !inTable {
			continue
		}

		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "```") {
			if strings.HasPrefix(trimmed, "#") || (trimmed == "" && len(rels) > 0) {
				// Could be end of table or just spacing
				continue
			}
			continue
		}

		matches := contextMapRowRe.FindStringSubmatch(trimmed)
		if matches == nil {
			continue
		}

		upstream := strings.TrimSpace(matches[1])
		downstream := strings.TrimSpace(matches[2])
		pattern := strings.TrimSpace(matches[3])

		// Skip header-like rows
		if upstream == "Upstream Context" || upstream == "Upstream" {
			continue
		}

		rel := vo.NewContextRelationship(upstream, downstream, pattern)
		rels = append(rels, rel)
	}

	return rels
}

func parseClassification(s string) (vo.SubdomainClassification, bool) {
	switch s {
	case "core":
		return vo.SubdomainCore, true
	case "supporting":
		return vo.SubdomainSupporting, true
	case "generic":
		return vo.SubdomainGeneric, true
	default:
		return "", false
	}
}

// isNonContextHeading returns true for ### headings that aren't bounded contexts.
func isNonContextHeading(name string) bool {
	lower := strings.ToLower(name)
	nonContextPrefixes := []string{
		"summary",
		"classification",
		"complexity",
		"story",
	}
	for _, prefix := range nonContextPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}
