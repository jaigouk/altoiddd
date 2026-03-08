package domain

import (
	"encoding/json"
	"fmt"
)

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

// MarshalJSON implements json.Marshaler for event bus serialization.
func (e TicketPlanApproved) MarshalJSON() ([]byte, error) {
	type proxy struct {
		PlanID             string   `json:"plan_id"`
		ApprovedTicketIDs  []string `json:"approved_ticket_ids"`
		DismissedTicketIDs []string `json:"dismissed_ticket_ids"`
	}
	data, err := json.Marshal(proxy{
		PlanID:             e.planID,
		ApprovedTicketIDs:  e.approvedTicketIDs,
		DismissedTicketIDs: e.dismissedTicketIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling TicketPlanApproved: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (e *TicketPlanApproved) UnmarshalJSON(data []byte) error {
	type proxy struct {
		PlanID             string   `json:"plan_id"`
		ApprovedTicketIDs  []string `json:"approved_ticket_ids"`
		DismissedTicketIDs []string `json:"dismissed_ticket_ids"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling TicketPlanApproved: %w", err)
	}
	e.planID = p.PlanID
	e.approvedTicketIDs = p.ApprovedTicketIDs
	e.dismissedTicketIDs = p.DismissedTicketIDs
	return nil
}
