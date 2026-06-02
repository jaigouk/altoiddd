package rules

import (
	"fmt"
	"regexp"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// canonicalSchemaFields enumerates the 8 required frontmatter fields per
// the alty-cli-766.2 spike (lines 224-244). Order matters only for stable
// table-driven test output.
var canonicalSchemaFields = []string{
	"name",
	"description",
	"kind",
	"phase",
	"when_to_use",
	"tools",
	"bash_substitution_policy",
	"license",
}

// validKinds enumerates accepted `kind:` values.
var validKinds = map[string]struct{}{
	"command":  {},
	"agent":    {},
	"template": {},
	"skill":    {},
}

// nameRegexFrontmatter validates `name:`. Linear (RE2) — no ReDoS surface.
var nameRegexFrontmatter = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// FrontmatterSchemaRule enforces the canonical 8-field schema on GENERIC
// scaffold assets. Overlays (.project.md) are exempt — they carry pure
// body content by design and inherit the parent's frontmatter.
type FrontmatterSchemaRule struct{}

// NewFrontmatterSchemaRule constructs the rule.
func NewFrontmatterSchemaRule() *FrontmatterSchemaRule { return &FrontmatterSchemaRule{} }

// Name returns the stable rule identifier.
func (r *FrontmatterSchemaRule) Name() string { return "FrontmatterSchemaRule" }

// Check enforces:
//   - All 8 canonical fields present + non-empty
//   - `name` matches ^[a-z][a-z0-9-]*$
//   - `kind` ∈ {command, agent, template, skill}
//
// Overlays are skipped (their parent .md carries the schema).
func (r *FrontmatterSchemaRule) Check(asset dochealthdomain.ScaffoldAsset, _ []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation {
	if asset.IsOverlay() {
		return nil
	}
	fm := asset.Frontmatter()
	var out []dochealthdomain.ScaffoldViolation
	for _, field := range canonicalSchemaFields {
		// `tools` is canonically polymorphic: inline-CSV string OR YAML
		// block-list ([]any). Use fieldPresent so the schema gate and the
		// downstream toolsList parser agree on which shapes are valid.
		// All other canonical fields are string-typed by spec — keep them
		// on stringValue. (alty-cli-tzw)
		present := false
		if field == "tools" {
			present = fieldPresent(fm, field)
		} else {
			_, present = stringValue(fm, field)
		}
		if !present {
			out = append(out, makeViolation(
				pathDisplay(asset.Path()),
				r.Name(),
				fmt.Sprintf("missing required frontmatter field %q", field),
			))
		}
	}
	if name, ok := stringValue(fm, "name"); ok && !nameRegexFrontmatter.MatchString(name) {
		out = append(out, makeViolation(
			pathDisplay(asset.Path()),
			r.Name(),
			fmt.Sprintf("name %q does not match %s", name, nameRegexFrontmatter.String()),
		))
	}
	if kind, ok := stringValue(fm, "kind"); ok {
		if _, valid := validKinds[kind]; !valid {
			out = append(out, makeViolation(
				pathDisplay(asset.Path()),
				r.Name(),
				fmt.Sprintf("kind %q not in {command, agent, template, skill}", kind),
			))
		}
	}
	return out
}
