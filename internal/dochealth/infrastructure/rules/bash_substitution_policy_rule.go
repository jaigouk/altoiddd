package rules

import (
	"fmt"
	"regexp"
	"strings"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// Recognised bash_substitution_policy values per spike §Frontmatter Schema.
const (
	policyNone         = "none"
	policyQuoted       = "quoted"
	policyUnrestricted = "unrestricted"
)

// unquotedSubstitutionRegex matches bare AND brace forms of shell
// substitutions: `$VAR`, `$ARGUMENTS`, `$ARGUMENTS[N]`, `$N`, `${VAR}`,
// `${ARGUMENTS}`, `${ARGUMENTS[N]}`, `${N}`. We then inspect quote
// state around each match via hasUnquotedSubstitution. RE2 — no
// backtracking.
//
// QA-SEC-DEFER-2 (Round 1) added the brace form: `${ARGUMENTS}` has the
// SAME shell semantics as `$ARGUMENTS` and was previously bypassable.
//
// Known limitation: ASCII `$` only. Full-width Unicode dollar (U+FF04 ＄)
// is NOT detected. Authors using IME or office-suite paste may
// accidentally introduce ＄ARGUMENTS, which renders as `$ARGUMENTS`
// visually but does not substitute under bash and is not flagged here
// (WH-LOW-2 — accepted as a documented gap).
var unquotedSubstitutionRegex = regexp.MustCompile(
	`\$(?:\{(?:ARGUMENTS(?:\[[0-9]+\])?|[0-9]+|[A-Za-z_][A-Za-z0-9_]*)(?:[^}]*)?\}` +
		`|(?:ARGUMENTS(?:\[[0-9]+\])?|[0-9]+|[A-Za-z_][A-Za-z0-9_]*))`,
)

// BashSubstitutionPolicyRule enforces the 3-mode `bash_substitution_policy`
// frontmatter directive:
//
//   - `none`         → any bash block in body is an ERROR
//   - `quoted`       → unquoted substitutions inside bash blocks are ERROR
//   - `unrestricted` → bash blocks emit a WARNING (escape hatch — flagged
//     for review, but allowed)
//
// Overlays are exempt (they carry no frontmatter, so no policy is
// declared at their level — the primary GENERIC sibling owns the rule).
type BashSubstitutionPolicyRule struct{}

// NewBashSubstitutionPolicyRule constructs the rule.
func NewBashSubstitutionPolicyRule() *BashSubstitutionPolicyRule {
	return &BashSubstitutionPolicyRule{}
}

// Name returns the stable rule identifier.
func (r *BashSubstitutionPolicyRule) Name() string { return "bash_substitution_policy" }

// Check applies the policy semantics above to asset.Body().
func (r *BashSubstitutionPolicyRule) Check(asset dochealthdomain.ScaffoldAsset, _ []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation {
	if asset.IsOverlay() {
		return nil
	}
	policy, ok := stringValue(asset.Frontmatter(), "bash_substitution_policy")
	if !ok {
		// FrontmatterSchemaRule reports missing field; we no-op here.
		return nil
	}
	body := asset.Body()
	switch policy {
	case policyNone:
		if hasAnyBashBlock(body) {
			return []dochealthdomain.ScaffoldViolation{
				makeViolation(
					pathDisplay(asset.Path()),
					r.Name(),
					"bash_substitution_policy=none but body contains a bash block",
				),
			}
		}
	case policyQuoted:
		for _, block := range extractBashBlocks(body) {
			if hasUnquotedSubstitution(block) {
				return []dochealthdomain.ScaffoldViolation{
					makeViolation(
						pathDisplay(asset.Path()),
						r.Name(),
						"bash_substitution_policy=quoted but bash block contains an unquoted substitution",
					),
				}
			}
		}
	case policyUnrestricted:
		if hasAnyBashBlock(body) {
			return []dochealthdomain.ScaffoldViolation{
				makeWarning(
					pathDisplay(asset.Path()),
					r.Name(),
					"bash_substitution_policy=unrestricted (escape hatch — review for safety)",
				),
			}
		}
	default:
		// Unknown policy value — surface as ERROR so authors notice typos.
		return []dochealthdomain.ScaffoldViolation{
			makeViolation(
				pathDisplay(asset.Path()),
				r.Name(),
				fmt.Sprintf("unknown bash_substitution_policy %q (expected none|quoted|unrestricted)", policy),
			),
		}
	}
	return nil
}

// hasUnquotedSubstitution reports whether `block` contains a `$…`
// substitution that is NOT inside a single quoted string.
//
// Two-stage quote evaluation (per QA-SEC-DEFER-1 + WH-MED-1, Round 1):
//
//  1. Single-quotes short-circuit. Bash semantics: `'...'` blocks ALL
//     substitution. A substitution that sits inside an OPEN single-quote
//     pair on the same line is literal — skip it (not flagged). Detection
//     uses parity of `'` counts to the left (within the line/block region):
//     odd = inside an open `'...'`, even = outside.
//
//  2. Double-quote parity. For substitutions NOT short-circuited by stage
//     1: count `"` characters before the substitution within the same
//     line. ODD parity means we're inside a `"..."` pair (the substitution
//     IS quoted — safe). EVEN parity (including zero) means we sit between
//     two distinct quoted strings or in pure unquoted shell text — flag.
//
// The previous adjacent-quote check (any `"` to the left AND any `"` to
// the right) was the QA-SEC-DEFER-1 bug: it treated
// `cmd "first" $ARG "second"` as quoted because both halves had a
// neighboring `"`, but bash sees `$ARG` in the unquoted gap.
//
// Block region: line-scoped to keep semantics tractable. Multi-line
// quoted strings (rare in bash one-liners) would defeat this — accepted
// as an open question for a future fast-follow if real bodies surface it.
func hasUnquotedSubstitution(block string) bool {
	matches := unquotedSubstitutionRegex.FindAllStringIndex(block, -1)
	for _, m := range matches {
		start := m[0]
		// Scan the prefix of the current line up to the match.
		lineStart := strings.LastIndexByte(block[:start], '\n') + 1
		prefix := block[lineStart:start]

		// Stage 1: single-quote short-circuit.
		// If odd number of `'` before match, we're inside an open single-
		// quote pair — substitution is literal, not flagged.
		if strings.Count(prefix, "'")%2 == 1 {
			continue
		}

		// Stage 2: double-quote parity. Odd count of `"` before match
		// means we're INSIDE a double-quoted string — quoted, safe.
		if strings.Count(prefix, `"`)%2 == 1 {
			continue
		}

		// Neither single- nor double-quoted → unquoted substitution.
		return true
	}
	return false
}
