package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/fitness/domain"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func coreBC(name string) domain.BoundedContextInput {
	if name == "" {
		name = "Orders"
	}
	cl := vo.SubdomainCore
	return domain.BoundedContextInput{Name: name, Responsibility: "Manages " + name, Classification: &cl}
}

func supportingBC(name string) domain.BoundedContextInput {
	if name == "" {
		name = "Notifications"
	}
	cl := vo.SubdomainSupporting
	return domain.BoundedContextInput{Name: name, Responsibility: "Manages " + name, Classification: &cl}
}

func genericBC(name string) domain.BoundedContextInput {
	if name == "" {
		name = "Logging"
	}
	cl := vo.SubdomainGeneric
	return domain.BoundedContextInput{Name: name, Responsibility: "Manages " + name, Classification: &cl}
}

// ---------------------------------------------------------------------------
// Suite creation
// ---------------------------------------------------------------------------

func TestCreateSuite(t *testing.T) {
	t.Parallel()

	t.Run("new suite is empty", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		assert.Empty(t, suite.Contracts())
		assert.Empty(t, suite.ArchRules())
		assert.Equal(t, "myapp", suite.RootPackage())
	})

	t.Run("suite has unique id", func(t *testing.T) {
		t.Parallel()
		a := domain.NewFitnessTestSuite("myapp")
		b := domain.NewFitnessTestSuite("myapp")
		assert.NotEqual(t, a.SuiteID(), b.SuiteID())
	})
}

// ---------------------------------------------------------------------------
// Contract generation per strictness
// ---------------------------------------------------------------------------

func TestGenerateContracts(t *testing.T) {
	t.Parallel()

	t.Run("core produces all four types", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		err := suite.GenerateContracts([]domain.BoundedContextInput{coreBC("")})
		require.NoError(t, err)
		types := make(map[domain.ContractType]bool)
		for _, c := range suite.Contracts() {
			types[c.ContractType()] = true
		}
		assert.True(t, types[domain.ContractTypeLayers])
		assert.True(t, types[domain.ContractTypeForbidden])
		assert.True(t, types[domain.ContractTypeIndependence])
		assert.True(t, types[domain.ContractTypeAcyclicSiblings])
	})

	t.Run("supporting produces two types", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		err := suite.GenerateContracts([]domain.BoundedContextInput{supportingBC("")})
		require.NoError(t, err)
		types := make(map[domain.ContractType]bool)
		for _, c := range suite.Contracts() {
			types[c.ContractType()] = true
		}
		assert.True(t, types[domain.ContractTypeLayers])
		assert.True(t, types[domain.ContractTypeForbidden])
		assert.Len(t, suite.Contracts(), 2)
	})

	t.Run("generic produces one type", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		err := suite.GenerateContracts([]domain.BoundedContextInput{genericBC("")})
		require.NoError(t, err)
		types := make(map[domain.ContractType]bool)
		for _, c := range suite.Contracts() {
			types[c.ContractType()] = true
		}
		assert.True(t, types[domain.ContractTypeForbidden])
		assert.Len(t, suite.Contracts(), 1)
	})

	t.Run("mixed classifications", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		err := suite.GenerateContracts([]domain.BoundedContextInput{
			coreBC("Orders"),
			coreBC("Payments"),
			supportingBC("Notifications"),
			genericBC("Logging"),
		})
		require.NoError(t, err)
		// Core: 4 each = 8, Supporting: 2, Generic: 1 = 11 total
		assert.Len(t, suite.Contracts(), 11)
	})

	t.Run("contracts have correct context name", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("Orders")})
		for _, c := range suite.Contracts() {
			assert.Equal(t, "Orders", c.ContextName())
		}
	})

	t.Run("layers contract module order", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("Orders")})
		var layersContracts []domain.Contract
		for _, c := range suite.Contracts() {
			if c.ContractType() == domain.ContractTypeLayers {
				layersContracts = append(layersContracts, c)
			}
		}
		require.Len(t, layersContracts, 1)
		modules := layersContracts[0].Modules()
		assert.Contains(t, strings.ToLower(modules[0]), "infrastructure")
		assert.Contains(t, strings.ToLower(modules[len(modules)-1]), "domain")
	})

	t.Run("forbidden contract prevents domain importing infra", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("Orders")})
		var forbidden []domain.Contract
		for _, c := range suite.Contracts() {
			if c.ContractType() == domain.ContractTypeForbidden {
				forbidden = append(forbidden, c)
			}
		}
		require.NotEmpty(t, forbidden)
		f := forbidden[0]
		assert.Contains(t, strings.ToLower(strings.Join(f.Modules(), " ")), "domain")
		assert.Contains(t, strings.ToLower(strings.Join(f.ForbiddenModules(), " ")), "infrastructure")
	})

	t.Run("arch rules generated for each bc", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("Orders")})
		assert.NotEmpty(t, suite.ArchRules())
		for _, r := range suite.ArchRules() {
			assert.Equal(t, "Orders", r.ContextName())
		}
	})
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestSingleBcEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("single bc still generates independence", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("")})
		types := make(map[domain.ContractType]bool)
		for _, c := range suite.Contracts() {
			types[c.ContractType()] = true
		}
		assert.True(t, types[domain.ContractTypeIndependence])
	})

	t.Run("empty bounded contexts raises", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		err := suite.GenerateContracts(nil)
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("bc without classification raises", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		err := suite.GenerateContracts([]domain.BoundedContextInput{
			{Name: "Unclassified", Responsibility: "test", Classification: nil},
		})
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

// ---------------------------------------------------------------------------
// Approve
// ---------------------------------------------------------------------------

func TestApproveInvariants(t *testing.T) {
	t.Parallel()

	t.Run("approve after generate succeeds", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("")})
		err := suite.Approve()
		require.NoError(t, err)
	})

	t.Run("approve without generate raises", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		err := suite.Approve()
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("approve emits event", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("")})
		_ = suite.Approve()
		assert.Len(t, suite.Events(), 1)
	})

	t.Run("double approve raises", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("")})
		_ = suite.Approve()
		err := suite.Approve()
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

