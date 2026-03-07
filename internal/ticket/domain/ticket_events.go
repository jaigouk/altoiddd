package domain

// TicketPlanApproved is emitted when a TicketPlan is approved and ready for output.
type TicketPlanApproved struct {
	planID             string
	approvedTicketIDs  []string
	dismissedTicketIDs []string
}

// NewTicketPlanApproved creates a TicketPlanApproved event.
func NewTicketPlanApproved(planID string, approvedIDs, dismissedIDs []string) TicketPlanApproved {
	a := make([]string, len(approvedIDs))
	copy(a, approvedIDs)
	d := make([]string, len(dismissedIDs))
	copy(d, dismissedIDs)
	return TicketPlanApproved{planID: planID, approvedTicketIDs: a, dismissedTicketIDs: d}
}

// PlanID returns the plan identifier.
func (e TicketPlanApproved) PlanID() string { return e.planID }

// ApprovedTicketIDs returns a defensive copy of approved ticket IDs.
func (e TicketPlanApproved) ApprovedTicketIDs() []string {
	out := make([]string, len(e.approvedTicketIDs))
	copy(out, e.approvedTicketIDs)
	return out
}

// DismissedTicketIDs returns a defensive copy of dismissed ticket IDs.
func (e TicketPlanApproved) DismissedTicketIDs() []string {
	out := make([]string, len(e.dismissedTicketIDs))
	copy(out, e.dismissedTicketIDs)
	return out
}
