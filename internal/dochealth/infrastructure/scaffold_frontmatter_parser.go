package infrastructure

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alto-cli/alto/internal/shared/infrastructure/markdown"
)

// scaffoldFrontmatterParser extracts YAML frontmatter from a scaffold
// asset. Delegates the split + generic unmarshal to the shared markdown
// kernel (alty-cli-1r0); the (raw, body, lineCount, hasFrontmatter, err)
// shape is dochealth-specific and stays here.
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
	rawFM, body, hasFrontmatter, err := markdown.ExtractFrontmatter(content)
	if err != nil {
		if errors.Is(err, markdown.ErrMissingClosingDelimiter) {
			// Treat unclosed frontmatter as no frontmatter; the walker
			// surfaces schema violations elsewhere.
			return map[string]any{}, content, lineCount(content), false, nil
		}
		return nil, "", 0, false, fmt.Errorf("extracting frontmatter: %w", err)
	}
	if !hasFrontmatter {
		return map[string]any{}, content, lineCount(content), false, nil
	}

	fm, err := markdown.ParseGeneric(rawFM)
	if err != nil {
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
