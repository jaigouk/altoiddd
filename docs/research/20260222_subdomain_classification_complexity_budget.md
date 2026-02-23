# Subdomain Classification and Complexity Budget

**Date:** 2026-02-22
**Type:** Spike Research
**Status:** Complete
**Related PRD Feature:** Complexity Budget (P0)

## Research Questions

1. What are the standard DDD definitions of Core, Supporting, and Generic subdomains?
2. What decision heuristics help classify subdomains?
3. How do you explain these concepts to non-technical users?
4. What treatment level does each classification receive in practice?
5. What real-world examples demonstrate classification across app types?

---

## 1. Authoritative Definitions

### Eric Evans (Domain-Driven Design, 2003)

Evans introduced the three subdomain types in the original Blue Book:

- **Core Domain**: "The distinguishing part of the model, central to the user's goals,
  that differentiates the application and makes it valuable." The core domain is where
  the organization must excel. It deserves the best developers and the most careful
  modeling. ([Evans, DDD, Ch. 15](https://www.amazon.com/Domain-Driven-Design-Tackling-Complexity-Software/dp/0321125215))

- **Supporting Subdomain**: Necessary for the business to function but not a
  differentiator. Requires custom development (no off-the-shelf solution fits), but
  the logic is relatively straightforward. Can be delegated to less experienced teams.
  ([Evans, DDD, Ch. 15](https://www.amazon.com/Domain-Driven-Design-Tackling-Complexity-Software/dp/0321125215))

- **Generic Subdomain**: A solved problem. Solutions exist as commercial or open-source
  products. Building this in-house wastes effort better spent on the core domain.
  ([Evans, DDD, Ch. 15](https://www.amazon.com/Domain-Driven-Design-Tackling-Complexity-Software/dp/0321125215))

### Vaughn Vernon (Implementing Domain-Driven Design, 2013)

Vernon reinforced Evans' definitions with emphasis on resource allocation:

- **Core Domain**: "What a company does differently from its competitors." Must be
  developed in-house with the best team. Receives highest priority and most resources.
  ([Vernon, IDDD, Ch. 2](https://www.happycoders.eu/books/implementing-domain-driven-design/);
  [Vaadin DDD overview](https://vaadin.com/blog/ddd-part-1-strategic-domain-driven-design))

- **Supporting Subdomain**: "Necessary for the organization to succeed, but does not
  fall into the core domain category." Custom development is needed because no
  off-the-shelf solution exists, but the logic is simpler.
  ([Vernon, IDDD, Ch. 2](https://www.happycoders.eu/books/implementing-domain-driven-design/))

- **Generic Subdomain**: "Does not contain anything special to the organization but is
  still needed for the overall solution." Use off-the-shelf software, SaaS, or
  open-source.
  ([Vernon, IDDD, Ch. 2](https://www.happycoders.eu/books/implementing-domain-driven-design/))

Vernon added a critical insight: **the same subdomain can be Core for one company and
Generic for another.** User identity management is Generic for an e-commerce company
but Core for an identity provider like Okta.
([Vaadin DDD overview](https://vaadin.com/blog/ddd-part-1-strategic-domain-driven-design))

### Vlad Khononov (Learning Domain-Driven Design, 2021)

Khononov added two key dimensions beyond Evans' definitions:

| Dimension         | Core              | Supporting         | Generic            |
|-------------------|-------------------|--------------------|--------------------|
| Competitive Advantage | High           | None               | None               |
| Complexity        | High              | Low                | Can be high*       |
| Volatility        | High (changes often) | Low-Medium       | Low                |
| Build/Buy         | Always build in-house | Build (simple)  | Buy/adopt          |
| Team Investment   | Best developers   | Standard team      | Avoid building     |

*Generic subdomains can be complex (e.g., encryption algorithms), but that complexity
has already been solved by others -- you should not re-solve it.

([Khononov, LDDD, Ch. 1](https://www.oreilly.com/library/view/learning-domain-driven-design/9781098100124/ch01.html);
[Book summary by Sodkiewicz](https://sodkiewiczm.medium.com/geek-read-learning-domain-driven-design-28e258a7e7b7))

Khononov's key contribution: **a design heuristics decision tree** (Ch. 10) that
cascades from subdomain type through business logic pattern, architecture, and testing
strategy.

### Nick Tune (Core Domain Charts, 2019-2024)

Tune introduced the **Core Domain Chart** -- a two-axis visualization tool:

- **X-axis: Business Differentiation** -- How much competitive advantage does this
  domain provide? Assessment questions:
  - How difficult is it for competitors to replicate this capability?
  - How much revenue, brand value, or engagement does it generate?
  - What is the potential future value?

- **Y-axis: Model Complexity** -- How difficult is this to implement and maintain?
  Three forms of complexity:
  - Essential domain complexity (conceptual modeling difficulty)
  - Accidental technical complexity (unnecessary technical complications)
  - Operational complexity (external business processes)

Domains are plotted on this chart and naturally cluster into Core (upper-right),
Supporting (lower-left to center), and Generic (lower-left).

([ddd-crew/core-domain-charts on GitHub](https://github.com/ddd-crew/core-domain-charts);
[esilva.net Tune visualization](https://esilva.net/tla_insights/core-domain-charts_tune))

### Wardley Mapping Alignment

The Wardley Map evolution axis (Genesis -> Custom -> Product -> Commodity) maps to
DDD subdomain types:

| Wardley Stage | DDD Subdomain Type | Characteristics |
|---------------|-------------------|-----------------|
| Genesis       | Core              | Novel, high uncertainty, requires exploration |
| Custom-built  | Core/Supporting   | Maturing, still valuable, becoming clearer |
| Product       | Supporting/Generic | Widely available, clear specifications |
| Commodity     | Generic           | Ubiquitous, standardized, utility |

([InfoQ: Architecture for Flow with Wardley Mapping, DDD, and Team Topologies](https://www.infoq.com/presentations/ddd-wardley-mapping-team-topology/);
[Tech Lead Journal #86](https://techleadjournal.dev/episodes/86/))

---

## 2. Decision Heuristics

### Primary Decision Tree (Khononov-derived)

This is the recommended decision tree for alty's guided classification:

```
Q1: Could you buy or adopt an off-the-shelf / open-source solution
    for this WITHOUT compromising your competitive advantage?
    |
    +-- YES --> GENERIC
    |           (Use existing solution: SaaS, library, open-source)
    |
    +-- NO --> Q2: Is the business logic complex?
               |
               +-- NO (simple, CRUD-like) --> SUPPORTING
               |   (Build in-house, keep it simple)
               |
               +-- YES (complex rules, algorithms, invariants) --> CORE
                   (Build in-house, invest heavily, best team)
```

Source: Khononov's three-step evaluation
([vladikk.com: Revisiting the Basics of DDD](https://vladikk.com/2018/01/26/revisiting-the-basics-of-ddd/))

### Complexity Assessment Sub-questions

When Q2 asks "is the business logic complex?", use these heuristics:

| Indicator | Points to... |
|-----------|-------------|
| Resembles a CRUD interface; domain experts describe it in CRUD terms | Simple (Supporting) |
| Business logic primarily revolves around input validation | Simple (Supporting) |
| Contains sophisticated algorithms or calculations | Complex (Core) |
| Has business rules and invariants that must be enforced | Complex (Core) |
| High cyclomatic complexity with many execution paths | Complex (Core) |
| Logic changes frequently as the business optimizes | Complex (Core) |
| Multiple domain experts disagree on the rules | Complex (Core) |

Source: Khononov, LDDD Ch. 10 design heuristics
([O'Reilly: LDDD Ch. 10](https://www.oreilly.com/library/view/learning-domain-driven-design/9781098100124/ch10.html))

### Differentiation Test Questions (Nick Tune)

Additional questions from Core Domain Charts for assessing the differentiation axis:

1. **Replication difficulty**: How hard would it be for a competitor to copy this?
2. **Revenue attribution**: How much revenue or engagement does this directly generate?
3. **Future value**: What is the potential growth or strategic value?
4. **Specialist expertise**: Does this require rare domain knowledge?
5. **Market entry barrier**: Would a new entrant struggle to build this?
6. **Failure impact**: What is the brand/security risk if this fails?

([ddd-crew/core-domain-charts](https://github.com/ddd-crew/core-domain-charts))

### Sanity Check: Reverse Validation

Khononov's heuristic for catching mis-classification:

> "If you assume a Core subdomain but find that a Transaction Script or Active Record
> pattern is sufficient, rethink your assumption. It is probably Supporting."

> "If you assumed Supporting but find yourself building complex domain models with
> aggregates and invariants, rethink -- it may be Core."

([Khononov summary via compiledconversations.com](https://compiledconversations.com/5/))

---

## 3. Non-Technical User Adaptation

### Plain-Language Equivalents

| DDD Term | Plain Language | Analogy | One-liner |
|----------|---------------|---------|-----------|
| Core Subdomain | "Your secret sauce" | The recipe that makes your restaurant famous | "This is what makes you YOU -- no one else does it this way" |
| Supporting Subdomain | "Necessary plumbing" | The kitchen equipment -- essential but not unique | "You need it, but it is not what customers pay for" |
| Generic Subdomain | "Off-the-shelf commodity" | Electricity -- you just plug in, don't build a power plant | "Everyone needs this, and everyone does it the same way" |

### Questions for a Product Owner or Domain Expert

Avoid DDD jargon entirely. Ask in business language:

**Question 1: "If a competitor copied this part of your business exactly, would you
lose your edge?"**
- YES -> This is your secret sauce (Core)
- NO -> Move to Question 2

**Question 2: "Could you replace this with a product you can buy today?"**
- YES -> This is a commodity (Generic)
- NO -> This is necessary plumbing (Supporting)

**Question 3 (validation): "Do your domain experts have strong, specific opinions
about how this should work, with lots of special cases and rules?"**
- YES -> Confirms Core
- NO -> Confirms Supporting

### Visual Metaphor: The Restaurant Analogy

This analogy works well with non-technical stakeholders:

| Part of Restaurant | Type | Why |
|-------------------|------|-----|
| The chef's signature recipes | Core | This is why people come to YOUR restaurant |
| Table reservation system | Supporting | You need it, but any restaurant has one |
| Electricity and water | Generic | You buy it from a utility; you don't generate your own |

### Additional Metaphors Tested in Practice

**The House Analogy** (works for construction/real estate audiences):
- Core = The architectural design and layout that makes this house special
- Supporting = The custom cabinets -- built for this house, but not the selling point
- Generic = Standard electrical wiring -- hire an electrician, don't invent it

**The Car Analogy** (works for manufacturing audiences):
- Core = The engine design and driving experience (what makes a BMW a BMW)
- Supporting = The infotainment system -- custom but not the core brand
- Generic = Tires -- buy from Michelin, do not manufacture rubber

---

## 4. Treatment Levels per Classification

### Khononov's Design Heuristics Decision Tree (LDDD, Ch. 10)

| Dimension | Core | Supporting | Generic |
|-----------|------|------------|---------|
| **Business Logic Pattern** | Domain Model or Event-Sourced Domain Model | Transaction Script or Active Record | N/A (use existing solution) |
| **Architecture** | Ports & Adapters (Hexagonal), CQRS, Event Sourcing | Layered Architecture or simple Service | Integration layer only (Anti-Corruption Layer) |
| **Testing Strategy** | Comprehensive: unit tests on aggregates, integration tests on use cases, property-based tests on invariants | Standard: unit tests on services, integration tests on persistence | Minimal: integration tests on the boundary/ACL only |
| **Team Investment** | Senior developers, domain experts, pair programming | Standard team, potentially outsourced | No development team -- configure/integrate only |
| **Modeling Effort** | Deep modeling: EventStorming, domain stories, ubiquitous language glossary | Light modeling: simple entity-relationship, basic flows | No modeling: use the vendor's model |

Sources:
- [Khononov, LDDD Ch. 10](https://www.oreilly.com/library/view/learning-domain-driven-design/9781098100124/ch10.html)
- [Geabunea: DDD Tactical Decision Tree](https://medium.com/@dangeabunea/the-domain-driven-design-tactical-decision-tree-i-wish-i-knew-10-years-ago-906b65905b33)
- [SAP DDD Resources](https://github.com/SAP/curated-resources-for-domain-driven-design/blob/main/blog/0002-core-concepts.md)

### alty Treatment Levels (Proposed)

Mapping the above to alty's concrete outputs:

| Dimension | Core | Supporting | Generic |
|-----------|------|------------|---------|
| **DDD Artifacts** | Full: domain stories, aggregate design, ubiquitous language glossary, bounded context canvas | Light: entity list, basic flow description | None: integration notes only |
| **Architecture** | Hexagonal (Ports & Adapters), domain layer with zero external deps | Simple layered or transaction script | Anti-Corruption Layer wrapping external lib/SaaS |
| **Fitness Functions** | Strict: layer dependency rules, aggregate isolation, forbidden cross-context imports, independence contracts | Moderate: layer dependency rules, basic import boundaries | Minimal: ACL boundary test only |
| **Ticket Generation** | Detailed: full acceptance criteria, TDD phases (RED/GREEN/REFACTOR), SOLID mapping, domain invariant tests, edge cases | Standard: acceptance criteria, basic test requirements | Stub: "integrate X library/service", verify boundary |
| **Ticket Tier** | Near-term with full specification | Near-term with standard specification | Far-term stubs until promoted |
| **Test Coverage Target** | >= 90% on domain logic, >= 80% overall | >= 80% | >= 60% (boundary tests only) |
| **Code Review Focus** | Domain model richness, ubiquitous language compliance, aggregate boundaries | Correctness, simplicity | Integration correctness |
| **Agent Persona Depth** | Full domain context in agent persona, ubiquitous language enforced | Basic domain context | Minimal -- just integration interface |
| **import-linter Contracts** | `layers` + `forbidden` + `independence` + `acyclic_siblings` | `layers` + basic `forbidden` | Single `forbidden` (no direct access, use ACL) |
| **pytestarch Rules** | Aggregate isolation, no cross-context imports, value object immutability | Layer dependency direction | ACL boundary only |

### Treatment Level Impact on Project Velocity

The complexity budget prevents over-engineering. Applying full DDD treatment to a
Generic subdomain (e.g., building a custom auth system when Auth0 exists) is a common
anti-pattern that wastes weeks. Conversely, under-investing in Core (e.g., using CRUD
for complex pricing logic) creates technical debt that compounds with every feature.

| Anti-pattern | Consequence | Detection |
|-------------|-------------|-----------|
| Full DDD on Generic | Wasted effort, slow delivery | If Transaction Script would suffice, subdomain is not Core |
| CRUD on Core | Anemic model, scattered business logic | If aggregates and invariants emerge, subdomain is not Supporting |
| Building Generic in-house | NIH syndrome, maintenance burden | If off-the-shelf solutions exist, do not build |

---

## 5. Real-World Classification Examples

### E-Commerce Platform

| Subdomain | Classification | Rationale |
|-----------|---------------|-----------|
| Product Catalog & Search | Core | Differentiates the shopping experience; recommendation algorithms, faceted search, merchandising rules |
| Pricing & Promotions | Core | Dynamic pricing, discount rules, bundle logic -- competitive differentiator |
| Order Fulfillment | Core or Supporting* | Complex for Amazon (Core), simpler for small shops (Supporting) |
| Inventory Management | Supporting | Needed for operations, not customer-facing differentiator |
| Customer Notifications | Supporting | Email/SMS/push -- custom templates but standard delivery |
| Shopping Cart | Supporting | Business-specific rules (expiry, merge) but not the differentiator |
| Payment Processing | Generic | Stripe, PayPal, Adyen -- solved problem, use SaaS |
| Authentication & User Management | Generic | Auth0, Cognito, Keycloak -- commodity |
| Tax Calculation | Generic | Avalara, TaxJar -- regulatory compliance, not differentiator |

*Demonstrates Vernon's insight: classification depends on the specific company.

Sources:
- [simonatta.medium.com: eCommerce by DDD](https://simonatta.medium.com/e-commerce-by-ddd-bf4459272188)
- [ttulka/ddd-example-ecommerce](https://github.com/ttulka/ddd-example-ecommerce)

### B2B SaaS Platform

| Subdomain | Classification | Rationale |
|-----------|---------------|-----------|
| Core Domain Logic (the "thing" the SaaS does) | Core | The entire value proposition -- what customers pay for |
| Tenant Management & Multi-tenancy | Supporting | Needed for SaaS but not what customers buy |
| Subscription & Billing | Supporting or Generic* | Stripe Billing handles most cases (Generic); custom usage-based billing may be Supporting |
| User Management & RBAC | Generic | Standard patterns, use existing libraries |
| Audit Logging | Supporting | Compliance requirement, custom but not differentiating |
| Email/Notifications | Generic | SendGrid, SES -- commodity |
| Analytics Dashboard | Supporting or Core* | Generic reporting = Supporting; AI-powered insights = potentially Core |

### Healthcare Scheduling App

| Subdomain | Classification | Rationale |
|-----------|---------------|-----------|
| Appointment Scheduling & Optimization | Core | The matching algorithm, constraint satisfaction, waitlist management |
| Provider Availability Management | Core | Complex rules: credentials, locations, time blocks, overrides |
| Patient Registration | Supporting | Custom fields but straightforward CRUD |
| Insurance Verification | Generic | Use Eligible API or similar |
| Notifications & Reminders | Generic | Twilio, SNS -- commodity |
| Payments & Co-pays | Generic | Stripe -- solved problem |

### Developer Tooling / CLI (like alty itself)

| Subdomain | Classification | Rationale |
|-----------|---------------|-----------|
| Guided DDD Discovery | Core | The conversational question framework, classification logic |
| Artifact Generation | Core | PRD/DDD/Architecture generation from domain model |
| Multi-tool Config Generation | Core | Domain-aware config for Claude Code, Cursor, etc. |
| Fitness Function Generation | Core | Novel: bounded context map to import-linter/pytestarch |
| Ticket Pipeline | Supporting | Dependency ordering, template filling -- important but algorithmic |
| Knowledge Base (RLM) | Supporting | Addressable docs, versioning -- custom but not the differentiator |
| File I/O & Git Operations | Generic | Use pathlib, gitpython, or subprocess -- commodity |
| CLI Framework | Generic | Use click, typer, or argparse -- commodity |

---

## 6. Guided Classification Flow for alty

### Recommended Implementation

During the DDD discovery conversation, after identifying subdomains (via domain stories
or EventStorming), alty should guide classification with this flow:

```
For each identified subdomain:

Step 1: "Let's figure out how important [subdomain name] is to your business."

Step 2: Ask the BUY question:
  "Could you use an existing product or service for [subdomain name]
   without losing what makes your business special?"
  - If YES -> classify as GENERIC
  - If NO -> continue

Step 3: Ask the COMPLEXITY question:
  "When your team talks about [subdomain name], do they describe it
   as mostly storing and retrieving data, or are there complex rules,
   special cases, and calculations involved?"
  - "Mostly storing/retrieving" -> classify as SUPPORTING
  - "Complex rules and special cases" -> classify as CORE

Step 4: Validate with the COMPETITOR question:
  "If a competitor copied your [subdomain name] exactly, would that
   threaten your business?"
  - If YES -> confirms CORE
  - If NO, and was classified CORE -> reconsider as SUPPORTING

Step 5: Present classification with plain-language explanation:
  "Based on your answers, [subdomain name] looks like your SECRET SAUCE /
   NECESSARY PLUMBING / OFF-THE-SHELF. Here's what that means for how
   we'll build it: [treatment summary]"

Step 6: Allow override with warning:
  User can reclassify. If upgrading (Supporting -> Core), warn about
  increased effort. If downgrading (Core -> Supporting), warn about
  under-investment risk.
```

### Output Format

The classification should be stored in the DDD artifacts (e.g., `docs/DDD.md`)
in a structured format:

```yaml
subdomains:
  - name: "Pricing Engine"
    classification: core
    rationale: "Dynamic pricing with customer-specific rules; competitors cannot replicate"
    treatment:
      architecture: "hexagonal"
      testing: "comprehensive"
      fitness_functions: "strict"
      ticket_detail: "full"

  - name: "Notification Delivery"
    classification: supporting
    rationale: "Custom templates but standard delivery mechanisms"
    treatment:
      architecture: "layered"
      testing: "standard"
      fitness_functions: "moderate"
      ticket_detail: "standard"

  - name: "User Authentication"
    classification: generic
    rationale: "Use Auth0/Keycloak; no competitive advantage in custom auth"
    treatment:
      architecture: "acl_wrapper"
      testing: "boundary"
      fitness_functions: "minimal"
      ticket_detail: "stub"
```

---

## 7. Key Risks and Open Questions

### Risk: Mis-classification by Non-Technical Users

Non-technical users may over-classify subdomains as Core ("everything is important!").
Mitigation: The competitor-copy question is the strongest corrective. If copying it
would not threaten the business, it is not Core.

### Risk: Classification Changes Over Time

Subdomains evolve. What starts as Core may become Generic as the market matures
(Wardley evolution). alty should support reclassification with a migration path
that adjusts treatment levels.

### Risk: Boundary Between Supporting and Generic

The hardest classification decision. The key differentiator: can you buy it? If yes,
Generic. If no existing solution fits, it is Supporting even if simple.

### Open Question: Automated Classification Assistance

No tool currently automates subdomain classification. Qlerify (commercial SaaS) is the
closest, offering DDD-aware domain discovery, but classification is still manual.
alty's guided question flow would be novel.
([Qlerify subdomain docs](https://www.qlerify.com/dddconcepts/subdomain))

### Open Question: Granularity of Treatment Levels

Should there be a spectrum (1-5 scale) rather than three discrete levels? Khononov's
approach is three types with sub-heuristics. Recommendation: start with three discrete
levels (simpler for users), add nuance later if needed.

---

## 8. Summary of Sources

| Source | Type | Key Contribution |
|--------|------|-----------------|
| Evans, DDD (2003) | Book | Original Core/Supporting/Generic definitions |
| Vernon, IDDD (2013) | Book | Context-dependent classification, resource allocation |
| Khononov, LDDD (2021) | Book | Design heuristics decision tree, complexity/volatility dimensions |
| Nick Tune, Core Domain Charts (2019) | Framework | Two-axis visualization (differentiation x complexity) |
| Wardley Mapping | Framework | Evolution axis alignment with subdomain types |
| SAP DDD Resources | Guide | Implementation guidance per subdomain type |
| Qlerify | Tool | Only commercial tool with DDD subdomain discovery |

Full URLs:
- https://vladikk.com/2018/01/26/revisiting-the-basics-of-ddd/
- https://vaadin.com/blog/ddd-part-1-strategic-domain-driven-design
- https://deviq.com/domain-driven-design/subdomain/
- https://www.qlerify.com/dddconcepts/subdomain
- https://github.com/ddd-crew/core-domain-charts
- https://github.com/SAP/curated-resources-for-domain-driven-design/blob/main/blog/0002-core-concepts.md
- https://esilva.net/tla_insights/core-domain-charts_tune
- https://www.infoq.com/presentations/ddd-wardley-mapping-team-topology/
- https://www.oreilly.com/library/view/learning-domain-driven-design/9781098100124/ch10.html
- https://medium.com/@dangeabunea/the-domain-driven-design-tactical-decision-tree-i-wish-i-knew-10-years-ago-906b65905b33
- https://simonatta.medium.com/e-commerce-by-ddd-bf4459272188
- https://github.com/ttulka/ddd-example-ecommerce
