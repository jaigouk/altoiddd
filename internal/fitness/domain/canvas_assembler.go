package domain

import (
	"fmt"
	"strings"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

var classificationToRole = map[vo.SubdomainClassification]Role{
	vo.SubdomainCore:       RoleExecution,
	vo.SubdomainSupporting: RoleSpecification,
	vo.SubdomainGeneric:    RoleGateway,
}

// Assemble builds one canvas per bounded context in the model.
func Assemble(model *ddd.DomainModel) []BoundedContextCanvas {
	var canvases []BoundedContextCanvas

	for _, ctx := range model.BoundedContexts() {
		domainCls := vo.SubdomainGeneric
		if ctx.Classification() != nil {
			domainCls = *ctx.Classification()
		}

		classification := NewStrategicClassification(domainCls, "unclassified", "unclassified")

		role, ok := classificationToRole[domainCls]
		if !ok {
			role = RoleDraft
		}

		var inbound []CommunicationMessage
		for _, rel := range model.ContextRelationships() {
			if rel.Downstream() == ctx.Name() {
				inbound = append(inbound, NewCommunicationMessage(
					rel.IntegrationPattern(), "Event", rel.Upstream()))
			}
		}

		var outbound []CommunicationMessage
		for _, rel := range model.ContextRelationships() {
			if rel.Upstream() == ctx.Name() {
				outbound = append(outbound, NewCommunicationMessage(
					rel.IntegrationPattern(), "Event", rel.Downstream()))
			}
		}

		var ulTerms [][2]string
		for _, entry := range model.UbiquitousLanguage().Terms() {
			if entry.ContextName() == ctx.Name() {
				ulTerms = append(ulTerms, [2]string{entry.Term(), entry.Definition()})
			}
		}

		var decisions []string
		for _, agg := range model.AggregateDesigns() {
			if agg.ContextName() == ctx.Name() {
				decisions = append(decisions, agg.Invariants()...)
			}
		}

		canvases = append(canvases, NewBoundedContextCanvas(
			ctx.Name(),
			ctx.Responsibility(),
			classification,
			[]Role{role},
			inbound, outbound, ulTerms, decisions,
			nil, nil,
		))
	}

	return canvases
}

// RenderMarkdown renders canvases to markdown following ddd-crew v5 format.
func RenderMarkdown(canvases []BoundedContextCanvas) string {
	if len(canvases) == 0 {
		return ""
	}

	sections := make([]string, len(canvases))
	for i, canvas := range canvases {
		sections[i] = renderSingleCanvas(canvas)
	}
	return strings.Join(sections, "\n---\n\n")
}

func renderSingleCanvas(canvas BoundedContextCanvas) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("# Bounded Context Canvas: %s", canvas.contextName), "")
	addPurpose(&lines, canvas)
	addClassification(&lines, canvas)
	addRoles(&lines, canvas)
	addCommunication(&lines, "Inbound Communication", "Sender", canvas.inboundCommunication)
	addCommunication(&lines, "Outbound Communication", "Receiver", canvas.outboundCommunication)
	addUL(&lines, canvas)
	addListSection(&lines, "Business Decisions & Rules", canvas.businessDecisions)
	addListSection(&lines, "Assumptions", canvas.assumptions)
	addListSection(&lines, "Open Questions", canvas.openQuestions)
	return strings.Join(lines, "\n")
}

func addPurpose(lines *[]string, canvas BoundedContextCanvas) {
	*lines = append(*lines, "## Purpose", "", canvas.purpose, "")
}

func addClassification(lines *[]string, canvas BoundedContextCanvas) {
	c := canvas.classification
	*lines = append(*lines,
		"## Strategic Classification", "",
		"| Aspect | Value |",
		"| --- | --- |",
		fmt.Sprintf("| Domain | %s |", string(c.domain)),
		fmt.Sprintf("| Business Model | %s |", c.businessModel),
		fmt.Sprintf("| Evolution | %s |", c.evolution),
		"",
	)
}

func addRoles(lines *[]string, canvas BoundedContextCanvas) {
	*lines = append(*lines, "## Domain Roles", "")
	for _, role := range canvas.domainRoles {
		*lines = append(*lines, fmt.Sprintf("- [x] %s", string(role)))
	}
	*lines = append(*lines, "")
}

func addCommunication(lines *[]string, heading, thirdCol string, messages []CommunicationMessage) {
	*lines = append(*lines, fmt.Sprintf("## %s", heading), "")
	if len(messages) > 0 {
		*lines = append(*lines,
			fmt.Sprintf("| Message | Type | %s |", thirdCol),
			"| --- | --- | --- |",
		)
		for _, m := range messages {
			*lines = append(*lines, fmt.Sprintf("| %s | %s | %s |", m.message, m.messageType, m.counterpart))
		}
	} else {
		*lines = append(*lines, "*None*")
	}
	*lines = append(*lines, "")
}

func addUL(lines *[]string, canvas BoundedContextCanvas) {
	*lines = append(*lines, "## Ubiquitous Language", "")
	if len(canvas.ubiquitousLanguage) > 0 {
		*lines = append(*lines, "| Term | Definition |", "| --- | --- |")
		for _, pair := range canvas.ubiquitousLanguage {
			*lines = append(*lines, fmt.Sprintf("| %s | %s |", pair[0], pair[1]))
		}
	} else {
		*lines = append(*lines, "*None*")
	}
	*lines = append(*lines, "")
}

func addListSection(lines *[]string, heading string, items []string) {
	*lines = append(*lines, fmt.Sprintf("## %s", heading), "")
	if len(items) > 0 {
		for _, item := range items {
			*lines = append(*lines, fmt.Sprintf("- %s", item))
		}
	} else {
		*lines = append(*lines, "*None*")
	}
	*lines = append(*lines, "")
}
