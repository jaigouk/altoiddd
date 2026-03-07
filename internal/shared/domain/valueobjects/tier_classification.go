package valueobjects

import "fmt"

// TicketDetailLevel maps subdomain classification to generation depth.
type TicketDetailLevel string

// Ticket detail level constants.
const (
	TicketDetailFull     TicketDetailLevel = "full"
	TicketDetailStandard TicketDetailLevel = "standard"
	TicketDetailStub     TicketDetailLevel = "stub"
)

// DetailLevelFromClassification maps a SubdomainClassification to its TicketDetailLevel.
func DetailLevelFromClassification(c SubdomainClassification) TicketDetailLevel {
	switch c {
	case SubdomainCore:
		return TicketDetailFull
	case SubdomainSupporting:
		return TicketDetailStandard
	case SubdomainGeneric:
		return TicketDetailStub
	default:
		return TicketDetailStub
	}
}

// Tier represents near-term vs far-term classification.
type Tier string

// Tier constants.
const (
	TierNearTerm Tier = "near_term"
	TierFarTerm  Tier = "far_term"
)

const nearTermMaxDepth = 2

// TierClassification is a depth-based tier with human-readable reason.
type TierClassification struct {
	tier   Tier
	reason string
}

// NewTierClassification creates a TierClassification value object.
func NewTierClassification(tier Tier, reason string) TierClassification {
	return TierClassification{tier: tier, reason: reason}
}

// Tier returns the tier value.
func (tc TierClassification) Tier() Tier { return tc.tier }

// Reason returns the classification reason.
func (tc TierClassification) Reason() string { return tc.reason }

// ClassifyTier classifies a ticket as near-term or far-term based on depth and subdomain.
// Core subdomain is always near-term. depth <= 2 is near-term. depth > 2 is far-term.
func ClassifyTier(depth int, classification SubdomainClassification) TierClassification {
	if classification == SubdomainCore {
		return NewTierClassification(TierNearTerm, "Core subdomain: always near-term")
	}
	if depth <= nearTermMaxDepth {
		return NewTierClassification(TierNearTerm, fmt.Sprintf("Depth %d <= %d", depth, nearTermMaxDepth))
	}
	return NewTierClassification(TierFarTerm, fmt.Sprintf("Depth %d > %d", depth, nearTermMaxDepth))
}

// TierToDetailLevel maps a TierClassification to a TicketDetailLevel.
// Core subdomain is always FULL. Far-term non-Core is always STUB.
func TierToDetailLevel(tier TierClassification, classification SubdomainClassification) TicketDetailLevel {
	if classification == SubdomainCore {
		return TicketDetailFull
	}
	if tier.tier == TierFarTerm {
		return TicketDetailStub
	}
	return DetailLevelFromClassification(classification)
}
