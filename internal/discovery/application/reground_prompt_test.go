package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
)

// ---------------------------------------------------------------------------
// BuildRegroundingPrompt — pure function tests
// ---------------------------------------------------------------------------

func TestBuildRegroundingPrompt_ZeroValue_ReturnsBasePromptUnchanged(t *testing.T) {
	t.Parallel()

	base := "Pick one"
	got := application.BuildRegroundingPrompt(application.RegroundingContext{}, base)
	assert.Equal(t, base, got)
}

func TestBuildRegroundingPrompt_ModeOnly_PrependsModeHeader(t *testing.T) {
	t.Parallel()

	ctx := application.RegroundingContext{Mode: "rapid"}
	got := application.BuildRegroundingPrompt(ctx, "Who?")
	assert.Equal(t, "[Mode: Rapid]\nWho?", got)
}

func TestBuildRegroundingPrompt_ModeAndPersona_PrependsBothFields(t *testing.T) {
	t.Parallel()

	ctx := application.RegroundingContext{Mode: "thorough", Persona: "Developer"}
	got := application.BuildRegroundingPrompt(ctx, "Continue?")
	assert.Equal(t, "[Mode: Thorough | Persona: Developer]\nContinue?", got)
}

func TestBuildRegroundingPrompt_AllFields_PrependsAllFields(t *testing.T) {
	t.Parallel()

	// Per the grooming update, StoryCount is dropped from the header.
	// AllFields = Mode + Persona (no StoryCount in header since 64x
	// already shows "Story N of M" in the base prompt).
	ctx := application.RegroundingContext{Mode: "thorough", Persona: "Domain Expert"}
	got := application.BuildRegroundingPrompt(ctx, "Tell another?")
	assert.Equal(t, "[Mode: Thorough | Persona: Domain Expert]\nTell another?", got)
}

func TestBuildRegroundingPrompt_EmptyBasePrompt_PrependsHeaderWithNewline(t *testing.T) {
	t.Parallel()

	ctx := application.RegroundingContext{Mode: "rapid"}
	got := application.BuildRegroundingPrompt(ctx, "")
	assert.Equal(t, "[Mode: Rapid]\n", got)
}

func TestBuildRegroundingPrompt_PersonaOnly_ReturnsBaseUnchanged(t *testing.T) {
	t.Parallel()

	// Mode is required for a header to appear — persona alone is not meaningful
	// without mode context. Zero-value Mode means zero-value context effectively.
	ctx := application.RegroundingContext{Persona: "Developer"}
	got := application.BuildRegroundingPrompt(ctx, "base")
	assert.Equal(t, "base", got)
}

// ---------------------------------------------------------------------------
// NewRegroundingContext — builds from a live session
// ---------------------------------------------------------------------------

func TestNewRegroundingContext_PopulatesFromSession(t *testing.T) {
	t.Parallel()

	session := domain.NewDiscoverySession("# Test Project")
	require.NoError(t, session.SetMode(domain.ModeRapid))
	require.NoError(t, session.DetectPersona("1")) // "1" = Developer

	ctx := application.NewRegroundingContext(session)
	assert.Equal(t, "rapid", ctx.Mode)
	assert.Equal(t, "Developer", ctx.Persona)
}

func TestNewRegroundingContext_UnsetPersona_LeavesPersonaEmpty(t *testing.T) {
	t.Parallel()

	session := domain.NewDiscoverySession("# Test Project")
	require.NoError(t, session.SetMode(domain.ModeThorough))
	// No DetectPersona called — persona is unset

	ctx := application.NewRegroundingContext(session)
	assert.Equal(t, "thorough", ctx.Mode)
	assert.Empty(t, ctx.Persona)
}

func TestNewRegroundingContext_NilSession_ReturnsZeroValue(t *testing.T) {
	t.Parallel()

	ctx := application.NewRegroundingContext(nil)
	assert.Equal(t, application.RegroundingContext{}, ctx)
}
