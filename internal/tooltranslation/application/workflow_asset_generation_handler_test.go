package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ttapp "github.com/alto-cli/alto/internal/tooltranslation/application"
	ttdomain "github.com/alto-cli/alto/internal/tooltranslation/domain"
)

// stubWorkflowAssetGeneration is an in-memory test double. The handler is
// a thin orchestrator, so it can be exercised without filesystem I/O.
type stubWorkflowAssetGeneration struct {
	calls int
	err   error
}

func (s *stubWorkflowAssetGeneration) GenerateFromAssets(_ context.Context, _ string, _ string) error {
	s.calls++
	return s.err
}

func TestWorkflowAssetGenerationHandler_SupportedTools_ReturnsRegistryKeys(t *testing.T) {
	t.Parallel()
	stub := &stubWorkflowAssetGeneration{}
	h := ttapp.NewWorkflowAssetGenerationHandler(
		map[ttdomain.SupportedTool]ttapp.WorkflowAssetGeneration{
			ttdomain.ToolOpenCode: stub,
		},
	)
	tools := h.SupportedTools()
	assert.Equal(t, []ttdomain.SupportedTool{ttdomain.ToolOpenCode}, tools)
}

func TestWorkflowAssetGenerationHandler_GenerateForTools_DelegatesToAdapter(t *testing.T) {
	t.Parallel()
	stub := &stubWorkflowAssetGeneration{}
	h := ttapp.NewWorkflowAssetGenerationHandler(
		map[ttdomain.SupportedTool]ttapp.WorkflowAssetGeneration{
			ttdomain.ToolOpenCode: stub,
		},
	)
	err := h.GenerateForTools(context.TODO(), []ttdomain.SupportedTool{ttdomain.ToolOpenCode}, "src", "dst")
	require.NoError(t, err)
	assert.Equal(t, 1, stub.calls)
}

func TestWorkflowAssetGenerationHandler_GenerateForTools_UnknownTool_ReturnsError(t *testing.T) {
	t.Parallel()
	h := ttapp.NewWorkflowAssetGenerationHandler(
		map[ttdomain.SupportedTool]ttapp.WorkflowAssetGeneration{},
	)
	err := h.GenerateForTools(context.TODO(), []ttdomain.SupportedTool{ttdomain.ToolOpenCode}, "src", "dst")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported tool")
}

func TestWorkflowAssetGenerationHandler_GenerateForTools_AggregatesErrors(t *testing.T) {
	t.Parallel()
	failing := &stubWorkflowAssetGeneration{err: ttdomain.ErrMissingTemplate}
	h := ttapp.NewWorkflowAssetGenerationHandler(
		map[ttdomain.SupportedTool]ttapp.WorkflowAssetGeneration{
			ttdomain.ToolOpenCode: failing,
		},
	)
	err := h.GenerateForTools(context.TODO(), []ttdomain.SupportedTool{ttdomain.ToolOpenCode}, "src", "dst")
	require.Error(t, err)
	assert.ErrorIs(t, err, ttdomain.ErrMissingTemplate)
}

func TestWorkflowAssetGenerationHandler_EmptyRegistry_EmptyToolList_Success(t *testing.T) {
	t.Parallel()
	h := ttapp.NewWorkflowAssetGenerationHandler(
		map[ttdomain.SupportedTool]ttapp.WorkflowAssetGeneration{},
	)
	assert.NoError(t, h.GenerateForTools(context.TODO(), nil, "src", "dst"))
}