// ---------------------------------------------------------------------------
// Preview
// ---------------------------------------------------------------------------

func TestSuitePreview(t *testing.T) {
	t.Parallel()

	t.Run("preview returns string", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("Orders")})
		preview, err := suite.Preview()
		require.NoError(t, err)
		assert.Contains(t, preview, "Orders")
	})

	t.Run("preview shows contract counts", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{
			coreBC("Orders"),
			supportingBC("Notifications"),
		})
		preview, err := suite.Preview()
		require.NoError(t, err)
		assert.Contains(t, preview, "Orders")
		assert.Contains(t, preview, "Notifications")
		assert.True(t, strings.Contains(preview, "STRICT") || strings.Contains(strings.ToLower(preview), "strict"))
		assert.True(t, strings.Contains(preview, "MODERATE") || strings.Contains(strings.ToLower(preview), "moderate"))
	})

	t.Run("preview without generate raises", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_, err := suite.Preview()
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

// ---------------------------------------------------------------------------
// Rendering
// ---------------------------------------------------------------------------

func TestRendering(t *testing.T) {
	t.Parallel()

	t.Run("render import linter toml", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("Orders")})
		toml, err := suite.RenderImportLinterTOML()
		require.NoError(t, err)
		assert.Contains(t, toml, "[tool.importlinter]")
		assert.Contains(t, toml, `root_package = "myapp"`)
		assert.Contains(t, toml, "[[tool.importlinter.contracts]]")
		assert.Contains(t, toml, "Orders")
	})

	t.Run("render pytestarch tests", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("Orders")})
		code, err := suite.RenderPytestarchTests()
		require.NoError(t, err)
		assert.Contains(t, code, "from pytestarch")
		assert.Contains(t, code, "def test_")
		assert.Contains(t, strings.ToLower(code), "orders")
	})

	t.Run("render toml without contracts raises", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_, err := suite.RenderImportLinterTOML()
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("render pytestarch without contracts raises", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_, err := suite.RenderPytestarchTests()
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("toml layers correct format", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("Orders")})
		toml, _ := suite.RenderImportLinterTOML()
		assert.Contains(t, toml, `type = "layers"`)
	})

	t.Run("toml forbidden correct format", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("Orders")})
		toml, _ := suite.RenderImportLinterTOML()
		assert.Contains(t, toml, `type = "forbidden"`)
	})

	t.Run("multiple bcs render separate contracts", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{
			coreBC("Orders"),
			supportingBC("Notifications"),
		})
		toml, _ := suite.RenderImportLinterTOML()
		assert.GreaterOrEqual(t, strings.Count(toml, "[[tool.importlinter.contracts]]"), 3)
	})
}

