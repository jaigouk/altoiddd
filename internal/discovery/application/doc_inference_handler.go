package application

import (
	"context"
	"errors"
	"fmt"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/shared/infrastructure/llm"
)

// DocInferenceHandler orchestrates LLM-powered doc inference with regex fallback.
type DocInferenceHandler struct {
	docReader     DocReader
	llmReader     LLMDocReader
	regexImporter RegexImporter
}

// NewDocInferenceHandler creates a DocInferenceHandler with the given dependencies.
func NewDocInferenceHandler(docReader DocReader, llmReader LLMDocReader, regexImporter RegexImporter) *DocInferenceHandler {
	return &DocInferenceHandler{
		docReader:     docReader,
		llmReader:     llmReader,
		regexImporter: regexImporter,
	}
}

// InferFromDocs reads documentation files from docsDir, sends them to the LLM for
// structured inference, and returns an InferenceResult. Falls back to regex-based
// import when the LLM is unavailable.
func (h *DocInferenceHandler) InferFromDocs(ctx context.Context, docsDir string) (*discoverydomain.InferenceResult, error) {
	docs, err := h.docReader.ReadDocs(ctx, docsDir)
	if err != nil {
		return nil, fmt.Errorf("reading docs from %q: %w", docsDir, err)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("no documentation found in %q", docsDir)
	}

	result, err := h.llmReader.InferModel(ctx, docs)
	if err == nil {
		return result, nil
	}

	// Only fallback on ErrLLMUnavailable; other errors are real failures.
	if !errors.Is(err, llm.ErrLLMUnavailable) {
		return nil, fmt.Errorf("LLM inference failed: %w", err)
	}

	// Fallback to regex-based import.
	model, regexErr := h.regexImporter.Import(ctx, docsDir)
	if regexErr != nil {
		return nil, fmt.Errorf("regex fallback failed: %w", regexErr)
	}

	sourceDocs := make([]string, 0, len(docs))
	for name := range docs {
		sourceDocs = append(sourceDocs, name)
	}

	fallbackResult, resultErr := discoverydomain.NewInferenceResult(model, "low", sourceDocs)
	if resultErr != nil {
		return nil, fmt.Errorf("creating fallback result: %w", resultErr)
	}
	return fallbackResult, nil
}
