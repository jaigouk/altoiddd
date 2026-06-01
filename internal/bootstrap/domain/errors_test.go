package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/bootstrap/domain"
)

func TestErrUnsafeTemplateParameter_IsError(t *testing.T) {
	t.Parallel()
	require.Error(t, domain.ErrUnsafeTemplateParameter)
	assert.Equal(t, "unsafe template parameter", domain.ErrUnsafeTemplateParameter.Error())

	wrapped := errors.New("wrap: " + domain.ErrUnsafeTemplateParameter.Error())
	assert.NotErrorIs(t, wrapped, domain.ErrUnsafeTemplateParameter, "string-equality must not pass errors.Is")
}
