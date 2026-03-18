package commands_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/cmd/alto/commands"
	"github.com/alto-cli/alto/internal/composition"
	rescueapp "github.com/alto-cli/alto/internal/rescue/application"
	rescuedomain "github.com/alto-cli/alto/internal/rescue/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type gapMockProfileDetector struct {
	profile vo.StackProfile
}

func (m *gapMockProfileDetector) DetectProfile(_ string) vo.StackProfile {
	return m.profile
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestGapCmd_WhenGapsExist_PrintsReport(t *testing.T) {
	// Given: a project with missing docs
	scan := rescuedomain.NewProjectScan(
		".",
		nil, // no docs
		nil, // no configs
		nil,
		false, false, false, false, false,
	)
	scanner := &mockProjectScan{scanResult: scan}
	handler := rescueapp.NewGapQueryHandler(scanner, &gapMockProfileDetector{})
	app := &composition.App{GapQueryHandler: handler}

	cmd := commands.NewGapCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"."})

	// When
	err := cmd.Execute()

	// Then — should error because required gaps exist
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required gaps found")

	output := buf.String()
	assert.Contains(t, output, "Gap Analysis Report")
	assert.Contains(t, output, "docs/PRD.md")
}

func TestGapCmd_WhenNoGaps_PrintsCompliant(t *testing.T) {
	// Given: a fully compliant project
	scan := rescuedomain.NewProjectScan(
		".",
		[]string{"docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md", "AGENTS.md"},
		[]string{".claude/CLAUDE.md", ".alto/config.toml"},
		nil,
		true, true, true, true, true,
	)
	scanner := &mockProjectScan{scanResult: scan}
	handler := rescueapp.NewGapQueryHandler(scanner, &gapMockProfileDetector{})
	app := &composition.App{GapQueryHandler: handler}

	cmd := commands.NewGapCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"."})

	// When
	err := cmd.Execute()

	// Then
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "No gaps found")
}

func TestGapCmd_WhenOnlyRecommended_ReturnsNoError(t *testing.T) {
	// Given: project with all required files but missing recommended .alto/config.toml
	scan := rescuedomain.NewProjectScan(
		".",
		[]string{"docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md", "AGENTS.md"},
		[]string{".claude/CLAUDE.md"},
		nil,
		true, true, true, false, true, // hasAltoConfig = false
	)
	scanner := &mockProjectScan{scanResult: scan}
	handler := rescueapp.NewGapQueryHandler(scanner, &gapMockProfileDetector{})
	app := &composition.App{GapQueryHandler: handler}

	cmd := commands.NewGapCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"."})

	// When
	err := cmd.Execute()

	// Then — no error because gaps are only recommended, not required
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, ".alto/config.toml")
	assert.Contains(t, output, "recommended")
}
