package commands_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/cmd/alty/commands"
	"github.com/alty-cli/alty/internal/composition"
)

const testDDDContent = `# DDD

## 3. Subdomain Classification

### Summary

| Subdomain | Type | Rationale | Architecture Approach |
|-----------|------|-----------|----------------------|
| Orders | **Core** | Main differentiator | Hexagonal |

## 4. Bounded Contexts

### Context: Orders

**Responsibility:** Owns order lifecycle management
`

func TestNewImportCmd_WhenValidDocs_PrintsSummary(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "DDD.md"), []byte(testDDDContent), 0o644))

	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	cmd := commands.NewImportCmd(app)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--docs-dir", docsDir})

	err = cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Imported 1 bounded context(s)")
	assert.Contains(t, output, "Orders (core)")
}

func TestNewImportCmd_WhenNoDocs_ReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	cmd := commands.NewImportCmd(app)
	cmd.SetArgs([]string{"--docs-dir", dir})

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DDD.md")
}
