package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBDCommentArgs_PinsCanonicalArgv pins the exact argv shape for
// posting a stdin-streamed comment via the bd CLI. Regression guard for
// alty-cli-olb — the previous code emitted ["bd", "comment", "add", "<id>"]
// which bd interpreted as two ticket IDs ("add" + the real id) and failed.
func TestBDCommentArgs_PinsCanonicalArgv(t *testing.T) {
	t.Parallel()

	got := bdCommentArgs("alty-cli-xyz")

	assert.Equal(t, []string{"bd", "comment", "alty-cli-xyz", "--stdin"}, got)
}
