package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	sharedapp "github.com/alty-cli/alty/internal/shared/application"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	ttdomain "github.com/alty-cli/alty/internal/tooltranslation/domain"
)

// adapterRegistry maps SupportedTool to their ToolAdapter factories.
var adapterRegistry = map[ttdomain.SupportedTool]func() ttdomain.ToolAdapter{
	ttdomain.ToolClaudeCode: func() ttdomain.ToolAdapter { return ttdomain.NewClaudeCodeAdapter() },
	ttdomain.ToolCursor:     func() ttdomain.ToolAdapter { return ttdomain.NewCursorAdapter() },
	ttdomain.ToolRooCode:    func() ttdomain.ToolAdapter { return ttdomain.NewRooCodeAdapter() },
	ttdomain.ToolOpenCode:   func() ttdomain.ToolAdapter { return ttdomain.NewOpenCodeAdapter() },
}

// ConfigPreview holds rendered tool configurations ready for user review.
type ConfigPreview struct {
	Configs []*ttdomain.ToolConfig
	Summary string
}

// ConfigGenerationHandler orchestrates tool config generation from a DomainModel.
type ConfigGenerationHandler struct {
	fileWriter sharedapp.FileWriter
	publisher  sharedapp.EventPublisher
}

// NewConfigGenerationHandler creates a new ConfigGenerationHandler.
func NewConfigGenerationHandler(fileWriter sharedapp.FileWriter, publisher sharedapp.EventPublisher) *ConfigGenerationHandler {
	return &ConfigGenerationHandler{fileWriter: fileWriter, publisher: publisher}
}

// BuildPreview generates tool configs for preview without writing files.
func (h *ConfigGenerationHandler) BuildPreview(
	model *ddd.DomainModel,
	tools []ttdomain.SupportedTool,
	profile vo.StackProfile,
) (*ConfigPreview, error) {
	if len(tools) == 0 {
		return nil, fmt.Errorf("no tools specified for config generation: %w",
			domainerrors.ErrInvariantViolation)
	}

	if profile == nil {
		profile = vo.PythonUvProfile{}
	}

	var configs []*ttdomain.ToolConfig
	var summaryLines []string
	summaryLines = append(summaryLines, "Config Generation Preview", "")

	for _, tool := range tools {
		factory, ok := adapterRegistry[tool]
		if !ok {
			return nil, fmt.Errorf("unsupported tool: %s: %w", tool, domainerrors.ErrInvariantViolation)
		}
		adapter := factory()
		config := ttdomain.NewToolConfig(tool)
		if err := config.BuildSections(model, adapter, profile); err != nil {
			return nil, fmt.Errorf("building sections for %s: %w", tool, err)
		}
		preview, err := config.Preview()
		if err != nil {
			return nil, fmt.Errorf("previewing config for %s: %w", tool, err)
		}
		configs = append(configs, config)
		summaryLines = append(summaryLines, preview, "")
	}

	return &ConfigPreview{
		Configs: configs,
		Summary: strings.Join(summaryLines, "\n"),
	}, nil
}

// ApproveAndWrite approves all configs (emitting domain events) and writes to disk.
func (h *ConfigGenerationHandler) ApproveAndWrite(
	ctx context.Context,
	preview *ConfigPreview,
	outputDir string,
) error {
	for _, config := range preview.Configs {
		if err := config.Approve(); err != nil {
			return fmt.Errorf("approving config for %s: %w", config.Tool(), err)
		}
		for _, section := range config.Sections() {
			target := filepath.Join(outputDir, section.FilePath())
			if err := h.fileWriter.WriteFile(ctx, target, section.Content()); err != nil {
				return fmt.Errorf("writing config file %s: %w", target, err)
			}
		}
		for _, event := range config.Events() {
			_ = h.publisher.Publish(ctx, event)
		}
	}
	return nil
}
