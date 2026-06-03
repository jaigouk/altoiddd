package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

func policyAsset(t *testing.T, policy, body string) dochealthdomain.ScaffoldAsset {
	t.Helper()
	fm := fullFrontmatter()
	fm["bash_substitution_policy"] = policy
	a, err := dochealthdomain.NewScaffoldAsset("foo.md", fm, body, 1, false)
	require.NoError(t, err)
	return a
}

func TestBashSubstitutionPolicyRule_Name(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "bash_substitution_policy", NewBashSubstitutionPolicyRule().Name())
}

func TestBashSubstitutionPolicyRule_NonePolicyRejectsAnyBashBlock(t *testing.T) {
	t.Parallel()
	body := "some text\n!`echo hi`\nmore"
	v := NewBashSubstitutionPolicyRule().Check(policyAsset(t, "none", body), nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
}

func TestBashSubstitutionPolicyRule_NonePolicyRejectsFencedBashBlock(t *testing.T) {
	t.Parallel()
	body := "intro\n```!\necho hi\n```\nend"
	v := NewBashSubstitutionPolicyRule().Check(policyAsset(t, "none", body), nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
}

func TestBashSubstitutionPolicyRule_NonePolicyCleanBody_NoViolation(t *testing.T) {
	t.Parallel()
	body := "no bash blocks here"
	assert.Empty(t, NewBashSubstitutionPolicyRule().Check(policyAsset(t, "none", body), nil))
}

func TestBashSubstitutionPolicyRule_QuotedPolicy_UnquotedVarRejected(t *testing.T) {
	t.Parallel()
	body := "see !`echo $VAR`"
	v := NewBashSubstitutionPolicyRule().Check(policyAsset(t, "quoted", body), nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
}

func TestBashSubstitutionPolicyRule_QuotedPolicy_UnquotedArgsRejected(t *testing.T) {
	t.Parallel()
	body := "run !`cmd $ARGUMENTS`"
	v := NewBashSubstitutionPolicyRule().Check(policyAsset(t, "quoted", body), nil)
	require.Len(t, v, 1)
}

func TestBashSubstitutionPolicyRule_QuotedPolicy_QuotedVarOk(t *testing.T) {
	t.Parallel()
	body := "ok !`echo \"$VAR\"`"
	assert.Empty(t, NewBashSubstitutionPolicyRule().Check(policyAsset(t, "quoted", body), nil))
}

func TestBashSubstitutionPolicyRule_UnrestrictedPolicy_Warning(t *testing.T) {
	t.Parallel()
	body := "ok !`echo $whatever`"
	v := NewBashSubstitutionPolicyRule().Check(policyAsset(t, "unrestricted", body), nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityWarning, v[0].Severity())
}

func TestBashSubstitutionPolicyRule_UnrestrictedPolicy_NoBashBlock_NoViolation(t *testing.T) {
	t.Parallel()
	assert.Empty(t, NewBashSubstitutionPolicyRule().Check(policyAsset(t, "unrestricted", "plain body"), nil))
}

func TestBashSubstitutionPolicyRule_OverlayExempt(t *testing.T) {
	t.Parallel()
	a, err := dochealthdomain.NewScaffoldAsset("foo.project.md", nil, "!`echo hi`", 1, true)
	require.NoError(t, err)
	assert.Empty(t, NewBashSubstitutionPolicyRule().Check(a, nil))
}

func TestBashSubstitutionPolicyRule_MissingPolicy_NoViolation(t *testing.T) {
	t.Parallel()
	fm := fullFrontmatter()
	delete(fm, "bash_substitution_policy")
	a, err := dochealthdomain.NewScaffoldAsset("foo.md", fm, "!`echo hi`", 1, false)
	require.NoError(t, err)
	// FrontmatterSchemaRule reports the missing field; this rule no-ops.
	assert.Empty(t, NewBashSubstitutionPolicyRule().Check(a, nil))
}

// --- Round 1 fix-cycle additions ---

// WH-HIGH-1 — standard Markdown fenced shell blocks must trigger `none`.

func TestBashSubstitutionPolicyRule_NonePolicyRejectsStandardBashFence(t *testing.T) {
	t.Parallel()
	body := "intro\n```bash\necho hi\n```\nend"
	v := NewBashSubstitutionPolicyRule().Check(policyAsset(t, "none", body), nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
}

func TestBashSubstitutionPolicyRule_NonePolicyRejectsShFence(t *testing.T) {
	t.Parallel()
	body := "```sh\necho hi\n```"
	v := NewBashSubstitutionPolicyRule().Check(policyAsset(t, "none", body), nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
}

func TestBashSubstitutionPolicyRule_NonePolicyRejectsZshFence(t *testing.T) {
	t.Parallel()
	body := "```zsh\necho hi\n```"
	v := NewBashSubstitutionPolicyRule().Check(policyAsset(t, "none", body), nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
}

func TestBashSubstitutionPolicyRule_NonePolicyAcceptsNonShellFence(t *testing.T) {
	t.Parallel()
	body := "```python\necho hi\n```"
	assert.Empty(t, NewBashSubstitutionPolicyRule().Check(policyAsset(t, "none", body), nil))
}

// QA-SEC-DEFER-2 — brace-form ${VAR} / ${ARGUMENTS} / ${N} must be flagged.

func TestBashSubstitutionPolicyRule_QuotedPolicy_BraceArgs_Error(t *testing.T) {
	t.Parallel()
	body := "```!\necho ${ARGUMENTS}\n```"
	v := NewBashSubstitutionPolicyRule().Check(policyAsset(t, "quoted", body), nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
}

func TestBashSubstitutionPolicyRule_QuotedPolicy_BraceVar_Error(t *testing.T) {
	t.Parallel()
	body := "```!\necho ${HOME}\n```"
	v := NewBashSubstitutionPolicyRule().Check(policyAsset(t, "quoted", body), nil)
	require.Len(t, v, 1)
}

func TestBashSubstitutionPolicyRule_QuotedPolicy_QuotedBraceArgs_NoViolation(t *testing.T) {
	t.Parallel()
	body := "```!\necho \"${ARGUMENTS}\"\n```"
	assert.Empty(t, NewBashSubstitutionPolicyRule().Check(policyAsset(t, "quoted", body), nil))
}

// WH-MED-1 — single-quotes block substitution entirely; treat as quoted.

func TestBashSubstitutionPolicyRule_QuotedPolicy_SingleQuotedArgs_NoViolation(t *testing.T) {
	t.Parallel()
	body := "```!\necho '$ARGUMENTS'\n```"
	assert.Empty(t, NewBashSubstitutionPolicyRule().Check(policyAsset(t, "quoted", body), nil))
}

func TestBashSubstitutionPolicyRule_QuotedPolicy_SingleQuotedBrace_NoViolation(t *testing.T) {
	t.Parallel()
	body := "```!\necho '${ARGUMENTS}'\n```"
	assert.Empty(t, NewBashSubstitutionPolicyRule().Check(policyAsset(t, "quoted", body), nil))
}

func TestBashSubstitutionPolicyRule_QuotedPolicy_MixedQuoting_FlagsOnlyUnquoted(t *testing.T) {
	t.Parallel()
	// '$SAFE' is literal (single-quoted); $UNSAFE is unquoted → ERROR.
	body := "```!\necho '$SAFE' && echo $UNSAFE\n```"
	v := NewBashSubstitutionPolicyRule().Check(policyAsset(t, "quoted", body), nil)
	require.NotEmpty(t, v)
	assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
}

// QA-SEC-DEFER-1 — parity check: gap between separate quoted strings is unquoted.

func TestBashSubstitutionPolicyRule_QuotedPolicy_GapBetweenQuotes_Error(t *testing.T) {
	t.Parallel()
	body := "```!\ncmd \"first\" $ARGS \"second\"\n```"
	v := NewBashSubstitutionPolicyRule().Check(policyAsset(t, "quoted", body), nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
}

// WH SUB-FINDING-1 (alty-cli-l8r) — parameter-expansion forms inside braces
// (`${VAR:-d}`, `${VAR#prefix}`, `${VAR//pat/repl}`, etc.) must be flagged
// when unquoted, and pass when quoted.

func TestBashSubstitutionPolicyRule_WhenUnquotedParameterExpansion_FlagsViolation(t *testing.T) {
	t.Parallel()
	body := "```!\necho ${ARGUMENTS:-d}\n```"
	v := NewBashSubstitutionPolicyRule().Check(policyAsset(t, "quoted", body), nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityError, v[0].Severity())
}

func TestBashSubstitutionPolicyRule_WhenQuotedParameterExpansion_NoViolation(t *testing.T) {
	t.Parallel()
	body := "```!\necho \"${ARGUMENTS:-d}\"\n```"
	assert.Empty(t, NewBashSubstitutionPolicyRule().Check(policyAsset(t, "quoted", body), nil))
}
