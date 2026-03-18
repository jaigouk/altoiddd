package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	knowledgedomain "github.com/alto-cli/alto/internal/knowledge/domain"
	"github.com/alto-cli/alto/internal/knowledge/infrastructure"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// -- Helpers --

func createMarkdownEntry(t *testing.T, base, category, topic, content string) {
	t.Helper()
	dir := filepath.Join(base, category)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, topic+".md"), []byte(content), 0o644))
}

func createTOMLEntry(t *testing.T, base, tool, version, topic, content string) {
	t.Helper()
	dir := filepath.Join(base, "tools", tool, version)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, topic+".toml"), []byte(content), 0o644))
}

func createCrossToolEntry(t *testing.T, base, topic, content string) {
	t.Helper()
	dir := filepath.Join(base, "cross-tool")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, topic+".toml"), []byte(content), 0o644))
}

// -- Markdown tests --

func TestFileKnowledgeReader_ReadsMarkdownEntry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createMarkdownEntry(t, dir, "ddd", "tactical-patterns", "# Tactical Patterns\n\nContent here.")

	reader := infrastructure.NewFileKnowledgeReader(dir)
	path, err := knowledgedomain.NewKnowledgePath("ddd/tactical-patterns")
	require.NoError(t, err)

	entry, err := reader.ReadEntry(context.Background(), path, "current")
	require.NoError(t, err)
	assert.Equal(t, "tactical-patterns", entry.Title())
	assert.Contains(t, entry.Content(), "Tactical Patterns")
	assert.Equal(t, "markdown", entry.Format())
}

func TestFileKnowledgeReader_ReadsConventionsMarkdown(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createMarkdownEntry(t, dir, "conventions", "tdd", "# TDD\n\nRed-Green-Refactor.")

	reader := infrastructure.NewFileKnowledgeReader(dir)
	path, err := knowledgedomain.NewKnowledgePath("conventions/tdd")
	require.NoError(t, err)

	entry, err := reader.ReadEntry(context.Background(), path, "current")
	require.NoError(t, err)
	assert.Equal(t, "tdd", entry.Title())
	assert.Contains(t, entry.Content(), "Red-Green-Refactor")
}

// -- TOML tests --

func TestFileKnowledgeReader_ReadsTOMLEntry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	tomlContent := "[format]\nfile_extension = \".md\"\ndescription = \"Agent definition format\"\n"
	createTOMLEntry(t, dir, "claude-code", "current", "agent-format", tomlContent)

	reader := infrastructure.NewFileKnowledgeReader(dir)
	path, err := knowledgedomain.NewKnowledgePath("tools/claude-code/agent-format")
	require.NoError(t, err)

	entry, err := reader.ReadEntry(context.Background(), path, "current")
	require.NoError(t, err)
	assert.Equal(t, "agent-format", entry.Title())
	assert.Equal(t, "toml", entry.Format())
	assert.Contains(t, entry.Content(), "Agent definition format")
}

func TestFileKnowledgeReader_ReadsTOMLMetadata(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	tomlContent := `[_meta]
last_verified = "2026-01-15"
verified_against = "v2.0"
confidence = "medium"
deprecated = true
next_review_date = "2026-07-01"
schema_version = 1
source_urls = ["https://example.com/docs"]

[format]
description = "Test entry"
`
	createTOMLEntry(t, dir, "claude-code", "current", "config-structure", tomlContent)

	reader := infrastructure.NewFileKnowledgeReader(dir)
	path, err := knowledgedomain.NewKnowledgePath("tools/claude-code/config-structure")
	require.NoError(t, err)

	entry, err := reader.ReadEntry(context.Background(), path, "current")
	require.NoError(t, err)

	meta := entry.Metadata()
	require.NotNil(t, meta)
	assert.Equal(t, "2026-01-15", meta.LastVerified())
	assert.Equal(t, "v2.0", meta.VerifiedAgainst())
	assert.Equal(t, "medium", meta.Confidence())
	assert.True(t, meta.Deprecated())
	assert.Equal(t, "2026-07-01", meta.NextReviewDate())
	assert.Equal(t, "1", meta.SchemaVersion())
	assert.Equal(t, []string{"https://example.com/docs"}, meta.SourceURLs())
}

