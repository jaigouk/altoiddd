package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

func TestQualityGateEnumHasFourMembers(t *testing.T) {
	t.Parallel()
	assert.Len(t, vo.AllQualityGates(), 4)
}

func TestQualityGateEnumValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, vo.QualityGateLint, vo.QualityGate("lint"))
	assert.Equal(t, vo.QualityGateTypes, vo.QualityGate("types"))
	assert.Equal(t, vo.QualityGateTests, vo.QualityGate("tests"))
	assert.Equal(t, vo.QualityGateFitness, vo.QualityGate("fitness"))
}

func TestGateResult(t *testing.T) {
	t.Parallel()

	t.Run("stores fields", func(t *testing.T) {
		t.Parallel()
		r := vo.NewGateResult(vo.QualityGateTypes, false, "error on line 5", 100)
		assert.Equal(t, vo.QualityGateTypes, r.Gate())
		assert.False(t, r.Passed())
		assert.Equal(t, "error on line 5", r.Output())
		assert.Equal(t, 100, r.DurationMS())
	})
}

func TestQualityReport(t *testing.T) {
	t.Parallel()

	t.Run("passed true when all pass", func(t *testing.T) {
		t.Parallel()
		results := []vo.GateResult{
			vo.NewGateResult(vo.QualityGateLint, true, "", 10),
			vo.NewGateResult(vo.QualityGateTypes, true, "", 20),
		}
		report := vo.NewQualityReport(results)
		assert.True(t, report.Passed())
	})

	t.Run("passed false when any fails", func(t *testing.T) {
		t.Parallel()
		results := []vo.GateResult{
			vo.NewGateResult(vo.QualityGateLint, true, "", 10),
			vo.NewGateResult(vo.QualityGateTests, false, "1 failed", 50),
		}
		report := vo.NewQualityReport(results)
		assert.False(t, report.Passed())
	})

	t.Run("passed true for empty results", func(t *testing.T) {
		t.Parallel()
		report := vo.NewQualityReport(nil)
		assert.True(t, report.Passed())
	})

	t.Run("defensive copy results", func(t *testing.T) {
		t.Parallel()
		results := []vo.GateResult{
			vo.NewGateResult(vo.QualityGateLint, true, "", 10),
		}
		report := vo.NewQualityReport(results)
		results[0] = vo.NewGateResult(vo.QualityGateTests, false, "changed", 0)
		assert.Equal(t, vo.QualityGateLint, report.Results()[0].Gate())
	})
}
