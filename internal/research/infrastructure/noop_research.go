// Package infrastructure provides adapters for the Research bounded context.
package infrastructure

import (
	"context"

	researchapp "github.com/alty-cli/alty/internal/research/application"
	researchdomain "github.com/alty-cli/alty/internal/research/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
)

// NoopResearchAdapter returns an empty ResearchBriefing with context names as no-data areas.
// Used as the default when no research infrastructure is configured.
type NoopResearchAdapter struct{}

// Compile-time interface check.
var _ researchapp.DomainResearch = (*NoopResearchAdapter)(nil)

// Research returns empty briefing -- all areas listed as no-data.
func (n *NoopResearchAdapter) Research(
	_ context.Context,
	model *ddd.DomainModel,
	_ int,
) (researchdomain.ResearchBriefing, error) {
	contexts := model.BoundedContexts()
	noData := make([]string, len(contexts))
	for i, ctx := range contexts {
		noData[i] = ctx.Name()
	}
	return researchdomain.NewResearchBriefing(nil, noData, ""), nil
}
