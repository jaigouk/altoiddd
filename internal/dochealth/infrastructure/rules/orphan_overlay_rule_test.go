package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

func TestOrphanOverlayRule_OverlayWithoutSibling_ReturnsError(t *testing.T) {
	t.Parallel()
	overlay := newBodyAsset(t, "commands/orphan.project.md", "x", true)
	corpus := []dochealthdomain.ScaffoldAsset{overlay}
	violations := NewOrphanOverlayRule().Check(overlay, corpus)
	assert.Len(t, violations, 1)
}

func TestOrphanOverlayRule_OverlayWithSibling_NoViolation(t *testing.T) {
	t.Parallel()
	primary := newBodyAsset(t, "commands/foo.md", "x", false)
	overlay := newBodyAsset(t, "commands/foo.project.md", "x", true)
	corpus := []dochealthdomain.ScaffoldAsset{primary, overlay}
	violations := NewOrphanOverlayRule().Check(overlay, corpus)
	assert.Empty(t, violations)
}

func TestOrphanOverlayRule_CorpusEmpty_NoViolation(t *testing.T) {
	t.Parallel()
	// An empty corpus contains no assets at all — rule fires on overlay
	// assets only, so this is vacuously clean.
	violations := NewOrphanOverlayRule().Check(
		newBodyAsset(t, "commands/foo.md", "x", false),
		nil,
	)
	assert.Empty(t, violations)
}

func TestOrphanOverlayRule_NonOverlayAsset_NoViolation(t *testing.T) {
	t.Parallel()
	primary := newBodyAsset(t, "commands/foo.md", "x", false)
	violations := NewOrphanOverlayRule().Check(primary, []dochealthdomain.ScaffoldAsset{primary})
	assert.Empty(t, violations, "rule must not fire on non-overlay assets")
}
