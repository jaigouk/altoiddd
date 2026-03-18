package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/fitness/domain"
)

func TestFitnessTestsGenerated(t *testing.T) {
	t.Parallel()

	t.Run("create event", func(t *testing.T) {
		t.Parallel()
		contracts := []domain.Contract{
			domain.NewContract("test", domain.ContractTypeLayers, "Orders",
				[]string{"a", "b"}, nil),
		}
		rules := []domain.ArchRule{
			domain.NewArchRule("Orders domain isolation",
				"modules in myapp.orders.domain should not import from myapp.orders.infrastructure",
				"Orders"),
		}
		event := domain.NewFitnessTestsGenerated("test-id", "myapp", contracts, rules)
		assert.Equal(t, "test-id", event.SuiteID())
		assert.Equal(t, "myapp", event.RootPackage())
		assert.Len(t, event.Contracts(), 1)
		assert.Len(t, event.ArchRules(), 1)
	})

	t.Run("defensive copy contracts", func(t *testing.T) {
		t.Parallel()
		contracts := []domain.Contract{
			domain.NewContract("test", domain.ContractTypeLayers, "Orders",
				[]string{"a"}, nil),
		}
		event := domain.NewFitnessTestsGenerated("id", "pkg", contracts, nil)
		contracts[0] = domain.NewContract("changed", domain.ContractTypeForbidden, "X",
			[]string{"b"}, nil)
		assert.Equal(t, "test", event.Contracts()[0].Name())
	})

	t.Run("defensive copy arch rules", func(t *testing.T) {
		t.Parallel()
		rules := []domain.ArchRule{
			domain.NewArchRule("test", "assertion", "Ctx"),
		}
		event := domain.NewFitnessTestsGenerated("id", "pkg", nil, rules)
		rules[0] = domain.NewArchRule("changed", "x", "Y")
		assert.Equal(t, "test", event.ArchRules()[0].Name())
	})
}

func TestFitnessTestsGenerated_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	contracts := []domain.Contract{
		domain.NewContract("test", domain.ContractTypeLayers, "Orders",
			[]string{"a", "b"}, nil),
	}
	rules := []domain.ArchRule{
		domain.NewArchRule("isolation", "domain should not import infra", "Orders"),
	}
	original := domain.NewFitnessTestsGenerated("suite-rt", "myapp", contracts, rules)

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"suite_id"`)
	assert.Contains(t, string(data), `"root_package"`)

	var restored domain.FitnessTestsGenerated
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, "suite-rt", restored.SuiteID())
	assert.Equal(t, "myapp", restored.RootPackage())
	require.Len(t, restored.Contracts(), 1)
	assert.Equal(t, "test", restored.Contracts()[0].Name())
	assert.Equal(t, domain.ContractTypeLayers, restored.Contracts()[0].ContractType())
	assert.Equal(t, "Orders", restored.Contracts()[0].ContextName())
	assert.Equal(t, []string{"a", "b"}, restored.Contracts()[0].Modules())
	require.Len(t, restored.ArchRules(), 1)
	assert.Equal(t, "isolation", restored.ArchRules()[0].Name())
	assert.Equal(t, "domain should not import infra", restored.ArchRules()[0].Assertion())
	assert.Equal(t, "Orders", restored.ArchRules()[0].ContextName())
}
