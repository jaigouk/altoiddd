package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/domain"
)

func TestInferenceFailedError_ErrorIncludesDocCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		docs []string
		want string
	}{
		{"zero docs", nil, "found 0 doc(s)"},
		{"one doc", []string{"README.md"}, "found 1 doc(s)"},
		{"three docs", []string{"a.md", "b.md", "c.md"}, "found 3 doc(s)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := &domain.InferenceFailedError{Docs: tt.docs, Reason: errors.New("boom")}
			msg := e.Error()
			assert.Contains(t, msg, tt.want)
		})
	}
}

func TestInferenceFailedError_IsMatchesSentinel(t *testing.T) {
	t.Parallel()

	e := &domain.InferenceFailedError{Docs: []string{"a.md"}, Reason: errors.New("boom")}

	require.ErrorIs(t, e, domain.ErrInferenceFailed, "must match ErrInferenceFailed sentinel")
	assert.NotErrorIs(t, e, domain.ErrNoDocsFound, "must NOT match unrelated sentinel")
}

func TestInferenceFailedError_UnwrapReturnsReason(t *testing.T) {
	t.Parallel()

	t.Run("wrapped reason is reachable via errors.Is", func(t *testing.T) {
		t.Parallel()

		sentinel := errors.New("upstream failure")
		e := &domain.InferenceFailedError{Docs: []string{"x.md"}, Reason: sentinel}

		assert.ErrorIs(t, e, sentinel, "errors.Is must walk through Unwrap to find the wrapped reason")
	})

	t.Run("nil reason returns nil from Unwrap", func(t *testing.T) {
		t.Parallel()

		e := &domain.InferenceFailedError{Docs: nil, Reason: nil}
		require.NoError(t, e.Unwrap())
	})

	t.Run("Error string contains reason text", func(t *testing.T) {
		t.Parallel()

		e := &domain.InferenceFailedError{Docs: []string{"a.md"}, Reason: errors.New("malformed json")}
		assert.Contains(t, e.Error(), "malformed json")
	})
}
