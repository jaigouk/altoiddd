// Package application defines ports for the ToolTranslation bounded context.
package application

import (
	"context"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
	ttdomain "github.com/alto-cli/alto/internal/tooltranslation/domain"
)

// ConfigGeneration generates tool-native configurations from a domain model
// for AI coding tools (Claude Code, Cursor, etc.).
type ConfigGeneration interface {
	// Generate generates tool-native configurations for the specified tools.
	Generate(ctx context.Context, model *ddd.DomainModel, tools []ttdomain.SupportedTool, outputDir string) error
}

// PersonaManager lists and generates AI agent persona configurations
// for supported coding tools.
type PersonaManager interface {
	// ListPersonas lists all available agent persona definitions.
	ListPersonas(ctx context.Context) ([]*vo.PersonaDefinition, error)

	// Generate generates persona configuration files for specified tools.
	Generate(ctx context.Context, personaName string, tools []string, outputDir string) error
}

// WorkflowAssetGeneration renders alto-scaffold/commands/*.md workflow assets
// into tool-native command formats. Sibling port to ConfigGeneration
// (which renders a DomainModel). Distinct concern — file-source-based,
// not model-based.
//
// Implementations read every primary `<name>.md` under sourceDir (skipping
// `<name>.project.md` overlay siblings), merge in any matching overlay,
// transform the body (bash-block stripping, template inlining), and emit
// per-tool command files under outputDir.
type WorkflowAssetGeneration interface {
	// GenerateFromAssets renders all workflow assets under sourceDir into
	// tool-native command files under outputDir. Non-fatal per-asset
	// errors (e.g. ErrInvocationProtectionNotSupported,
	// ErrMissingTemplate) are aggregated and returned alongside any
	// successfully written files.
	GenerateFromAssets(ctx context.Context, sourceDir string, outputDir string) error
}
