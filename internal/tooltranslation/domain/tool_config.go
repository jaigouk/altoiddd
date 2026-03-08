// Package domain provides the ToolTranslation bounded context's core domain model.
// It contains value objects and aggregate roots for generating tool-native configurations
// from a DomainModel: SupportedTool, ConfigSection, ToolConfig, and adapters.
package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	"github.com/alty-cli/alty/internal/shared/domain/identity"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// SupportedTool enumerates AI coding tools that alty can generate configurations for.
type SupportedTool string

// Supported tool constants.
const (
	ToolClaudeCode SupportedTool = "claude-code"
	ToolCursor     SupportedTool = "cursor"
	ToolRooCode    SupportedTool = "roo-code"
	ToolOpenCode   SupportedTool = "opencode"
)

// AllSupportedTools returns all valid SupportedTool values.
func AllSupportedTools() []SupportedTool {
	return []SupportedTool{ToolClaudeCode, ToolCursor, ToolRooCode, ToolOpenCode}
}

// ConfigSection is a single file section within a tool configuration.
type ConfigSection struct {
	filePath    string
	content     string
	sectionName string
}

// NewConfigSection creates a ConfigSection value object.
func NewConfigSection(filePath, content, sectionName string) ConfigSection {
	return ConfigSection{filePath: filePath, content: content, sectionName: sectionName}
}

// FilePath returns the output file path.
func (s ConfigSection) FilePath() string { return s.filePath }

// Content returns the section content.
func (s ConfigSection) Content() string { return s.content }

// SectionName returns the section name.
func (s ConfigSection) SectionName() string { return s.sectionName }

// ToolAdapter is the interface for translating a DomainModel into tool-native config sections.
type ToolAdapter interface {
	Translate(model *ddd.DomainModel, profile vo.StackProfile) []ConfigSection
}

// ConfigsGeneratedEvent is a domain event emitted when configs are approved.
type ConfigsGeneratedEvent struct {
	toolNames   []string
	outputPaths []string
}

func newConfigsGeneratedEvent(toolNames, outputPaths []string) ConfigsGeneratedEvent {
	tn := make([]string, len(toolNames))
	copy(tn, toolNames)
	op := make([]string, len(outputPaths))
	copy(op, outputPaths)
	return ConfigsGeneratedEvent{toolNames: tn, outputPaths: op}
}

// ToolNames returns a defensive copy.
func (e ConfigsGeneratedEvent) ToolNames() []string {
	out := make([]string, len(e.toolNames))
	copy(out, e.toolNames)
	return out
}

// OutputPaths returns a defensive copy.
func (e ConfigsGeneratedEvent) OutputPaths() []string {
	out := make([]string, len(e.outputPaths))
	copy(out, e.outputPaths)
	return out
}

// MarshalJSON implements json.Marshaler for event bus serialization.
func (e ConfigsGeneratedEvent) MarshalJSON() ([]byte, error) {
	type proxy struct {
		ToolNames   []string `json:"tool_names"`
		OutputPaths []string `json:"output_paths"`
	}
	data, err := json.Marshal(proxy{
		ToolNames:   e.toolNames,
		OutputPaths: e.outputPaths,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling ConfigsGeneratedEvent: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (e *ConfigsGeneratedEvent) UnmarshalJSON(data []byte) error {
	type proxy struct {
		ToolNames   []string `json:"tool_names"`
		OutputPaths []string `json:"output_paths"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling ConfigsGeneratedEvent: %w", err)
	}
	e.toolNames = p.ToolNames
	e.outputPaths = p.OutputPaths
	return nil
}

// ToolConfig is the aggregate root: generates and manages tool-native configuration sections.
type ToolConfig struct {
	configID string
	tool     SupportedTool
	sections []ConfigSection
	events   []ConfigsGeneratedEvent
	approved bool
}

// NewToolConfig creates a new ToolConfig aggregate root.
func NewToolConfig(tool SupportedTool) *ToolConfig {
	return &ToolConfig{
		configID: identity.NewID(),
		tool:     tool,
	}
}

// ConfigID returns the unique config identifier.
func (tc *ToolConfig) ConfigID() string { return tc.configID }

// Tool returns the supported tool.
func (tc *ToolConfig) Tool() SupportedTool { return tc.tool }

// Sections returns a defensive copy of all generated config sections.
func (tc *ToolConfig) Sections() []ConfigSection {
	out := make([]ConfigSection, len(tc.sections))
	copy(out, tc.sections)
	return out
}

// Events returns a defensive copy of domain events.
func (tc *ToolConfig) Events() []ConfigsGeneratedEvent {
	out := make([]ConfigsGeneratedEvent, len(tc.events))
	copy(out, tc.events)
	return out
}

// BuildSections generates config sections from a DomainModel using the given adapter.
func (tc *ToolConfig) BuildSections(model *ddd.DomainModel, adapter ToolAdapter, profile vo.StackProfile) error {
	if tc.approved {
		return fmt.Errorf("cannot regenerate sections on an approved config: %w",
			domainerrors.ErrInvariantViolation)
	}
	tc.sections = adapter.Translate(model, profile)
	return nil
}

// Preview returns a human-readable preview of generated sections.
func (tc *ToolConfig) Preview() (string, error) {
	if len(tc.sections) == 0 {
		return "", fmt.Errorf("no sections generated yet — call BuildSections() first: %w",
			domainerrors.ErrInvariantViolation)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Tool: %s\nTotal sections: %d\n\n", tc.tool, len(tc.sections))
	for _, s := range tc.sections {
		fmt.Fprintf(&b, "  %s: %s", s.sectionName, s.filePath)
	}
	return b.String(), nil
}

// Approve approves the config, emitting ConfigsGeneratedEvent.
func (tc *ToolConfig) Approve() error {
	if tc.approved {
		return fmt.Errorf("config already approved: %w", domainerrors.ErrInvariantViolation)
	}
	if len(tc.sections) == 0 {
		return fmt.Errorf("cannot approve config with no sections: %w",
			domainerrors.ErrInvariantViolation)
	}
	tc.approved = true
	paths := make([]string, len(tc.sections))
	for i, s := range tc.sections {
		paths[i] = s.filePath
	}
	tc.events = append(tc.events, newConfigsGeneratedEvent(
		[]string{string(tc.tool)}, paths,
	))
	return nil
}
