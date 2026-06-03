// Package markdown provides primitive markdown helpers shared by every
// bounded context. The frontmatter splitter is intentionally schema-agnostic:
// callers own the typed unmarshal into their domain-specific struct.
//
// arch-go permits importing this package from any bounded context because
// internal/shared/infrastructure/... carries no domain semantics and is
// implementation reused across contexts (CQRS-lite shared kernel).
package markdown

import (
	"errors"
	"strings"
)

// ErrMissingClosingDelimiter signals that the input opens with `---` but
// never carries a `\n---` closing line. Consumers decide whether to treat
// this as fatal (tooltranslation) or as "no frontmatter" (challenge,
// dochealth) by inspecting the returned error.
var ErrMissingClosingDelimiter = errors.New("missing closing frontmatter delimiter")

const delimiter = "---"

// ExtractFrontmatter splits a markdown document into its raw frontmatter
// (without delimiters) and body.
//
// Contract:
//   - Empty input -> (raw="", body="", hasFrontmatter=false, err=nil)
//   - Input without a leading "---" -> (raw="", body=content, hasFrontmatter=false, err=nil)
//   - "---\n<yaml>\n---\n<body>" -> (raw="<yaml>" TrimSpaced, body="<body>" with leading "\n" stripped, hasFrontmatter=true, err=nil)
//   - "---" with no closing "\n---" -> (raw="", body=content, hasFrontmatter=false, err=ErrMissingClosingDelimiter)
//   - "---\n---\n<body>" (empty frontmatter) -> (raw="", body="<body>", hasFrontmatter=true, err=nil)
//
// The split is implemented with strings.Index over a stripped-prefix slice
// — no regex, no allocation beyond the returned strings.
func ExtractFrontmatter(content string) (raw string, body string, hasFrontmatter bool, err error) {
	if content == "" {
		return "", "", false, nil
	}
	if !strings.HasPrefix(content, delimiter) {
		return "", content, false, nil
	}

	rest := content[len(delimiter):]
	idx := strings.Index(rest, "\n"+delimiter)
	if idx == -1 {
		return "", content, false, ErrMissingClosingDelimiter
	}

	raw = strings.TrimSpace(rest[:idx])
	afterClosing := rest[idx+len("\n"+delimiter):]
	body = strings.TrimPrefix(afterClosing, "\n")
	return raw, body, true, nil
}
