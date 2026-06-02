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
		"tools":           "Read",
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
