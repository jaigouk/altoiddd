package rules

import (
	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// BashWithParametersWarnRule warns when an asset declares both `Bash` in
// `tools` AND `parameters:` AND `disable_model_invocation` is
// not `true`. This catches authors who forgot the normative rule from
// §Frontmatter Schema: parameter-accepting commands that can invoke Bash
// must opt out of model invocation OR explicitly accept the trust
// boundary.
//
// WARNING rather than ERROR — soft path during initial migration per
// spike L825.
type BashWithParametersWarnRule struct{}

// NewBashWithParametersWarnRule constructs the rule.
func NewBashWithParametersWarnRule() *BashWithParametersWarnRule {
	return &BashWithParametersWarnRule{}
}

// Name returns the stable rule identifier.
func (r *BashWithParametersWarnRule) Name() string { return "bash_with_parameters_warn" }

// Check emits one WARNING when the three conditions all hold.
func (r *BashWithParametersWarnRule) Check(asset dochealthdomain.ScaffoldAsset, _ []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation {
	if asset.IsOverlay() {
		return nil
	}
	fm := asset.Frontmatter()
	tools := toolsList(fm)
	hasBash := false
	for _, t := range tools {
		if t == "Bash" {
			hasBash = true
			break
		}
	}
	if !hasBash {
		return nil
	}
	if _, hasParams := fm["parameters"]; !hasParams {
		return nil
	}
	if protected, ok := boolValue(fm, "disable_model_invocation"); ok && protected {
		return nil
	}
	return []dochealthdomain.ScaffoldViolation{
		makeWarning(
			pathDisplay(asset.Path()),
			r.Name(),
			"tools contains Bash AND parameters are declared but disable_model_invocation is not true — model could invoke Bash with attacker-controlled parameters",
		),
	}
}
