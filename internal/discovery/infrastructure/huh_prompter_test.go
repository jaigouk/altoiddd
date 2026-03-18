package infrastructure_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
)

// Compile-time interface satisfaction check.
var _ application.Prompter = (*infrastructure.HuhPrompter)(nil)

func TestHuhPrompter_New(t *testing.T) {
	t.Parallel()
	prompter := infrastructure.NewHuhPrompter()
	assert.NotNil(t, prompter)
}
