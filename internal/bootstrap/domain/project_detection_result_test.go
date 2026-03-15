package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProjectDetectionResult_WhenAllFieldsProvided_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	result := NewProjectDetectionResult(true, "go", true, true, true, "go.mod")

	assert.True(t, result.HasSourceCode())
	assert.Equal(t, "go", result.Language())
	assert.True(t, result.HasDocsFolder())
	assert.True(t, result.HasAltyConfig())
	assert.True(t, result.HasAIToolConfig())
	assert.Equal(t, "go.mod", result.ManifestPath())
}

func TestNewProjectDetectionResult_WhenEmpty_ReturnsZeroValues(t *testing.T) {
	t.Parallel()

	result := NewProjectDetectionResult(false, "", false, false, false, "")

	assert.False(t, result.HasSourceCode())
	assert.Empty(t, result.Language())
	assert.False(t, result.HasDocsFolder())
	assert.False(t, result.HasAltyConfig())
	assert.False(t, result.HasAIToolConfig())
	assert.Empty(t, result.ManifestPath())
}

func TestProjectDetectionResult_IsExistingProject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		hasSourceCode bool
		hasDocsFolder bool
		want          bool
	}{
		{"source code present", true, false, true},
		{"docs folder present", false, true, true},
		{"both present", true, true, true},
		{"neither present", false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := NewProjectDetectionResult(tt.hasSourceCode, "", tt.hasDocsFolder, false, false, "")
			assert.Equal(t, tt.want, result.IsExistingProject())
		})
	}
}

func TestProjectDetectionResult_IsAmbiguous(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		hasSourceCode bool
		hasDocsFolder bool
		want          bool
	}{
		{"docs but no source", false, true, true},
		{"source and docs", true, true, false},
		{"source only", true, false, false},
		{"neither", false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := NewProjectDetectionResult(tt.hasSourceCode, "", tt.hasDocsFolder, false, false, "")
			assert.Equal(t, tt.want, result.IsAmbiguous())
		})
	}
}
