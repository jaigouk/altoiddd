package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

func bashParamsAsset(t *testing.T, fmExtras map[string]any) dochealthdomain.ScaffoldAsset {
	t.Helper()
	fm := fullFrontmatter()
	for k, v := range fmExtras {
		fm[k] = v
	}
	a, err := dochealthdomain.NewScaffoldAsset("foo.md", fm, "", 0, false)
	require.NoError(t, err)
	return a
}

func TestBashWithParametersWarnRule_Name(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "bash_with_parameters_warn", NewBashWithParametersWarnRule().Name())
}

func TestBashWithParametersWarnRule_BashWithParamsNoProtection_ReturnsWarning(t *testing.T) {
	t.Parallel()
	a := bashParamsAsset(t, map[string]any{
		"tools":      "Read, Bash",
		"parameters": []any{"x"},
	})
	v := NewBashWithParametersWarnRule().Check(a, nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityWarning, v[0].Severity())
}

func TestBashWithParametersWarnRule_BashWithParamsProtected_NoViolation(t *testing.T) {
	t.Parallel()
	a := bashParamsAsset(t, map[string]any{
		"tools":                    "Read, Bash",
		"parameters":               []any{"x"},
		"disable_model_invocation": true,
	})
	assert.Empty(t, NewBashWithParametersWarnRule().Check(a, nil))
}

func TestBashWithParametersWarnRule_NoBash_NoViolation(t *testing.T) {
	t.Parallel()
	a := bashParamsAsset(t, map[string]any{
		"tools":      "Read, Write",
		"parameters": []any{"x"},
	})
	assert.Empty(t, NewBashWithParametersWarnRule().Check(a, nil))
}

func TestBashWithParametersWarnRule_NoParameters_NoViolation(t *testing.T) {
	t.Parallel()
	a := bashParamsAsset(t, map[string]any{
		"tools": "Read, Bash",
	})
	assert.Empty(t, NewBashWithParametersWarnRule().Check(a, nil))
}

func TestBashWithParametersWarnRule_OverlayExempt(t *testing.T) {
	t.Parallel()
	a, err := dochealthdomain.NewScaffoldAsset("foo.project.md", nil, "", 0, true)
	require.NoError(t, err)
	assert.Empty(t, NewBashWithParametersWarnRule().Check(a, nil))
}
