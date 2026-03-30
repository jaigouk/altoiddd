package commands

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/composition"
	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	"github.com/alto-cli/alto/internal/shared/infrastructure/eventbus"
)

// stubToolDetector is a minimal ToolDetector for precondition tests.
type stubToolDetector struct{}

func (s *stubToolDetector) Detect(_ string) ([]string, error) {
	return nil, nil
}

func (s *stubToolDetector) ScanConflicts(_ string) ([]domain.SettingsConflict, error) {
	return nil, nil
}

var _ application.ToolDetector = (*stubToolDetector)(nil)

func TestGuide_WhenMissingConfig_ReturnsInitError(t *testing.T) {
	tmpDir := t.TempDir()
	altoDir := filepath.Join(tmpDir, ".alto")
	require.NoError(t, os.MkdirAll(altoDir, 0o755))

	t.Chdir(tmpDir)

	app := &composition.App{}
	cmd := NewGuideCmd(app)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestGuide_WhenPresentConfig_NoInitError(t *testing.T) {
	tmpDir := t.TempDir()
	altoDir := filepath.Join(tmpDir, ".alto")
	require.NoError(t, os.MkdirAll(altoDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(altoDir, "config.toml"), []byte{}, 0o644))

	t.Chdir(tmpDir)

	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	publisher := eventbus.NewPublisher(bus)
	discoveryHandler := application.NewDiscoveryHandler(publisher)
	detectionHandler := application.NewDetectionHandler(&stubToolDetector{})
	app := &composition.App{
		DiscoveryHandler: discoveryHandler,
		DetectionHandler: detectionHandler,
	}

	cmd := NewGuideCmd(app)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		assert.NotContains(t, err.Error(), "not initialized")
	}
}

func TestGuideIngest_WhenMissingConfig_ReturnsInitError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	altoDir := filepath.Join(tmpDir, ".alto")
	require.NoError(t, os.MkdirAll(altoDir, 0o755))

	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	publisher := eventbus.NewPublisher(bus)
	sessionRepo := infrastructure.NewFileSystemSessionRepository(altoDir)
	handler := application.NewDiscoveryHandler(publisher, application.WithSessionRepository(sessionRepo))

	session, err := handler.StartSession("test readme")
	require.NoError(t, err)
	require.NoError(t, sessionRepo.Save(context.Background(), session))

	app := &composition.App{DiscoveryHandler: handler}

	var out bytes.Buffer
	stdinReader := bytes.NewReader([]byte{})
	err = runGuideAgentIngestFromReader(context.Background(), app, stdinReader, altoDir, &out)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestGuideLegacy_WhenMissingConfig_ReturnsInitError(t *testing.T) {
	tmpDir := t.TempDir()
	altoDir := filepath.Join(tmpDir, ".alto")
	require.NoError(t, os.MkdirAll(altoDir, 0o755))

	t.Chdir(tmpDir)

	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	publisher := eventbus.NewPublisher(bus)
	handler := application.NewDiscoveryHandler(publisher)
	app := &composition.App{DiscoveryHandler: handler}

	cmd := NewGuideCmd(app)
	cmd.SetArgs([]string{"--legacy"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}
