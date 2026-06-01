package rules

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// lifecycleInProgressMarker is the path segment that scopes this rule.
// Only assets under `lifecycle/in-progress/` are checked for staleness —
// assets in `commands/` and `agents/` have no staleness contract.
const lifecycleInProgressMarker = "lifecycle/in-progress/"

// LifecycleStalenessRule warns when an asset under
// `alto-scaffold/lifecycle/in-progress/` has not been modified for more than
// `thresholdDays` days. Threshold injected via ScaffoldParams — never
// hardcoded.
//
// Rule uses ScaffoldAsset.ModTime() (captured by the walker) so the rule
// itself remains pure (no I/O — DDD: infrastructure concerns live in the
// walker, not the rule).
type LifecycleStalenessRule struct {
	thresholdDays int
	now           func() time.Time
}

// NewLifecycleStalenessRule constructs the rule with the configured
// staleness threshold. thresholdDays must be >= 1 (validated upstream by
// NewScaffoldParams); values <= 0 are treated as "disabled".
func NewLifecycleStalenessRule(thresholdDays int) *LifecycleStalenessRule {
	return &LifecycleStalenessRule{thresholdDays: thresholdDays, now: time.Now}
}

// Name returns the stable rule identifier.
func (r *LifecycleStalenessRule) Name() string { return "lifecycle_staleness" }

// Check emits one WARNING when the asset is under lifecycle/in-progress/
// AND its mtime is older than thresholdDays. A separate WARNING fires
// when mtime is in the future (clock skew or filesystem tamper — the
// author should know rather than silently absorb).
//
// WH-MED-3 (Round 1) added the future-mtime branch: the previous "skip
// silently" behavior hid clock-drift and tamper from operators.
func (r *LifecycleStalenessRule) Check(asset dochealthdomain.ScaffoldAsset, _ []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation {
	if r.thresholdDays <= 0 {
		return nil
	}
	if !strings.Contains(filepath.ToSlash(asset.Path()), lifecycleInProgressMarker) {
		return nil
	}
	mod := asset.ModTime()
	if mod.IsZero() {
		// Walker didn't capture mtime — skip rather than spuriously flag.
		return nil
	}
	age := r.now().Sub(mod)
	if age < 0 {
		return []dochealthdomain.ScaffoldViolation{
			makeWarning(
				pathDisplay(asset.Path()),
				r.Name(),
				"asset mtime is in the future — clock skew or filesystem tamper",
			),
		}
	}
	threshold := time.Duration(r.thresholdDays) * 24 * time.Hour
	if age <= threshold {
		return nil
	}
	days := int(age.Hours() / 24)
	return []dochealthdomain.ScaffoldViolation{
		makeWarning(
			pathDisplay(asset.Path()),
			r.Name(),
			fmt.Sprintf("in-progress asset is %d days stale (threshold %d) — promote or deprecate", days, r.thresholdDays),
		),
	}
}
