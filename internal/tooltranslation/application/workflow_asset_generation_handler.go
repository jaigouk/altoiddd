package application

import (
	"context"
	"errors"
	"fmt"
	"maps"

	ttdomain "github.com/alto-cli/alto/internal/tooltranslation/domain"
)

// workflowAssetAdapterRegistry maps each SupportedTool to a
// WorkflowAssetGeneration adapter. Mirrors the adapterRegistry pattern in
// config_generation_handler.go (line 17-22): factories registered once,
// looked up by tool. Today the registry contains one entry (OpenCode) per
// epic §Scope Update; Cursor / Roo Code reintroduction is a registry edit.
type workflowAssetAdapterRegistry map[ttdomain.SupportedTool]WorkflowAssetGeneration

// WorkflowAssetGenerationHandler orchestrates per-tool workflow-asset
// rendering. Distinct from ConfigGenerationHandler (DomainModel-based) —
// this handler operates on a filesystem source under `.alto/commands/`.
type WorkflowAssetGenerationHandler struct {
	registry workflowAssetAdapterRegistry
}

// NewWorkflowAssetGenerationHandler constructs a handler from an explicit
// registry. The registry must be non-nil; an empty registry is allowed and
// makes every GenerateForTools call a no-op (handler ignores tools).
func NewWorkflowAssetGenerationHandler(
	registry map[ttdomain.SupportedTool]WorkflowAssetGeneration,
) *WorkflowAssetGenerationHandler {
	r := make(workflowAssetAdapterRegistry, len(registry))
	maps.Copy(r, registry)
	return &WorkflowAssetGenerationHandler{registry: r}
}

// SupportedTools returns the tools for which an adapter is registered.
// Order is unspecified — callers should not depend on it.
func (h *WorkflowAssetGenerationHandler) SupportedTools() []ttdomain.SupportedTool {
	tools := make([]ttdomain.SupportedTool, 0, len(h.registry))
	for t := range h.registry {
		tools = append(tools, t)
	}
	return tools
}

// GenerateForTools renders the assets under sourceDir for every tool in
// the request. The projectRoot argument is forwarded verbatim to each
// adapter; per-tool output subdirectories (`.opencode/commands/` etc.) are
// the adapter's responsibility.
//
// Errors from individual adapters are aggregated with errors.Join — a
// failure in one tool does not prevent rendering the others. A
// non-registered tool surfaces ttdomain.ErrInvocationProtectionNotSupported
// is reserved for the per-asset case; unknown tools return a wrapped
// invariant-violation error (caller has misconfigured the registry).
func (h *WorkflowAssetGenerationHandler) GenerateForTools(
	ctx context.Context,
	tools []ttdomain.SupportedTool,
	sourceDir string,
	projectRoot string,
) error {
	var errs []error
	for _, tool := range tools {
		adapter, ok := h.registry[tool]
		if !ok {
			errs = append(errs, fmt.Errorf("unsupported tool %q (registry size %d)", tool, len(h.registry)))
			continue
		}
		if err := adapter.GenerateFromAssets(ctx, sourceDir, projectRoot); err != nil {
			errs = append(errs, fmt.Errorf("tool %s: %w", tool, err))
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
