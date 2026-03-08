package application_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	"github.com/alty-cli/alty/internal/tooltranslation/application"
)

// ---------------------------------------------------------------------------
// Fakes
// ---------------------------------------------------------------------------

type fakeFileWriterP struct {
	written map[string]string
}

func newFakeFileWriterP() *fakeFileWriterP {
	return &fakeFileWriterP{written: make(map[string]string)}
}

func (f *fakeFileWriterP) WriteFile(_ context.Context, path, content string) error {
	f.written[path] = content
	return nil
}

// ---------------------------------------------------------------------------
// Tests — List Personas
// ---------------------------------------------------------------------------

func TestPersonaHandler_ListPersonas(t *testing.T) {
	t.Parallel()

	t.Run("returns five personas", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterP()
		handler := application.NewPersonaHandler(writer)

		result := handler.ListPersonas()

		assert.Equal(t, 5, len(result))
	})

	t.Run("correct registers", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterP()
		handler := application.NewPersonaHandler(writer)

		result := handler.ListPersonas()

		technical := 0
		nonTechnical := 0
		for _, p := range result {
			if p.Register() == vo.RegisterTechnical {
				technical++
			}
			if p.Register() == vo.RegisterNonTechnical {
				nonTechnical++
			}
		}
		assert.Equal(t, 3, technical)
		assert.Equal(t, 2, nonTechnical)
	})
}

// ---------------------------------------------------------------------------
// Tests — Build Preview (valid)
// ---------------------------------------------------------------------------

func TestPersonaHandler_BuildPreview(t *testing.T) {
	t.Parallel()

	t.Run("valid persona", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterP()
		handler := application.NewPersonaHandler(writer)

		preview, err := handler.BuildPreview("Solo Developer", "claude-code")

		require.NoError(t, err)
		assert.NotEmpty(t, preview.Content)
		assert.NotEmpty(t, preview.TargetPath)
		assert.NotEmpty(t, preview.Summary)
	})

	t.Run("case insensitive name", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterP()
		handler := application.NewPersonaHandler(writer)

		preview, err := handler.BuildPreview("solo developer", "claude-code")

		require.NoError(t, err)
		assert.Equal(t, "Solo Developer", preview.Persona.Name())
	})

	t.Run("by persona type value", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterP()
		handler := application.NewPersonaHandler(writer)

		preview, err := handler.BuildPreview("solo_developer", "claude-code")

		require.NoError(t, err)
		assert.Equal(t, "Solo Developer", preview.Persona.Name())
	})

	t.Run("does not write", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterP()
		handler := application.NewPersonaHandler(writer)

		handler.BuildPreview("Solo Developer", "claude-code")

		assert.Equal(t, 0, len(writer.written))
	})
}

// ---------------------------------------------------------------------------
// Tests — Build Preview (invalid)
// ---------------------------------------------------------------------------

func TestPersonaHandler_BuildPreviewInvalid(t *testing.T) {
	t.Parallel()

	t.Run("unknown persona raises", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterP()
		handler := application.NewPersonaHandler(writer)

		_, err := handler.BuildPreview("nonexistent", "claude-code")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Unknown persona")
	})

	t.Run("unknown tool raises", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterP()
		handler := application.NewPersonaHandler(writer)

		_, err := handler.BuildPreview("Solo Developer", "unknown-tool")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Unsupported tool")
	})
}

// ---------------------------------------------------------------------------
// Tests — Target Paths
// ---------------------------------------------------------------------------

func TestPersonaHandler_TargetPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		tool       string
		prefix     string
		suffix     string
	}{
		{"claude-code", "claude-code", ".claude/agents/", ".md"},
		{"cursor", "cursor", ".cursor/rules/", ".mdc"},
		{"roo-code", "roo-code", ".roo-code/modes/", ".md"},
		{"opencode", "opencode", ".opencode/agents/", ".md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			writer := newFakeFileWriterP()
			handler := application.NewPersonaHandler(writer)

			preview, err := handler.BuildPreview("Solo Developer", tt.tool)

			require.NoError(t, err)
			assert.True(t, strings.HasPrefix(preview.TargetPath, tt.prefix),
				"expected prefix %q, got %q", tt.prefix, preview.TargetPath)
			assert.True(t, strings.HasSuffix(preview.TargetPath, tt.suffix),
				"expected suffix %q, got %q", tt.suffix, preview.TargetPath)
		})
	}
}

// ---------------------------------------------------------------------------
// Tests — Approve and Write
// ---------------------------------------------------------------------------

func TestPersonaHandler_ApproveAndWrite(t *testing.T) {
	t.Parallel()

	t.Run("calls writer", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterP()
		handler := application.NewPersonaHandler(writer)
		preview, _ := handler.BuildPreview("Solo Developer", "claude-code")

		err := handler.ApproveAndWrite(context.Background(), preview, "/project")

		require.NoError(t, err)
		assert.Equal(t, 1, len(writer.written))
		for path, content := range writer.written {
			assert.Contains(t, path, preview.TargetPath)
			assert.Equal(t, preview.Content, content)
		}
	})

	t.Run("uses output dir", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterP()
		handler := application.NewPersonaHandler(writer)
		preview, _ := handler.BuildPreview("Team Lead", "cursor")

		handler.ApproveAndWrite(context.Background(), preview, "/my/project")

		for path := range writer.written {
			assert.True(t, strings.HasPrefix(path, "/my/project/"),
				"expected path to start with /my/project/, got %q", path)
		}
	})
}
