package rules

import (
	"fmt"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// validPhases enumerates accepted `phase:` values per the alty-cli-766.2
// spike lifecycle taxonomy.
var validPhases = map[string]struct{}{
	"design":    {},
	"groom":     {},
	"implement": {},
	"review":    {},
	"close":     {},
}

// PhaseEnumRule validates `phase:` ∈ {design, groom, implement, review, close}.
//
// Split from FrontmatterSchemaRule per ISP — individually testable and
// (future) individually suppressable when an asset legitimately needs a
// non-standard phase value.
type PhaseEnumRule struct{}

// NewPhaseEnumRule constructs the rule.
func NewPhaseEnumRule() *PhaseEnumRule { return &PhaseEnumRule{} }

// Name returns the stable rule identifier.
func (r *PhaseEnumRule) Name() string { return "PhaseEnumRule" }

// Check validates `phase`. Overlays are exempt (no frontmatter).
// Missing `phase` is reported by FrontmatterSchemaRule, not here — this
// rule only fires when `phase` IS present but invalid.
func (r *PhaseEnumRule) Check(asset dochealthdomain.ScaffoldAsset, _ []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation {
	if asset.IsOverlay() {
		return nil
	}
	phase, ok := stringValue(asset.Frontmatter(), "phase")
	if !ok {
		return nil
	}
	if _, valid := validPhases[phase]; valid {
		return nil
	}
	return []dochealthdomain.ScaffoldViolation{
		makeViolation(
			pathDisplay(asset.Path()),
			r.Name(),
			fmt.Sprintf("phase %q not in {design, groom, implement, review, close}", phase),
		),
	}
}
