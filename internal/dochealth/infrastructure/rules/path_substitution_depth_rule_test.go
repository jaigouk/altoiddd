package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

func TestPathSubstitutionDepthRule_Name(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "path_substitution_depth", NewPathSubstitutionDepthRule().Name())
}

func TestPathSubstitutionDepthRule_TwoSegments_NoViolation(t *testing.T) {
	t.Parallel()
	body := "ref ${CLAUDE_SKILL_DIR}/../templates/foo.md"
	assert.Empty(t, NewPathSubstitutionDepthRule().Check(bodyAsset(t, body, false), nil))
}

func TestPathSubstitutionDepthRule_OneSegment_NoViolation(t *testing.T) {
	t.Parallel()
	body := "ref ${CLAUDE_SKILL_DIR}/templates/foo.md"
	assert.Empty(t, NewPathSubstitutionDepthRule().Check(bodyAsset(t, body, false), nil))
}

func TestPathSubstitutionDepthRule_TwoSegments_AtLimit_NoViolation(t *testing.T) {
	t.Parallel()
	body := "ref ${CLAUDE_SKILL_DIR}/../../etc/passwd"
	assert.Empty(t, NewPathSubstitutionDepthRule().Check(bodyAsset(t, body, false), nil),
		"two segments is the documented maximum — no violation")
}

func TestPathSubstitutionDepthRule_ThreeSegments_Error(t *testing.T) {
	t.Parallel()
	body := "ref ${CLAUDE_SKILL_DIR}/../../../etc/passwd"
	v := NewPathSubstitutionDepthRule().Check(bodyAsset(t, body, false), nil)
	if assert.Len(t, v, 1) {
		assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
	}
}

func TestPathSubstitutionDepthRule_FourSegments_Error(t *testing.T) {
	t.Parallel()
	body := "ref ${CLAUDE_SKILL_DIR}/../../../../etc/passwd"
	v := NewPathSubstitutionDepthRule().Check(bodyAsset(t, body, false), nil)
	assert.NotEmpty(t, v)
}

func TestPathSubstitutionDepthRule_NoSubstitution_NoViolation(t *testing.T) {
	t.Parallel()
	body := "../../something/else.md (literal markdown, not a substitution)"
	assert.Empty(t, NewPathSubstitutionDepthRule().Check(bodyAsset(t, body, false), nil))
}

func TestPathSubstitutionDepthRule_OverlayExempt(t *testing.T) {
	t.Parallel()
	body := "ref ${CLAUDE_SKILL_DIR}/../../../escape"
	assert.Empty(t, NewPathSubstitutionDepthRule().Check(bodyAsset(t, body, true), nil))
}
