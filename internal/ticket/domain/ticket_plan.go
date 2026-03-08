package domain

import (
	"fmt"
	"strings"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	"github.com/alty-cli/alty/internal/shared/domain/identity"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// TicketPlan is the aggregate root for the ticket pipeline.
type TicketPlan struct {
	profile         vo.StackProfile
	dependencyOrder *DependencyOrder
	planID          string
	epics           []GeneratedEpic
	tickets         []GeneratedTicket
	events          []TicketPlanApproved
	approved        bool
}

// NewTicketPlan creates a new TicketPlan aggregate root.
func NewTicketPlan() *TicketPlan {
	return &TicketPlan{planID: identity.NewID()}
}

// PlanID returns the plan identifier.
func (p *TicketPlan) PlanID() string { return p.planID }

// Epics returns a defensive copy of all generated epics.
func (p *TicketPlan) Epics() []GeneratedEpic {
	out := make([]GeneratedEpic, len(p.epics))
	copy(out, p.epics)
	return out
}

// Tickets returns a defensive copy of all generated tickets.
func (p *TicketPlan) Tickets() []GeneratedTicket {
	out := make([]GeneratedTicket, len(p.tickets))
	copy(out, p.tickets)
	return out
}

// DependencyOrder returns the topological order, or nil if not computed.
func (p *TicketPlan) DependencyOrder() *DependencyOrder {
	if p.dependencyOrder == nil {
		return nil
	}
	d := *p.dependencyOrder
	return &d
}

// Events returns a defensive copy of domain events.
func (p *TicketPlan) Events() []TicketPlanApproved {
	out := make([]TicketPlanApproved, len(p.events))
	copy(out, p.events)
	return out
}

// GeneratePlan generates epics and tickets from a finalized DomainModel.
func (p *TicketPlan) GeneratePlan(model *ddd.DomainModel, profile vo.StackProfile) error {
	if p.approved {
		return fmt.Errorf("cannot regenerate plan on an approved TicketPlan: %w",
			domainerrors.ErrInvariantViolation)
	}
	if profile == nil {
		profile = vo.PythonUvProfile{}
	}
	p.profile = profile

	contexts := model.BoundedContexts()
	if len(contexts) == 0 {
		return fmt.Errorf("no bounded contexts to generate tickets for: %w",
			domainerrors.ErrInvariantViolation)
	}

	p.epics = nil
	p.tickets = nil

	// Build lookup of aggregates by context name.
	aggsByCtx := map[string][]vo.AggregateDesign{}
	for _, agg := range model.AggregateDesigns() {
		aggsByCtx[agg.ContextName()] = append(aggsByCtx[agg.ContextName()], agg)
	}

	// Build upstream lookup from context relationships.
	upstreamCtxs := map[string]map[string]bool{}
	for _, rel := range model.ContextRelationships() {
		if upstreamCtxs[rel.Downstream()] == nil {
			upstreamCtxs[rel.Downstream()] = map[string]bool{}
		}
		upstreamCtxs[rel.Downstream()][rel.Upstream()] = true
	}

	epicIDByCtx := map[string]string{}

	for _, bc := range contexts {
		if bc.Classification() == nil {
			return fmt.Errorf("bounded context '%s' has no subdomain classification: %w",
				bc.Name(), domainerrors.ErrInvariantViolation)
		}

		epicID := identity.NewID()
		epicIDByCtx[bc.Name()] = epicID
		classification := *bc.Classification()

		p.epics = append(p.epics, NewGeneratedEpic(
			epicID,
			bc.Name()+" Epic",
			fmt.Sprintf("Implement the %s bounded context (%s subdomain).",
				bc.Name(), string(classification)),
			bc.Name(),
			classification,
		))

		detailLevel := vo.DetailLevelFromClassification(classification)
		ctxAggs := aggsByCtx[bc.Name()]

		if len(ctxAggs) == 0 {
			stubAgg := vo.NewAggregateDesign(bc.Name(), bc.Name(), bc.Name(), nil, nil, nil, nil)
			stubDesc := RenderTicketDetail(stubAgg, vo.TicketDetailStub, profile)
			p.tickets = append(p.tickets, NewGeneratedTicket(
				identity.NewID(),
				"Integrate "+bc.Name()+" boundary",
				stubDesc,
				vo.TicketDetailStub,
				epicID, bc.Name(), bc.Name(), nil, 0,
			))
		} else {
			for _, agg := range ctxAggs {
				desc := RenderTicketDetail(agg, detailLevel, profile)
				p.tickets = append(p.tickets, NewGeneratedTicket(
					identity.NewID(),
					"Implement "+agg.Name()+" aggregate",
					desc,
					detailLevel,
					epicID, bc.Name(), agg.Name(), nil, 0,
				))
			}
		}
	}

	p.assignCrossBCDependencies(upstreamCtxs)
	order, err := p.computeDependencyOrder()
	if err != nil {
		return err
	}
	p.dependencyOrder = &order

	classificationByCtx := map[string]vo.SubdomainClassification{}
	for _, bc := range contexts {
		classificationByCtx[bc.Name()] = *bc.Classification()
	}
	p.reclassifyByDepth(classificationByCtx, profile)
	return nil
}

// Preview returns a human-readable preview of the generated plan.
func (p *TicketPlan) Preview() (string, error) {
	if len(p.epics) == 0 {
		return "", fmt.Errorf("no plan generated yet -- call GeneratePlan() first: %w",
			domainerrors.ErrInvariantViolation)
	}

	fullCount, stdCount, stubCount := 0, 0, 0
	for _, t := range p.tickets {
		switch t.DetailLevel() {
		case vo.TicketDetailFull:
			fullCount++
		case vo.TicketDetailStandard:
			stdCount++
		case vo.TicketDetailStub:
			stubCount++
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Ticket Plan: %s\n", p.planID)
	fmt.Fprintf(&b, "Epics: %d\n", len(p.epics))
	fmt.Fprintf(&b, "Tickets: %d (FULL=%d, STANDARD=%d, STUB=%d)\n\n",
		len(p.tickets), fullCount, stdCount, stubCount)

	for _, epic := range p.epics {
		fmt.Fprintf(&b, "  %s (%s):\n", epic.Title(), strings.ToUpper(string(epic.Classification())))
		for _, ticket := range p.tickets {
			if ticket.EpicID() == epic.EpicID() {
				fmt.Fprintf(&b, "    - [%s] %s\n", string(ticket.DetailLevel()), ticket.Title())
			}
		}
		b.WriteString("\n")
	}
	return b.String(), nil
}

// PromoteStub promotes a STUB ticket to FULL detail.
func (p *TicketPlan) PromoteStub(ticketID string, profile vo.StackProfile) error {
	for i, ticket := range p.tickets {
		if ticket.TicketID() == ticketID {
			if ticket.DetailLevel() != vo.TicketDetailStub {
				return fmt.Errorf("ticket '%s' is %s, not STUB -- cannot promote: %w",
					ticketID, string(ticket.DetailLevel()), domainerrors.ErrInvariantViolation)
			}
			resolvedProfile := profile
			if resolvedProfile == nil {
				resolvedProfile = p.profile
			}
			if resolvedProfile == nil {
				resolvedProfile = vo.PythonUvProfile{}
			}
			agg := vo.NewAggregateDesign(
				ticket.AggregateName(), ticket.BoundedContextName(),
				ticket.AggregateName(), nil, nil, nil, nil,
			)
			newDesc := RenderTicketDetail(agg, vo.TicketDetailFull, resolvedProfile)
			p.tickets[i] = NewGeneratedTicket(
				ticket.TicketID(), ticket.Title(), newDesc,
				vo.TicketDetailFull, ticket.EpicID(),
				ticket.BoundedContextName(), ticket.AggregateName(),
				ticket.Dependencies(), ticket.Depth(),
			)
			return nil
		}
	}
	return fmt.Errorf("ticket '%s' not found: %w", ticketID, domainerrors.ErrInvariantViolation)
}

// Approve approves the plan (all or a subset), emitting TicketPlanApproved.
func (p *TicketPlan) Approve(approvedIDs []string) error {
	if p.approved {
		return fmt.Errorf("plan already approved: %w", domainerrors.ErrInvariantViolation)
	}
	if len(p.tickets) == 0 {
		return fmt.Errorf("cannot approve plan with no tickets: %w", domainerrors.ErrInvariantViolation)
	}

	allIDs := map[string]bool{}
	for _, t := range p.tickets {
		allIDs[t.TicketID()] = true
	}

	var finalApproved, finalDismissed []string
	if approvedIDs == nil {
		for _, t := range p.tickets {
			finalApproved = append(finalApproved, t.TicketID())
		}
	} else {
		for _, id := range approvedIDs {
			if !allIDs[id] {
				return fmt.Errorf("unknown ticket ID: %s: %w", id, domainerrors.ErrInvariantViolation)
			}
		}
		approvedSet := map[string]bool{}
		for _, id := range approvedIDs {
			approvedSet[id] = true
		}
		finalApproved = approvedIDs
		for id := range allIDs {
			if !approvedSet[id] {
				finalDismissed = append(finalDismissed, id)
			}
		}
	}

	p.approved = true
	p.events = append(p.events, NewTicketPlanApproved(p.planID, finalApproved, finalDismissed))
	return nil
}

// PromotionEligibleIDs returns IDs of STUB tickets whose dependencies are all resolved.
func (p *TicketPlan) PromotionEligibleIDs(resolvedIDs map[string]bool) map[string]bool {
	eligible := map[string]bool{}
	for _, ticket := range p.tickets {
		if ticket.DetailLevel() != vo.TicketDetailStub {
			continue
		}
		deps := ticket.Dependencies()
		if len(deps) == 0 {
			continue
		}
		allResolved := true
		for _, depID := range deps {
			if !resolvedIDs[depID] {
				allResolved = false
				break
			}
		}
		if allResolved {
			eligible[ticket.TicketID()] = true
		}
	}
	return eligible
}

// -- Private helpers ----------------------------------------------------------

func (p *TicketPlan) computeTicketDepths() (map[string]int, error) {
	if p.dependencyOrder == nil {
		return nil, fmt.Errorf("dependency order must be computed before depths: %w",
			domainerrors.ErrInvariantViolation)
	}

	allIDs := map[string]bool{}
	for _, t := range p.tickets {
		allIDs[t.TicketID()] = true
	}
	depsMap := map[string][]string{}
	for _, t := range p.tickets {
		var inPlanDeps []string
		for _, d := range t.Dependencies() {
			if allIDs[d] {
				inPlanDeps = append(inPlanDeps, d)
			}
		}
		depsMap[t.TicketID()] = inPlanDeps
	}

	depths := map[string]int{}
	for _, tid := range p.dependencyOrder.OrderedIDs() {
		maxDepth := -1
		for _, d := range depsMap[tid] {
			if dd, ok := depths[d]; ok && dd > maxDepth {
				maxDepth = dd
			}
		}
		if maxDepth >= 0 {
			depths[tid] = maxDepth + 1
		} else {
			depths[tid] = 0
		}
	}
	return depths, nil
}

func (p *TicketPlan) reclassifyByDepth(classificationByCtx map[string]vo.SubdomainClassification, profile vo.StackProfile) {
	depths, err := p.computeTicketDepths()
	if err != nil {
		return
	}

	updated := make([]GeneratedTicket, 0, len(p.tickets))
	for _, ticket := range p.tickets {
		depth := depths[ticket.TicketID()]
		classification := classificationByCtx[ticket.BoundedContextName()]
		tier := vo.ClassifyTier(depth, classification)
		newLevel := vo.TierToDetailLevel(tier, classification)

		if newLevel != ticket.DetailLevel() {
			agg := vo.NewAggregateDesign(
				ticket.AggregateName(), ticket.BoundedContextName(),
				ticket.AggregateName(), nil, nil, nil, nil,
			)
			newDesc := RenderTicketDetail(agg, newLevel, profile)
			updated = append(updated, NewGeneratedTicket(
				ticket.TicketID(), ticket.Title(), newDesc,
				newLevel, ticket.EpicID(),
				ticket.BoundedContextName(), ticket.AggregateName(),
				ticket.Dependencies(), depth,
			))
		} else {
			updated = append(updated, NewGeneratedTicket(
				ticket.TicketID(), ticket.Title(), ticket.Description(),
				ticket.DetailLevel(), ticket.EpicID(),
				ticket.BoundedContextName(), ticket.AggregateName(),
				ticket.Dependencies(), depth,
			))
		}
	}
	p.tickets = updated
}

func (p *TicketPlan) assignCrossBCDependencies(upstreamCtxs map[string]map[string]bool) {
	ticketsByCtx := map[string][]string{}
	for _, t := range p.tickets {
		ticketsByCtx[t.BoundedContextName()] = append(ticketsByCtx[t.BoundedContextName()], t.TicketID())
	}

	updated := make([]GeneratedTicket, 0, len(p.tickets))
	for _, ticket := range p.tickets {
		ctx := ticket.BoundedContextName()
		upstreams := upstreamCtxs[ctx]
		depIDs := make([]string, len(ticket.Dependencies()))
		copy(depIDs, ticket.Dependencies())

		for upstream := range upstreams {
			depIDs = append(depIDs, ticketsByCtx[upstream]...)
		}

		if len(depIDs) != len(ticket.Dependencies()) {
			updated = append(updated, NewGeneratedTicket(
				ticket.TicketID(), ticket.Title(), ticket.Description(),
				ticket.DetailLevel(), ticket.EpicID(),
				ticket.BoundedContextName(), ticket.AggregateName(),
				depIDs, ticket.Depth(),
			))
		} else {
			updated = append(updated, ticket)
		}
	}
	p.tickets = updated
}

func (p *TicketPlan) computeDependencyOrder() (DependencyOrder, error) {
	allIDs := map[string]bool{}
	for _, t := range p.tickets {
		allIDs[t.TicketID()] = true
	}

	inDegree := map[string]int{}
	dependents := map[string][]string{}
	for id := range allIDs {
		inDegree[id] = 0
		dependents[id] = nil
	}

	for _, ticket := range p.tickets {
		for _, depID := range ticket.Dependencies() {
			if allIDs[depID] {
				inDegree[ticket.TicketID()]++
				dependents[depID] = append(dependents[depID], ticket.TicketID())
			}
		}
	}

	// Kahn's algorithm
	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	var ordered []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		ordered = append(ordered, current)
		for _, dep := range dependents[current] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if len(ordered) != len(allIDs) {
		return DependencyOrder{}, fmt.Errorf("circular dependency detected in ticket plan: %w",
			domainerrors.ErrInvariantViolation)
	}

	return NewDependencyOrder(ordered), nil
}
