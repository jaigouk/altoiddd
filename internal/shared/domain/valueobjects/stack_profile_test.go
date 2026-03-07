package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Protocol compliance
// ---------------------------------------------------------------------------

func TestPythonUvProfileSatisfiesInterface(t *testing.T) {
	t.Parallel()
	var _ vo.StackProfile = vo.PythonUvProfile{}
}

func TestGenericProfileSatisfiesInterface(t *testing.T) {
	t.Parallel()
	var _ vo.StackProfile = vo.GenericProfile{}
}

// ---------------------------------------------------------------------------
// PythonUvProfile
// ---------------------------------------------------------------------------

func TestPythonUvProfileStackID(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "python-uv", vo.PythonUvProfile{}.StackID())
}

func TestPythonUvProfileFileGlob(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "**/*.py", vo.PythonUvProfile{}.FileGlob())
}

func TestPythonUvProfileProjectManifest(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "pyproject.toml", vo.PythonUvProfile{}.ProjectManifest())
}

func TestPythonUvProfileSourceLayout(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{
		"src/domain/",
		"src/application/",
		"src/infrastructure/",
	}, vo.PythonUvProfile{}.SourceLayout())
}

func TestPythonUvProfileQualityGateCommandsLint(t *testing.T) {
	t.Parallel()
	cmds := vo.PythonUvProfile{}.QualityGateCommands()
	assert.Equal(t, []string{"uv", "run", "ruff", "check", "."}, cmds[vo.QualityGateLint])
}

func TestPythonUvProfileQualityGateCommandsTypes(t *testing.T) {
	t.Parallel()
	cmds := vo.PythonUvProfile{}.QualityGateCommands()
	assert.Equal(t, []string{"uv", "run", "mypy", "."}, cmds[vo.QualityGateTypes])
}

func TestPythonUvProfileQualityGateCommandsTests(t *testing.T) {
	t.Parallel()
	cmds := vo.PythonUvProfile{}.QualityGateCommands()
	assert.Equal(t, []string{"uv", "run", "pytest"}, cmds[vo.QualityGateTests])
}

func TestPythonUvProfileQualityGateCommandsFitness(t *testing.T) {
	t.Parallel()
	cmds := vo.PythonUvProfile{}.QualityGateCommands()
	assert.Equal(t, []string{"uv", "run", "pytest", "tests/architecture/"}, cmds[vo.QualityGateFitness])
}

func TestPythonUvProfileQualityGateCommandsCoversAllGates(t *testing.T) {
	t.Parallel()
	cmds := vo.PythonUvProfile{}.QualityGateCommands()
	allGates := vo.AllQualityGates()
	for _, gate := range allGates {
		_, ok := cmds[gate]
		require.True(t, ok, "missing command for gate %s", gate)
	}
}

func TestPythonUvProfileQualityGateDisplay(t *testing.T) {
	t.Parallel()
	display := vo.PythonUvProfile{}.QualityGateDisplay()
	assert.Contains(t, display, "uv run ruff check .")
	assert.Contains(t, display, "uv run mypy .")
	assert.Contains(t, display, "uv run pytest")
	assert.True(t, len(display) > 0 && display[:len("## Quality Gates")] == "## Quality Gates")
}

func TestPythonUvProfileFitnessAvailable(t *testing.T) {
	t.Parallel()
	assert.True(t, vo.PythonUvProfile{}.FitnessAvailable())
}

func TestPythonUvProfileToRootPackage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"hyphenated", "my-app", "my_app"},
		{"multi hyphen", "my-cool-app", "my_cool_app"},
		{"already underscored", "my_app", "my_app"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, vo.PythonUvProfile{}.ToRootPackage(tt.input))
		})
	}
}

// ---------------------------------------------------------------------------
// GenericProfile
// ---------------------------------------------------------------------------

func TestGenericProfileStackID(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "generic", vo.GenericProfile{}.StackID())
}

func TestGenericProfileFileGlob(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "*", vo.GenericProfile{}.FileGlob())
}

func TestGenericProfileProjectManifest(t *testing.T) {
	t.Parallel()
	assert.Empty(t, vo.GenericProfile{}.ProjectManifest())
}

func TestGenericProfileSourceLayout(t *testing.T) {
	t.Parallel()
	assert.Nil(t, vo.GenericProfile{}.SourceLayout())
}

func TestGenericProfileQualityGateCommandsEmpty(t *testing.T) {
	t.Parallel()
	assert.Nil(t, vo.GenericProfile{}.QualityGateCommands())
}

func TestGenericProfileQualityGateDisplayEmpty(t *testing.T) {
	t.Parallel()
	assert.Empty(t, vo.GenericProfile{}.QualityGateDisplay())
}

func TestGenericProfileFitnessNotAvailable(t *testing.T) {
	t.Parallel()
	assert.False(t, vo.GenericProfile{}.FitnessAvailable())
}

func TestGenericProfileToRootPackage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"passthrough hyphenated", "my-app", "my-app"},
		{"passthrough multi hyphen", "my-cool-app", "my-cool-app"},
		{"passthrough underscored", "my_app", "my_app"},
		{"passthrough empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, vo.GenericProfile{}.ToRootPackage(tt.input))
		})
	}
}