func TestFileKnowledgeReader_ReadsCrossToolTOML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createCrossToolEntry(t, dir, "concept-mapping", "[mapping]\ndescription = \"Cross-tool concept mapping\"\n")

	reader := infrastructure.NewFileKnowledgeReader(dir)
	path, err := knowledgedomain.NewKnowledgePath("cross-tool/concept-mapping")
	require.NoError(t, err)

	entry, err := reader.ReadEntry(context.Background(), path, "current")
	require.NoError(t, err)
	assert.Equal(t, "concept-mapping", entry.Title())
	assert.Equal(t, "toml", entry.Format())
}

func TestFileKnowledgeReader_ResolvesToolVersionPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createTOMLEntry(t, dir, "cursor", "v1", "rules-format", "[format]\ndescription = 'v1 content'\n")

	reader := infrastructure.NewFileKnowledgeReader(dir)
	path, err := knowledgedomain.NewKnowledgePath("tools/cursor/rules-format")
	require.NoError(t, err)

	entry, err := reader.ReadEntry(context.Background(), path, "v1")
	require.NoError(t, err)
	assert.Equal(t, "rules-format", entry.Title())
	assert.Contains(t, entry.Content(), "v1 content")
}

// -- Error tests --

func TestFileKnowledgeReader_NotFoundRaises(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	reader := infrastructure.NewFileKnowledgeReader(dir)
	path, err := knowledgedomain.NewKnowledgePath("ddd/nonexistent-topic")
	require.NoError(t, err)

	_, err = reader.ReadEntry(context.Background(), path, "current")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "not found")
}

func TestFileKnowledgeReader_ToolNotFoundRaises(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	reader := infrastructure.NewFileKnowledgeReader(dir)
	path, err := knowledgedomain.NewKnowledgePath("tools/nonexistent-tool/some-topic")
	require.NoError(t, err)

	_, err = reader.ReadEntry(context.Background(), path, "current")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// -- ListTopics tests --

func TestFileKnowledgeReader_ListTopics(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createMarkdownEntry(t, dir, "ddd", "tactical-patterns", "# TP")
	createMarkdownEntry(t, dir, "ddd", "strategic-patterns", "# SP")

	reader := infrastructure.NewFileKnowledgeReader(dir)
	topics, err := reader.ListTopics(context.Background(), knowledgedomain.CategoryDDD, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"strategic-patterns", "tactical-patterns"}, topics)
}

func TestFileKnowledgeReader_ListToolTopics(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createTOMLEntry(t, dir, "claude-code", "current", "agent-format", "[f]\nx = 1\n")
	createTOMLEntry(t, dir, "claude-code", "current", "config-structure", "[f]\nx = 1\n")

	reader := infrastructure.NewFileKnowledgeReader(dir)
	tool := "claude-code"
	topics, err := reader.ListTopics(context.Background(), knowledgedomain.CategoryTools, &tool)
	require.NoError(t, err)
	assert.Equal(t, []string{"agent-format", "config-structure"}, topics)
}

func TestFileKnowledgeReader_ListTopicsEmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	reader := infrastructure.NewFileKnowledgeReader(dir)
	topics, err := reader.ListTopics(context.Background(), knowledgedomain.CategoryDDD, nil)
	require.NoError(t, err)
	assert.Empty(t, topics)
}

func TestFileKnowledgeReader_ListCrossToolTopics(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createCrossToolEntry(t, dir, "agents-md", "[f]\nx = 1\n")
	createCrossToolEntry(t, dir, "concept-mapping", "[f]\nx = 1\n")

	reader := infrastructure.NewFileKnowledgeReader(dir)
	topics, err := reader.ListTopics(context.Background(), knowledgedomain.CategoryCrossTool, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"agents-md", "concept-mapping"}, topics)
}
