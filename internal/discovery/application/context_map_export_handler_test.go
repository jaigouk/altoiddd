package application_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Mock: ContextMapWriter
// ---------------------------------------------------------------------------

type mockContextMapWriter struct {
	writtenMap  *discoverydomain.ContextMap
	writtenPath string
	err         error
}

func (m *mockContextMapWriter) Write(_ context.Context, path string, cm *discoverydomain.ContextMap) error {
	if m.err != nil {
		return m.err
	}

	m.writtenPath = path
	m.writtenMap = cm

	return nil
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func buildExportSketch(t *testing.T, name string, workObjects []string) discoverydomain.BoundedContextSketch {
	t.Helper()

	sketch, err := discoverydomain.NewBoundedContextSketch(
		name,
		vo.SubdomainCore,
		0.8,
		[]string{"Actor1"},
		workObjects,
		[]string{"Story1"},
		nil,
		vo.AIInferred,
	)
	require.NoError(t, err)

	return sketch
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestContextMapExportHandler_Export_HappyPath(t *testing.T) {
	t.Parallel()

	sketchA := buildExportSketch(t, "ContextA", []string{"Order"})
	sketchB := buildExportSketch(t, "ContextB", []string{"Order"})
	writer := &mockContextMapWriter{}

	handler := application.NewContextMapExportHandler(writer)
	err := handler.Export(
		context.TODO(),
		"myproject",
		[]discoverydomain.BoundedContextSketch{sketchA, sketchB},
		"output.yaml",
	)

	require.NoError(t, err)
	assert.Equal(t, "output.yaml", writer.writtenPath)
	require.NotNil(t, writer.writtenMap)
	assert.Len(t, writer.writtenMap.Relationships(), 1)
}

func TestContextMapExportHandler_Export_WriteError_Propagates(t *testing.T) {
	t.Parallel()

	sketchA := buildExportSketch(t, "ContextA", []string{"Order"})
	sketchB := buildExportSketch(t, "ContextB", []string{"Order"})
	writer := &mockContextMapWriter{err: fmt.Errorf("disk full")}

	handler := application.NewContextMapExportHandler(writer)
	err := handler.Export(
		context.TODO(),
		"myproject",
		[]discoverydomain.BoundedContextSketch{sketchA, sketchB},
		"output.yaml",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "disk full")
}

func TestContextMapExportHandler_Export_EmptySketches(t *testing.T) {
	t.Parallel()

	writer := &mockContextMapWriter{}

	handler := application.NewContextMapExportHandler(writer)
	err := handler.Export(
		context.TODO(),
		"myproject",
		[]discoverydomain.BoundedContextSketch{},
		"output.yaml",
	)

	require.NoError(t, err)
	require.NotNil(t, writer.writtenMap)
	assert.Empty(t, writer.writtenMap.Relationships())
}

func TestContextMapExportHandler_Export_NilSketches(t *testing.T) {
	t.Parallel()

	writer := &mockContextMapWriter{}

	handler := application.NewContextMapExportHandler(writer)
	err := handler.Export(
		context.TODO(),
		"myproject",
		nil,
		"output.yaml",
	)

	require.NoError(t, err)
	require.NotNil(t, writer.writtenMap)
	assert.Empty(t, writer.writtenMap.Relationships())
}
