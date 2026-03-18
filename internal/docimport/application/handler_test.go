package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/docimport/application"
	"github.com/alto-cli/alto/internal/docimport/domain"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
)

// mockDocImporter is a test double for the DocImporter port.
type mockDocImporter struct {
	result *domain.ImportResult
	err    error
}

func (m *mockDocImporter) Import(_ context.Context, _ string) (*domain.ImportResult, error) {
	return m.result, m.err
}

func TestDocImportHandler_Import(t *testing.T) {
	t.Parallel()

	t.Run("success returns import result", func(t *testing.T) {
		t.Parallel()
		model := ddd.NewDomainModel("test-model")
		importResult, resultErr := domain.NewImportResult(model, nil)
		require.NoError(t, resultErr)
		mock := &mockDocImporter{result: importResult}
		handler := application.NewDocImportHandler(mock)

		result, err := handler.Import(context.Background(), "docs/")
		require.NoError(t, err)
		assert.Equal(t, "test-model", result.Model().ModelID())
		assert.Empty(t, result.Warnings())
	})

	t.Run("propagates importer error", func(t *testing.T) {
		t.Parallel()
		importErr := errors.New("file not found")
		mock := &mockDocImporter{err: importErr}
		handler := application.NewDocImportHandler(mock)

		_, err := handler.Import(context.Background(), "docs/")
		require.Error(t, err)
		assert.ErrorIs(t, err, importErr)
	})
}
