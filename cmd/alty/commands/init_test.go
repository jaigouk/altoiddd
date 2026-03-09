package commands_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/cmd/alty/commands"
	"github.com/alty-cli/alty/internal/composition"
	rescueapp "github.com/alty-cli/alty/internal/rescue/application"
	rescuedomain "github.com/alty-cli/alty/internal/rescue/domain"
	sharedapp "github.com/alty-cli/alty/internal/shared/application"
	"github.com/alty-cli/alty/internal/shared/domain/identity"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// --- Mocks ---

// mockProjectScan implements rescueapp.ProjectScan for testing.
type mockProjectScan struct {
	scanResult rescuedomain.ProjectScan
	scanErr    error
}

func (m *mockProjectScan) Scan(_ context.Context, _ string, _ vo.StackProfile) (rescuedomain.ProjectScan, error) {
	return m.scanResult, m.scanErr
}

// mockGitOps implements rescueapp.GitOps for testing.
type mockGitOps struct {
	hasGit       bool
	hasGitErr    error
	isClean      bool
	isCleanErr   error
	branchExists bool
	branchErr    error
	createErr    error
}

func (m *mockGitOps) HasGit(_ context.Context, _ string) (bool, error) {
	return m.hasGit, m.hasGitErr
}

func (m *mockGitOps) IsClean(_ context.Context, _ string) (bool, error) {
	return m.isClean, m.isCleanErr
}

func (m *mockGitOps) BranchExists(_ context.Context, _ string, _ string) (bool, error) {
	return m.branchExists, m.branchErr
}

func (m *mockGitOps) CreateBranch(_ context.Context, _ string, _ string) error {
	return m.createErr
}

// mockFileWriter implements sharedapp.FileWriter for testing.
type mockFileWriter struct {
	writtenFiles map[string]string
	writeErr     error
}

func newMockFileWriter() *mockFileWriter {
	return &mockFileWriter{writtenFiles: make(map[string]string)}
}

func (m *mockFileWriter) WriteFile(_ context.Context, path, content string) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.writtenFiles[path] = content
	return nil
}

// mockEventPublisher implements sharedapp.EventPublisher for testing.
type mockEventPublisher struct{}

func (m *mockEventPublisher) Publish(_ context.Context, _ any) error {
	return nil
}

// --- Helper ---

func newTestRescueHandler(
	projectScan rescueapp.ProjectScan,
	gitOps rescueapp.GitOps,
	fileWriter sharedapp.FileWriter,
	publisher sharedapp.EventPublisher,
) *rescueapp.RescueHandler {
	return rescueapp.NewRescueHandler(projectScan, gitOps, fileWriter, publisher)
}

// --- Tests ---

func TestRunRescue_PrintsGapReport(t *testing.T) {
	// Setup: project with gaps (missing docs).
	scan := rescuedomain.NewProjectScan(
		".",
		[]string{},                    // no existing docs
		[]string{".claude/CLAUDE.md"}, // has CLAUDE.md
		[]string{},
		false, // no knowledge dir
		false, // no agents.md
		true,  // has git
		false, // no alty config
		false, // no maintenance dir
	)

	gitOps := &mockGitOps{
		hasGit:       true,
		isClean:      true,
		branchExists: false,
	}
	projectScan := &mockProjectScan{scanResult: scan}
	fileWriter := newMockFileWriter()
	publisher := &mockEventPublisher{}

	handler := newTestRescueHandler(projectScan, gitOps, fileWriter, publisher)
	app := &composition.App{RescueHandler: handler}

	cmd := commands.NewInitCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--existing"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should print gap report with missing docs.
	assert.Contains(t, output, "docs/PRD.md")
	assert.Contains(t, output, "docs/DDD.md")
	assert.Contains(t, output, "docs/ARCHITECTURE.md")
}

func TestRunRescue_PreconditionFailure_NotGitRepo(t *testing.T) {
	gitOps := &mockGitOps{
		hasGit: false, // Not a git repo.
	}
	projectScan := &mockProjectScan{}
	fileWriter := newMockFileWriter()
	publisher := &mockEventPublisher{}

	handler := newTestRescueHandler(projectScan, gitOps, fileWriter, publisher)
	app := &composition.App{RescueHandler: handler}

	cmd := commands.NewInitCmd(app)
	var buf bytes.Buffer
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--existing"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
}

func TestRunRescue_PreconditionFailure_DirtyTree(t *testing.T) {
	gitOps := &mockGitOps{
		hasGit:  true,
		isClean: false, // Dirty working tree.
	}
	projectScan := &mockProjectScan{}
	fileWriter := newMockFileWriter()
	publisher := &mockEventPublisher{}

	handler := newTestRescueHandler(projectScan, gitOps, fileWriter, publisher)
	app := &composition.App{RescueHandler: handler}

	cmd := commands.NewInitCmd(app)
	var buf bytes.Buffer
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--existing"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dirty")
}

