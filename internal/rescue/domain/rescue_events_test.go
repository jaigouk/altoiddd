package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alty-cli/alty/internal/rescue/domain"
)

func TestGapAnalysisCompleted(t *testing.T) {
	t.Parallel()

	t.Run("fields", func(t *testing.T) {
		t.Parallel()
		e := domain.NewGapAnalysisCompleted("abc-123", "/tmp/proj", 5, 3)
		assert.Equal(t, "abc-123", e.AnalysisID())
		assert.Equal(t, "/tmp/proj", e.ProjectDir())
		assert.Equal(t, 5, e.GapsFound())
		assert.Equal(t, 3, e.GapsResolved())
	})

	t.Run("equality", func(t *testing.T) {
		t.Parallel()
		e1 := domain.NewGapAnalysisCompleted("abc", "/tmp/proj", 1, 1)
		e2 := domain.NewGapAnalysisCompleted("abc", "/tmp/proj", 1, 1)
		assert.Equal(t, e1, e2)
	})

	t.Run("inequality", func(t *testing.T) {
		t.Parallel()
		e1 := domain.NewGapAnalysisCompleted("abc", "/tmp/proj", 1, 1)
		e2 := domain.NewGapAnalysisCompleted("def", "/tmp/proj", 1, 1)
		assert.NotEqual(t, e1, e2)
	})
}
