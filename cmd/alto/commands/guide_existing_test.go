package commands

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/composition"
	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	"github.com/alto-cli/alto/internal/shared/infrastructure/eventbus"
)

// ---------------------------------------------------------------------------
// Test stubs
// ---------------------------------------------------------------------------

// stubDocInferenceResult holds the canned return for stubDocReader + stubLLMDocReader.
type stubDocReader struct {
	docs map[string]string
	err  error
}

func (s *stubDocReader) ReadDocs(_ context.Context, _ string) (map[string]string, error) {
	return s.docs, s.err
}

type stubLLMDocReader struct {
	result *domain.InferenceResult
	err    error
}

func (s *stubLLMDocReader) InferModel(_ context.Context, _ map[string]string) (*domain.InferenceResult, error) {
	return s.result, s.err
}

type stubRegexImporter struct {
	model *ddd.DomainModel
	err   error
}

func (s *stubRegexImporter) Import(_ context.Context, _ string) (*ddd.DomainModel, error) {
	return s.model, s.err
}

// stubPrompterChoice is a StorytellingPrompter that returns a fixed choice.
type stubPrompterChoice struct {
	choice string
	err    error
}

func (s *stubPrompterChoice) SelectMode(_ context.Context) (domain.DiscoveryMode, error) {
	return "", nil
}

func (s *stubPrompterChoice) ProposeStory(_ context.Context, _ *domain.DomainStory) (*domain.DomainStory, error) {
	return nil, nil
}

func (s *stubPrompterChoice) AskNarration(_ context.Context, _ string, _ string) (string, error) {
	return "", nil
}

func (s *stubPrompterChoice) ConfirmSentence(_ context.Context, _ domain.StorySentence) (domain.StorySentence, bool, error) {
	return domain.StorySentence{}, false, nil
}

func (s *stubPrompterChoice) AskChoice(_ context.Context, _ string, _ []application.Choice, _ string) (string, error) {
	return s.choice, s.err
}

func (s *stubPrompterChoice) DisplayStory(_ context.Context, _ *domain.DomainStory) error {
	return nil
}

func (s *stubPrompterChoice) SynthesisCheckpoint(_ context.Context, _ application.SynthesisSummary) (bool, error) {
	return false, nil
}

func (s *stubPrompterChoice) AskAnnotation(_ context.Context) (string, int, error) {
	return "", 0, nil
}

var _ application.StorytellingPrompter = (*stubPrompterChoice)(nil)

// stubArtifactRenderer renders empty strings for all artifacts.
type stubArtifactRenderer struct {
	err error
}

func (s *stubArtifactRenderer) RenderPRD(_ context.Context, _ *ddd.DomainModel) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return "# PRD\n", nil
}

func (s *stubArtifactRenderer) RenderDDD(_ context.Context, _ *ddd.DomainModel) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return "# DDD\n", nil
}

func (s *stubArtifactRenderer) RenderArchitecture(_ context.Context, _ *ddd.DomainModel) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return "# ARCH\n", nil
}

var _ application.ArtifactRenderer = (*stubArtifactRenderer)(nil)

// stubFileWriter records files written.
type stubFileWriter struct {
	files map[string]string
}

func newStubFileWriter() *stubFileWriter {
	return &stubFileWriter{files: make(map[string]string)}
}

func (s *stubFileWriter) WriteFile(_ context.Context, path string, content string) error {
	s.files[path] = content
	return nil
}

// buildInferenceResult creates an InferenceResult with a simple model.
func buildInferenceResult(t *testing.T) *domain.InferenceResult {
	t.Helper()
	model := ddd.NewDomainModel("test-project")
	result, err := domain.NewInferenceResult(model, "high", []string{"README.md"})
	require.NoError(t, err)
	return result
}

// setupExistingTest builds an ArtifactGenerationHandler and DocInferenceHandler
// from stubs, returning the handlers and a file writer for verification.
func setupExistingTest(t *testing.T, inferResult *domain.InferenceResult, inferErr error) (
	*application.DocInferenceHandler,
	*application.ArtifactGenerationHandler,
	*stubFileWriter,
) {
	t.Helper()

	// Wire DocInferenceHandler with stubs that return the given result/err.
	var docReader *stubDocReader
	var llmReader *stubLLMDocReader
	if inferErr != nil {
		docReader = &stubDocReader{err: inferErr}
		llmReader = &stubLLMDocReader{err: inferErr}
	} else {
		docReader = &stubDocReader{docs: map[string]string{"README.md": "# My Project"}}
		llmReader = &stubLLMDocReader{result: inferResult}
	}
	regexImporter := &stubRegexImporter{}
	inferenceHandler := application.NewDocInferenceHandler(docReader, llmReader, regexImporter)

	// Wire ArtifactGenerationHandler with stubs.
	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	publisher := eventbus.NewPublisher(bus)
	renderer := &stubArtifactRenderer{}
	fw := newStubFileWriter()
	artifactHandler := application.NewArtifactGenerationHandler(renderer, fw, publisher)

	return inferenceHandler, artifactHandler, fw
}

