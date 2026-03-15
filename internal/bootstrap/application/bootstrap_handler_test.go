package application_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/bootstrap/application"
	"github.com/alty-cli/alty/internal/bootstrap/domain"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Mock GitCommitter
// ---------------------------------------------------------------------------

type fakeGitCommitter struct {
	hasGit      bool
	hasGitErr   error
	stagedPaths []string
	stageErr    error
	commitMsg   string
	commitErr   error
}

func (f *fakeGitCommitter) HasGit(_ context.Context, _ string) (bool, error) {
	return f.hasGit, f.hasGitErr
}

func (f *fakeGitCommitter) StageFiles(_ context.Context, _ string, paths []string) error {
	if f.stageErr != nil {
		return f.stageErr
	}
	f.stagedPaths = append(f.stagedPaths, paths...)
	return nil
}

func (f *fakeGitCommitter) Commit(_ context.Context, _ string, message string) error {
	if f.commitErr != nil {
		return f.commitErr
	}
	f.commitMsg = message
	return nil
}

// fakePublisher is a spy that records published events.
type fakePublisher struct {
	published []any
}

func (f *fakePublisher) Publish(_ context.Context, event any) error {
	f.published = append(f.published, event)
	return nil
}

// ---------------------------------------------------------------------------
// Mock tool detection
// ---------------------------------------------------------------------------

type fakeToolDetection struct {
	tools     []string
	conflicts []string
}

func (f *fakeToolDetection) Detect(projectDir string) ([]string, error) {
	return f.tools, nil
}

func (f *fakeToolDetection) ScanConflicts(projectDir string) ([]string, error) {
	return f.conflicts, nil
}

func newFakeToolDetection(tools []string, conflicts []string) *fakeToolDetection {
	if tools == nil {
		tools = []string{"claude"}
	}
	if conflicts == nil {
		conflicts = []string{}
	}
	return &fakeToolDetection{tools: tools, conflicts: conflicts}
}

// ---------------------------------------------------------------------------
// Mock file writer (spy that records calls)
// ---------------------------------------------------------------------------

type writtenFile struct {
	path    string
	content string
}

type fakeFileWriter struct {
	written []writtenFile
	err     error
}

func (f *fakeFileWriter) WriteFile(_ context.Context, path string, content string) error {
	if f.err != nil {
		return f.err
	}
	f.written = append(f.written, writtenFile{path: path, content: content})
	return nil
}

// ---------------------------------------------------------------------------
// Mock content provider
// ---------------------------------------------------------------------------

type fakeContentProvider struct{}

func (f *fakeContentProvider) ContentFor(path string, config domain.ProjectConfig) string {
	return "test content for " + path
}

// ---------------------------------------------------------------------------
// OS-backed file checker (uses real filesystem in temp dirs)
// ---------------------------------------------------------------------------

type osFileChecker struct{}

func (osFileChecker) Exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestBootstrapHandler_Preview(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tools       []string
		conflicts   []string
		setupDir    func(t *testing.T, dir string)
		wantStatus  domain.SessionStatus
		wantErr     string
		wantToolCnt int
	}{
		{
			name:  "calls tool detection and sets previewed status",
			tools: []string{"claude", "cursor"},
			setupDir: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("My project idea"), 0o644))
			},
			wantStatus:  domain.SessionStatusPreviewed,
			wantToolCnt: 2,
		},
		{
			name: "creates preview with file actions",
			setupDir: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("My project idea"), 0o644))
			},
			wantStatus: domain.SessionStatusPreviewed,
		},
		{
			name:     "raises on missing readme",
			setupDir: func(t *testing.T, dir string) { t.Helper() /* no README */ },
			wantErr:  "README.md",
		},
		{
			name: "skips existing files",
			setupDir: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "PRD.md"), []byte("existing"), 0o644))
			},
			wantStatus: domain.SessionStatusPreviewed,
		},
		{
			name:  "stores detected tools on session",
			tools: []string{"claude", "cursor"},
			setupDir: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))
			},
			wantStatus:  domain.SessionStatusPreviewed,
			wantToolCnt: 2,
		},
		{
			name:      "stores conflict descriptions on preview",
			conflicts: []string{"Global cursor setting overrides local"},
			setupDir: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))
			},
			wantStatus: domain.SessionStatusPreviewed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			tt.setupDir(t, dir)

			fake := newFakeToolDetection(tt.tools, tt.conflicts)
			handler := application.NewBootstrapHandler(fake, osFileChecker{}, &fakePublisher{}, &fakeFileWriter{}, &fakeContentProvider{})

			session, err := handler.Preview(dir)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, session.Status())

			if tt.wantToolCnt > 0 {
				assert.Len(t, session.DetectedTools(), tt.wantToolCnt)
			}
		})
	}
}

