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
