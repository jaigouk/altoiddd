package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	sharedapp "github.com/alty-cli/alty/internal/shared/application"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// PersonaPreview holds rendered persona content ready for user review.
type PersonaPreview struct {
	Persona    *vo.PersonaDefinition
	Tool       string
	Content    string
	TargetPath string
	Summary    string
}

// PersonaHandler orchestrates persona listing, preview, and file writing.
type PersonaHandler struct {
	fileWriter sharedapp.FileWriter
}

// NewPersonaHandler creates a new PersonaHandler.
func NewPersonaHandler(fileWriter sharedapp.FileWriter) *PersonaHandler {
	return &PersonaHandler{fileWriter: fileWriter}
}

// ListPersonas returns all registered persona definitions.
func (h *PersonaHandler) ListPersonas() []*vo.PersonaDefinition {
	registry := vo.PersonaRegistry()
	result := make([]*vo.PersonaDefinition, 0, len(registry))
	for _, ptype := range vo.AllPersonaTypes() {
		if defn, ok := registry[ptype]; ok {
			result = append(result, defn)
		}
	}
	return result
}

// BuildPreview builds a preview for the given persona and tool without writing.
func (h *PersonaHandler) BuildPreview(personaName, tool string) (*PersonaPreview, error) {
	persona, err := resolvePersona(personaName)
	if err != nil {
		return nil, err
	}
	if err := validateTool(tool); err != nil {
		return nil, err
	}

	slug := strings.ToLower(strings.ReplaceAll(persona.Name(), " ", "-"))
	targetPaths := vo.ToolTargetPaths()
	targetPath := strings.ReplaceAll(targetPaths[tool], "{name}", slug)

	content := persona.InstructionsTemplate()

	summary := fmt.Sprintf("Persona: %s (%s)\nTool: %s\nTarget: %s",
		persona.Name(), persona.Register(), tool, targetPath)

	return &PersonaPreview{
		Persona:    persona,
		Tool:       tool,
		Content:    content,
		TargetPath: targetPath,
		Summary:    summary,
	}, nil
}

// ApproveAndWrite writes a previously previewed persona configuration to disk.
func (h *PersonaHandler) ApproveAndWrite(
	ctx context.Context,
	preview *PersonaPreview,
	outputDir string,
) error {
	target := filepath.Join(outputDir, preview.TargetPath)
	return h.fileWriter.WriteFile(ctx, target, preview.Content)
}

func resolvePersona(personaName string) (*vo.PersonaDefinition, error) {
	lower := strings.ToLower(strings.TrimSpace(personaName))
	registry := vo.PersonaRegistry()

	// Match by display name (case-insensitive)
	for _, defn := range registry {
		if strings.ToLower(defn.Name()) == lower {
			return defn, nil
		}
	}

	// Match by PersonaType value
	for ptype, defn := range registry {
		if string(ptype) == lower {
			return defn, nil
		}
	}

	names := make([]string, 0, len(registry))
	for _, defn := range registry {
		names = append(names, "'"+defn.Name()+"'")
	}
	return nil, fmt.Errorf("Unknown persona '%s'. Valid personas: %s: %w",
		personaName, strings.Join(names, ", "), domainerrors.ErrInvariantViolation)
}

func validateTool(tool string) error {
	for _, supported := range vo.SupportedTools() {
		if supported == tool {
			return nil
		}
	}
	return fmt.Errorf("Unsupported tool '%s'. Valid tools: %s: %w",
		tool, strings.Join(vo.SupportedTools(), ", "), domainerrors.ErrInvariantViolation)
}
