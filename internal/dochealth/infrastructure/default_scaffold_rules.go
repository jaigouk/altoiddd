package infrastructure

import (
	dochealthapp "github.com/alto-cli/alto/internal/dochealth/application"
	"github.com/alto-cli/alto/internal/dochealth/infrastructure/rules"
)

// DefaultScaffoldRules returns the 4 ERROR-severity ValidationRules that
// ship in alty-cli-766.7. The fast-follow ticket adds 4+ WARNING rules to
// this slice without touching the handler — Open/Closed by construction.
func DefaultScaffoldRules() []dochealthapp.ValidationRule {
	return []dochealthapp.ValidationRule{
		rules.NewFrontmatterSchemaRule(),
		rules.NewPhaseEnumRule(),
		rules.NewNoInternalLeaksRule(),
		rules.NewOrphanOverlayRule(),
	}
}
