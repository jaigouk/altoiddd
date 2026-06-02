package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

func TestSecretsGrepRule_Name(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "secrets_grep", NewSecretsGrepRule(nil).Name())
}

func TestSecretsGrepRule_DefaultPatterns_KeywordHit(t *testing.T) {
	t.Parallel()
	cases := []string{
		"the password is hunter2",
		"export SECRET=foo",
		"API_KEY: bar",
		"api-key=baz",
		"bearer abcdef",
		"contains a JWT token",
		"client_secret=xyz",
		"private_key here",
		"credentials inside",
		"aws_access_key=visible",
	}
	rule := NewSecretsGrepRule(nil)
	for _, body := range cases {
		a, err := dochealthdomain.NewScaffoldAsset("foo.md", nil, body, 1, false)
		require.NoError(t, err)
		v := rule.Check(a, nil)
		require.NotEmpty(t, v, "body %q should match", body)
		assert.Equal(t, dochealthdomain.SeverityWarning, v[0].Severity())
	}
}

func TestSecretsGrepRule_DefaultPatterns_AwsKey(t *testing.T) {
	t.Parallel()
	body := "leaked AKIAABCDEFGHIJKLMNOP here"
	a, err := dochealthdomain.NewScaffoldAsset("foo.md", nil, body, 1, false)
	require.NoError(t, err)
	v := NewSecretsGrepRule(nil).Check(a, nil)
	require.NotEmpty(t, v)
	assert.Equal(t, dochealthdomain.SeverityWarning, v[0].Severity())
}

func TestSecretsGrepRule_DefaultPatterns_GithubPat(t *testing.T) {
	t.Parallel()
	body := "token ghp_abcdefghijklmnopqrstuvwxyz0123456789 here"
	a, err := dochealthdomain.NewScaffoldAsset("foo.md", nil, body, 1, false)
	require.NoError(t, err)
	v := NewSecretsGrepRule(nil).Check(a, nil)
	require.NotEmpty(t, v)
}

func TestSecretsGrepRule_CleanBody_NoViolation(t *testing.T) {
	t.Parallel()
	body := "this is a plain markdown body about cats"
	a, err := dochealthdomain.NewScaffoldAsset("foo.md", nil, body, 1, false)
	require.NoError(t, err)
	assert.Empty(t, NewSecretsGrepRule(nil).Check(a, nil))
}

func TestSecretsGrepRule_FrontmatterExcluded(t *testing.T) {
	t.Parallel()
	// Frontmatter values like description: "Manages API keys" must NOT
	// trigger — rule scans body only.
	fm := fullFrontmatter()
	fm["description"] = "Manages API keys and tokens for services"
	a, err := dochealthdomain.NewScaffoldAsset("foo.md", fm, "plain body", 1, false)
	require.NoError(t, err)
	assert.Empty(t, NewSecretsGrepRule(nil).Check(a, nil))
}

func TestSecretsGrepRule_CustomPatterns_Override(t *testing.T) {
	t.Parallel()
	custom, err := dochealthdomain.NewSecretPattern("custom", `XXXX-[0-9]+`)
	require.NoError(t, err)
	rule := NewSecretsGrepRule([]dochealthdomain.SecretPattern{custom})

	hit, err := dochealthdomain.NewScaffoldAsset("foo.md", nil, "value XXXX-12345 here", 1, false)
	require.NoError(t, err)
	assert.NotEmpty(t, rule.Check(hit, nil))

	// Default keyword that would have triggered the binding-floor set must
	// NOT trigger when overrides are supplied (override, not append).
	miss, err := dochealthdomain.NewScaffoldAsset("bar.md", nil, "the password is hunter2", 1, false)
	require.NoError(t, err)
	assert.Empty(t, rule.Check(miss, nil))
}

func TestSecretsGrepRule_CaseInsensitiveKeywords(t *testing.T) {
	t.Parallel()
	a, err := dochealthdomain.NewScaffoldAsset("foo.md", nil, "the PASSWORD is hidden", 1, false)
	require.NoError(t, err)
	assert.NotEmpty(t, NewSecretsGrepRule(nil).Check(a, nil))
}

func TestSecretsGrepRule_ExemptFrontmatter_SkipsKeywordHit(t *testing.T) {
	t.Parallel()
	fm := fullFrontmatter()
	fm["secrets_grep_exempt"] = "security-review agent — domain vocabulary"
	a, err := dochealthdomain.NewScaffoldAsset("agents/sec.md", fm, "credentials live here", 1, false)
	require.NoError(t, err)
	assert.Empty(t, NewSecretsGrepRule(nil).Check(a, nil),
		"non-empty secrets_grep_exempt must suppress all warnings")
}

func TestSecretsGrepRule_ExemptFrontmatter_EmptyReason_StillFires(t *testing.T) {
	t.Parallel()
	fm := fullFrontmatter()
	fm["secrets_grep_exempt"] = ""
	a, err := dochealthdomain.NewScaffoldAsset("agents/sec.md", fm, "credentials inside", 1, false)
	require.NoError(t, err)
	assert.NotEmpty(t, NewSecretsGrepRule(nil).Check(a, nil),
		"empty exempt reason must NOT suppress — the field exists to capture audit decisions, not silence")
}

func TestSecretsGrepRule_Overlay_InheritsExemptFromPrimary(t *testing.T) {
	t.Parallel()
	primaryFM := fullFrontmatter()
	primaryFM["secrets_grep_exempt"] = "security-review agent"
	primary, err := dochealthdomain.NewScaffoldAsset("agents/sec.md", primaryFM, "x", 1, false)
	require.NoError(t, err)
	overlay, err := dochealthdomain.NewScaffoldAsset("agents/sec.project.md", nil, "credentials here", 1, true)
	require.NoError(t, err)
	corpus := []dochealthdomain.ScaffoldAsset{primary, overlay}
	assert.Empty(t, NewSecretsGrepRule(nil).Check(overlay, corpus),
		"overlay must inherit secrets_grep_exempt from its primary sibling")
}

func TestSecretsGrepRule_Overlay_NoPrimaryInCorpus_StillFires(t *testing.T) {
	t.Parallel()
	// Defensive: orphan overlay (no primary in corpus) cannot inherit
	// an exemption — it must still be scanned. Belt + braces with
	// OrphanOverlayRule, which would flag the orphan separately.
	overlay, err := dochealthdomain.NewScaffoldAsset("agents/sec.project.md", nil, "credentials here", 1, true)
	require.NoError(t, err)
	assert.NotEmpty(t, NewSecretsGrepRule(nil).Check(overlay, nil))
}
