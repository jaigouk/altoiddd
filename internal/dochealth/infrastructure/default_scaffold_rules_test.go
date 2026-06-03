package infrastructure_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
	"github.com/alto-cli/alto/internal/dochealth/infrastructure"
)

// TestDefaultScaffoldRules_TemplateRenderabilityRule_IsSlot13 locks the
// alty-cli-zc9 registry contract: the TemplateRenderabilityRule MUST sit in
// slot 13 (index 12), AFTER LifecycleStalenessRule. Contract-5 from
// alty-cli-ihk owns slots 0..11; slot 12 is the alty-cli-fln staleness rule;
// slot 13 is this rule. Reordering breaks downstream expectations.
func TestDefaultScaffoldRules_TemplateRenderabilityRule_IsSlot13(t *testing.T) {
	t.Parallel()

	params, err := dochealthdomain.NewScaffoldParams(30, nil)
	require.NoError(t, err)

	rs := infrastructure.DefaultScaffoldRules(params)
	require.Len(t, rs, 13, "DefaultScaffoldRules must contain exactly 13 slots after alty-cli-zc9 append")

	// Slot 12 (the previously-last slot) must still be lifecycle_staleness.
	assert.Equal(t, "lifecycle_staleness", rs[11].Name(),
		"slot 12 must remain LifecycleStalenessRule — contract from alty-cli-fln")

	// Slot 13 (newly appended) is TemplateRenderabilityRule.
	assert.Equal(t, "template_renderability", rs[12].Name(),
		"slot 13 must be TemplateRenderabilityRule — alty-cli-zc9 append-only contract")
}

// TestDefaultScaffoldRules_LockedOrder_Slots1To12 asserts the Contract-5
// locked slot order from alty-cli-ihk, guarding against accidental
// reordering when future rules are added.
func TestDefaultScaffoldRules_LockedOrder_Slots1To12(t *testing.T) {
	t.Parallel()

	params, err := dochealthdomain.NewScaffoldParams(30, nil)
	require.NoError(t, err)

	rs := infrastructure.DefaultScaffoldRules(params)
	require.GreaterOrEqual(t, len(rs), 12, "registry must contain at least the 12 locked slots")

	// Names are returned VERBATIM by each rule's Name() method. Slots 1-4
	// (the legacy alty-cli-766.7 rules) ship with PascalCase identifiers;
	// slots 5+ (alty-cli-ihk + alty-cli-fln + alty-cli-zc9) use snake_case.
	// The mixed casing is preserved exactly as a Contract-5 lock.
	want := []string{
		"FrontmatterSchemaRule",
		"PhaseEnumRule",
		"NoInternalLeaksRule",
		"OrphanOverlayRule",
		"bash_substitution_policy",
		"bash_arguments",
		"path_substitution_depth",
		"secrets_grep",
		"body_size",
		"unknown_tools",
		"bash_with_parameters_warn",
		"lifecycle_staleness",
	}
	for i, name := range want {
		assert.Equal(t, name, rs[i].Name(), "slot %d must remain %q (Contract-5 lock)", i+1, name)
	}
}
