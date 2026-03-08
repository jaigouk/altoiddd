// Package infrastructure provides adapters for the Knowledge bounded context.
package infrastructure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	knowledgedomain "github.com/alty-cli/alty/internal/knowledge/domain"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// FileKnowledgeReader reads knowledge entries from the local filesystem.
//
// Directory layout:
//
//	{knowledge_dir}/
//	  ddd/{topic}.md
//	  conventions/{topic}.md
//	  tools/{tool}/{version}/{topic}.toml
//	  cross-tool/{topic}.toml
type FileKnowledgeReader struct {
	root string
}

// NewFileKnowledgeReader creates a FileKnowledgeReader.
func NewFileKnowledgeReader(knowledgeDir string) *FileKnowledgeReader {
	return &FileKnowledgeReader{root: knowledgeDir}
}

// ReadEntry reads a knowledge entry from the filesystem.
func (r *FileKnowledgeReader) ReadEntry(_ context.Context, path knowledgedomain.KnowledgePath, version string) (knowledgedomain.KnowledgeEntry, error) {
	filePath := r.resolveFilePath(path, version)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return knowledgedomain.KnowledgeEntry{}, fmt.Errorf(
			"knowledge entry not found: %s (looked at %s): %w",
			path.Raw(), filePath, domainerrors.ErrInvariantViolation)
	}

	if strings.HasSuffix(filePath, ".toml") {
		return r.readTOMLEntry(path, filePath)
	}
	return r.readMarkdownEntry(path, filePath)
}

// ListTopics lists available topics within a category.
func (r *FileKnowledgeReader) ListTopics(_ context.Context, category knowledgedomain.KnowledgeCategory, tool *string) ([]string, error) {
	var scanDir, pattern string

	switch category {
	case knowledgedomain.CategoryTools:
		if tool == nil {
			return nil, nil
		}
		scanDir = filepath.Join(r.root, "tools", *tool, "current")
		pattern = "*.toml"
	case knowledgedomain.CategoryCrossTool:
		scanDir = filepath.Join(r.root, "cross-tool")
		pattern = "*.toml"
	case knowledgedomain.CategoryDDD:
		scanDir = filepath.Join(r.root, "ddd")
		pattern = "*.md"
	case knowledgedomain.CategoryConventions:
		scanDir = filepath.Join(r.root, "conventions")
		pattern = "*.md"
	}

	if _, err := os.Stat(scanDir); os.IsNotExist(err) {
		return nil, nil
	}

	matches, err := filepath.Glob(filepath.Join(scanDir, pattern))
	if err != nil {
		return nil, fmt.Errorf("globbing %s: %w", scanDir, err)
	}

	topics := make([]string, 0, len(matches))
	for _, m := range matches {
		base := filepath.Base(m)
		ext := filepath.Ext(base)
		topics = append(topics, strings.TrimSuffix(base, ext))
	}
	sort.Strings(topics)
	return topics, nil
}

func (r *FileKnowledgeReader) resolveFilePath(path knowledgedomain.KnowledgePath, version string) string {
	category := path.Category()

	if category == knowledgedomain.CategoryTools {
		tool := path.Tool()
		subtopic := path.Subtopic()
		toolStr := ""
		if tool != nil {
			toolStr = *tool
		}
		subtopicStr := ""
		if subtopic != nil {
			subtopicStr = *subtopic
		}
		return filepath.Join(r.root, "tools", toolStr, version, subtopicStr+".toml")
	}

	if category == knowledgedomain.CategoryCrossTool {
		return filepath.Join(r.root, "cross-tool", path.Topic()+".toml")
	}

	// DDD and conventions: {category}/{topic}.md
	return filepath.Join(r.root, string(category), path.Topic()+".md")
}

func (r *FileKnowledgeReader) readTOMLEntry(path knowledgedomain.KnowledgePath, filePath string) (knowledgedomain.KnowledgeEntry, error) {
	rawText, err := os.ReadFile(filePath)
	if err != nil {
		return knowledgedomain.KnowledgeEntry{}, fmt.Errorf("reading %s: %w", filePath, err)
	}

	var data map[string]any
	if _, err := toml.Decode(string(rawText), &data); err != nil {
		return knowledgedomain.KnowledgeEntry{}, fmt.Errorf("decoding TOML %s: %w", filePath, err)
	}

	metadata := extractMetadata(data)

	title := path.Topic()
	if sub := path.Subtopic(); sub != nil {
		title = *sub
	}

	return knowledgedomain.NewKnowledgeEntry(path, title, string(rawText), metadata, "toml"), nil
}

func (r *FileKnowledgeReader) readMarkdownEntry(path knowledgedomain.KnowledgePath, filePath string) (knowledgedomain.KnowledgeEntry, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return knowledgedomain.KnowledgeEntry{}, fmt.Errorf("reading %s: %w", filePath, err)
	}
	return knowledgedomain.NewKnowledgeEntry(path, path.Topic(), string(content), nil, "markdown"), nil
}

func extractMetadata(data map[string]any) *knowledgedomain.EntryMetadata {
	metaRaw, ok := data["_meta"]
	if !ok {
		return nil
	}
	metaMap, ok := metaRaw.(map[string]any)
	if !ok {
		return nil
	}

	var sourceURLs []string
	if urlsRaw, ok := metaMap["source_urls"]; ok {
		if urlsList, ok := urlsRaw.([]any); ok {
			for _, u := range urlsList {
				sourceURLs = append(sourceURLs, fmt.Sprintf("%v", u))
			}
		}
	}

	schemaVersion := ""
	if sv, ok := metaMap["schema_version"]; ok && sv != nil {
		schemaVersion = fmt.Sprintf("%v", sv)
	}

	confidence := "high"
	if c, ok := metaMap["confidence"]; ok {
		confidence = fmt.Sprintf("%v", c)
	}

	deprecated := false
	if d, ok := metaMap["deprecated"]; ok {
		if b, ok := d.(bool); ok {
			deprecated = b
		}
	}

	meta := knowledgedomain.NewEntryMetadata(
		strOrEmpty(metaMap["last_verified"]),
		strOrEmpty(metaMap["verified_against"]),
		confidence,
		deprecated,
		strOrEmpty(metaMap["next_review_date"]),
		schemaVersion,
		sourceURLs,
	)
	return &meta
}

func strOrEmpty(val any) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}
