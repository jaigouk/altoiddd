package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ttdomain "github.com/alto-cli/alto/internal/tooltranslation/domain"
)

func TestSentinels_AreDistinctNonNilErrors(t *testing.T) {
	t.Parallel()

	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrInvocationProtectionNotSupported", ttdomain.ErrInvocationProtectionNotSupported},
		{"ErrMissingTemplate", ttdomain.ErrMissingTemplate},
		{"ErrInvalidAssetName", ttdomain.ErrInvalidAssetName},
		{"ErrPathTraversal", ttdomain.ErrPathTraversal},
		{"ErrInvalidFrontmatter", ttdomain.ErrInvalidFrontmatter},
		{"ErrOrphanOverlay", ttdomain.ErrOrphanOverlay},
	}

	for _, tc := range sentinels {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Error(t, tc.err)
			assert.NotEmpty(t, tc.err.Error())
		})
	}

	// Distinctness — every pair compares unequal under errors.Is.
	for i := range sentinels {
		for j := range sentinels {
			if i == j {
				continue
			}
			assert.NotErrorIs(t,
				sentinels[i].err, sentinels[j].err,
				"%s and %s must be distinct sentinels", sentinels[i].name, sentinels[j].name)
		}
	}
}

func TestSentinels_WrappedDetectableViaErrorsIs(t *testing.T) {
	t.Parallel()

	wrapped := errors.Join(ttdomain.ErrInvocationProtectionNotSupported, ttdomain.ErrMissingTemplate)
	require.ErrorIs(t, wrapped, ttdomain.ErrInvocationProtectionNotSupported)
	require.ErrorIs(t, wrapped, ttdomain.ErrMissingTemplate)
	assert.NotErrorIs(t, wrapped, ttdomain.ErrPathTraversal)
}
