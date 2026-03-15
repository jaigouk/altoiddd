package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProjectConfig_WhenAllFieldsProvided_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	config := NewProjectConfig("my-service", "go", "github.com/user/my-service", []string{"claude", "cursor"})

	assert.Equal(t, "my-service", config.Name())
	assert.Equal(t, "go", config.Language())
	assert.Equal(t, "github.com/user/my-service", config.ModulePath())
	assert.Equal(t, []string{"claude", "cursor"}, config.DetectedTools())
}

func TestNewProjectConfig_WhenEmpty_ReturnsZeroValues(t *testing.T) {
	t.Parallel()

	config := NewProjectConfig("", "", "", nil)

	assert.Empty(t, config.Name())
	assert.Empty(t, config.Language())
	assert.Empty(t, config.ModulePath())
	assert.Empty(t, config.DetectedTools())
}

func TestProjectConfig_DetectedTools_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()

	tools := []string{"claude", "cursor"}
	config := NewProjectConfig("svc", "go", "mod", tools)

	// Mutate the returned slice
	got := config.DetectedTools()
	got[0] = "mutated"

	// Original must be unaffected
	assert.Equal(t, "claude", config.DetectedTools()[0])
}

func TestProjectConfig_DetectedTools_WhenNil_ReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	config := NewProjectConfig("svc", "go", "mod", nil)

	assert.NotNil(t, config.DetectedTools())
	assert.Empty(t, config.DetectedTools())
}