func TestRunRescue_PreconditionFailure_BranchExists(t *testing.T) {
	gitOps := &mockGitOps{
		hasGit:       true,
		isClean:      true,
		branchExists: true, // Branch already exists.
	}
	projectScan := &mockProjectScan{}
	fileWriter := newMockFileWriter()
	publisher := &mockEventPublisher{}

	handler := newTestRescueHandler(projectScan, gitOps, fileWriter, publisher)
	app := &composition.App{RescueHandler: handler}

	cmd := commands.NewInitCmd(app)
	var buf bytes.Buffer
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--existing"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRunRescue_NoGaps_PrintsSuccess(t *testing.T) {
	// Setup: compliant project with all required files.
	scan := rescuedomain.NewProjectScan(
		".",
		[]string{"docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md"}, // all docs present
		[]string{".claude/CLAUDE.md", ".alty/config.toml"},             // all configs
		[]string{},
		true, // has knowledge dir
		true, // has agents.md
		true, // has git
		true, // has alty config
		true, // has maintenance dir
	)

	gitOps := &mockGitOps{
		hasGit:       true,
		isClean:      true,
		branchExists: false,
	}
	projectScan := &mockProjectScan{scanResult: scan}
	fileWriter := newMockFileWriter()
	publisher := &mockEventPublisher{}

	handler := newTestRescueHandler(projectScan, gitOps, fileWriter, publisher)
	app := &composition.App{RescueHandler: handler}

	cmd := commands.NewInitCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--existing"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should print success message, no files written.
	assert.Contains(t, strings.ToLower(output), "no gaps")
	assert.Empty(t, fileWriter.writtenFiles)
}

func TestRunRescue_DryRun_NoFilesWritten(t *testing.T) {
	// Setup: project with gaps.
	scan := rescuedomain.NewProjectScan(
		".",
		[]string{},                    // no existing docs
		[]string{".claude/CLAUDE.md"}, // has CLAUDE.md
		[]string{},
		false, false, true, false, false,
	)

	gitOps := &mockGitOps{
		hasGit:       true,
		isClean:      true,
		branchExists: false,
	}
	projectScan := &mockProjectScan{scanResult: scan}
	fileWriter := newMockFileWriter()
	publisher := &mockEventPublisher{}

	handler := newTestRescueHandler(projectScan, gitOps, fileWriter, publisher)
	app := &composition.App{RescueHandler: handler}

	cmd := commands.NewInitCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--existing", "--dry-run"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should show plan but NOT write files.
	assert.Contains(t, output, "docs/PRD.md")
	assert.Empty(t, fileWriter.writtenFiles, "dry-run should not write files")
}

func TestRunRescue_ExecutesPlan_WritesFiles(t *testing.T) {
	// Setup: project with gaps, no dry-run.
	scan := rescuedomain.NewProjectScan(
		".",
		[]string{},                    // no existing docs
		[]string{".claude/CLAUDE.md"}, // has CLAUDE.md
		[]string{},
		true, // has knowledge dir
		true, // has agents.md (so it won't create one)
		true, // has git
		true, // has alty config
		true, // has maintenance dir
	)

	gitOps := &mockGitOps{
		hasGit:       true,
		isClean:      true,
		branchExists: false,
	}
	projectScan := &mockProjectScan{scanResult: scan}
	fileWriter := newMockFileWriter()
	publisher := &mockEventPublisher{}

	handler := newTestRescueHandler(projectScan, gitOps, fileWriter, publisher)
	app := &composition.App{RescueHandler: handler}

	cmd := commands.NewInitCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--existing"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Should have written files for missing docs.
	assert.NotEmpty(t, fileWriter.writtenFiles)
	// Check that at least one of the required docs was created.
	foundDoc := false
	for path := range fileWriter.writtenFiles {
		if strings.Contains(path, "PRD.md") || strings.Contains(path, "DDD.md") || strings.Contains(path, "ARCHITECTURE.md") {
			foundDoc = true
			break
		}
	}
	assert.True(t, foundDoc, "should have written at least one required doc")
}

// Compile-time interface checks.
var (
	_ rescueapp.ProjectScan    = (*mockProjectScan)(nil)
	_ rescueapp.GitOps         = (*mockGitOps)(nil)
	_ sharedapp.FileWriter     = (*mockFileWriter)(nil)
	_ sharedapp.EventPublisher = (*mockEventPublisher)(nil)
)

// Suppress unused import warning.
var (
	_ = identity.NewID
	_ = errors.New
)