// ---------------------------------------------------------------------------
// Flag tests
// ---------------------------------------------------------------------------

func TestGuideExisting_FlagRegistered(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	app := &composition.App{}
	cmd := NewGuideCmd(app)

	existingFlag := cmd.Flags().Lookup("existing")
	require.NotNil(t, existingFlag, "--existing flag must exist")
	assert.Contains(t, existingFlag.Usage, "docs")
}

// ---------------------------------------------------------------------------
// Mutual exclusion tests
// ---------------------------------------------------------------------------

func TestGuideExisting_MutualExclusion_Agent(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	cmd := NewGuideCmd(app)
	cmd.SetArgs([]string{"--existing", "--agent"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--existing and --agent are mutually exclusive")
}

func TestGuideExisting_MutualExclusion_Continue(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	cmd := NewGuideCmd(app)
	cmd.SetArgs([]string{"--existing", "--continue"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--existing and --continue are mutually exclusive")
}

func TestGuideExisting_MutualExclusion_Legacy(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	cmd := NewGuideCmd(app)
	cmd.SetArgs([]string{"--existing", "--legacy"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--existing and --legacy are mutually exclusive")
}

// ---------------------------------------------------------------------------
// Init guard test
// ---------------------------------------------------------------------------

func TestGuideExisting_NoConfig_ReturnsInitError(t *testing.T) {
	tmpDir := t.TempDir()
	altoDir := filepath.Join(tmpDir, "alto-scaffold")
	require.NoError(t, os.MkdirAll(altoDir, 0o755))
	// No config.toml written

	t.Chdir(tmpDir)

	app := &composition.App{}
	cmd := NewGuideCmd(app)
	cmd.SetArgs([]string{"--existing"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// ---------------------------------------------------------------------------
// runGuideExistingWithDeps tests
// ---------------------------------------------------------------------------

func TestGuideExisting_UserConfirms_WritesArtifacts(t *testing.T) {
	t.Parallel()

	inferResult := buildInferenceResult(t)
	inferenceHandler, artifactHandler, fw := setupExistingTest(t, inferResult, nil)
	prompter := &stubPrompterChoice{choice: "y"}
	var out bytes.Buffer

	err := runGuideExistingWithDeps(
		context.Background(),
		inferenceHandler, artifactHandler, prompter, &out,
	)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Artifacts generated")
	assert.Contains(t, fw.files, filepath.Join("docs", "PRD.md"))
	assert.Contains(t, fw.files, filepath.Join("docs", "DDD.md"))
	assert.Contains(t, fw.files, filepath.Join("docs", "ARCHITECTURE.md"))
}

func TestGuideExisting_UserConfirms_PrintsSummary(t *testing.T) {
	t.Parallel()

	inferResult := buildInferenceResult(t)
	inferenceHandler, artifactHandler, _ := setupExistingTest(t, inferResult, nil)
	prompter := &stubPrompterChoice{choice: "y"}
	var out bytes.Buffer

	err := runGuideExistingWithDeps(
		context.Background(),
		inferenceHandler, artifactHandler, prompter, &out,
	)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Confidence: high")
	assert.Contains(t, out.String(), "README.md")
}

func TestGuideExisting_UserSkips_DoesNotWriteArtifacts(t *testing.T) {
	t.Parallel()

	inferResult := buildInferenceResult(t)
	inferenceHandler, artifactHandler, fw := setupExistingTest(t, inferResult, nil)
	prompter := &stubPrompterChoice{choice: "s"}
	var out bytes.Buffer

	err := runGuideExistingWithDeps(
		context.Background(),
		inferenceHandler, artifactHandler, prompter, &out,
	)
	require.NoError(t, err)
	assert.Empty(t, fw.files, "no artifacts should be written on skip")
}

func TestGuideExisting_UserDeclines_ReturnsErrInferenceDismissed(t *testing.T) {
	t.Parallel()

	inferResult := buildInferenceResult(t)
	inferenceHandler, artifactHandler, _ := setupExistingTest(t, inferResult, nil)
	prompter := &stubPrompterChoice{choice: "n"}
	var out bytes.Buffer

	err := runGuideExistingWithDeps(
		context.Background(),
		inferenceHandler, artifactHandler, prompter, &out,
	)
	require.ErrorIs(t, err, domain.ErrInferenceDismissed)
}

func TestGuideExisting_NoDocsFound_ReturnsErrInferenceDismissed(t *testing.T) {
	t.Parallel()

	// alty-cli-dfd: this test exercises the genuine "no docs anywhere" path,
	// which under the new error semantics is signalled by an empty doc map
	// (handler returns ErrNoDocsFound) — not by a ReadDocs error. The CLI then
	// retries "." (also empty here) and prints the generic message.
	docReader := &stubDocReader{docs: map[string]string{}}
	llmReader := &stubLLMDocReader{}
	regexImporter := &stubRegexImporter{}
	inferenceHandler := application.NewDocInferenceHandler(docReader, llmReader, regexImporter)

	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	publisher := eventbus.NewPublisher(bus)
	fw := newStubFileWriter()
	artifactHandler := application.NewArtifactGenerationHandler(&stubArtifactRenderer{}, fw, publisher)

	prompter := &stubPrompterChoice{choice: "y"}
	var out bytes.Buffer

	err := runGuideExistingWithDeps(
		context.Background(),
		inferenceHandler, artifactHandler, prompter, &out,
	)
	require.ErrorIs(t, err, domain.ErrInferenceDismissed)
	assert.Contains(t, out.String(), "No documentation found")
}

func TestGuideExisting_PromptError_PropagatesError(t *testing.T) {
	t.Parallel()

	inferResult := buildInferenceResult(t)
	inferenceHandler, artifactHandler, _ := setupExistingTest(t, inferResult, nil)
	prompter := &stubPrompterChoice{err: fmt.Errorf("user cancelled")}
	var out bytes.Buffer

	err := runGuideExistingWithDeps(
		context.Background(),
		inferenceHandler, artifactHandler, prompter, &out,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "confirmation prompt")
	assert.Contains(t, err.Error(), "user cancelled")
}

// ---------------------------------------------------------------------------
// Doc directory resolution tests
// ---------------------------------------------------------------------------

func TestGuideExisting_DocsDir_TriesDocsThenRoot(t *testing.T) {
	t.Parallel()

	// alty-cli-dfd: the CLI now retries "." only when "docs" reports
	// ErrNoDocsFound (i.e. handler returned empty), not for arbitrary ReadDocs
	// errors. Return an empty map for "docs" to trigger the fallback.
	callCount := 0
	docsDir := ""
	docReader := &trackingDocReader{
		readFn: func(_ context.Context, dir string) (map[string]string, error) {
			callCount++
			docsDir = dir
			if dir == "docs" {
				return map[string]string{}, nil
			}
			return map[string]string{"README.md": "# Root README"}, nil
		},
	}

	model := ddd.NewDomainModel("fallback-test")
	inferResult, err := domain.NewInferenceResult(model, "medium", []string{"README.md"})
	require.NoError(t, err)

	llmReader := &stubLLMDocReader{result: inferResult}
	regexImporter := &stubRegexImporter{}
	inferenceHandler := application.NewDocInferenceHandler(docReader, llmReader, regexImporter)

	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	publisher := eventbus.NewPublisher(bus)
	fw := newStubFileWriter()
	artifactHandler := application.NewArtifactGenerationHandler(&stubArtifactRenderer{}, fw, publisher)

	prompter := &stubPrompterChoice{choice: "y"}
	var out bytes.Buffer

	runErr := runGuideExistingWithDeps(
		context.Background(),
		inferenceHandler, artifactHandler, prompter, &out,
	)
	require.NoError(t, runErr)
	// Doc reader should have been called twice: first with "docs", then with "."
	assert.Equal(t, 2, callCount)
	assert.Equal(t, ".", docsDir) // last call was to "."
}

// trackingDocReader tracks calls for directory resolution tests.
type trackingDocReader struct {
	readFn func(ctx context.Context, dir string) (map[string]string, error)
}

func (t *trackingDocReader) ReadDocs(ctx context.Context, dir string) (map[string]string, error) {
	return t.readFn(ctx, dir)
}

var _ application.DocReader = (*trackingDocReader)(nil)

// ---------------------------------------------------------------------------
// ErrInferenceDismissed sentinel test
// ---------------------------------------------------------------------------

func TestErrInferenceDismissed_IsSentinel(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("wrapped: %w", domain.ErrInferenceDismissed)
	assert.ErrorIs(t, wrapped, domain.ErrInferenceDismissed)
}

// ---------------------------------------------------------------------------
// GenerateFromModel test
// ---------------------------------------------------------------------------

func TestGenerateFromModel_NilModel_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	publisher := eventbus.NewPublisher(bus)
	renderer := &stubArtifactRenderer{}
	fw := newStubFileWriter()
	handler := application.NewArtifactGenerationHandler(renderer, fw, publisher)

	err := handler.GenerateFromModel(context.Background(), nil, "docs", ".")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestGenerateFromModel_ValidModel_WritesArtifacts(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	publisher := eventbus.NewPublisher(bus)
	renderer := &stubArtifactRenderer{}
	fw := newStubFileWriter()
	handler := application.NewArtifactGenerationHandler(renderer, fw, publisher)

	model := ddd.NewDomainModel("test")
	err := handler.GenerateFromModel(context.Background(), model, "docs", ".")
	require.NoError(t, err)
	assert.Contains(t, fw.files, filepath.Join("docs", "PRD.md"))
	assert.Contains(t, fw.files, filepath.Join("docs", "DDD.md"))
	assert.Contains(t, fw.files, filepath.Join("docs", "ARCHITECTURE.md"))
}

// ---------------------------------------------------------------------------
// Ctrl-C / context.Canceled at confirmation prompt
// ---------------------------------------------------------------------------

func TestGuideExisting_CtrlC_AtConfirmation_PropagatesContextCanceled(t *testing.T) {
	t.Parallel()

	inferResult := buildInferenceResult(t)
	inferenceHandler, artifactHandler, _ := setupExistingTest(t, inferResult, nil)
	prompter := &stubPrompterChoice{err: context.Canceled}
	var out bytes.Buffer

	err := runGuideExistingWithDeps(
		context.Background(),
		inferenceHandler, artifactHandler, prompter, &out,
	)
	require.ErrorIs(t, err, context.Canceled, "Ctrl-C should propagate as context.Canceled")
	assert.Contains(t, err.Error(), "confirmation prompt")
}

// ---------------------------------------------------------------------------
// GenerateFromModel render error propagated through runGuideExistingWithDeps
// ---------------------------------------------------------------------------

func TestGuideExisting_GenerateFromModel_RenderError_PropagatesWrapped(t *testing.T) {
	t.Parallel()

	inferResult := buildInferenceResult(t)

	// Build inference handler with working stubs.
	docReader := &stubDocReader{docs: map[string]string{"README.md": "# My Project"}}
	llmReader := &stubLLMDocReader{result: inferResult}
	regexImporter := &stubRegexImporter{}
	inferenceHandler := application.NewDocInferenceHandler(docReader, llmReader, regexImporter)

	// Build artifact handler with a renderer that fails.
	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	publisher := eventbus.NewPublisher(bus)
	renderer := &stubArtifactRenderer{err: fmt.Errorf("LLM unavailable")}
	fw := newStubFileWriter()
	artifactHandler := application.NewArtifactGenerationHandler(renderer, fw, publisher)

	prompter := &stubPrompterChoice{choice: "y"}
	var out bytes.Buffer

	err := runGuideExistingWithDeps(
		context.Background(),
		inferenceHandler, artifactHandler, prompter, &out,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "generating artifacts from inferred model")
	assert.Contains(t, err.Error(), "LLM unavailable")
	assert.Empty(t, fw.files, "no artifacts should be written on render error")
}

// ---------------------------------------------------------------------------
// "skip" does NOT mark discovery.completed = true
// ---------------------------------------------------------------------------

func TestGuideExisting_UserSkips_DoesNotMarkDiscoveryCompleted(t *testing.T) {
	t.Parallel()

	inferResult := buildInferenceResult(t)
	inferenceHandler, artifactHandler, _ := setupExistingTest(t, inferResult, nil)
	prompter := &stubPrompterChoice{choice: "s"}
	var out bytes.Buffer

	err := runGuideExistingWithDeps(
		context.Background(),
		inferenceHandler, artifactHandler, prompter, &out,
	)
	require.NoError(t, err)
	// runGuideExistingWithDeps returns nil for "skip" without calling
	// markDiscoveryCompleted. Verify output does NOT contain the completion marker.
	assert.NotContains(t, out.String(), "Artifacts generated",
		"skip should not print artifact completion message")
	assert.NotContains(t, out.String(), "Discovery complete",
		"skip should not print discovery complete message")
}
