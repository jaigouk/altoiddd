package infrastructure_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
)

// errorReader returns an error on Read (simulates I/O failure).
type errorReader struct {
	err error
}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, r.err
}

// Compile-time interface check.
var _ application.Prompter = (*infrastructure.StdinPrompter)(nil)

// --- SelectPersona Tests ---

func TestStdinPrompter_SelectPersona_ReturnsChoice(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("2\n")
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	choice, err := p.SelectPersona(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "2", choice)
}

func TestStdinPrompter_SelectPersona_DisplaysOptions(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("1\n")
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	_, err := p.SelectPersona(context.Background())

	require.NoError(t, err)
	assert.Contains(t, output.String(), "1.")
	assert.Contains(t, output.String(), "Developer")
	assert.Contains(t, output.String(), "2.")
	assert.Contains(t, output.String(), "Product Owner")
}

func TestStdinPrompter_SelectPersona_EOF_ReturnsCanceled(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("") // EOF
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	_, err := p.SelectPersona(context.Background())

	assert.ErrorIs(t, err, context.Canceled)
}

// --- AskQuestion Tests ---

func TestStdinPrompter_AskQuestion_ReturnsAnswer(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("Users and admins\n")
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	answer, err := p.AskQuestion(context.Background(), "Who are the actors?")

	require.NoError(t, err)
	assert.Equal(t, "Users and admins", answer)
}

func TestStdinPrompter_AskQuestion_DisplaysQuestion(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("answer\n")
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	_, err := p.AskQuestion(context.Background(), "Who are the actors?")

	require.NoError(t, err)
	assert.Contains(t, output.String(), "Who are the actors?")
}

func TestStdinPrompter_AskQuestion_EmptyInput_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("\n") // Just Enter
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	answer, err := p.AskQuestion(context.Background(), "Question?")

	require.NoError(t, err)
	assert.Empty(t, answer)
}

func TestStdinPrompter_AskQuestion_EOF_ReturnsCanceled(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("")
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	_, err := p.AskQuestion(context.Background(), "Question?")

	assert.ErrorIs(t, err, context.Canceled)
}

// --- AskSkipReason Tests ---

func TestStdinPrompter_AskSkipReason_ReturnsReason(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("Not relevant\n")
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	reason, err := p.AskSkipReason(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "Not relevant", reason)
}

func TestStdinPrompter_AskSkipReason_DisplaysPrompt(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("reason\n")
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	_, err := p.AskSkipReason(context.Background())

	require.NoError(t, err)
	assert.Contains(t, output.String(), "Reason")
}

func TestStdinPrompter_AskSkipReason_EOF_ReturnsCanceled(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("")
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	_, err := p.AskSkipReason(context.Background())

	assert.ErrorIs(t, err, context.Canceled)
}

// --- ConfirmPlayback Tests ---

func TestStdinPrompter_ConfirmPlayback_Yes_ReturnsTrue(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("y\n")
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	confirmed, err := p.ConfirmPlayback(context.Background(), "Q: Actor?\nA: Users")

	require.NoError(t, err)
	assert.True(t, confirmed)
}

func TestStdinPrompter_ConfirmPlayback_No_ReturnsFalse(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("n\n")
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	confirmed, err := p.ConfirmPlayback(context.Background(), "summary")

	require.NoError(t, err)
	assert.False(t, confirmed)
}

func TestStdinPrompter_ConfirmPlayback_DisplaysSummary(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("y\n")
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	_, err := p.ConfirmPlayback(context.Background(), "Q: Actor?\nA: Users and admins")

	require.NoError(t, err)
	assert.Contains(t, output.String(), "Q: Actor?")
	assert.Contains(t, output.String(), "A: Users and admins")
}

func TestStdinPrompter_ConfirmPlayback_EOF_ReturnsCanceled(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("")
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	_, err := p.ConfirmPlayback(context.Background(), "summary")

	assert.ErrorIs(t, err, context.Canceled)
}

func TestStdinPrompter_ConfirmPlayback_YesVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		{"y\n", true},
		{"Y\n", true},
		{"yes\n", true},
		{"YES\n", true},
		{"n\n", false},
		{"no\n", false},
		{"anything\n", false},
	}

	for _, tc := range tests {
		input := bytes.NewBufferString(tc.input)
		output := &bytes.Buffer{}
		p := infrastructure.NewStdinPrompter(input, output)

		confirmed, err := p.ConfirmPlayback(context.Background(), "summary")

		require.NoError(t, err)
		assert.Equal(t, tc.expected, confirmed, "input: %q", tc.input)
	}
}

// --- Scanner Error Tests (Issue #1 fix verification) ---

func TestStdinPrompter_SelectPersona_ScannerError_ReturnsWrappedError(t *testing.T) {
	t.Parallel()

	readErr := errors.New("broken pipe")
	input := &errorReader{err: readErr}
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	_, err := p.SelectPersona(context.Background())

	require.Error(t, err)
	require.NotErrorIs(t, err, context.Canceled, "should not be context.Canceled for I/O error")
	require.ErrorIs(t, err, readErr, "should wrap the original error")
	assert.Contains(t, err.Error(), "reading input")
}

func TestStdinPrompter_AskQuestion_ScannerError_ReturnsWrappedError(t *testing.T) {
	t.Parallel()

	readErr := errors.New("connection reset")
	input := &errorReader{err: readErr}
	output := &bytes.Buffer{}
	p := infrastructure.NewStdinPrompter(input, output)

	_, err := p.AskQuestion(context.Background(), "Question?")

	require.Error(t, err)
	assert.ErrorIs(t, err, readErr)
}

// Verify io.Reader interface compliance for errorReader.
var _ io.Reader = (*errorReader)(nil)
