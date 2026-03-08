package domain

import (
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// Role is a role a bounded context plays in the domain.
type Role string

// Domain role constants.
const (
	RoleExecution     Role = "execution"
	RoleAnalysis      Role = "analysis"
	RoleGateway       Role = "gateway"
	RoleSpecification Role = "specification"
	RoleDraft         Role = "draft"
)

// AllRoles returns all defined Role values.
func AllRoles() []Role {
	return []Role{RoleExecution, RoleAnalysis, RoleGateway, RoleSpecification, RoleDraft}
}

// StrategicClassification is the strategic classification of a bounded context.
type StrategicClassification struct {
	domain        vo.SubdomainClassification
	businessModel string
	evolution     string
}

// NewStrategicClassification creates a StrategicClassification value object.
func NewStrategicClassification(domain vo.SubdomainClassification, businessModel, evolution string) StrategicClassification {
	return StrategicClassification{domain: domain, businessModel: businessModel, evolution: evolution}
}

// Domain returns the subdomain classification.
func (sc StrategicClassification) Domain() vo.SubdomainClassification { return sc.domain }

// BusinessModel returns the business model classification.
func (sc StrategicClassification) BusinessModel() string { return sc.businessModel }

// Evolution returns the evolution stage.
func (sc StrategicClassification) Evolution() string { return sc.evolution }

// CommunicationMessage is a message flowing into or out of a bounded context.
type CommunicationMessage struct {
	message     string
	messageType string
	counterpart string
}

// NewCommunicationMessage creates a CommunicationMessage value object.
func NewCommunicationMessage(message, messageType, counterpart string) CommunicationMessage {
	return CommunicationMessage{message: message, messageType: messageType, counterpart: counterpart}
}

// Message returns the message name.
func (m CommunicationMessage) Message() string { return m.message }

// MessageType returns the message type.
func (m CommunicationMessage) MessageType() string { return m.messageType }

// Counterpart returns the counterpart context name.
func (m CommunicationMessage) Counterpart() string { return m.counterpart }

// BoundedContextCanvas is a Bounded Context Canvas following the ddd-crew v5 format.
type BoundedContextCanvas struct {
	contextName           string
	purpose               string
	classification        StrategicClassification
	domainRoles           []Role
	inboundCommunication  []CommunicationMessage
	outboundCommunication []CommunicationMessage
	ubiquitousLanguage    [][2]string
	businessDecisions     []string
	assumptions           []string
	openQuestions         []string
}

// NewBoundedContextCanvas creates a BoundedContextCanvas value object.
func NewBoundedContextCanvas(
	contextName, purpose string,
	classification StrategicClassification,
	domainRoles []Role,
	inbound, outbound []CommunicationMessage,
	ul [][2]string,
	businessDecisions, assumptions, openQuestions []string,
) BoundedContextCanvas {
	dr := make([]Role, len(domainRoles))
	copy(dr, domainRoles)
	ib := make([]CommunicationMessage, len(inbound))
	copy(ib, inbound)
	ob := make([]CommunicationMessage, len(outbound))
	copy(ob, outbound)
	ulCopy := make([][2]string, len(ul))
	copy(ulCopy, ul)
	bd := make([]string, len(businessDecisions))
	copy(bd, businessDecisions)
	as := make([]string, len(assumptions))
	copy(as, assumptions)
	oq := make([]string, len(openQuestions))
	copy(oq, openQuestions)
	return BoundedContextCanvas{
		contextName:           contextName,
		purpose:               purpose,
		classification:        classification,
		domainRoles:           dr,
		inboundCommunication:  ib,
		outboundCommunication: ob,
		ubiquitousLanguage:    ulCopy,
		businessDecisions:     bd,
		assumptions:           as,
		openQuestions:         oq,
	}
}

// ContextName returns the bounded context name.
func (c BoundedContextCanvas) ContextName() string { return c.contextName }

// Purpose returns the context purpose.
func (c BoundedContextCanvas) Purpose() string { return c.purpose }

// Classification returns the strategic classification.
func (c BoundedContextCanvas) Classification() StrategicClassification { return c.classification }

// Roles returns a defensive copy.
func (c BoundedContextCanvas) Roles() []Role {
	out := make([]Role, len(c.domainRoles))
	copy(out, c.domainRoles)
	return out
}

// InboundCommunication returns a defensive copy.
func (c BoundedContextCanvas) InboundCommunication() []CommunicationMessage {
	out := make([]CommunicationMessage, len(c.inboundCommunication))
	copy(out, c.inboundCommunication)
	return out
}

// OutboundCommunication returns a defensive copy.
func (c BoundedContextCanvas) OutboundCommunication() []CommunicationMessage {
	out := make([]CommunicationMessage, len(c.outboundCommunication))
	copy(out, c.outboundCommunication)
	return out
}

// UbiquitousLanguage returns a defensive copy of (term, definition) pairs.
func (c BoundedContextCanvas) UbiquitousLanguage() [][2]string {
	out := make([][2]string, len(c.ubiquitousLanguage))
	copy(out, c.ubiquitousLanguage)
	return out
}

// BusinessDecisions returns a defensive copy.
func (c BoundedContextCanvas) BusinessDecisions() []string {
	out := make([]string, len(c.businessDecisions))
	copy(out, c.businessDecisions)
	return out
}

// Assumptions returns a defensive copy.
func (c BoundedContextCanvas) Assumptions() []string {
	out := make([]string, len(c.assumptions))
	copy(out, c.assumptions)
	return out
}

// OpenQuestions returns a defensive copy.
func (c BoundedContextCanvas) OpenQuestions() []string {
	out := make([]string, len(c.openQuestions))
	copy(out, c.openQuestions)
	return out
}
