package rules

import (
	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// BashArgumentsRule rejects ANY unquoted shell-variable substitution
// inside a bash block (inline, fenced ` ```! `, or standard ` ```bash `
// fence). This includes argument-specific patterns (`$ARGUMENTS`,
// `$ARGUMENTS[N]`, `$N`) AND environment variables (`$HOME`, `$PATH`,
// `${USER}`, etc.) — both bare and brace forms. Reuses
// `extractBashBlocks` + `hasUnquotedSubstitution` from rules.go so we
// don't duplicate the parse.
//
// This is a SUPERSET of the `quoted`-policy check from
// BashSubstitutionPolicyRule — it fires regardless of declared policy,
// because the substitution is a hard security boundary (per spike L827-
// 831). Overlap with the `quoted` policy is intentional defense in depth.
//
// Overlays are exempt — they have no frontmatter declaring intent and
// inherit the parent's safety contract.
type BashArgumentsRule struct{}

// NewBashArgumentsRule constructs the rule.
func NewBashArgumentsRule() *BashArgumentsRule { return &BashArgumentsRule{} }

// Name returns the stable rule identifier.
func (r *BashArgumentsRule) Name() string { return "bash_arguments" }

// Check returns one ERROR per bash block containing an unquoted
// substitution. Mitigation hint per ticket AC.
func (r *BashArgumentsRule) Check(asset dochealthdomain.ScaffoldAsset, _ []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation {
	if asset.IsOverlay() {
		return nil
	}
	const mitigation = "wrap argument substitutions in double quotes inside the bash block, or move the bash call outside the block and use a separate validation step"
	var out []dochealthdomain.ScaffoldViolation
	for _, block := range extractBashBlocks(asset.Body()) {
		if hasUnquotedSubstitution(block) {
			// QA-MIN-4 (Round 1) — message must reflect that the rule fires
			// on env vars too (`$HOME`, `$PATH`) and on brace forms
			// (`${ARGUMENTS}`, `${VAR}`), not only argument-named patterns.
			out = append(out, makeViolation(
				pathDisplay(asset.Path()),
				r.Name(),
				"unquoted variable substitution ($ARGUMENTS, $VAR, $N, or ${...} form) inside bash block — "+mitigation,
			))
		}
	}
	return out
}
