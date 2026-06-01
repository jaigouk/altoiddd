package rules

import (
	"fmt"
	"regexp"
	"strings"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// claudeSkillDirSubRegex matches `${CLAUDE_SKILL_DIR}/<rest-of-path>`
// references. We then count `..` segments in the captured tail to enforce
// the 2-segment limit (documented maximum: `${CLAUDE_SKILL_DIR}/../
// templates/`). RE2 — no backtracking.
var claudeSkillDirSubRegex = regexp.MustCompile(`\$\{CLAUDE_SKILL_DIR\}((?:/[^\s)]*)?)`)

// maxParentSegments is the binding limit on `..` segments inside a
// `${CLAUDE_SKILL_DIR}/` substitution. More than this is treated as an
// author-trust escape attempt per spike L832.
const maxParentSegments = 2

// PathSubstitutionDepthRule rejects path substitutions with more than
// `maxParentSegments` `..` segments — an escape attempt out of the
// canonical scaffold root.
type PathSubstitutionDepthRule struct{}

// NewPathSubstitutionDepthRule constructs the rule.
func NewPathSubstitutionDepthRule() *PathSubstitutionDepthRule {
	return &PathSubstitutionDepthRule{}
}

// Name returns the stable rule identifier.
func (r *PathSubstitutionDepthRule) Name() string { return "path_substitution_depth" }

// Check returns one ERROR per offending substitution. Overlays are exempt.
func (r *PathSubstitutionDepthRule) Check(asset dochealthdomain.ScaffoldAsset, _ []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation {
	if asset.IsOverlay() {
		return nil
	}
	var out []dochealthdomain.ScaffoldViolation
	for _, m := range claudeSkillDirSubRegex.FindAllStringSubmatch(asset.Body(), -1) {
		tail := m[1] // includes leading `/` if present
		// Count `..` path segments.
		dots := 0
		for _, seg := range strings.Split(tail, "/") {
			if seg == ".." {
				dots++
			}
		}
		if dots > maxParentSegments {
			out = append(out, makeViolation(
				pathDisplay(asset.Path()),
				r.Name(),
				fmt.Sprintf("${CLAUDE_SKILL_DIR}%s has %d `..` segments (max %d)", tail, dots, maxParentSegments),
			))
		}
	}
	return out
}
