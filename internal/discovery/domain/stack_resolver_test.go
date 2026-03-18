package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

func TestResolveProfilePythonUv(t *testing.T) {
	t.Parallel()
	ts := vo.NewTechStack("python", "uv")
	profile := ResolveProfile(&ts)
	_, ok := profile.(vo.PythonUvProfile)
	assert.True(t, ok)
}

func TestResolveProfilePythonPip(t *testing.T) {
	t.Parallel()
	ts := vo.NewTechStack("python", "pip")
	profile := ResolveProfile(&ts)
	_, ok := profile.(vo.PythonUvProfile)
	assert.True(t, ok)
}

func TestResolveProfilePythonEmptyManager(t *testing.T) {
	t.Parallel()
	ts := vo.NewTechStack("python", "")
	profile := ResolveProfile(&ts)
	_, ok := profile.(vo.PythonUvProfile)
	assert.True(t, ok)
}

func TestResolveProfileUnknownLanguage(t *testing.T) {
	t.Parallel()
	ts := vo.NewTechStack("unknown", "")
	profile := ResolveProfile(&ts)
	_, ok := profile.(vo.GenericProfile)
	assert.True(t, ok)
}

func TestResolveProfileRust(t *testing.T) {
	t.Parallel()
	ts := vo.NewTechStack("rust", "cargo")
	profile := ResolveProfile(&ts)
	_, ok := profile.(vo.GenericProfile)
	assert.True(t, ok)
}

func TestResolveProfileJavaScript(t *testing.T) {
	t.Parallel()
	ts := vo.NewTechStack("javascript", "npm")
	profile := ResolveProfile(&ts)
	_, ok := profile.(vo.GenericProfile)
	assert.True(t, ok)
}

func TestResolveProfileNilReturnsGeneric(t *testing.T) {
	t.Parallel()
	profile := ResolveProfile(nil)
	_, ok := profile.(vo.GenericProfile)
	assert.True(t, ok)
}

func TestResolveProfilePythonSatisfiesProtocol(t *testing.T) {
	t.Parallel()
	ts := vo.NewTechStack("python", "uv")
	profile := ResolveProfile(&ts)
	assert.NotEmpty(t, profile.StackID())
}

func TestResolveProfileGenericSatisfiesProtocol(t *testing.T) {
	t.Parallel()
	profile := ResolveProfile(nil)
	assert.NotEmpty(t, profile.StackID())
}