func TestBootstrapHandler_FullFlow(t *testing.T) {
	t.Parallel()

	t.Run("preview confirm execute", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("My project idea"), 0o644))

		handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{}, &fakeFileWriter{}, &fakeContentProvider{})

		session, err := handler.Preview(dir)
		require.NoError(t, err)

		session, err = handler.Confirm(session.SessionID())
		require.NoError(t, err)
		assert.Equal(t, domain.SessionStatusConfirmed, session.Status())

		session, err = handler.Execute(session.SessionID())
		require.NoError(t, err)
		assert.Equal(t, domain.SessionStatusCompleted, session.Status())
		assert.Len(t, session.Events(), 1)
	})

	t.Run("cancel flow", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))

		handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{}, &fakeFileWriter{}, &fakeContentProvider{})

		session, err := handler.Preview(dir)
		require.NoError(t, err)

		session, err = handler.Cancel(session.SessionID())
		require.NoError(t, err)
		assert.Equal(t, domain.SessionStatusCancelled, session.Status())
	})
}

func TestBootstrapHandler_SessionNotFound(t *testing.T) {
	t.Parallel()

	handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{}, &fakeFileWriter{}, &fakeContentProvider{})

	tests := []struct {
		name string
		fn   func() (*domain.BootstrapSession, error)
	}{
		{"confirm", func() (*domain.BootstrapSession, error) { return handler.Confirm("no-such-id") }},
		{"cancel", func() (*domain.BootstrapSession, error) { return handler.Cancel("no-such-id") }},
		{"execute", func() (*domain.BootstrapSession, error) { return handler.Execute("no-such-id") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := tt.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "no-such-id")
		})
	}
}

func TestBootstrapHandler_PlannedFiles(t *testing.T) {
	t.Parallel()

	t.Run("includes alty config", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))

		handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{}, &fakeFileWriter{}, &fakeContentProvider{})
		session, err := handler.Preview(dir)
		require.NoError(t, err)
		preview := session.Preview()
		require.NotNil(t, preview)

		paths := make([]string, 0)
		for _, a := range preview.FileActions() {
			paths = append(paths, a.Path())
		}
		assert.Contains(t, paths, ".alty/config.toml")
	})

	t.Run("includes alty maintenance", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))

		handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{}, &fakeFileWriter{}, &fakeContentProvider{})
		session, err := handler.Preview(dir)
		require.NoError(t, err)
		preview := session.Preview()
		require.NotNil(t, preview)

		paths := make([]string, 0)
		for _, a := range preview.FileActions() {
			paths = append(paths, a.Path())
		}
		assert.Contains(t, paths, ".alty/maintenance/doc-registry.toml")
	})
}

func TestBootstrapHandler_SkipsExistingFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "PRD.md"), []byte("existing"), 0o644))

	handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{}, &fakeFileWriter{}, &fakeContentProvider{})
	session, err := handler.Preview(dir)
	require.NoError(t, err)
	preview := session.Preview()
	require.NotNil(t, preview)

	for _, a := range preview.FileActions() {
		if a.Path() == "docs/PRD.md" {
			assert.Equal(t, vo.FileActionSkip, a.ActionType())
			assert.Equal(t, "already exists", a.Reason())
		}
	}
}

func TestBootstrapHandler_Execute_PublishesEvent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))

	pub := &fakePublisher{}
	handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, pub, &fakeFileWriter{}, &fakeContentProvider{})

	session, err := handler.Preview(dir)
	require.NoError(t, err)
	_, err = handler.Confirm(session.SessionID())
	require.NoError(t, err)
	_, err = handler.Execute(session.SessionID())
	require.NoError(t, err)

	require.Len(t, pub.published, 1)
	_, ok := pub.published[0].(domain.BootstrapCompletedEvent)
	assert.True(t, ok, "expected BootstrapCompletedEvent, got %T", pub.published[0])
}

func TestBootstrapHandler_ConflictDescriptions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))

	handler := application.NewBootstrapHandler(newFakeToolDetection(nil, []string{"Global cursor setting overrides local"}), osFileChecker{}, &fakePublisher{}, &fakeFileWriter{}, &fakeContentProvider{})
	session, err := handler.Preview(dir)
	require.NoError(t, err)
	preview := session.Preview()
	require.NotNil(t, preview)
	assert.Equal(t, []string{"Global cursor setting overrides local"}, preview.ConflictDescriptions())
}

// ---------------------------------------------------------------------------
// Execute file-writing tests
// ---------------------------------------------------------------------------

func TestBootstrapHandler_Execute_WritesCreateFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))

	fw := &fakeFileWriter{}
	handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{}, fw, &fakeContentProvider{})

	session, err := handler.Preview(dir)
	require.NoError(t, err)

	// Count expected CREATE actions
	preview := session.Preview()
	require.NotNil(t, preview)
	var createCount int
	for _, a := range preview.FileActions() {
		if a.ActionType() == vo.FileActionCreate {
			createCount++
		}
	}
	require.Positive(t, createCount, "should have CREATE actions")

	_, err = handler.Confirm(session.SessionID())
	require.NoError(t, err)
	_, err = handler.Execute(session.SessionID())
	require.NoError(t, err)

	assert.Len(t, fw.written, createCount, "should write one file per CREATE action")

	// Verify .alty/config.toml was written via the content provider
	var foundConfig bool
	for _, w := range fw.written {
		if filepath.Base(w.path) == "config.toml" {
			assert.Contains(t, w.content, ".alty/config.toml")
			foundConfig = true
		}
	}
	assert.True(t, foundConfig, "expected .alty/config.toml to be written")
}

