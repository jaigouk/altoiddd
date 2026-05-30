package domain

// TemplateType represents the type of beads template to use when creating a ticket.
// Maps to templates in docs/beads_templates/:
//   - epic: beads-epic-template.md
//   - task: beads-ticket-template.md
//   - spike: beads-spike-template.md
type TemplateType string

const (
	// TemplateEpic maps to beads-epic-template.md (BoundedContext → epic).
	TemplateEpic TemplateType = "epic"

	// TemplateTask maps to beads-ticket-template.md (AggregateDesign → task).
	TemplateTask TemplateType = "task"

	// TemplateSpike maps to beads-spike-template.md (unresolved DomainStory → spike).
	TemplateSpike TemplateType = "spike"
)

// String returns the string representation of the template type.
func (t TemplateType) String() string {
	return string(t)
}

// IsValid returns true if the template type is a known type.
func (t TemplateType) IsValid() bool {
	switch t {
	case TemplateEpic, TemplateTask, TemplateSpike:
		return true
	default:
		return false
	}
}
