package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
)

// Compile-time interface satisfaction checks.
var (
	_ application.Discovery            = (*mockDiscovery)(nil)
	_ application.ArtifactRenderer     = (*mockArtifactRenderer)(nil)
	_ application.ToolDetection        = (*mockToolDetection)(nil)
	_ application.StorytellingPrompter = (*mockStorytellingPrompter)(nil)
)

// --- mockDiscovery ---

type mockDiscovery struct{}

func (m *mockDiscovery) StartSession(_ string) (*discoverydomain.DiscoverySession, error) {
	return nil, nil
}

func (m *mockDiscovery) DetectPersona(_ string, _ string) (*discoverydomain.DiscoverySession, error) {
	return nil, nil
}

func (m *mockDiscovery) AnswerQuestion(_ string, _ string, _ string) (*discoverydomain.DiscoverySession, error) {
	return nil, nil
}

func (m *mockDiscovery) SkipQuestion(_ string, _ string, _ string) (*discoverydomain.DiscoverySession, error) {
	return nil, nil
}

func (m *mockDiscovery) ConfirmPlayback(_ string, _ bool) (*discoverydomain.DiscoverySession, error) {
	return nil, nil
}

func (m *mockDiscovery) Complete(_ string) (*discoverydomain.DiscoverySession, error) {
	return nil, nil
}

// --- mockArtifactRenderer ---

type mockArtifactRenderer struct{}

func (m *mockArtifactRenderer) RenderPRD(_ context.Context, _ *ddd.DomainModel) (string, error) {
	return "", nil
}

func (m *mockArtifactRenderer) RenderDDD(_ context.Context, _ *ddd.DomainModel) (string, error) {
	return "", nil
}

func (m *mockArtifactRenderer) RenderArchitecture(_ context.Context, _ *ddd.DomainModel) (string, error) {
	return "", nil
}

// --- mockToolDetection ---

type mockToolDetection struct{}

func (m *mockToolDetection) Detect(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func (m *mockToolDetection) ScanConflicts(_ context.Context, _ string) ([]discoverydomain.SettingsConflict, error) {
	return nil, nil
}

// --- mockStorytellingPrompter ---

type mockStorytellingPrompter struct{}

func (m *mockStorytellingPrompter) SelectMode(_ context.Context) (discoverydomain.DiscoveryMode, error) {
	return "", nil
}

func (m *mockStorytellingPrompter) ProposeStory(_ context.Context, _ *discoverydomain.DomainStory) (*discoverydomain.DomainStory, error) {
	return nil, nil
}

func (m *mockStorytellingPrompter) AskNarration(_ context.Context, _ string, _ string) (string, error) {
	return "", nil
}

func (m *mockStorytellingPrompter) ConfirmSentence(_ context.Context, _ discoverydomain.StorySentence) (discoverydomain.StorySentence, bool, error) {
	return discoverydomain.StorySentence{}, false, nil
}

func (m *mockStorytellingPrompter) AskChoice(_ context.Context, _ string, _ []application.Choice, _ string) (string, error) {
	return "", nil
}

func (m *mockStorytellingPrompter) DisplayStory(_ context.Context, _ *discoverydomain.DomainStory) error {
	return nil
}

func (m *mockStorytellingPrompter) SynthesisCheckpoint(_ context.Context, _ application.SynthesisSummary) (bool, error) {
	return false, nil
}

func (m *mockStorytellingPrompter) AskAnnotation(_ context.Context) (string, int, error) {
	return "", 0, nil
}

// --- StorytellingPrompter structural tests ---

func TestStorytellingPrompterInterface_Exists(t *testing.T) {
	t.Parallel()

	// The compile-time check above (var _ application.StorytellingPrompter = (*mockStorytellingPrompter)(nil))
	// already verifies the interface exists and the mock satisfies it.
	// This test ensures the mock is assignable at runtime as well.
	var p application.StorytellingPrompter = &mockStorytellingPrompter{}
	assert.NotNil(t, p)
}

func TestChoice_Fields(t *testing.T) {
	t.Parallel()

	c := application.Choice{
		Key:         "a",
		Label:       "Option A",
		Description: "The first option",
	}

	assert.Equal(t, "a", c.Key)
	assert.Equal(t, "Option A", c.Label)
	assert.Equal(t, "The first option", c.Description)
}

func TestSynthesisSummary_Fields(t *testing.T) {
	t.Parallel()

	s := application.SynthesisSummary{
		StoriesSoFar:    []*discoverydomain.DomainStory{},
		ActorInventory:  []discoverydomain.StoryActor{},
		ObjectInventory: []discoverydomain.WorkObject{},
		BoundarySignals: []discoverydomain.BoundarySignal{},
		GlossaryTerms:   []string{"term1"},
	}

	assert.Empty(t, s.StoriesSoFar)
	assert.Empty(t, s.ActorInventory)
	assert.Empty(t, s.ObjectInventory)
	assert.Empty(t, s.BoundarySignals)
	assert.Len(t, s.GlossaryTerms, 1)
}
