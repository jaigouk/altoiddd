package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

func fullFrontmatter() map[string]any {
	return map[string]any{
		"name":                     "foo",
		"description":              "x",
		"kind":                     "command",
		"phase":                    "groom",
		"when_to_use":              "test",
		"tools":                    "Read",
		"bash_substitution_policy": "none",
		"license":                  "Apache-2.0",
	}
}

func newTestAsset(t *testing.T, path string, fm map[string]any, isOverlay bool) dochealthdomain.ScaffoldAsset {
	t.Helper()
	a, err := dochealthdomain.NewScaffoldAsset(path, fm, "", 0, isOverlay)
	require.NoError(t, err)
	return a
}

func TestFrontmatterSchemaRule_AllEightFieldsRequired(t *testing.T) {
	t.Parallel()
	cases := []string{
		"name",
		"description",
		"kind",
		"phase",
		"when_to_use",
		"tools",
		"bash_substitution_policy",
		"license",
	}
	rule := NewFrontmatterSchemaRule()
	for _, field := range cases {
		t.Run("missing_"+field, func(t *testing.T) {
			t.Parallel()
			fm := fullFrontmatter()
			delete(fm, field)
			asset := newTestAsset(t, "foo.md", fm, false)
			violations := rule.Check(asset, nil)
			require.NotEmpty(t, violations, "missing %s must produce a violation", field)
			found := false
			for _, v := range violations {
				if strings.Contains(v.Message(), field) {
					found = true
					break
				}
			}
			assert.True(t, found, "violation message must name the missing field %q", field)
		})
	}
}

func TestFrontmatterSchemaRule_AllFieldsPresent_NoViolation(t *testing.T) {
	t.Parallel()
	asset := newTestAsset(t, "foo.md", fullFrontmatter(), false)
	violations := NewFrontmatterSchemaRule().Check(asset, nil)
	assert.Empty(t, violations)
}

func TestFrontmatterSchemaRule_OverlayExempt(t *testing.T) {
	t.Parallel()
	// Overlay with empty frontmatter — must NOT trigger.
	asset := newTestAsset(t, "foo.project.md", map[string]any{}, true)
	violations := NewFrontmatterSchemaRule().Check(asset, nil)
	assert.Empty(t, violations)
}

func TestFrontmatterSchemaRule_InvalidNameRegex_ReturnsError(t *testing.T) {
	t.Parallel()
	fm := fullFrontmatter()
	fm["name"] = "BadName"
	asset := newTestAsset(t, "foo.md", fm, false)
	violations := NewFrontmatterSchemaRule().Check(asset, nil)
	require.NotEmpty(t, violations)
	found := false
	for _, v := range violations {
		if strings.Contains(v.Message(), "does not match") {
			found = true
			break
		}
	}
	assert.True(t, found, "violation must mention name regex")
}

func TestFrontmatterSchemaRule_InvalidKindEnum_ReturnsError(t *testing.T) {
	t.Parallel()
	fm := fullFrontmatter()
	fm["kind"] = "weird"
	asset := newTestAsset(t, "foo.md", fm, false)
	violations := NewFrontmatterSchemaRule().Check(asset, nil)
	require.NotEmpty(t, violations)
	found := false
	for _, v := range violations {
		if strings.Contains(v.Message(), "kind") {
			found = true
			break
		}
	}
	assert.True(t, found)
}

// hasMissingToolsViolation reports whether any violation message claims the
// `tools` field is missing. Used by the polymorphic-presence tests below.
func hasMissingToolsViolation(violations []dochealthdomain.ScaffoldViolation) bool {
	for _, v := range violations {
		msg := v.Message()
		if strings.Contains(msg, "missing") && strings.Contains(msg, `"tools"`) {
			return true
		}
	}
	return false
}

// TestFrontmatterSchemaRule_ToolsAsYAMLList_PassesPresenceCheck — bug repro
// for alty-cli-tzw. A YAML block list parses to []any; the schema gate must
// treat it as present, not as missing. See toolsList in rules.go for the
// canonical polymorphic semantics this test pins.
func TestFrontmatterSchemaRule_ToolsAsYAMLList_PassesPresenceCheck(t *testing.T) {
	t.Parallel()
	fm := fullFrontmatter()
	fm["tools"] = []any{"Read", "Write"}
	asset := newTestAsset(t, "foo.md", fm, false)
	violations := NewFrontmatterSchemaRule().Check(asset, nil)
	assert.False(t, hasMissingToolsViolation(violations),
		"YAML block-list form of `tools` must pass the presence check; got: %v", violations)
}

