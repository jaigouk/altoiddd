package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alty-cli/alty/internal/fitness/domain"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

func TestGateResult(t *testing.T) {
	t.Parallel()

	t.Run("stores fields", func(t *testing.T) {
		t.Parallel()
		r := domain.NewGateResult(vo.QualityGateTypes, false, "error on line 5", 100)
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
		results := []domain.GateResult{
			domain.NewGateResult(vo.QualityGateLint, true, "", 10),
			domain.NewGateResult(vo.QualityGateTypes, true, "", 20),
		}
		report := domain.NewQualityReport(results)
		assert.True(t, report.Passed())
	})

	t.Run("passed false when any fails", func(t *testing.T) {
		t.Parallel()
		results := []domain.GateResult{
			domain.NewGateResult(vo.QualityGateLint, true, "", 10),
			domain.NewGateResult(vo.QualityGateTests, false, "1 failed", 50),
		}
		report := domain.NewQualityReport(results)
		assert.False(t, report.Passed())
	})

	t.Run("passed true for empty results", func(t *testing.T) {
		t.Parallel()
		report := domain.NewQualityReport(nil)
		assert.True(t, report.Passed())
	})

	t.Run("defensive copy results", func(t *testing.T) {
		t.Parallel()
		results := []domain.GateResult{
			domain.NewGateResult(vo.QualityGateLint, true, "", 10),
		}
		report := domain.NewQualityReport(results)
		results[0] = domain.NewGateResult(vo.QualityGateTests, false, "changed", 0)
		assert.Equal(t, vo.QualityGateLint, report.Results()[0].Gate())
	})
}
