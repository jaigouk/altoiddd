package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// -- Compile-time interface check --

func TestStorytellingFlow_CompileTimeCheck(t *testing.T) {
	t.Parallel()
	// var _ DiscoveryFlow = (*StorytellingFlow)(nil) is in storytelling_flow.go
	// This test verifies the type assertion compiles.
	var flow DiscoveryFlow = &StorytellingFlow{}
	_ = flow
}

// -- Constructor tests --

func TestNewStorytellingFlow_Rapid_RequiredStoryCount3(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	assert.Equal(t, 3, flow.RequiredStoryCount())
	assert.Equal(t, ModeRapid, flow.Mode())
}

func TestNewStorytellingFlow_Thorough_RequiredStoryCount5(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeThorough)
	require.NoError(t, err)
	assert.Equal(t, 5, flow.RequiredStoryCount())
	assert.Equal(t, ModeThorough, flow.Mode())
}

func TestNewStorytellingFlow_Express_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewStorytellingFlow(ModeExpress)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStorytellingFlow_Deep_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewStorytellingFlow(ModeDeep)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStorytellingFlow_Empty_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewStorytellingFlow(DiscoveryMode(""))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStorytellingFlow_Conversational_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewStorytellingFlow(ModeConversational)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// -- ValidateQuestionOrder --

func TestStorytellingFlow_ValidateQuestionOrder_AlwaysNil(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)

	ref := NewQuestionRef("Q1", PhaseActors)
	answered := []Answer{NewAnswer("Q0", "seed answer")}
	skipped := map[string]bool{"Q2": true}

	err = flow.ValidateQuestionOrder(ref, answered, skipped)
	assert.NoError(t, err)
}

// -- IsPlaybackDue --

func TestStorytellingFlow_IsPlaybackDue_ZeroStories_False(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	assert.False(t, flow.IsPlaybackDue(0))
}

func TestStorytellingFlow_IsPlaybackDue_OneStory_True(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	assert.True(t, flow.IsPlaybackDue(1))
}

func TestStorytellingFlow_IsPlaybackDue_ThreeStories_True(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeThorough)
	require.NoError(t, err)
	assert.True(t, flow.IsPlaybackDue(3))
}

// -- PlaybackInterval --

func TestStorytellingFlow_PlaybackInterval_Returns1(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	assert.Equal(t, 1, flow.PlaybackInterval())
}

// -- CheckCompleteness --

func TestStorytellingFlow_CheckCompleteness_AlwaysNil(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)

	// With answers
	answers := []Answer{NewAnswer("Q1", "a"), NewAnswer("Q2", "b")}
	skipped := map[string]bool{"Q3": true}
	err = flow.CheckCompleteness(answers, skipped)
	require.NoError(t, err)

	// With nil
	err = flow.CheckCompleteness(nil, nil)
	require.NoError(t, err)
}

// -- CheckStoryCompleteness --

func TestStorytellingFlow_CheckStoryCompleteness_Rapid_BelowThreshold_Error(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	err = flow.CheckStoryCompleteness(2)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestStorytellingFlow_CheckStoryCompleteness_Rapid_AtThreshold_Nil(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	err = flow.CheckStoryCompleteness(3)
	assert.NoError(t, err)
}

func TestStorytellingFlow_CheckStoryCompleteness_Rapid_AboveThreshold_Nil(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	err = flow.CheckStoryCompleteness(5)
	assert.NoError(t, err)
}

func TestStorytellingFlow_CheckStoryCompleteness_Thorough_BelowThreshold_Error(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeThorough)
	require.NoError(t, err)
	err = flow.CheckStoryCompleteness(4)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestStorytellingFlow_CheckStoryCompleteness_Thorough_AtThreshold_Nil(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeThorough)
	require.NoError(t, err)
	err = flow.CheckStoryCompleteness(5)
	assert.NoError(t, err)
}