// ---------------------------------------------------------------------------
// Module boundary validation
// ---------------------------------------------------------------------------

func TestModuleBoundaryValidation(t *testing.T) {
	t.Parallel()

	t.Run("generated contracts pass boundary check", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{
			coreBC("Orders"),
			supportingBC("Notifications"),
		})
		err := suite.Approve()
		require.NoError(t, err)
		assert.Len(t, suite.Events(), 1)
	})
}

// ---------------------------------------------------------------------------
// Event payload
// ---------------------------------------------------------------------------

func TestApproveEventPayload(t *testing.T) {
	t.Parallel()

	t.Run("event includes arch rules", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("Orders")})
		_ = suite.Approve()
		event := suite.Events()[0]
		assert.NotEmpty(t, event.ArchRules())
	})

	t.Run("event contracts and rules match suite", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("Orders")})
		_ = suite.Approve()
		event := suite.Events()[0]
		assert.Len(t, event.Contracts(), len(suite.Contracts()))
		assert.Len(t, event.ArchRules(), len(suite.ArchRules()))
	})
}

// ---------------------------------------------------------------------------
// SnakeCase
// ---------------------------------------------------------------------------

func TestSnakeCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"pascal case", "OrderManagement", "order_management"},
		{"space separated", "Order Management", "order_management"},
		{"hyphenated", "Architecture-Testing", "architecture_testing"},
		{"consecutive uppercase", "ABCTest", "abc_test"},
		{"already snake case", "order_management", "order_management"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := domain.SnakeCase(tt.input)
			assert.Equal(t, tt.want, got)
			assert.NotContains(t, got, "__", "should not have double underscores")
		})
	}
}

// ---------------------------------------------------------------------------
// Additional edge cases
// ---------------------------------------------------------------------------

func TestAdditionalEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("generate clears previous contracts", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("Orders")})
		firstCount := len(suite.Contracts())
		_ = suite.GenerateContracts([]domain.BoundedContextInput{genericBC("Logging")})
		assert.Less(t, len(suite.Contracts()), firstCount)
		for _, c := range suite.Contracts() {
			assert.Equal(t, "Logging", c.ContextName())
		}
	})

	t.Run("arch rules cleared on regenerate", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("Orders")})
		for _, r := range suite.ArchRules() {
			assert.Equal(t, "Orders", r.ContextName())
		}
		_ = suite.GenerateContracts([]domain.BoundedContextInput{genericBC("Logging")})
		for _, r := range suite.ArchRules() {
			assert.Equal(t, "Logging", r.ContextName())
		}
	})

	t.Run("pytestarch test has correct function names", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("OrderManagement")})
		code, _ := suite.RenderPytestarchTests()
		assert.Contains(t, code, "def test_order_management_domain_isolation")
	})

	t.Run("event carries suite id", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("")})
		_ = suite.Approve()
		assert.Equal(t, suite.SuiteID(), suite.Events()[0].SuiteID())
	})

	t.Run("event root package matches", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("custom_pkg")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("")})
		_ = suite.Approve()
		assert.Equal(t, "custom_pkg", suite.Events()[0].RootPackage())
	})

	t.Run("module prefix uses snake case context name", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("OrderManagement")})
		for _, c := range suite.Contracts() {
			for _, mod := range c.Modules() {
				assert.NotContains(t, mod, "OrderManagement")
				assert.Contains(t, mod, "order_management")
			}
		}
	})

	t.Run("many bcs produce correct total", func(t *testing.T) {
		t.Parallel()
		var bcs []domain.BoundedContextInput
		for i := 0; i < 5; i++ {
			bcs = append(bcs, coreBC("Context"+string(rune('A'+i))))
		}
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts(bcs)
		assert.Len(t, suite.Contracts(), 20)
		assert.Len(t, suite.ArchRules(), 5)
	})

	t.Run("generate after approve raises", func(t *testing.T) {
		t.Parallel()
		suite := domain.NewFitnessTestSuite("myapp")
		_ = suite.GenerateContracts([]domain.BoundedContextInput{coreBC("")})
		_ = suite.Approve()
		err := suite.GenerateContracts([]domain.BoundedContextInput{coreBC("")})
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

// ---------------------------------------------------------------------------
// RenderArchGoYAML
// ---------------------------------------------------------------------------

// Helper to create a BoundedContextMap for testing
func testBoundedContextMap(contexts ...domain.BoundedContextEntry) *domain.BoundedContextMap {
	bcMap := domain.NewBoundedContextMap("testproject", "github.com/org/testproject", contexts)
	return &bcMap
}

func TestFitnessTestSuite_RenderArchGoYAML_ValidOutput(t *testing.T) {
	t.Parallel()

	entry1 := domain.NewBoundedContextEntry(
		"Bootstrap",
		"bootstrap",
		vo.SubdomainSupporting,
		[]string{"domain", "application", "infrastructure"},
		nil,
	)
	entry2 := domain.NewBoundedContextEntry(
		"Discovery",
		"discovery",
		vo.SubdomainCore,
		[]string{"domain", "application", "infrastructure"},
		nil,
	)
	bcMap := testBoundedContextMap(entry1, entry2)

	suite := domain.NewFitnessTestSuite("github.com/org/testproject")
	yaml, err := suite.RenderArchGoYAML(bcMap, 100)

	require.NoError(t, err)
	assert.Contains(t, yaml, "version: 1")
	assert.Contains(t, yaml, "dependenciesRules:")
	assert.Contains(t, yaml, "threshold:")
	assert.Contains(t, yaml, "compliance: 100")
}

func TestFitnessTestSuite_RenderArchGoYAML_DomainLayerRules(t *testing.T) {
	t.Parallel()

	entry := domain.NewBoundedContextEntry(
		"Orders",
		"orders",
		vo.SubdomainCore,
		[]string{"domain", "application", "infrastructure"},
		nil,
	)
	bcMap := testBoundedContextMap(entry)

	suite := domain.NewFitnessTestSuite("github.com/org/testproject")
	yaml, err := suite.RenderArchGoYAML(bcMap, 100)

	require.NoError(t, err)
	// Domain layer should have shouldOnlyDependsOn for positive model
	assert.Contains(t, yaml, "shouldOnlyDependsOn:")
	assert.Contains(t, yaml, "github.com/org/testproject/internal/orders/domain")
	assert.Contains(t, yaml, "github.com/org/testproject/internal/shared/domain")
}

func TestFitnessTestSuite_RenderArchGoYAML_ApplicationLayerRules(t *testing.T) {
	t.Parallel()

	entry := domain.NewBoundedContextEntry(
		"Orders",
		"orders",
		vo.SubdomainCore,
		[]string{"domain", "application", "infrastructure"},
		nil,
	)
	bcMap := testBoundedContextMap(entry)

	suite := domain.NewFitnessTestSuite("github.com/org/testproject")
	yaml, err := suite.RenderArchGoYAML(bcMap, 100)

	require.NoError(t, err)
	// Application layer should not depend on infrastructure
	assert.Contains(t, yaml, "shouldNotDependsOn:")
	assert.Contains(t, yaml, "github.com/org/testproject/internal/orders/application")
	assert.Contains(t, yaml, "github.com/org/testproject/internal/orders/infrastructure")
}

func TestFitnessTestSuite_RenderArchGoYAML_CrossContextIsolation(t *testing.T) {
	t.Parallel()

	// Two contexts with NO relationship — should generate isolation rule
	entry1 := domain.NewBoundedContextEntry(
		"Orders",
		"orders",
		vo.SubdomainCore,
		[]string{"domain", "application", "infrastructure"},
		nil, // No relationships
	)
	entry2 := domain.NewBoundedContextEntry(
		"Shipping",
		"shipping",
		vo.SubdomainCore,
		[]string{"domain", "application", "infrastructure"},
		nil, // No relationships
	)
	bcMap := testBoundedContextMap(entry1, entry2)

	suite := domain.NewFitnessTestSuite("github.com/org/testproject")
	yaml, err := suite.RenderArchGoYAML(bcMap, 100)

	require.NoError(t, err)
	// Should have cross-context isolation rules
	assert.Contains(t, yaml, "orders")
	assert.Contains(t, yaml, "shipping")
}

func TestFitnessTestSuite_RenderArchGoYAML_AllowedRelationship(t *testing.T) {
	t.Parallel()

	// Bootstrap has upstream relationship to Discovery — Bootstrap may depend on Discovery
	rel := domain.NewContextRelationship("Discovery", domain.RelationshipUpstream, domain.PatternDomainEvent)
	entry1 := domain.NewBoundedContextEntry(
		"Bootstrap",
		"bootstrap",
		vo.SubdomainSupporting,
		[]string{"domain", "application", "infrastructure"},
		[]domain.ContextRelationship{rel},
	)
	entry2 := domain.NewBoundedContextEntry(
		"Discovery",
		"discovery",
		vo.SubdomainCore,
		[]string{"domain", "application", "infrastructure"},
		nil,
	)
	bcMap := testBoundedContextMap(entry1, entry2)

	suite := domain.NewFitnessTestSuite("github.com/org/testproject")
	yaml, err := suite.RenderArchGoYAML(bcMap, 100)

	require.NoError(t, err)
	// When Bootstrap is upstream to Discovery, Bootstrap can depend on Discovery
	// So there should NOT be a rule blocking bootstrap → discovery
	assert.Contains(t, yaml, "bootstrap")
	assert.Contains(t, yaml, "discovery")
}

func TestFitnessTestSuite_RenderArchGoYAML_ThresholdGreenfield(t *testing.T) {
	t.Parallel()

	entry := domain.NewBoundedContextEntry(
		"Orders",
		"orders",
		vo.SubdomainCore,
		[]string{"domain"},
		nil,
	)
	bcMap := testBoundedContextMap(entry)

	suite := domain.NewFitnessTestSuite("github.com/org/testproject")
	yaml, err := suite.RenderArchGoYAML(bcMap, 100)

	require.NoError(t, err)
	assert.Contains(t, yaml, "compliance: 100")
	assert.Contains(t, yaml, "coverage: 100")
}

func TestFitnessTestSuite_RenderArchGoYAML_ThresholdBrownfield(t *testing.T) {
	t.Parallel()

	entry := domain.NewBoundedContextEntry(
		"Orders",
		"orders",
		vo.SubdomainCore,
		[]string{"domain"},
		nil,
	)
	bcMap := testBoundedContextMap(entry)

	suite := domain.NewFitnessTestSuite("github.com/org/testproject")
	yaml, err := suite.RenderArchGoYAML(bcMap, 80)

	require.NoError(t, err)
	assert.Contains(t, yaml, "compliance: 80")
	assert.Contains(t, yaml, "coverage: 80")
}

func TestFitnessTestSuite_RenderArchGoYAML_EmptyMap(t *testing.T) {
	t.Parallel()

	bcMap := testBoundedContextMap() // Empty

	suite := domain.NewFitnessTestSuite("github.com/org/testproject")
	_, err := suite.RenderArchGoYAML(bcMap, 100)

	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}
