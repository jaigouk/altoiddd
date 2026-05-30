package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dochealthapp "github.com/alto-cli/alto/internal/dochealth/application"
	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

type stubWalker struct {
	corpus []dochealthdomain.ScaffoldAsset
	err    error
}

func (s *stubWalker) Walk(_ context.Context, _ string) ([]dochealthdomain.ScaffoldAsset, error) {
	return s.corpus, s.err
}

type stubRule struct {
	name string
	hits []dochealthdomain.ScaffoldViolation
}

func (s *stubRule) Name() string { return s.name }
func (s *stubRule) Check(_ dochealthdomain.ScaffoldAsset, _ []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation {
	return s.hits
}

func mustAsset(t *testing.T, path string) dochealthdomain.ScaffoldAsset {
	t.Helper()
	a, err := dochealthdomain.NewScaffoldAsset(path, nil, "", 0, false)
	require.NoError(t, err)
	return a
}

func mustViolation(t *testing.T, file, rule string) dochealthdomain.ScaffoldViolation {
	t.Helper()
	v, err := dochealthdomain.NewScaffoldViolation(file, rule, "test", dochealthdomain.SeverityError, 0)
	require.NoError(t, err)
	return v
}

func TestScaffoldHealthHandler_Handle_NoRules_NoViolations(t *testing.T) {
	t.Parallel()
	w := &stubWalker{corpus: []dochealthdomain.ScaffoldAsset{mustAsset(t, "foo.md")}}
	h := dochealthapp.NewScaffoldHealthHandler(w, nil)
	report, err := h.Handle(context.TODO(), ".alto/")
	require.NoError(t, err)
	assert.Equal(t, 0, report.TotalCount())
}

func TestScaffoldHealthHandler_Handle_AggregatesAcrossRulesAndAssets(t *testing.T) {
	t.Parallel()
	corpus := []dochealthdomain.ScaffoldAsset{
		mustAsset(t, "foo.md"),
		mustAsset(t, "bar.md"),
	}
	rules := []dochealthapp.ValidationRule{
		&stubRule{name: "r1", hits: []dochealthdomain.ScaffoldViolation{mustViolation(t, "foo.md", "r1")}},
		&stubRule{name: "r2", hits: []dochealthdomain.ScaffoldViolation{mustViolation(t, "foo.md", "r2")}},
	}
	w := &stubWalker{corpus: corpus}
	h := dochealthapp.NewScaffoldHealthHandler(w, rules)
	report, err := h.Handle(context.TODO(), ".alto/")
	require.NoError(t, err)
	// 2 rules × 2 assets = 4 violations.
	assert.Equal(t, 4, report.TotalCount())
}

func TestScaffoldHealthHandler_Handle_WalkerError_Propagates(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("walker boom")
	w := &stubWalker{err: sentinel}
	h := dochealthapp.NewScaffoldHealthHandler(w, nil)
	_, err := h.Handle(context.TODO(), ".alto/")
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
}
