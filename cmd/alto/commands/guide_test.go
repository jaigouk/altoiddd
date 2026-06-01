package commands

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/composition"
	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	"github.com/alto-cli/alto/internal/shared/infrastructure/llm"
)

// pathAwareDocReader returns different docs per requested directory. Used to
// exercise the "docs" -> "." fallback branch in runGuideExistingWithDeps.
// (Other tests in this package use the path-agnostic stubDocReader.)
type pathAwareDocReader struct {
	byDir map[string]map[string]string
}

func (r *pathAwareDocReader) ReadDocs(_ context.Context, dir string) (map[string]string, error) {
	if docs, ok := r.byDir[dir]; ok {
		return docs, nil
	}
	// Default: empty map => handler returns ErrNoDocsFound for that dir.
	return map[string]string{}, nil
}

func TestNewGuideCmd_ContinueFlagMentionsAgent(t *testing.T) {
	t.Parallel()

	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	cmd := NewGuideCmd(app)

	continueFlag := cmd.Flags().Lookup("continue")
	require.NotNil(t, continueFlag, "--continue flag must exist")
	assert.Contains(t, continueFlag.Usage, "agent",
		"--continue flag usage should mention agent-mode sessions")
}

func TestNewGuideCmd_LongHelpMentionsContinueRequiresAgent(t *testing.T) {
	t.Parallel()

	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	cmd := NewGuideCmd(app)

	assert.Contains(t, cmd.Long, "--continue",
		"Long help text should mention --continue")
	assert.Contains(t, cmd.Long, "started with --agent",
		"Long help text should clarify that --continue requires a session started with --agent")
}

func TestNewGuideCmd_ContinueNoSession_ErrorMentionsAgent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create alto-scaffold/config.toml so the init guard passes
	altoDir := filepath.Join(tmpDir, "alto-scaffold")
	require.NoError(t, os.MkdirAll(altoDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(altoDir, "config.toml"), nil, 0o644))

	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	cmd := NewGuideCmd(app)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--continue"})

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "alto guide --agent",
		"error message should tell users to start with alto guide --agent")
}

// ---------------------------------------------------------------------------
// alty-cli-dfd: runGuideExistingWithDeps error-mapping tests
// ---------------------------------------------------------------------------

// TestRunGuideExistingWithDeps_InferenceFailedWithDocs_PrintsDocNames verifies
// that when the underlying handler returns *InferenceFailedError, the CLI
// prints the discovered doc names and the underlying reason — not a generic
// "no documentation found" message.
func TestRunGuideExistingWithDeps_InferenceFailedWithDocs_PrintsDocNames(t *testing.T) {
	t.Parallel()

	// Given: "docs" contains files but LLM returns a non-unavailable error,
	// so the handler wraps the result as InferenceFailedError with sorted Docs.
	docReader := &stubDocReader{docs: map[string]string{
		"a.md": "# A",
		"b.md": "# B",
	}}
	llmReader := &stubLLMDocReader{err: errors.New("boom")}
	regexImporter := &stubRegexImporter{}
	inferenceHandler := application.NewDocInferenceHandler(docReader, llmReader, regexImporter)

	prompter := &stubPrompterChoice{choice: "s"}
	var out bytes.Buffer

	// When
	err := runGuideExistingWithDeps(
		context.Background(),
		inferenceHandler, nil, prompter, &out,
	)

	// Then: errored attempt falls through to storytelling (sentinel returned).
	require.ErrorIs(t, err, domain.ErrInferenceDismissed)

	output := out.String()
	assert.Contains(t, output, "a.md", "doc names must appear in the message")
	assert.Contains(t, output, "b.md", "doc names must appear in the message")
	assert.Contains(t, output, "boom", "underlying reason must appear in the message")
	assert.Contains(t, output, "inference failed", "message must distinguish from 'no docs found'")
}

// TestRunGuideExistingWithDeps_NoDocsFound_PrintsGenericMessage verifies the
// original generic-message path is preserved when BOTH "docs" and "." return
// ErrNoDocsFound — the legitimate "no docs anywhere" signal.
func TestRunGuideExistingWithDeps_NoDocsFound_PrintsGenericMessage(t *testing.T) {
	t.Parallel()

	// Given: doc reader returns empty for any directory.
	docReader := &stubDocReader{docs: map[string]string{}}
	llmReader := &stubLLMDocReader{}
	regexImporter := &stubRegexImporter{}
	inferenceHandler := application.NewDocInferenceHandler(docReader, llmReader, regexImporter)

	prompter := &stubPrompterChoice{choice: "s"}
	var out bytes.Buffer

	// When
	err := runGuideExistingWithDeps(
		context.Background(),
		inferenceHandler, nil, prompter, &out,
	)

	// Then
	require.ErrorIs(t, err, domain.ErrInferenceDismissed)
	assert.Contains(t, out.String(), "No documentation found for inference",
		"the generic message must remain for the genuine no-docs-anywhere case")
}

// TestRunGuideExistingWithDeps_DocsFails_DotSucceeds verifies that
// ErrNoDocsFound on "docs" triggers a fallback to "." and the success
// path is reached when "." has docs.
func TestRunGuideExistingWithDeps_DocsFails_DotSucceeds(t *testing.T) {
	t.Parallel()

	// Given: "docs" is empty, "." has README.md, LLM unavailable, regex succeeds.
	docReader := &pathAwareDocReader{byDir: map[string]map[string]string{
		"docs": {}, // empty -> ErrNoDocsFound
		".":    {"README.md": "# Root README"},
	}}
	llmReader := &stubLLMDocReader{err: llm.ErrLLMUnavailable}
	regexImporter := &stubRegexImporter{model: buildRegexModel(t)}
	inferenceHandler := application.NewDocInferenceHandler(docReader, llmReader, regexImporter)

	// Use a prompter that records being called and returns "s" so the test
	// doesn't depend on artifact-generation wiring.
	prompter := &stubPrompterChoice{choice: "s"}
	var out bytes.Buffer

	// When
	err := runGuideExistingWithDeps(
		context.Background(),
		inferenceHandler, nil, prompter, &out,
	)

	// Then: success path was reached — no error, no "no documentation" message.
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Confidence:", "summary must be printed on success")
	assert.NotContains(t, out.String(), "No documentation found",
		"fallback succeeded, so the no-docs message must not appear")
}

// buildRegexModel constructs a minimal DomainModel for tests that need the
// regex fallback to succeed.
func buildRegexModel(t *testing.T) *ddd.DomainModel {
	t.Helper()
	return ddd.NewDomainModel("regex-parsed")
}
