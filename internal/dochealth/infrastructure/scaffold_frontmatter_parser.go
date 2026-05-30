package infrastructure

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// scaffoldFrontmatterParser extracts YAML frontmatter from a scaffold
// asset. Duplicates a small subset of internal/challenge/infrastructure/
// yaml_frontmatter_parser.go — arch-go blocks cross-context imports; the
// shared-kernel refactor stub (filed at 766.7 close) hoists both into
// internal/shared/infrastructure/markdown/.
type scaffoldFrontmatterParser struct{}

func newScaffoldFrontmatterParser() *scaffoldFrontmatterParser {
	return &scaffoldFrontmatterParser{}
}

// Parse returns (frontmatter map, body, bodyLineCount, hasFrontmatter, err).
//
// `hasFrontmatter` is false for files without a leading `---` delimiter
// (e.g. .project.md overlays, which are pure body content by design). In
// that case the frontmatter map is empty and body equals the full input.
//
// Unclosed frontmatter is treated as "no frontmatter" rather than an error
// — defensive parsing keeps the walker resilient to author typos; the
// FrontmatterSchemaRule will then flag the missing required fields.
func (p *scaffoldFrontmatterParser) Parse(content string) (map[string]any, string, int, bool, error) {
	if !strings.HasPrefix(content, "---") {
		return map[string]any{}, content, lineCount(content), false, nil
	}
	rest := content[3:]
	end, ok := strings.CutPrefix(rest, "\n")
	if !ok {
		end = rest // tolerate `---\n...` without leading newline
	}
	closeIdx := strings.Index(end, "\n---")
	if closeIdx == -1 {
		// Unclosed — treat as no frontmatter.
		return map[string]any{}, content, lineCount(content), false, nil
	}
	rawFM := strings.TrimSpace(end[:closeIdx])
	body := strings.TrimPrefix(end[closeIdx+len("\n---"):], "\n")

	fm := map[string]any{}
	if err := yaml.Unmarshal([]byte(rawFM), &fm); err != nil {
		return nil, "", 0, false, fmt.Errorf("yaml unmarshal: %w", err)
	}
	return fm, body, lineCount(body), true, nil
}

// lineCount returns the number of `\n`-separated lines in s. Empty input
// counts as 0 lines (rather than 1) so empty bodies are reported as zero.
func lineCount(s string) int {
	if s == "" {
		return 0
	}
	n := strings.Count(s, "\n")
	if !strings.HasSuffix(s, "\n") {
		n++
	}
	return n
}
