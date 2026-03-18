# DDD Collaborative Modeling — State of the Art (2026)

> **Date:** 2026-03-05
> **Type:** Research
> **Purpose:** Survey of DDD meetings, techniques, and artifacts that practitioners actually use in 2025-2026. Informs alto's guided discovery flow and generated artifacts.

## The DDD Starter Modelling Process (8 Steps)

The [ddd-crew](https://github.com/ddd-crew) process is the de facto standard for DDD practitioners. Source: [ddd-starter-modelling-process](https://ddd-crew.github.io/ddd-starter-modelling-process/)

| Step | Meeting/Technique | Artifact Produced |
|------|-------------------|-------------------|
| 1. Understand | Business Model Canvas, User Story Mapping | Business goals & user needs alignment |
| 2. Discover | EventStorming, Domain Storytelling, Example Mapping | Visual domain model, shared understanding |
| 3. Decompose | EventStorming with sub-domains, Business Capability Modelling | Sub-domain map |
| 4. Strategize | Core Domain Charts, Wardley Mapping | Core/Supporting/Generic classification (build/buy/outsource) |
| 5. Connect | Domain Message Flow Modelling | Interaction architecture between sub-domains |
| 6. Organise | Team Topologies + Context Maps | Team structures aligned with bounded contexts |
| 7. Define | Bounded Context Canvas | Per-context roles, responsibilities, interfaces |
| 8. Code | Aggregate Design Canvas, Design-Level EventStorming | Working domain model |

**Key insight:** On a real project, teams switch between all 8 steps iteratively based on new insights. The linear order is a learning aid, not a rigid process.

## The Big Three Collaborative Techniques

### 1. EventStorming (Alberto Brandolini)

Still the most popular collaborative modeling technique. Three levels:

- **Big Picture** — 15-20 people, maps the complete business domain, surfaces bottlenecks and risks. Starting point for most DDD initiatives.
- **Process Level** — Focuses on a single business process, uncovers ambiguities and redesign opportunities.
- **Design Level** — Deep dive into one bounded context, produces aggregates, commands, events, policies. Closest to code.

**Sticky note colors:** Orange (domain events), Blue (commands), Yellow (actors), Purple (policies/business rules), Pink (hotspots/questions), Lilac (external systems), Green (read models).

### 2. Domain Storytelling (Stefan Hofer & Henning Schwentner)

Rising fast in adoption. Uses actors + work objects + activities as visual stories.

- Better for engaging non-technical stakeholders (narrative format vs sticky notes)
- Used *before* EventStorming to establish domain narrative
- Then hand off to design-level EventStorming for the "how"
- Stories are recorded as diagrams with simple icons, arrows, and text
- Helps uncover misunderstandings, contradictions, and plot holes

**When to use:** Early stages, when working with domain experts who aren't comfortable with EventStorming's abstract format.

### 3. Example Mapping

Quick (30 min) sessions for specification by example.

- Used *after* EventStorming to detail individual stories
- Produces: Rules, Examples, Questions per story
- Yellow cards (stories), Blue (rules), Green (examples), Red (questions)

**When to use:** After EventStorming or Domain Storytelling, when you need to specify the details of a specific business rule or policy.

## Key Canvases (Structured Artifacts)

### Bounded Context Canvas (v4)

Source: [ddd-crew/bounded-context-canvas](https://github.com/ddd-crew/bounded-context-canvas)

The standard structured document for each bounded context:
- **Name** and **Purpose** (short description)
- **Strategic Classification** — domain (core/supporting/generic), business model (revenue/engagement/compliance), evolution (genesis/custom/product/commodity)
- **Domain Roles** — specification model, execution model, analysis model, gateway, etc.
- **Inbound Communication** — messages/queries this context accepts
- **Outbound Communication** — messages/queries this context sends
- **Ubiquitous Language** — key terms defined within this context
- **Business Decisions** — key rules and policies
- **Assumptions** — what we believe but haven't verified

Templates available: Miro (v4), draw.io, Excalidraw, Lucidchart.

### Aggregate Design Canvas

Source: [ddd-crew/aggregate-design-canvas](https://github.com/ddd-crew/aggregate-design-canvas)

Structured approach to designing aggregates:
- **Name** and **Description**
- **State Transitions** — lifecycle of the aggregate
- **Enforced Invariants** — consistency rules
- **Handled Commands** — what triggers state changes
- **Created Events** — what the aggregate emits
- **Throughput** — expected load characteristics

### Core Domain Chart

Visual quadrant placing subdomains by:
- X-axis: Model complexity (simple → complex)
- Y-axis: Business differentiation (commodity → competitive advantage)

Quadrants map to: Generic (buy/outsource), Supporting (build simple), Core (invest heavily).

### Architecture for Flow Canvas (Nick Tune, 2025)

Source: [architectureforflow.com/canvas](https://architectureforflow.com/canvas/)

Integrates three frameworks:
- **Wardley Mapping** — evolution and value chain positioning
- **DDD** — bounded contexts and domain model
- **Team Topologies** — stream-aligned, platform, enabling, complicated-subsystem teams

## New in 2025-2026

### Books

- **Annegret Junker, "Mastering Domain-Driven Design"** (Jan 2025) — First book treating Domain Storytelling + EventStorming + Context Mapping as a unified workflow. Covers canvases, capability maps, visual glossaries, and API design. Author is Chief Software Architect at codecentric AG with 30+ years experience.
  - Source: [Amazon](https://www.amazon.com/Mastering-Domain-Driven-Design-Collaborative-storytelling/dp/936589252X)

### Conferences

- **DDD Europe 2026** — Workshops June 8-10, conference June 10-12, Antwerp. Includes EventStorming Masterclass with Alberto Brandolini, Domain Storytelling workshops, Architecture for Flow.
  - Source: [2026.dddeurope.com](https://2026.dddeurope.com/)
- **Explore DDD 2026** — September 23-25, Denver. Focus on how AI reshapes software design, modeling, and architecture.
  - Source: [exploreddd.com](https://exploreddd.com/)

### Trends

- **AI-assisted modelling** is a major theme at both conferences — how AI tools can participate in collaborative modeling sessions
- **Architecture for Flow** (Nick Tune) — Wardley Mapping as a first-class DDD artifact
- **Storystorming** ([storystorming.com](https://storystorming.com/)) — emerging hybrid of Domain Storytelling and EventStorming

## Gap Analysis: alto vs 2026 State of the Art

| alto currently generates | Status |
|--------------------------|--------|
| Domain Stories (step 2) | Done |
| Ubiquitous Language glossary | Done |
| Subdomain classification — Core/Supporting/Generic (step 4) | Done |
| Context Map with relationships (step 5-6) | Done |
| Aggregate design with invariants (step 8) | Done |

| alto is missing | Priority | Notes |
|-----------------|----------|-------|
| Business Model Canvas (step 1) | Low | Overkill for solo dev / small project bootstrapping |
| Core Domain Chart visual | Medium | alto classifies subdomains but doesn't produce the visual chart |
| Domain Message Flow (step 5) | Medium | alto has context map but not explicit message flow diagrams |
| Bounded Context Canvas (step 7) | **High** | The standard structured artifact for documenting each context. alto generates text-based context descriptions but not the canvas format. |
| Example Mapping | Low | Too granular for bootstrapping — useful during implementation |
| Team Topologies alignment (step 6) | Low | Relevant for teams, not solo devs |
| Wardley Mapping integration | Low | Valuable but complex; better as a future spike |

### Recommendation

The **Bounded Context Canvas** is the highest-value gap. It's the industry standard for documenting bounded contexts and would make alto's output immediately recognizable to DDD practitioners. Consider adding canvas generation to `alto generate` output.

## Sources

- [DDD Starter Modelling Process](https://ddd-crew.github.io/ddd-starter-modelling-process/)
- [DDD Crew GitHub](https://github.com/ddd-crew)
- [Bounded Context Canvas](https://github.com/ddd-crew/bounded-context-canvas)
- [Bounded Context Canvas V3 — Nick Tune](https://medium.com/nick-tune-tech-strategy-blog/bounded-context-canvas-v2-simplifications-and-additions-229ed35f825f)
- [Aggregate Design Canvas](https://github.com/ddd-crew/aggregate-design-canvas)
- [Architecture for Flow Canvas](https://architectureforflow.com/canvas/)
- [Mastering DDD — Annegret Junker](https://www.amazon.com/Mastering-Domain-Driven-Design-Collaborative-storytelling/dp/936589252X)
- [DDD Europe 2026](https://2026.dddeurope.com/)
- [Explore DDD 2026](https://exploreddd.com/)
- [EventStorming for DDD — Jakub Lambrych](https://medium.com/@lambrych/eventstorming-for-domain-driven-design-strengths-and-limitations-3f0b49009c38)
- [Domain Storytelling (Open Practice Library)](https://openpracticelibrary.com/practice/domain-storytelling/)
- [Why EventStorming Practitioners Should Try Domain Storytelling](https://kalele.io/why-eventstorming-practitioners-should-try-domain-storytelling/)
- [Event Storming & Domain Storytelling (Axxes)](https://www.axxes.com/en/insights/event-storming-domain-storytelling)
- [From Event Storming to User Stories](https://www.qlerify.com/post/from-event-storming-to-user-stories)
- [Context Mapper — Bounded Context](https://contextmapper.org/docs/bounded-context/)
- [Martin Fowler — Bounded Context](https://www.martinfowler.com/bliki/BoundedContext.html)
- [DDD Academy](https://ddd.academy/)