// TestFrontmatterSchemaRule_ToolsAsInlineCSV_PassesPresenceCheck — regression
// guard for the form that already works. Ensures the polymorphic carve-out
// doesn't break the string-CSV path.
func TestFrontmatterSchemaRule_ToolsAsInlineCSV_PassesPresenceCheck(t *testing.T) {
	t.Parallel()
	fm := fullFrontmatter()
	fm["tools"] = "Read, Write"
	asset := newTestAsset(t, "foo.md", fm, false)
	violations := NewFrontmatterSchemaRule().Check(asset, nil)
	assert.False(t, hasMissingToolsViolation(violations),
		"inline-CSV form of `tools` must pass the presence check; got: %v", violations)
}

// TestFrontmatterSchemaRule_ToolsAsEmptyList_FailsAsMissing — an empty list
// is semantically the same as no tools; the presence check must reject it.
func TestFrontmatterSchemaRule_ToolsAsEmptyList_FailsAsMissing(t *testing.T) {
	t.Parallel()
	fm := fullFrontmatter()
	fm["tools"] = []any{}
	asset := newTestAsset(t, "foo.md", fm, false)
	violations := NewFrontmatterSchemaRule().Check(asset, nil)
	assert.True(t, hasMissingToolsViolation(violations),
		"empty list for `tools` must be treated as missing; got: %v", violations)
}

// TestFrontmatterSchemaRule_ToolsAsEmptyString_FailsAsMissing — already
// covered for other fields via stringValue's empty-string semantics, but
// the polymorphic carve-out for `tools` must preserve that behaviour.
func TestFrontmatterSchemaRule_ToolsAsEmptyString_FailsAsMissing(t *testing.T) {
	t.Parallel()
	fm := fullFrontmatter()
	fm["tools"] = ""
	asset := newTestAsset(t, "foo.md", fm, false)
	violations := NewFrontmatterSchemaRule().Check(asset, nil)
	assert.True(t, hasMissingToolsViolation(violations),
		"empty string for `tools` must be treated as missing; got: %v", violations)
}

// TestFrontmatterSchemaRule_ToolsMissingKey_FailsAsMissing — regression guard
// for the basic case. Already covered by the table-driven
// TestFrontmatterSchemaRule_AllEightFieldsRequired, but enumerated here so
// the polymorphic tests form a self-contained suite.
func TestFrontmatterSchemaRule_ToolsMissingKey_FailsAsMissing(t *testing.T) {
	t.Parallel()
	fm := fullFrontmatter()
	delete(fm, "tools")
	asset := newTestAsset(t, "foo.md", fm, false)
	violations := NewFrontmatterSchemaRule().Check(asset, nil)
	assert.True(t, hasMissingToolsViolation(violations),
		"absent `tools` key must be treated as missing; got: %v", violations)
}

// TestFrontmatterSchemaRule_NonToolsFields_StillStringOnly — guard that the
// polymorphic carve-out applies ONLY to `tools`. `name`, `kind`, `phase`,
// etc. are string-typed by design and must reject non-string values just
// as they did before the fix.
func TestFrontmatterSchemaRule_NonToolsFields_StillStringOnly(t *testing.T) {
	t.Parallel()
	cases := []struct {
		field string
		value any
	}{
		{"name", 42},                      // int
		{"description", []any{"a"}},       // list
		{"kind", []any{"command"}},        // list
		{"phase", true},                   // bool
		{"when_to_use", map[string]any{}}, // map
		{"bash_substitution_policy", 1.5}, // float
		{"license", []any{"Apache-2.0"}},  // list
	}
	for _, tc := range cases {
		t.Run(tc.field, func(t *testing.T) {
			t.Parallel()
			fm := fullFrontmatter()
			fm[tc.field] = tc.value
			asset := newTestAsset(t, "foo.md", fm, false)
			violations := NewFrontmatterSchemaRule().Check(asset, nil)
			require.NotEmpty(t, violations, "non-string %s must produce a violation", tc.field)
			found := false
			for _, v := range violations {
				msg := v.Message()
				if strings.Contains(msg, "missing") && strings.Contains(msg, tc.field) {
					found = true
					break
				}
			}
			assert.True(t, found,
				"non-string %s must be flagged as missing (polymorphic carve-out is `tools`-only); got: %v",
				tc.field, violations)
		})
	}
}
