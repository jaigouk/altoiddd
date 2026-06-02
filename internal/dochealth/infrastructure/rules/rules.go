// Package rules — ValidationRule implementations for ScaffoldHealthHandler.
//
// Each rule lives in its own file (ISP — testable + suppressable
// individually). Rules are stateless and deterministic; the handler
// composes them via DefaultScaffoldRules in
// internal/dochealth/infrastructure/default_scaffold_rules.go.
//
// Ubiquitous-language: every rule's Name() returns a stable string used
// in violation reports and (future) per-rule suppression configuration.
package rules

import (
	"path/filepath"
	"regexp"
	"strings"

	dochealthapp "github.com/alto-cli/alto/internal/dochealth/application"
	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// makeViolation panics if construction fails — every input here is
// controlled by the rule itself, never by attacker input, so a panic
// signals a programmer error (the rule built an invalid violation).
func makeViolation(file, ruleName, message string) dochealthdomain.ScaffoldViolation {
	v, err := dochealthdomain.NewScaffoldViolation(file, ruleName, message, dochealthdomain.SeverityError, 0)
	if err != nil {
		panic("rule built invalid violation: " + err.Error())
	}
	return v
}

// makeWarning is the WARNING-severity sibling of makeViolation. Same panic
// contract — rule-controlled inputs only.
func makeWarning(file, ruleName, message string) dochealthdomain.ScaffoldViolation {
	v, err := dochealthdomain.NewScaffoldViolation(file, ruleName, message, dochealthdomain.SeverityWarning, 0)
	if err != nil {
		panic("rule built invalid warning: " + err.Error())
	}
	return v
}

// pathDisplay returns a repo-relative slash-normalised path for display.
func pathDisplay(p string) string { return filepath.ToSlash(p) }

// Compile-time interface checks for every concrete rule.
var (
	_ dochealthapp.ValidationRule = (*FrontmatterSchemaRule)(nil)
	_ dochealthapp.ValidationRule = (*PhaseEnumRule)(nil)
	_ dochealthapp.ValidationRule = (*NoInternalLeaksRule)(nil)
	_ dochealthapp.ValidationRule = (*OrphanOverlayRule)(nil)
	_ dochealthapp.ValidationRule = (*BodySizeRule)(nil)
	_ dochealthapp.ValidationRule = (*UnknownToolsRule)(nil)
	_ dochealthapp.ValidationRule = (*SecretsGrepRule)(nil)
	_ dochealthapp.ValidationRule = (*LifecycleStalenessRule)(nil)
	_ dochealthapp.ValidationRule = (*BashSubstitutionPolicyRule)(nil)
	_ dochealthapp.ValidationRule = (*BashArgumentsRule)(nil)
	_ dochealthapp.ValidationRule = (*BashWithParametersWarnRule)(nil)
	_ dochealthapp.ValidationRule = (*PathSubstitutionDepthRule)(nil)
)

// bashBlocks extracts every bash-execution block from a body. Returns:
//   - inline blocks `!` cmd “ (inner text between the backticks) — Claude Code form
//   - fenced blocks ` ```! ... ``` ` (inner text) — Claude Code form
//   - standard Markdown shell fences ` ```bash ... ``` `, ` ```sh `,
//     ` ```zsh `, ` ```shell `, ` ```console ` (case-insensitive lang tag)
//
// Used by BashSubstitutionPolicyRule and BashArgumentsRule. RE2 — no
// catastrophic backtracking. Both regexes are anchored to begin/end markers
// to avoid greedy over-capture across markdown paragraphs.
//
// WH-HIGH-1 (Round 1) added the standard-shell-fence form: scaffold authors
// who copy bash from external docs naturally use ` ```bash ` rather than
// the Claude-specific ` ```! `; the rules MUST inspect both.
var (
	inlineBashBlockRegex  = regexp.MustCompile("!`([^`]*)`")
	fencedBashBlockRegex  = regexp.MustCompile("(?s)```!\\s*\\n(.*?)\\n```")
	fencedShellBlockRegex = regexp.MustCompile("(?s)```(?i:bash|sh|zsh|shell|console)\\s*\\n(.*?)\\n```")
)

// extractBashBlocks returns the inner text of every bash block in body.
// Inline `!`x“, fenced ```` ```! ... ``` ````, and standard fenced
// ```` ```bash ... ``` ```` (or sh/zsh/shell/console) are returned in a
// single slice; callers don't care which form produced each string.
func extractBashBlocks(body string) []string {
	var out []string
	for _, m := range inlineBashBlockRegex.FindAllStringSubmatch(body, -1) {
		out = append(out, m[1])
	}
	for _, m := range fencedBashBlockRegex.FindAllStringSubmatch(body, -1) {
		out = append(out, m[1])
	}
	for _, m := range fencedShellBlockRegex.FindAllStringSubmatch(body, -1) {
		out = append(out, m[1])
	}
	return out
}

// hasAnyBashBlock reports whether body contains at least one bash block.
func hasAnyBashBlock(body string) bool {
	return inlineBashBlockRegex.MatchString(body) ||
		fencedBashBlockRegex.MatchString(body) ||
		fencedShellBlockRegex.MatchString(body)
}

// toolsList parses the `tools` frontmatter value (string or
// []any) into a normalised []string of trimmed entries.
func toolsList(fm map[string]any) []string {
	raw, ok := fm["tools"]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case string:
		var out []string
		for _, p := range strings.Split(v, ",") {
			if t := strings.TrimSpace(p); t != "" {
				out = append(out, t)
			}
		}
		return out
	case []any:
		var out []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				if t := strings.TrimSpace(s); t != "" {
					out = append(out, t)
				}
			}
		}
		return out
	default:
		return nil
	}
}

// boolValue returns the typed boolean for key, or (false, false) when the
// field is missing or not a bool.
func boolValue(fm map[string]any, key string) (bool, bool) {
	raw, ok := fm[key]
	if !ok {
		return false, false
	}
	b, ok := raw.(bool)
	if !ok {
		return false, false
	}
	return b, true
}

// stringValue extracts a non-empty string from a frontmatter map. Returns
// (value, true) when key exists, value is a string AND non-empty. Empty
// strings are treated as missing — frontmatter rules want "present and
// non-empty" semantics across the board.
func stringValue(fm map[string]any, key string) (string, bool) {
	raw, ok := fm[key]
	if !ok {
		return "", false
	}
	s, ok := raw.(string)
	if !ok {
		return "", false
	}
	if strings.TrimSpace(s) == "" {
		return "", false
	}
	return s, true
}

// fieldPresent reports whether a frontmatter field is canonically present
// and non-empty, accepting BOTH the inline-CSV string form AND the YAML
// block-list ([]any) form. Used by FrontmatterSchemaRule's presence-loop
// for polymorphic fields whose downstream parser accepts both shapes
// (today: `tools` only, via toolsList).
//
// Returns true when:
//   - value is a non-empty (post-TrimSpace) string, OR
//   - value is a non-empty []any
//
// All other shapes (missing key, nil, empty string, empty list, bool, int,
// map) return false. This is intentionally narrower than "key exists" —
// frontmatter rules want "present AND non-empty" semantics.
//
// Polymorphic carve-out alty-cli-tzw: scope is `tools` today. If a future
// field needs the same treatment, add it here, do NOT broaden the helper
// to accept all shapes.
func fieldPresent(fm map[string]any, key string) bool {
	raw, ok := fm[key]
	if !ok || raw == nil {
		return false
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v) != ""
	case []any:
		return len(v) > 0
	default:
		return false
	}
}
