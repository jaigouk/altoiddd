package application_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/bootstrap/application"
	"github.com/alty-cli/alty/internal/bootstrap/domain"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

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
			handler := application.NewBootstrapHandler(fake, osFileChecker{}, &fakePublisher{})

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

		handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{})

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

		handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{})

		session, err := handler.Preview(dir)
		require.NoError(t, err)

		session, err = handler.Cancel(session.SessionID())
		require.NoError(t, err)
		assert.Equal(t, domain.SessionStatusCancelled, session.Status())
	})
}

func TestBootstrapHandler_SessionNotFound(t *testing.T) {
	t.Parallel()

	handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{})

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

		handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{})
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

		handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{})
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

	handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, &fakePublisher{})
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
	handler := application.NewBootstrapHandler(newFakeToolDetection(nil, nil), osFileChecker{}, pub)

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

	handler := application.NewBootstrapHandler(newFakeToolDetection(nil, []string{"Global cursor setting overrides local"}), osFileChecker{}, &fakePublisher{})
	session, err := handler.Preview(dir)
	require.NoError(t, err)
	preview := session.Preview()
	require.NotNil(t, preview)
	assert.Equal(t, []string{"Global cursor setting overrides local"}, preview.ConflictDescriptions())
}
