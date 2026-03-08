package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTierPolicy_AllowsPermittedTool(t *testing.T) {
	t.Parallel()
	policy := NewTierPolicy(TierWrite, map[string]Tier{
		"detect_tools": TierRead,
		"init_project": TierWrite,
	})

	assert.True(t, policy.IsAllowed("detect_tools"), "read tool should be allowed at write tier")
	assert.True(t, policy.IsAllowed("init_project"), "write tool should be allowed at write tier")
}

func TestTierPolicy_RejectsRestrictedTool(t *testing.T) {
	t.Parallel()
	policy := NewTierPolicy(TierRead, map[string]Tier{
		"detect_tools":  TierRead,
		"init_project":  TierWrite,
		"check_quality": TierExec,
	})

	assert.True(t, policy.IsAllowed("detect_tools"), "read tool at read tier should be allowed")
	assert.False(t, policy.IsAllowed("init_project"), "write tool at read tier should be rejected")
	assert.False(t, policy.IsAllowed("check_quality"), "exec tool at read tier should be rejected")
}

func TestTierPolicy_DefaultAllowsUnknown(t *testing.T) {
	t.Parallel()
	policy := NewTierPolicy(TierRead, map[string]Tier{
		"check_quality": TierExec,
	})

	assert.True(t, policy.IsAllowed("unknown_tool"), "unlisted tools should be allowed by default")
}

func TestTierPolicy_ExecAllowsAll(t *testing.T) {
	t.Parallel()
	policy := NewTierPolicy(TierExec, map[string]Tier{
		"detect_tools":  TierRead,
		"init_project":  TierWrite,
		"check_quality": TierExec,
	})

	assert.True(t, policy.IsAllowed("detect_tools"))
	assert.True(t, policy.IsAllowed("init_project"))
	assert.True(t, policy.IsAllowed("check_quality"))
}

func TestTier_Ordering(t *testing.T) {
	t.Parallel()
	assert.Less(t, TierRead, TierWrite)
	assert.Less(t, TierWrite, TierExec)
}
