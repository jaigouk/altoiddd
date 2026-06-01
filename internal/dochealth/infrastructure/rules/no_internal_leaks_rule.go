package rules

import (
	"fmt"
	"regexp"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// internalLeaksRegex matches the alto-internal references the canonical
// scaffold must NOT contain in GENERIC files (so the scaffold can be
// adopted by downstream projects without leaking alto's internals).
//
// Word boundaries on `internal/`, `alto-cli`, `alty-cli`, `cmd/alto`,
// `Watermill`, `golangci`. The `/` after `internal` prevents the prose
// word "internally" from triggering the rule — verified by
// TestNoInternalLeaksRule_ProseInternallyNoSlash_NoFalsePositive. The
// `alto-cli` form (not bare `alto-`) is required so user-facing references
// to the `alto-scaffold/` directory in GENERIC bodies do not false-positive.
//
// RE2 (no backtracking) — no ReDoS surface even on adversarial input.
var internalLeaksRegex = regexp.MustCompile(`\binternal/|\balto-cli\b|\balty-cli\b|\bcmd/alto\b|\bWatermill\b|\bgolangci`)

// NoInternalLeaksRule rejects alto-internal references in GENERIC asset
// bodies. Overlays carry alto-internal content by design and are exempt.
type NoInternalLeaksRule struct{}

// NewNoInternalLeaksRule constructs the rule.
func NewNoInternalLeaksRule() *NoInternalLeaksRule { return &NoInternalLeaksRule{} }

// Name returns the stable rule identifier.
func (r *NoInternalLeaksRule) Name() string { return "NoInternalLeaksRule" }

// Check scans the asset body for the fitness regex. Frontmatter is NOT
// scanned (separate concern; the schema rule covers it).
//
// One violation per asset (not one per match) — the report stays readable
// even when an author drops multiple internal refs. The first match's
// matched substring is included in the message for actionable feedback.
func (r *NoInternalLeaksRule) Check(asset dochealthdomain.ScaffoldAsset, _ []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation {
	if asset.IsOverlay() {
		return nil
	}
	match := internalLeaksRegex.FindString(asset.Body())
	if match == "" {
		return nil
	}
	return []dochealthdomain.ScaffoldViolation{
		makeViolation(
			pathDisplay(asset.Path()),
			r.Name(),
			fmt.Sprintf("alto-internal reference %q in GENERIC body (move to .project.md overlay)", match),
		),
	}
}
