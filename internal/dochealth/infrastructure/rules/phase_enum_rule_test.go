package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhaseEnumRule_InvalidPhase_ReturnsError(t *testing.T) {
	t.Parallel()
	fm := fullFrontmatter()
	fm["phase"] = "weird"
	asset := newTestAsset(t, "foo.md", fm, false)
	violations := NewPhaseEnumRule().Check(asset, nil)
	assert.Len(t, violations, 1)
}

func TestPhaseEnumRule_AllValidPhases_NoViolation(t *testing.T) {
	t.Parallel()
	for _, p := range []string{"design", "groom", "implement", "review", "close"} {
		t.Run("phase_"+p, func(t *testing.T) {
			t.Parallel()
			fm := fullFrontmatter()
			fm["phase"] = p
			asset := newTestAsset(t, "foo.md", fm, false)
			violations := NewPhaseEnumRule().Check(asset, nil)
			assert.Empty(t, violations)
		})
	}
}

func TestPhaseEnumRule_OverlayExempt(t *testing.T) {
	t.Parallel()
	asset := newTestAsset(t, "foo.project.md", map[string]any{}, true)
	violations := NewPhaseEnumRule().Check(asset, nil)
	assert.Empty(t, violations)
}

func TestPhaseEnumRule_MissingPhase_NoViolation(t *testing.T) {
	t.Parallel()
	// FrontmatterSchemaRule reports missing fields; PhaseEnumRule only
	// validates present values — keeps responsibilities single (SRP).
	fm := fullFrontmatter()
	delete(fm, "phase")
	asset := newTestAsset(t, "foo.md", fm, false)
	violations := NewPhaseEnumRule().Check(asset, nil)
	assert.Empty(t, violations)
}
