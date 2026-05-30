package application

import (
	"context"
	"fmt"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// ScaffoldHealthHandler runs every registered ValidationRule against every
// asset returned by the walker, aggregates the violations, and packages
// them into a ScaffoldHealthReport. Sibling to DocHealthHandler — they
// share a bounded context but address distinct concerns (ISP).
type ScaffoldHealthHandler struct {
	walker ScaffoldWalker
	rules  []ValidationRule
}

// NewScaffoldHealthHandler constructs a handler. `walker` must be non-nil;
// `rules` may be empty (the handler then reports zero violations for any
// non-empty corpus — useful when callers want to defer wiring).
func NewScaffoldHealthHandler(walker ScaffoldWalker, rules []ValidationRule) *ScaffoldHealthHandler {
	r := make([]ValidationRule, len(rules))
	copy(r, rules)
	return &ScaffoldHealthHandler{walker: walker, rules: r}
}

// Handle walks altoDir, runs every rule against every asset, and returns
// the aggregated report. Walker errors propagate as wrapped errors (fatal);
// rule violations are values in the report (non-fatal at the layer level —
// the CLI converts ERROR severity into a non-zero exit code).
func (h *ScaffoldHealthHandler) Handle(ctx context.Context, altoDir string) (dochealthdomain.ScaffoldHealthReport, error) {
	corpus, err := h.walker.Walk(ctx, altoDir)
	if err != nil {
		return dochealthdomain.ScaffoldHealthReport{}, fmt.Errorf("walking scaffold %s: %w", altoDir, err)
	}

	var all []dochealthdomain.ScaffoldViolation
	for i := range corpus {
		for _, rule := range h.rules {
			all = append(all, rule.Check(corpus[i], corpus)...)
		}
	}
	return dochealthdomain.NewScaffoldHealthReport(all), nil
}