func TestBootstrapHandler_Execute_SkipsExistingFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))
	// Create docs/PRD.md so it gets a SKIP action
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "PRD.md"), []byte("existing"), 0o644))

	fw := &fakeFileWriter{}
	handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{}, fw, &fakeContentProvider{})

	session, err := handler.Preview(dir)
	require.NoError(t, err)
	_, err = handler.Confirm(session.SessionID())
	require.NoError(t, err)
	_, err = handler.Execute(session.SessionID())
	require.NoError(t, err)

	// Verify docs/PRD.md was NOT written (it was skipped)
	for _, w := range fw.written {
		assert.NotEqual(t, filepath.Join(dir, "docs", "PRD.md"), w.path,
			"should not write skipped file docs/PRD.md")
	}
}

func TestBootstrapHandler_Execute_WriteFailureReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))

	fw := &fakeFileWriter{err: fmt.Errorf("disk full")}
	handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{}, fw, &fakeContentProvider{})

	session, err := handler.Preview(dir)
	require.NoError(t, err)
	_, err = handler.Confirm(session.SessionID())
	require.NoError(t, err)

	_, err = handler.Execute(session.SessionID())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disk full")
}

// ---------------------------------------------------------------------------
// GitCommitter integration tests
// ---------------------------------------------------------------------------

func TestBootstrapHandler_Execute_WhenGitCommitter_ExpectFilesCommitted(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))

	gc := &fakeGitCommitter{hasGit: true}
	handler := application.NewBootstrapHandler(
		newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{},
		&fakeFileWriter{}, &fakeContentProvider{},
		application.WithGitCommitter(gc),
	)

	session, err := handler.Preview(dir)
	require.NoError(t, err)
	_, err = handler.Confirm(session.SessionID())
	require.NoError(t, err)
	_, err = handler.Execute(session.SessionID())
	require.NoError(t, err)

	assert.NotEmpty(t, gc.stagedPaths, "expected files to be staged")
	assert.Equal(t, application.ScaffoldCommitMessage, gc.commitMsg)
}

func TestBootstrapHandler_Execute_WhenNoGitCommitter_ExpectSkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))

	// No WithGitCommitter option — gitCommitter is nil.
	handler := application.NewBootstrapHandler(
		newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{},
		&fakeFileWriter{}, &fakeContentProvider{},
	)

	session, err := handler.Preview(dir)
	require.NoError(t, err)
	_, err = handler.Confirm(session.SessionID())
	require.NoError(t, err)
	_, err = handler.Execute(session.SessionID())
	require.NoError(t, err)
	// No panic, no error — gracefully skipped.
}

func TestBootstrapHandler_Execute_WhenNotGitRepo_ExpectSkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))

	gc := &fakeGitCommitter{hasGit: false}
	handler := application.NewBootstrapHandler(
		newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{},
		&fakeFileWriter{}, &fakeContentProvider{},
		application.WithGitCommitter(gc),
	)

	session, err := handler.Preview(dir)
	require.NoError(t, err)
	_, err = handler.Confirm(session.SessionID())
	require.NoError(t, err)
	_, err = handler.Execute(session.SessionID())
	require.NoError(t, err)

	assert.Empty(t, gc.stagedPaths, "should not stage when not a git repo")
	assert.Empty(t, gc.commitMsg, "should not commit when not a git repo")
}

func TestBootstrapHandler_Execute_WhenAllFilesSkipped_ExpectNoCommit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))

	// Pre-create ALL planned files so every action is SKIP.
	for _, planned := range []string{
		"docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md",
		"AGENTS.md", ".alty/config.toml", ".alty/knowledge/_index.toml",
		".alty/maintenance/doc-registry.toml",
	} {
		full := filepath.Join(dir, planned)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte("existing"), 0o644))
	}

	gc := &fakeGitCommitter{hasGit: true}
	handler := application.NewBootstrapHandler(
		newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{},
		&fakeFileWriter{}, &fakeContentProvider{},
		application.WithGitCommitter(gc),
	)

	session, err := handler.Preview(dir)
	require.NoError(t, err)
	_, err = handler.Confirm(session.SessionID())
	require.NoError(t, err)
	_, err = handler.Execute(session.SessionID())
	require.NoError(t, err)

	assert.Empty(t, gc.stagedPaths, "should not stage when no files created")
	assert.Empty(t, gc.commitMsg, "should not commit when no files created")
}

func TestBootstrapHandler_Execute_WhenStageError_ExpectError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("idea"), 0o644))

	gc := &fakeGitCommitter{hasGit: true, stageErr: fmt.Errorf("git add failed")}
	handler := application.NewBootstrapHandler(
		newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{},
		&fakeFileWriter{}, &fakeContentProvider{},
		application.WithGitCommitter(gc),
	)

	session, err := handler.Preview(dir)
	require.NoError(t, err)
	_, err = handler.Confirm(session.SessionID())
	require.NoError(t, err)
	_, err = handler.Execute(session.SessionID())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "staging files")
}
