package infrastructure

import (
	dochealthapp "github.com/alto-cli/alto/internal/dochealth/application"
	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
	"github.com/alto-cli/alto/internal/dochealth/infrastructure/rules"
)

// DefaultScaffoldRules returns the canonical scaffold-health rule
// registry. Slot order is LOCKED by the alty-cli-ihk Phase 1 broadcast
// (Contract 5) — DO NOT reorder without a fresh tech-lead lock.
//
// Signature accepts ScaffoldParams so rule factories that need
// configuration (LifecycleStalenessRule, SecretsGrepRule) can be wired
// from a single composition-root call. Open/Closed-preserved — adding a
// new rule means appending to this slice; the handler stays untouched.
//
// Comment-block grouping below reflects SEVERITY (not slot order). The
// locked Contract-5 slot for SecretsGrepRule is 8 (mid-block), but its
// severity is WARNING — so the comment groups it with the WARNING rules
// rather than mis-label it as ERROR (QA-MIN-1 fix, Round 1).
func DefaultScaffoldRules(params dochealthdomain.ScaffoldParams) []dochealthapp.ValidationRule {
	return []dochealthapp.ValidationRule{
		// ERROR rules — existing 4 (alty-cli-766.7). DO NOT REORDER.
		rules.NewFrontmatterSchemaRule(),
		rules.NewPhaseEnumRule(),
		rules.NewNoInternalLeaksRule(),
		rules.NewOrphanOverlayRule(),

		// ERROR rules — new 3 (alty-cli-ihk). Append-only order.
		rules.NewBashSubstitutionPolicyRule(),
		rules.NewBashArgumentsRule(),
		rules.NewPathSubstitutionDepthRule(),

		// WARNING rules — new 5 (alty-cli-ihk). Append-only order.
		// Slot 8 (SecretsGrepRule) is locked by Contract 5; severity is
		// WARNING per Contract 7. Grouping comment matches severity.
		rules.NewSecretsGrepRule(params.SecretPatterns()),
		rules.NewBodySizeRule(),
		rules.NewUnknownToolsRule(),
		rules.NewBashWithParametersWarnRule(),
		rules.NewLifecycleStalenessRule(params.DefaultStalenessDays()),
	}
}
