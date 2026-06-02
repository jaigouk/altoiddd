package rules

import (
	"fmt"
	"regexp"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// toolNameRegex matches a known native tool (PascalCase: `Read`,
// `WebSearch`) OR an MCP tool name of the form
// `mcp__<server>__<tool>` where server + tool are lowercase
// snake-style identifiers. RE2 — no backtracking.
var toolNameRegex = regexp.MustCompile(`^[A-Z][A-Za-z]+$|^mcp__[a-z0-9_]+__[a-z0-9_]+$`)

// UnknownToolsRule warns when `tools` contains an entry that
// does not match the native PascalCase form or the canonical MCP
// `mcp__<server>__<tool>` form. Catches author typos before the asset
// reaches a tool that would silently ignore the unknown name.
type UnknownToolsRule struct{}

// NewUnknownToolsRule constructs the rule.
func NewUnknownToolsRule() *UnknownToolsRule { return &UnknownToolsRule{} }

// Name returns the stable rule identifier.
func (r *UnknownToolsRule) Name() string { return "unknown_tools" }

// Check emits one WARNING per unrecognised tool name. Overlays carry no
// frontmatter and are exempt.
func (r *UnknownToolsRule) Check(asset dochealthdomain.ScaffoldAsset, _ []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation {
	if asset.IsOverlay() {
		return nil
	}
	var out []dochealthdomain.ScaffoldViolation
	for _, name := range toolsList(asset.Frontmatter()) {
		if !toolNameRegex.MatchString(name) {
			out = append(out, makeWarning(
				pathDisplay(asset.Path()),
				r.Name(),
				fmt.Sprintf("tool %q does not match %s", name, toolNameRegex.String()),
			))
		}
	}
	return out
}
