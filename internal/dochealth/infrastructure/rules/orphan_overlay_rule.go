package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// OrphanOverlayRule rejects `<name>.project.md` files whose `<name>.md`
// primary sibling is absent. This is a cross-asset check — the only rule
// in this ticket that consumes the `corpus` parameter.
type OrphanOverlayRule struct{}

// NewOrphanOverlayRule constructs the rule.
func NewOrphanOverlayRule() *OrphanOverlayRule { return &OrphanOverlayRule{} }

// Name returns the stable rule identifier.
func (r *OrphanOverlayRule) Name() string { return "OrphanOverlayRule" }

// Check fires only when `asset` is an overlay with no matching primary in
// `corpus`. Returning empty for non-overlay assets means the handler can
// invoke this rule on every asset without double-counting (each orphan
// produces exactly one violation, on its own iteration).
func (r *OrphanOverlayRule) Check(asset dochealthdomain.ScaffoldAsset, corpus []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation {
	if !asset.IsOverlay() {
		return nil
	}
	primaryPath := primaryPathOf(asset.Path())
	for _, candidate := range corpus {
		if !candidate.IsOverlay() && filepath.ToSlash(candidate.Path()) == primaryPath {
			return nil
		}
	}
	return []dochealthdomain.ScaffoldViolation{
		makeViolation(
			pathDisplay(asset.Path()),
			r.Name(),
			fmt.Sprintf("overlay has no primary at %s", primaryPath),
		),
	}
}

// primaryPathOf returns the `.md` sibling path for a given `.project.md`
// overlay path. Idempotent on paths that don't end in `.project.md` (so
// callers don't need to pre-check).
func primaryPathOf(overlayPath string) string {
	p := filepath.ToSlash(overlayPath)
	if !strings.HasSuffix(p, ".project.md") {
		return p
	}
	return strings.TrimSuffix(p, ".project.md") + ".md"
}
