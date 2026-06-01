package rules

import (
	"fmt"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// bodySizeBudget is the soft cap on a single asset's body length in lines.
// Above the budget the rule emits a WARNING — scaffold authors should split
// large commands/agents into smaller composable units (DDD: keep the bounded
// context for each asset small enough to read in one sitting).
const bodySizeBudget = 500

// BodySizeRule warns when an asset's body exceeds bodySizeBudget lines.
// Both GENERIC and OVERLAY assets are evaluated — defense in depth, since
// the merged body is the relevant size for downstream consumers.
type BodySizeRule struct{}

// NewBodySizeRule constructs the rule.
func NewBodySizeRule() *BodySizeRule { return &BodySizeRule{} }

// Name returns the stable rule identifier.
func (r *BodySizeRule) Name() string { return "body_size" }

// Check returns a single WARNING violation when bodyLineCount > budget.
func (r *BodySizeRule) Check(asset dochealthdomain.ScaffoldAsset, _ []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation {
	if asset.BodyLineCount() <= bodySizeBudget {
		return nil
	}
	return []dochealthdomain.ScaffoldViolation{
		makeWarning(
			pathDisplay(asset.Path()),
			r.Name(),
			fmt.Sprintf("body has %d lines (budget %d) — consider splitting", asset.BodyLineCount(), bodySizeBudget),
		),
	}
}
