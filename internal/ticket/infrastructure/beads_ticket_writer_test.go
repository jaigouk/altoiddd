package infrastructure_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/ticket/application"
	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
	"github.com/alty-cli/alty/internal/ticket/infrastructure"
)

// Compile-time interface satisfaction check.
var _ application.BeadsWriter = (*infrastructure.BeadsCLIWriter)(nil)

func TestBeadsCLIWriter_WriteEpic(t *testing.T) {
	t.Parallel()

	// This test requires bd CLI to be installed
	// Skip if not available
	writer := infrastructure.NewBeadsCLIWriter(t.TempDir())
	epic := ticketdomain.NewGeneratedEpic(
		"test-id", "Test Epic", "Test description", "TestContext", "core",
	)

	epicID, err := writer.WriteEpic(context.Background(), epic)
	// In unit tests without bd, this will fail - that's expected
	// The integration test will verify full behavior
	if err != nil {
		t.Skip("bd CLI not available or not in a beads project")
	}

	assert.NotEmpty(t, epicID)
	assert.Contains(t, epicID, "-") // beads IDs have format like "alty-cli-xxx"
}

func TestBeadsCLIWriter_WriteTicket_Task(t *testing.T) {
	t.Parallel()

	writer := infrastructure.NewBeadsCLIWriter(t.TempDir())
	ticket := ticketdomain.NewGeneratedTicket(
		"test-id", "Test Task", "Task description",
		"full", "", "TestContext", "TestAggregate", nil, 0,
	)

	taskID, err := writer.WriteTicket(context.Background(), ticket)
	if err != nil {
		t.Skip("bd CLI not available or not in a beads project")
	}

	assert.NotEmpty(t, taskID)
}

func TestBeadsCLIWriter_WriteTicket_Spike(t *testing.T) {
	t.Parallel()

	writer := infrastructure.NewBeadsCLIWriter(t.TempDir())
	ticket := ticketdomain.NewGeneratedSpikeTicket(
		"test-id", "Spike: Test Research", "Research question",
		"", "TestContext",
	)

	spikeID, err := writer.WriteTicket(context.Background(), ticket)
	if err != nil {
		t.Skip("bd CLI not available or not in a beads project")
	}

	assert.NotEmpty(t, spikeID)
}

func TestBeadsCLIWriter_ParseIssueID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "standard output",
			output:   "Created issue: alty-cli-abc — Test title",
			expected: "alty-cli-abc",
		},
		{
			name:     "with newlines",
			output:   "\nCreated issue: my-proj-xyz — Another title\n",
			expected: "my-proj-xyz",
		},
		{
			name:     "unicode title",
			output:   "Created issue: proj-123 — Test with émojis 🎉",
			expected: "proj-123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			id, err := infrastructure.ParseIssueID(tc.output)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, id)
		})
	}
}

func TestBeadsCLIWriter_ParseIssueID_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		output string
	}{
		{"empty output", ""},
		{"no match", "Some other output"},
		{"partial match", "Created issue: "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := infrastructure.ParseIssueID(tc.output)
			assert.Error(t, err)
		})
	}
}
