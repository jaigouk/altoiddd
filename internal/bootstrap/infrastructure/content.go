// Package infrastructure provides adapters for the Bootstrap bounded context.
package infrastructure

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alto-cli/alto/internal/bootstrap/application"
	"github.com/alto-cli/alto/internal/bootstrap/domain"
)

// Compile-time check that ContentProviderAdapter satisfies ContentProvider.
var _ application.ContentProvider = (*ContentProviderAdapter)(nil)

// ContentProviderAdapter implements ContentProvider using plain string generators.
type ContentProviderAdapter struct{}

// ContentFor returns the generated content for a planned file path.
func (c *ContentProviderAdapter) ContentFor(path string, config domain.ProjectConfig) string {
	switch path {
	case "alto-scaffold/config.toml":
		return AltoConfigContent(config)
	case "alto-scaffold/knowledge/_index.toml":
		return KnowledgeIndexContent()
	case "alto-scaffold/maintenance/doc-registry.toml":
		return DocRegistryContent()
	default:
		// AGENTS.md is the only remaining default-case file.
		// docs/PRD.md, DDD.md, ARCHITECTURE.md are owned by the artifact pipeline.
		stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		return fmt.Sprintf("# %s\n\n> TODO: Fill in content.\n", stem)
	}
}

// AltoConfigContent returns valid TOML for alto-scaffold/config.toml.
func AltoConfigContent(config domain.ProjectConfig) string {
	var b strings.Builder
	b.WriteString("# alto project configuration\n\n[project]\n")
	fmt.Fprintf(&b, "name = %q\n", config.Name())

	if config.Language() != "" {
		fmt.Fprintf(&b, "language = %q\n", config.Language())
	}

	if config.ModulePath() != "" {
		fmt.Fprintf(&b, "module_path = %q\n", config.ModulePath())
	}

	b.WriteString("\n[tools]\n")
	tools := config.DetectedTools()
	if len(tools) == 0 {
		b.WriteString("detected = []\n")
	} else {
		quoted := make([]string, len(tools))
		for i, t := range tools {
			quoted[i] = fmt.Sprintf("%q", t)
		}
		fmt.Fprintf(&b, "detected = [%s]\n", strings.Join(quoted, ", "))
	}

	b.WriteString("\n[discovery]\ncompleted = false\n")

	b.WriteString(`
# [llm]
# provider = ""
# model = ""
# api_key_env = ""
# Uncomment and configure when LLM features are enabled.
`)

	return b.String()
}

// KnowledgeIndexContent returns valid TOML for alto-scaffold/knowledge/_index.toml.
func KnowledgeIndexContent() string {
	return `# alto knowledge base index
#
# Sections map to subdirectories under alto-scaffold/knowledge/.
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

// DocRegistryContent returns valid TOML for alto-scaffold/maintenance/doc-registry.toml.
// NOTE: The docs/PRD.md, docs/DDD.md, docs/ARCHITECTURE.md entries below are
// forward references — these files do not exist at bootstrap time. They are
// created later by the artifact pipeline (ArtifactGenerationHandler). The
// registry is pre-populated so doc-health monitoring is ready once artifacts
// are generated. See alty-cli-18l.
func DocRegistryContent() string {
	return `# alto document registry
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
