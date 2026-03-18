package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

func TestNewInferenceResult_WhenModelAndConfidence_ExpectAccessors(t *testing.T) {
	t.Parallel()

	model := ddd.NewDomainModel("test-model")
	sourceDocs := []string{"README.md", "DDD.md"}

	result, err := discoverydomain.NewInferenceResult(model, "high", sourceDocs)

	require.NoError(t, err)
	assert.Equal(t, model, result.Model())
	assert.Equal(t, "high", result.Confidence())
	assert.Equal(t, []string{"README.md", "DDD.md"}, result.SourceDocs())
}

func TestNewInferenceResult_WhenNoModel_ExpectError(t *testing.T) {
	t.Parallel()

	result, err := discoverydomain.NewInferenceResult(nil, "high", []string{"README.md"})

	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Nil(t, result)
}

func TestNewInferenceResult_WhenEmptyConfidence_ExpectError(t *testing.T) {
	t.Parallel()

	model := ddd.NewDomainModel("test-model")
	result, err := discoverydomain.NewInferenceResult(model, "", []string{"README.md"})

	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Nil(t, result)
}

func TestInferenceResult_SourceDocs_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()

	model := ddd.NewDomainModel("test-model")
	sourceDocs := []string{"README.md", "DDD.md"}

	result, err := discoverydomain.NewInferenceResult(model, "medium", sourceDocs)
	require.NoError(t, err)

	// Mutate the returned slice — should not affect the original.
	docs := result.SourceDocs()
	docs[0] = "MUTATED"
	assert.Equal(t, "README.md", result.SourceDocs()[0])
}
