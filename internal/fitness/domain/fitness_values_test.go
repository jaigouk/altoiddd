package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alto-cli/alto/internal/fitness/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// ContractType enum
// ---------------------------------------------------------------------------

func TestContractTypeValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "layers", string(domain.ContractTypeLayers))
	assert.Equal(t, "forbidden", string(domain.ContractTypeForbidden))
	assert.Equal(t, "independence", string(domain.ContractTypeIndependence))
	assert.Equal(t, "acyclic_siblings", string(domain.ContractTypeAcyclicSiblings))
}

// ---------------------------------------------------------------------------
// ContractStrictness enum
// ---------------------------------------------------------------------------

func TestContractStrictnessValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "strict", string(domain.ContractStrictnessStrict))
	assert.Equal(t, "moderate", string(domain.ContractStrictnessModerate))
	assert.Equal(t, "minimal", string(domain.ContractStrictnessMinimal))
}

func TestContractStrictnessFromClassification(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		classification vo.SubdomainClassification
		want           domain.ContractStrictness
	}{
		{"core maps to strict", vo.SubdomainCore, domain.ContractStrictnessStrict},
		{"supporting maps to moderate", vo.SubdomainSupporting, domain.ContractStrictnessModerate},
		{"generic maps to minimal", vo.SubdomainGeneric, domain.ContractStrictnessMinimal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := domain.StrictnessFromClassification(tt.classification)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRequiredContractTypes(t *testing.T) {
	t.Parallel()

	t.Run("strict requires all four", func(t *testing.T) {
		t.Parallel()
		types := domain.RequiredContractTypes(domain.ContractStrictnessStrict)
		typeSet := make(map[domain.ContractType]bool)
		for _, ct := range types {
			typeSet[ct] = true
		}
		assert.True(t, typeSet[domain.ContractTypeLayers])
		assert.True(t, typeSet[domain.ContractTypeForbidden])
		assert.True(t, typeSet[domain.ContractTypeIndependence])
		assert.True(t, typeSet[domain.ContractTypeAcyclicSiblings])
	})

	t.Run("moderate requires layers and forbidden", func(t *testing.T) {
		t.Parallel()
		types := domain.RequiredContractTypes(domain.ContractStrictnessModerate)
		typeSet := make(map[domain.ContractType]bool)
		for _, ct := range types {
			typeSet[ct] = true
		}
		assert.True(t, typeSet[domain.ContractTypeLayers])
		assert.True(t, typeSet[domain.ContractTypeForbidden])
		assert.Len(t, types, 2)
	})

	t.Run("minimal requires forbidden only", func(t *testing.T) {
		t.Parallel()
		types := domain.RequiredContractTypes(domain.ContractStrictnessMinimal)
		typeSet := make(map[domain.ContractType]bool)
		for _, ct := range types {
			typeSet[ct] = true
		}
		assert.True(t, typeSet[domain.ContractTypeForbidden])
		assert.Len(t, types, 1)
	})
}

// ---------------------------------------------------------------------------
// Contract value object
// ---------------------------------------------------------------------------

func TestContract(t *testing.T) {
	t.Parallel()

	t.Run("create layers contract", func(t *testing.T) {
		t.Parallel()
		c := domain.NewContract("DDD layers", domain.ContractTypeLayers, "Orders",
			[]string{"orders.infrastructure", "orders.application", "orders.domain"}, nil)
		assert.Equal(t, "DDD layers", c.Name())
		assert.Equal(t, domain.ContractTypeLayers, c.ContractType())
		assert.Equal(t, "Orders", c.ContextName())
		assert.Len(t, c.Modules(), 3)
	})

	t.Run("equality", func(t *testing.T) {
		t.Parallel()
		a := domain.NewContract("test", domain.ContractTypeForbidden, "Ctx", []string{"a"}, nil)
		b := domain.NewContract("test", domain.ContractTypeForbidden, "Ctx", []string{"a"}, nil)
		assert.Equal(t, a, b)
	})

	t.Run("forbidden with target modules", func(t *testing.T) {
		t.Parallel()
		c := domain.NewContract("domain isolation", domain.ContractTypeForbidden, "Orders",
			[]string{"orders.domain"}, []string{"orders.infrastructure"})
		assert.Equal(t, []string{"orders.infrastructure"}, c.ForbiddenModules())
	})

	t.Run("defensive copy modules", func(t *testing.T) {
		t.Parallel()
		mods := []string{"a", "b"}
		c := domain.NewContract("test", domain.ContractTypeLayers, "Ctx", mods, nil)
		mods[0] = "changed"
		assert.Equal(t, "a", c.Modules()[0])
	})
}

// ---------------------------------------------------------------------------
// ArchRule value object
// ---------------------------------------------------------------------------

func TestArchRule(t *testing.T) {
	t.Parallel()

	t.Run("create arch rule", func(t *testing.T) {
		t.Parallel()
		r := domain.NewArchRule("domain isolation",
			"modules in orders.domain should not import from orders.infrastructure",
			"Orders")
		assert.Equal(t, "domain isolation", r.Name())
		assert.Equal(t, "Orders", r.ContextName())
	})

	t.Run("equality", func(t *testing.T) {
		t.Parallel()
		a := domain.NewArchRule("test", "x", "Ctx")
		b := domain.NewArchRule("test", "x", "Ctx")
		assert.Equal(t, a, b)
	})
}
