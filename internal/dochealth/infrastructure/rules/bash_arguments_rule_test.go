package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

func bodyAsset(t *testing.T, body string, isOverlay bool) dochealthdomain.ScaffoldAsset {
	t.Helper()
	a, err := dochealthdomain.NewScaffoldAsset("foo.md", fullFrontmatter(), body, 1, isOverlay)
	require.NoError(t, err)
	return a
}

func TestBashArgumentsRule_Name(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "bash_arguments", NewBashArgumentsRule().Name())
}

func TestBashArgumentsRule_InlineBareArguments_Error(t *testing.T) {
	t.Parallel()
	v := NewBashArgumentsRule().Check(bodyAsset(t, "run !`cmd $ARGUMENTS`", false), nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
	// QA-MIN-3 — tightened from `assert.True(Contains() || Contains())`.
	// Mitigation hint always includes both "wrap" and "quote" — the rule's
	// canonical hint reads "wrap argument substitutions in double quotes …".
	assert.Contains(t, v[0].Message(), "wrap")
	assert.Contains(t, v[0].Message(), "quote")
	// QA-MIN-5 / Fix 5 — message must NOT mislead about scope. Rule fires
	// on env vars too, so the message should not name only $ARGUMENTS/$N.
	assert.Contains(t, v[0].Message(), "unquoted")
}

func TestBashArgumentsRule_InlineBracketedArguments_Error(t *testing.T) {
	t.Parallel()
	v := NewBashArgumentsRule().Check(bodyAsset(t, "run !`cmd $ARGUMENTS[0]`", false), nil)
	require.NotEmpty(t, v)
}

func TestBashArgumentsRule_InlinePositional_Error(t *testing.T) {
	t.Parallel()
	v := NewBashArgumentsRule().Check(bodyAsset(t, "run !`cmd $1`", false), nil)
	require.NotEmpty(t, v)
}

func TestBashArgumentsRule_InlineNamed_Error(t *testing.T) {
	t.Parallel()
	v := NewBashArgumentsRule().Check(bodyAsset(t, "run !`cmd $myvar`", false), nil)
	require.NotEmpty(t, v)
}

func TestBashArgumentsRule_InlineQuoted_NoViolation(t *testing.T) {
	t.Parallel()
	cases := []string{
		"run !`cmd \"$ARGUMENTS\"`",
		"run !`cmd \"$1\"`",
		"run !`cmd \"$myvar\"`",
	}
	rule := NewBashArgumentsRule()
	for _, body := range cases {
		assert.Empty(t, rule.Check(bodyAsset(t, body, false), nil), "body %q should be clean", body)
	}
}

func TestBashArgumentsRule_FencedBareArguments_Error(t *testing.T) {
	t.Parallel()
	body := "```!\ncmd $ARGUMENTS\n```"
	v := NewBashArgumentsRule().Check(bodyAsset(t, body, false), nil)
	require.NotEmpty(t, v)
	assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
}

func TestBashArgumentsRule_FencedQuoted_NoViolation(t *testing.T) {
	t.Parallel()
	body := "```!\ncmd \"$ARGUMENTS\"\n```"
	assert.Empty(t, NewBashArgumentsRule().Check(bodyAsset(t, body, false), nil))
}

func TestBashArgumentsRule_OverlayExempt(t *testing.T) {
	t.Parallel()
	body := "!`cmd $ARGUMENTS`"
	assert.Empty(t, NewBashArgumentsRule().Check(bodyAsset(t, body, true), nil))
}

func TestBashArgumentsRule_OutsideBashBlock_NoViolation(t *testing.T) {
	t.Parallel()
	// Prose `$ARGUMENTS` outside a bash block is allowed (it documents the
	// variable rather than executes it).
	body := "the `$ARGUMENTS` token expands at runtime"
	assert.Empty(t, NewBashArgumentsRule().Check(bodyAsset(t, body, false), nil))
}

// --- Round 1 fix-cycle additions ---

// WH-HIGH-1 — standard Markdown fenced shell blocks must trigger.

func TestBashArgumentsRule_StandardBashFence_BareArguments_Error(t *testing.T) {
	t.Parallel()
	body := "```bash\necho $ARGUMENTS\n```"
	v := NewBashArgumentsRule().Check(bodyAsset(t, body, false), nil)
	require.NotEmpty(t, v)
	assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
}

func TestBashArgumentsRule_ShFence_BareArguments_Error(t *testing.T) {
	t.Parallel()
	body := "```sh\necho $ARGUMENTS\n```"
	v := NewBashArgumentsRule().Check(bodyAsset(t, body, false), nil)
	require.NotEmpty(t, v)
}

func TestBashArgumentsRule_BashFenceQuoted_NoViolation(t *testing.T) {
	t.Parallel()
	body := "```bash\necho \"$ARGUMENTS\"\n```"
	assert.Empty(t, NewBashArgumentsRule().Check(bodyAsset(t, body, false), nil))
}

func TestBashArgumentsRule_BashFenceSingleQuoted_NoViolation(t *testing.T) {
	t.Parallel()
	// Single-quoted blocks substitution → literal $ARGUMENTS, safe.
	body := "```bash\necho '$ARGUMENTS'\n```"
	assert.Empty(t, NewBashArgumentsRule().Check(bodyAsset(t, body, false), nil))
}

func TestBashArgumentsRule_PythonFence_BareArguments_NoViolation(t *testing.T) {
	t.Parallel()
	// Rule is bash-scoped — python fence is not a bash block.
	body := "```python\necho $ARGUMENTS\n```"
	assert.Empty(t, NewBashArgumentsRule().Check(bodyAsset(t, body, false), nil))
}

// QA-SEC-DEFER-2 — brace-form substitutions.

func TestBashArgumentsRule_BraceArguments_Error(t *testing.T) {
	t.Parallel()
	body := "```!\necho ${ARGUMENTS}\n```"
	v := NewBashArgumentsRule().Check(bodyAsset(t, body, false), nil)
	require.NotEmpty(t, v)
}

func TestBashArgumentsRule_BraceArgumentsBracketed_Error(t *testing.T) {
	t.Parallel()
	body := "```!\necho ${ARGUMENTS[3]}\n```"
	v := NewBashArgumentsRule().Check(bodyAsset(t, body, false), nil)
	require.NotEmpty(t, v)
}

// QA-SEC-DEFER-1 — gap-between-quotes parity check.

func TestBashArgumentsRule_GapBetweenQuotes_Error(t *testing.T) {
	t.Parallel()
	body := "```!\ncmd \"x\" $ARGS \"y\"\n```"
	v := NewBashArgumentsRule().Check(bodyAsset(t, body, false), nil)
	require.NotEmpty(t, v)
}