func TestStorytellingFlow_CheckStoryCompleteness_ZeroStories_Error(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	err = flow.CheckStoryCompleteness(0)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// -- IsSynthesisCheckpointDue --

func TestStorytellingFlow_IsSynthesisCheckpointDue_0_False(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	assert.False(t, flow.IsSynthesisCheckpointDue(0))
}

func TestStorytellingFlow_IsSynthesisCheckpointDue_1_True(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	assert.True(t, flow.IsSynthesisCheckpointDue(1))
}

func TestStorytellingFlow_IsSynthesisCheckpointDue_3_True(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeThorough)
	require.NoError(t, err)
	assert.True(t, flow.IsSynthesisCheckpointDue(3))
}

// -- IsMidStoryCheckpointDue --

func TestStorytellingFlow_IsMidStoryCheckpointDue_2_False(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	assert.False(t, flow.IsMidStoryCheckpointDue(2))
}

func TestStorytellingFlow_IsMidStoryCheckpointDue_3_True(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	assert.True(t, flow.IsMidStoryCheckpointDue(3))
}

func TestStorytellingFlow_IsMidStoryCheckpointDue_4_True(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeThorough)
	require.NoError(t, err)
	assert.True(t, flow.IsMidStoryCheckpointDue(4))
}

func TestStorytellingFlow_IsMidStoryCheckpointDue_0_False(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	assert.False(t, flow.IsMidStoryCheckpointDue(0))
}

// -- CanRunBoundaryDetection --

func TestStorytellingFlow_CanRunBoundaryDetection_0_False(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	assert.False(t, flow.CanRunBoundaryDetection(0))
}

func TestStorytellingFlow_CanRunBoundaryDetection_1_False(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	assert.False(t, flow.CanRunBoundaryDetection(1))
}

func TestStorytellingFlow_CanRunBoundaryDetection_2_True(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	assert.True(t, flow.CanRunBoundaryDetection(2))
}

func TestStorytellingFlow_CanRunBoundaryDetection_5_True(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeThorough)
	require.NoError(t, err)
	assert.True(t, flow.CanRunBoundaryDetection(5))
}

// -- ShouldSuggestBranchingSplit --

func TestStorytellingFlow_ShouldSuggestBranchingSplit_WithBranching_True(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)

	story, err := NewDomainStory("Order Flow", StoryTypeCoarseGrained, TimeTypeAsIs, PurityTypePure, "Customer places order")
	require.NoError(t, err)

	actor, err := NewStoryActor("Customer", ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := NewWorkObject("Order", WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	// Activity contains branching keyword "sometimes"
	sentence, err := NewStorySentence(1, "Customer", "sometimes creates", "Order", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	assert.True(t, flow.ShouldSuggestBranchingSplit(story))
}

func TestStorytellingFlow_ShouldSuggestBranchingSplit_NoBranching_False(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)

	story, err := NewDomainStory("Order Flow", StoryTypeCoarseGrained, TimeTypeAsIs, PurityTypePure, "Customer places order")
	require.NoError(t, err)

	actor, err := NewStoryActor("Customer", ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := NewWorkObject("Order", WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	// Activity does NOT contain branching keywords
	sentence, err := NewStorySentence(1, "Customer", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	assert.False(t, flow.ShouldSuggestBranchingSplit(story))
}

func TestStorytellingFlow_ShouldSuggestBranchingSplit_NilStory_False(t *testing.T) {
	t.Parallel()
	flow, err := NewStorytellingFlow(ModeRapid)
	require.NoError(t, err)
	assert.False(t, flow.ShouldSuggestBranchingSplit(nil))
}

// -- Mode --

func TestStorytellingFlow_Mode_ReturnsCorrectMode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mode DiscoveryMode
	}{
		{"rapid", ModeRapid},
		{"thorough", ModeThorough},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flow, err := NewStorytellingFlow(tt.mode)
			require.NoError(t, err)
			assert.Equal(t, tt.mode, flow.Mode())
		})
	}
}
