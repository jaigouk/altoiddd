package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

func TestBuildEditedSentence_HappyPath_NoPreposition(t *testing.T) {
	t.Parallel()

	original, err := discoverydomain.NewStorySentence(1, "User", "submits", "Order", vo.UserStated, "")
	require.NoError(t, err)

	result, err := buildEditedSentence(original, "Admin", "reviews", "Invoice", "", "")
	require.NoError(t, err)

	assert.Equal(t, 1, result.Step())
	assert.Equal(t, "Admin", result.Subject())
	assert.Equal(t, "reviews", result.Activity())
	assert.Equal(t, "Invoice", result.Object())
	assert.Equal(t, vo.UserStated, result.Trust())
	assert.Empty(t, result.Source())
	assert.False(t, result.HasIndirectObject())
}

func TestBuildEditedSentence_HappyPath_WithPreposition(t *testing.T) {
	t.Parallel()

	original, err := discoverydomain.NewStorySentence(1, "User", "submits", "Order", vo.UserStated, "")
	require.NoError(t, err)

	result, err := buildEditedSentence(original, "Clerk", "sends", "Invoice", "for", "Manager")
	require.NoError(t, err)

	assert.Equal(t, 1, result.Step())
	assert.Equal(t, "Clerk", result.Subject())
	assert.Equal(t, "sends", result.Activity())
	assert.Equal(t, "Invoice", result.Object())
	assert.True(t, result.HasIndirectObject())
	assert.Equal(t, "for", result.Preposition())
	assert.Equal(t, "Manager", result.IndirectObject())
}

func TestBuildEditedSentence_EmptySubject_Error(t *testing.T) {
	t.Parallel()

	original, err := discoverydomain.NewStorySentence(1, "User", "submits", "Order", vo.UserStated, "")
	require.NoError(t, err)

	_, err = buildEditedSentence(original, "", "reviews", "Invoice", "", "")
	require.Error(t, err)
}

func TestBuildEditedSentence_EmptyActivity_Error(t *testing.T) {
	t.Parallel()

	original, err := discoverydomain.NewStorySentence(1, "User", "submits", "Order", vo.UserStated, "")
	require.NoError(t, err)

	_, err = buildEditedSentence(original, "Admin", "", "Invoice", "", "")
	require.Error(t, err)
}

func TestBuildEditedSentence_EmptyObject_Error(t *testing.T) {
	t.Parallel()

	original, err := discoverydomain.NewStorySentence(1, "User", "submits", "Order", vo.UserStated, "")
	require.NoError(t, err)

	_, err = buildEditedSentence(original, "Admin", "reviews", "", "", "")
	require.Error(t, err)
}

func TestBuildEditedSentence_InvalidPreposition_Error(t *testing.T) {
	t.Parallel()

	original, err := discoverydomain.NewStorySentence(1, "User", "submits", "Order", vo.UserStated, "")
	require.NoError(t, err)

	_, err = buildEditedSentence(original, "Admin", "reviews", "Invoice", "toward", "Manager")
	require.Error(t, err)
}

func TestBuildEditedSentence_PrepositionWithEmptyIndirectObject_Error(t *testing.T) {
	t.Parallel()

	original, err := discoverydomain.NewStorySentence(1, "User", "submits", "Order", vo.UserStated, "")
	require.NoError(t, err)

	_, err = buildEditedSentence(original, "Admin", "reviews", "Invoice", "for", "")
	require.Error(t, err)
}
