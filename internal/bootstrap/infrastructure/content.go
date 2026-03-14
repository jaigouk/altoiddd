// Package infrastructure provides adapters for the Bootstrap bounded context.
package infrastructure

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alty-cli/alty/internal/bootstrap/application"
)

// Compile-time check that ContentProviderAdapter satisfies ContentProvider.
var _ application.ContentProvider = (*ContentProviderAdapter)(nil)

// ContentProviderAdapter implements ContentProvider using plain string generators.
type ContentProviderAdapter struct{}

// ContentFor returns the generated content for a planned file path.
func (c *ContentProviderAdapter) ContentFor(path string, projectName string) string {
	switch path {
	case ".alty/config.toml":
		return AltyConfigContent(projectName)
	case ".alty/knowledge/_index.toml":
		return KnowledgeIndexContent()
	case ".alty/maintenance/doc-registry.toml":
		return DocRegistryContent()
	default:
		// Doc stubs (PRD.md, DDD.md, ARCHITECTURE.md, AGENTS.md)
		stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		return fmt.Sprintf("# %s\n\n> TODO: Fill in content.\n", stem)
	}
}

// AltyConfigContent returns valid TOML for .alty/config.toml.
func AltyConfigContent(projectName string) string {
	return fmt.Sprintf(`# alty project configuration
project_name = %q
version = "0.1.0"
`, projectName)
}

// KnowledgeIndexContent returns valid TOML for .alty/knowledge/_index.toml.
func KnowledgeIndexContent() string {
	return `# alty knowledge base index
#
# Sections map to subdirectories under .alty/knowledge/.
# Each section contains RLM-addressable documents.

[knowledge]
version = 1

[[sections]]
name = "ddd"
description = "DDD patterns, tactical and strategic references"

[[sections]]
name = "tools"
description = "AI coding tool conventions (versioned per tool)"

[[sections]]
name = "conventions"
description = "TDD, SOLID, quality gate references"
`
}

// DocRegistryContent returns valid TOML for .alty/maintenance/doc-registry.toml.
func DocRegistryContent() string {
	return `# alty document registry
#
# Tracks which docs to monitor, their owners, and review cadence.

[registry]
version = 1

[[docs]]
path = "docs/PRD.md"
owner = "product"
review_days = 90

[[docs]]
path = "docs/DDD.md"
owner = "architecture"
review_days = 90

[[docs]]
path = "docs/ARCHITECTURE.md"
owner = "architecture"
review_days = 90
`
}
