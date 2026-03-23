---
last_reviewed: 2026-03-23
owner: researcher
status: complete
---

# Domain Storytelling Methodology — Research Report

**Date:** 2026-03-23
**Purpose:** Evaluate Domain Storytelling (Hoppe & Schwentner) as a lighter-weight alternative to Event Storming for AI-guided DDD discovery in alto.
**Context:** This research supports spike alty-cli-jcf (Conversational UX + Domain Storytelling) and the alto PRD P0 requirement for "Conversational DDD discovery" (docs/PRD.md, line 129).

---

## Table of Contents

1. [Notation and Format](#1-notation-and-format)
2. [Workshop Format](#2-workshop-format)
3. [Mapping to DDD Concepts](#3-mapping-to-ddd-concepts)
4. [Domain Storytelling vs Event Storming](#4-domain-storytelling-vs-event-storming)
5. [Stories to Bounded Contexts](#5-stories-to-bounded-contexts)
6. [Minimum Viable Story Set](#6-minimum-viable-story-set)
7. [Suitability for AI-Guided Discovery](#7-suitability-for-ai-guided-discovery)
8. [Implications for alto](#8-implications-for-alto)
9. [Sources](#9-sources)

---

## 1. Notation and Format

### 1.1 The Sentence Structure

Domain Storytelling captures domain knowledge as concrete stories following a natural-language sentence pattern:

```
Actor  -->  Activity  -->  Work Object  [-->  (with) Actor]
(who)       (does what)    (with what)        (with whom)
```

This mirrors the Subject-Verb-Object structure of natural speech. Each story is a single concrete scenario — never an abstract process covering all possible paths.

**Source:** [domainstorytelling.org/quick-start-guide](https://domainstorytelling.org/quick-start-guide)

### 1.2 Pictographic Notation Elements

| Element | Representation | Purpose | Example |
|---------|---------------|---------|---------|
| **Actors** | Icons (person, group, system) | Who performs the action | Cashier, Customer, Billing System |
| **Work Objects** | Document/item icons | What is acted upon or exchanged | Reservation, Seating Plan, Invoice |
| **Activities** | Labeled arrows with domain verbs | What action is performed | "suggests", "searches", "confirms", "cancels" |
| **Sequence Numbers** | Numbers at arrow origins | Temporal ordering of sentences | 1, 2, 3... (chronological) |
| **Annotations** | Text callouts | Variations, assumptions, domain terms | "Only on weekdays", "Requires approval" |
| **Groups** | Outlined clusters | Repeated actions, phases, locations, org boundaries | "Payment Phase", "Back Office" |

**Key rule:** Each actor appears only once per story (as an icon). Activities are always labeled with domain-language verbs, never technical implementation terms.

**Source:** [domainstorytelling.org/quick-start-guide](https://domainstorytelling.org/quick-start-guide), Hofer & Schwentner, "Domain Storytelling" (Addison-Wesley, 2022)

### 1.3 Sequential Numbering

Activities are numbered in chronological order. Without numbering, you have a graph, not a story. Numbers ensure:
- Reproducible narrative (anyone can "replay" the story)
- Clear temporal dependencies
- Easy identification of handoff points

**Source:** [emmanuelvalverderamos.substack.com/p/introduction-to-domain-storytelling](https://emmanuelvalverderamos.substack.com/p/introduction-to-domain-storytelling)

### 1.4 One Story at a Time

Domain Storytelling uses scenario-based modeling. Each story captures ONE concrete scenario:
- No branching logic or conditionals within a single story
- "Sometimes" or "or" in a story signals need for a separate story
- Alternatives are modeled as separate domain stories
- The constraint to single storylines is intentional: "we are doing scenario-based modeling. Every domain story is telling one scenario."

**Source:** [InfoQ Podcast — Domain Storytelling with Hofer & Schwentner](https://www.infoq.com/podcasts/domain-storytelling/)

### 1.5 Three Scope Dimensions

Each domain story can be characterized along three dimensions:

| Dimension | Options | When to Use |
|-----------|---------|-------------|
| **Granularity** | Coarse-grained (overview) to fine-grained (detailed) | Coarse for bounded context discovery; fine for implementation design |
| **Time** | As-Is (current process) vs To-Be (future state) | As-Is first to ground reality; To-Be for design |
| **Domain Purity** | Pure (no software mentioned) vs Digitalized (includes systems) | Pure first to capture intent; Digitalized for implementation |

Coarse-grained stories "use one icon just for the whole company, without going down into different roles." Fine-grained stories detail individual steps within a specific process. The granularity you choose determines what DDD artifacts you can extract.

**Source:** [domainstorytelling.org/quick-start-guide](https://domainstorytelling.org/quick-start-guide), [Tech Lead Journal #75 — Stefan Hofer](https://techleadjournal.dev/episodes/75/)

### 1.6 What Domain Storytelling is NOT

- NOT exhaustive process specification covering all paths
- NOT technical sequence diagrams (HTTP calls, DB writes)
- NOT BPMN with gateways and exhaustive branching
- NOT UI click-tracking ("user clicks button X")

It prefers business-level intent: "places order" not "sends POST /checkout".

**Source:** [emmanuelvalverderamos.substack.com/p/introduction-to-domain-storytelling](https://emmanuelvalverderamos.substack.com/p/introduction-to-domain-storytelling)

---

## 2. Workshop Format

### 2.1 Participants

| Role | Responsibility |
|------|---------------|
| **Domain Experts** (storytellers) | Tell concrete stories about how they work |
| **Developers / Analysts** (listeners) | Ask clarifying questions, learn the domain |
| **Moderator** (facilitator) | Guides conversation, records stories in pictographic language, validates understanding |

The moderator records the stories visually while domain experts narrate, allowing experts to immediately verify whether they are understood correctly. This "campfire experience" — being present during creation — conveys understanding that finished diagrams alone cannot.

**Source:** [openpracticelibrary.com/practice/domain-storytelling/](https://openpracticelibrary.com/practice/domain-storytelling/), [Tech Lead Journal #75](https://techleadjournal.dev/episodes/75/)

### 2.2 Typical Workshop Flow

1. **Pick one concrete scenario** (not vague requirements)
2. **Narrate in sentences**: "who does what with what, in what order"
3. **Draw and number each step** using pictographic language
4. **Capture assumptions** as annotations
5. **Stop** when the story makes end-to-end sense
6. **List variations**, then model significant ones as separate stories
7. **Repeat** for the next scenario

**Facilitator questions that drive elicitation:**
- "Can you walk me through a typical [process] from start to finish?"
- "Who starts this process? What triggers it?"
- "What happens next?"
- "Who does that?"
- "What do they use to do it?" / "Where do they get this information?"
- "What do they produce / hand off?"
- "How do they determine what to do next?"
- "What vocabulary do you use for this?"

**Source:** [domainstorytelling.org/quick-start-guide](https://domainstorytelling.org/quick-start-guide), [docs/research/20260222_ddd_question_framework.md lines 48-72](/home/kusanagi/Alty/alty-cli/docs/research/20260222_ddd_question_framework.md)

### 2.3 Session Duration and Story Count

- **Per story:** 15-30 minutes depending on complexity
- **Per workshop session:** 1-2 hours (aligns with typical meeting slots)
- **"Usually a handful of examples is enough to understand even complex business processes."** — Stefan Hofer
- **"Maybe two, three examples are enough to really understand a business process."** — Stefan Hofer
- **3-5 workshops** should produce enough knowledge to communicate with domain experts about their needs, requirements, and problems
- After very few stories, participants can talk about the people, tasks, tools, work objects, and events in that domain

**Source:** [Tech Lead Journal #75](https://techleadjournal.dev/episodes/75/), [InfoQ Podcast](https://www.infoq.com/podcasts/domain-storytelling/), [openpracticelibrary.com](https://openpracticelibrary.com/practice/domain-storytelling/)

### 2.4 No Upfront Training Needed

Participants do not need to learn the notation beforehand. The moderator teaches by doing — "let participants learn through doing." The notation is simple enough that domain experts grasp it within the first story.

**Source:** [InfoQ Podcast](https://www.infoq.com/podcasts/domain-storytelling/)

---

## 3. Mapping to DDD Concepts

### 3.1 Element-to-Artifact Mapping

| Domain Story Element | DDD Artifact | How to Identify |
|---------------------|-------------|-----------------|
| **Actors** | User roles, bounded context entry points, external systems | Who appears in stories; each distinct actor often maps to a role that interacts with one or more bounded contexts |
| **Work Objects** | Entities / Aggregates | Objects that are created, modified, or exchanged; "reservation...will probably become a class" (Hofer) |
| **Work Object properties** | Value Objects | Discoverable attributes: "movie or time could become value objects" (Hofer) |
| **Activities on Work Objects** | Commands / Methods / Use Cases | "they can make a reservation, they can cancel it" — verbs become the API surface |
| **Sequence of activities** | Domain Events (implicit) | State transitions between numbered steps; Event Storming makes these explicit |
| **Groups / spatial clusters** | Bounded Contexts (candidates) | "group parts of the story together that belong in the same subdomain" |
| **Handoffs between groups** | Context Map relationships | Where actors pass work objects across group boundaries |
| **Vocabulary differences** | Context boundary signals | Same term used differently = different bounded context |
| **Annotations** | Business rules / Invariants (implicit) | "Only if...", "Must be...", "Cannot..." patterns |

**Source:** [Tech Lead Journal #75](https://techleadjournal.dev/episodes/75/), [domainstorytelling.org/domain-driven-design](https://domainstorytelling.org/domain-driven-design), [docs/research/20260222_ddd_question_framework.md lines 88-103](/home/kusanagi/Alty/alty-cli/docs/research/20260222_ddd_question_framework.md)

### 3.2 Coarse-Grained vs Fine-Grained Mapping

**Coarse-grained stories** (strategic design):
- Actors map to organizational roles or entire departments
- Work objects are high-level business concepts (Order, Customer, Product)
- Groups reveal bounded context candidates
- Handoffs reveal context map relationships
- Best for: identifying subdomains, bounded contexts, and context maps

**Fine-grained stories** (tactical design):
- Actors map to specific user roles within a bounded context
- Work objects become aggregate candidates
- Work object properties become value objects
- Activities become commands/methods on aggregates
- Best for: designing aggregates, entities, value objects, and commands

The progression: "When finding boundaries, it's typically a good idea to be more coarse-grained to look from a bird's eye perspective." Then zoom into core subdomains with fine-grained stories for implementation design.

**Source:** [domainstorytelling.org/domain-driven-design](https://domainstorytelling.org/domain-driven-design), [Tech Lead Journal #75](https://techleadjournal.dev/episodes/75/)

### 3.3 From Stories to Ubiquitous Language

Domain Storytelling naturally builds ubiquitous language:
- Work object labels become the vocabulary (nouns)
- Activity labels become the verbs
- Annotations capture definitions and disambiguation
- The language emerges from domain expert narration, not developer invention
- "After very few stories, we are able to talk about the people, tasks, tools, work objects, and events in that domain"

**Source:** [domainstorytelling.org](https://domainstorytelling.org/), [openpracticelibrary.com](https://openpracticelibrary.com/practice/domain-storytelling/)

### 3.4 From Stories to Acceptance Tests

Each domain story directly seeds an acceptance test:
- The story IS a Given/When/Then scenario
- Given: initial actors and work objects in a state
- When: the sequence of activities occurs
- Then: the resulting state of work objects

**Source:** [emmanuelvalverderamos.substack.com/p/introduction-to-domain-storytelling](https://emmanuelvalverderamos.substack.com/p/introduction-to-domain-storytelling), [richard-seidl.com/en/blog/domain-storytelling](https://www.richard-seidl.com/en/blog/domain-storytelling)

---

## 4. Domain Storytelling vs Event Storming

### 4.1 Side-by-Side Comparison

| Dimension | Domain Storytelling | Event Storming |
|-----------|-------------------|----------------|
| **Format** | Narrative (sentence-based) | Timeline (event-based) |
| **Starting point** | "Who does what with what?" | "What events happen?" |
| **Notation** | Actors, arrows, work objects, sequence numbers | Colored sticky notes (7+ colors) |
| **Facilitation** | Moderator-led, one storyteller at a time | Group chaotic exploration, then structured |
| **Iteration size** | Very small (one sentence at a time, immediate consolidation) | Large (brainstorm events, then organize) |
| **Complexity for participants** | Low — natural language, no special training | Medium — must learn sticky note color scheme |
| **Scope per session** | One concrete scenario | Entire business domain (Big Picture) or one process |
| **Non-technical accessibility** | High — narrative format familiar to everyone | Medium — abstract event taxonomy requires explanation |
| **Output artifact** | Visual narrative diagrams with domain vocabulary | Event timelines revealing behavior patterns and dependencies |
| **Structural model** | Actors + work objects naturally suggest entities | Events + commands suggest behavior; entities require supplementary modeling |
| **Bounded context discovery** | Via story groups, handoffs, vocabulary shifts | Via event clustering, pivot events, swimlanes |
| **Session size** | 2-8 people (small group, focused) | 5-20+ people (larger, cross-functional) |
| **Time per session** | 15-30 min per story, 1-2 hours per workshop | 2-4 hours for Big Picture, 1-2 hours for Process Level |

**Source:** [axxes.com/en/insights/event-storming-domain-storytelling](https://www.axxes.com/en/insights/event-storming-domain-storytelling), [kalele.io/why-eventstorming-practitioners-should-try-domain-storytelling/](https://kalele.io/why-eventstorming-practitioners-should-try-domain-storytelling/), [thoughtworks.com/radar/techniques/domain-storytelling](https://www.thoughtworks.com/radar/techniques/domain-storytelling)

### 4.2 Domain Knowledge Requirements

Domain Storytelling requires LESS specialized knowledge from participants:
- No need to learn sticky note color coding
- No need to understand "domain event" as a concept
- Stories use natural language
- Domain experts tell stories they already know

Event Storming requires participants to:
- Understand what a "domain event" is (orange sticky)
- Distinguish commands (blue) from events (orange) from policies (purple)
- Think in terms of past-tense state changes
- Participate in chaotic brainstorming phases

**Source:** [kalele.io](https://kalele.io/why-eventstorming-practitioners-should-try-domain-storytelling/), [axxes.com](https://www.axxes.com/en/insights/event-storming-domain-storytelling)

### 4.3 Output Quality for DDD

**Domain Storytelling strengths:**
- Directly produces actors, work objects, activities (natural mapping to entities, aggregates, commands)
- Builds ubiquitous language organically
- Each story seeds an acceptance test
- Better at capturing structural aspects (who works with what)

**Event Storming strengths:**
- Better at capturing temporal causality (what triggers what)
- Naturally produces domain events (not just implied by story sequence)
- Better at discovering policies and automated reactions
- Better at identifying hotspots and unknown unknowns (Big Picture level)

**Event Storming weakness:** "One notable limitation of Event Storming is its inability to fully capture the structural aspects of domain models, particularly Aggregates (Entities and Value Objects), which often necessitates supplementing Event Storming outputs with UML diagrams or other structural modeling techniques."

**Source:** [medium.com/@lambrych — EventStorming Strengths and Limitations](https://medium.com/@lambrych/eventstorming-for-domain-driven-design-strengths-and-limitations-3f0b49009c38), [kalele.io](https://kalele.io/why-eventstorming-practitioners-should-try-domain-storytelling/)

### 4.4 Combining Both Methods

The community consensus is clear: **Domain Storytelling first, then Event Storming for complex areas.**

**Recommended sequence:**
1. **Domain Storytelling** (coarse-grained) — establish the narrative, actors, work objects, vocabulary
2. **Event Storming** (Big Picture) — map the entire event flow, discover hotspots
3. **Domain Storytelling** (fine-grained) — detail specific bounded contexts
4. **Event Storming** (Design Level) — design aggregates, commands, events for implementation

"Starting with Domain Storytelling first allows teams to establish foundational knowledge before progressing to EventStorming's more technical event-focused analysis."

Domain Storytelling is "an excellent precursor to EventStorming or Event Modeling, stabilizing 'what really happens' before jumping into events, commands, projections, or architecture."

**Source:** [kalele.io](https://kalele.io/why-eventstorming-practitioners-should-try-domain-storytelling/), [axxes.com](https://www.axxes.com/en/insights/event-storming-domain-storytelling), [emmanuelvalverderamos.substack.com](https://emmanuelvalverderamos.substack.com/p/introduction-to-domain-storytelling)

### 4.5 Which is Better for AI-Guided Discovery?

Domain Storytelling is significantly better suited for AI-guided discovery for these reasons:

1. **Sentence-based format is text-native.** Domain stories are naturally expressible as text: "Actor does Activity using Work Object." No visual wall or sticky notes required. AI can generate, validate, and iterate on text-based stories.

2. **Moderator role maps to AI agent.** The Domain Storytelling moderator asks structured questions ("What happens next?", "Who does that?", "What do they use?") and records the story. An AI agent can play this role in a CLI or chat interface.

3. **One story at a time prevents overwhelm.** The scenario-based approach fits the constraint of a CLI interaction where the user sees one thing at a time.

4. **No visual notation dependency.** Event Storming relies heavily on spatial arrangement of sticky notes on a wall — difficult to replicate in text. Domain Storytelling's narrative structure works without visuals (though visuals enhance it).

5. **Lower cognitive load on users.** Users tell stories they already know. They do not need to learn DDD terminology, sticky note colors, or event-first thinking.

6. **Direct artifact extraction.** An AI agent can parse "Actor does Activity using Work Object" sentences and extract DDD artifacts programmatically.

---

## 5. Stories to Bounded Contexts

### 5.1 Boundary Detection Heuristics

Hofer and Schwentner identified three primary indicators of bounded context boundaries within stories (presented at DDD Europe 2018):

| Signal | Description | Example |
|--------|-------------|---------|
| **One-way information flow** | Information flows between parts of a story in only one direction | Orders flow from Sales to Fulfillment, never back |
| **Language differences** | Same concept has different names or meanings across sections | "Account" in Billing vs "Account" in Identity |
| **Different triggers** | Work patterns vary by time, event type, or frequency | Daily batch processing vs on-demand real-time |

**The rule of three:** Finding three indicators suggests a valid boundary, though it requires verification through additional stories.

**Additional boundary signals (from broader research):**
- **Organizational boundary** — different departments own different story sections
- **Handoff points** — where work objects are passed between actor groups
- **Different granularity needs** — parts of the story that need much more detail than others
- **Different lifecycle stages** — what happens before vs after a pivotal event

**Source:** [InfoQ — Finding Bounded Contexts Using Domain Storytelling (2018)](https://www.infoq.com/news/2018/02/storytelling-domain-contexts/), [domainstorytelling.org/domain-driven-design](https://domainstorytelling.org/domain-driven-design)

### 5.2 The "Same Work Object, Different Context" Signal

When the same work object appears in multiple stories but is used differently, described differently, or has different relevant properties — this is a strong signal for a bounded context boundary.

Example: A "Customer" in a Sales context has name, email, and preferences. The same "Customer" in a Billing context has account number, payment method, and credit limit. Different stories reveal that "Customer" means different things in different contexts.

This directly maps to the DDD concept of the same real-world entity having different representations in different bounded contexts.

**Source:** [InfoQ — Finding Bounded Contexts Using Domain Storytelling](https://www.infoq.com/news/2018/02/storytelling-domain-contexts/), [gist.ly/youtube-summarizer/unveiling-domain-storytelling](https://gist.ly/youtube-summarizer/unveiling-domain-storytelling-a-journey-into-modeling-context-boundaries)

### 5.3 Progression from Stories to Bounded Contexts

The methodology follows this sequence:

1. **Tell coarse-grained stories** — capture the big picture with high-level actors and work objects
2. **Identify activity clusters** — group activities that relate closely from an actor's perspective
3. **Look for boundary signals** — one-way flow, language differences, different triggers
4. **Draw tentative boundaries** — outline potential bounded contexts around clusters
5. **Verify with more stories** — tell additional stories that cross suspected boundaries to validate or refute them
6. **Refine** — adjust boundaries based on accumulated evidence
7. **Name the contexts** — using the ubiquitous language that emerged from the stories

**Key insight from Hofer:** "The goal is not to build walls, but rather to build models that separate contexts while allowing for people to work together."

**Source:** [InfoQ — Finding Bounded Contexts Using Domain Storytelling](https://www.infoq.com/news/2018/02/storytelling-domain-contexts/), [gist.ly/youtube-summarizer/unveiling-domain-storytelling](https://gist.ly/youtube-summarizer/unveiling-domain-storytelling-a-journey-into-modeling-context-boundaries)

### 5.4 From Boundaries to Subdomain Classification

Once bounded contexts are identified from stories, subdomain classification can be derived by examining:

| Signal in Stories | Likely Classification |
|-------------------|-----------------------|
| Complex stories with many actors, rich work objects, business-critical outcomes | **Core** subdomain |
| Supporting stories that enable core workflows but are not themselves differentiating | **Supporting** subdomain |
| Standard processes (payment, auth, email) that could be bought/outsourced | **Generic** subdomain |

The complexity budget (Evans/Vernon/Khononov) maps naturally: core subdomains get fine-grained stories and full DDD treatment; supporting get moderate detail; generic get coarse-grained stories and library recommendations.

**Source:** alto's own [docs/research/20260222_subdomain_classification_complexity_budget.md](/home/kusanagi/Alty/alty-cli/docs/research/20260222_subdomain_classification_complexity_budget.md), [domainstorytelling.org/domain-driven-design](https://domainstorytelling.org/domain-driven-design)

### 5.5 Domain Message Flow Modeling

Once bounded contexts are identified, Domain Storytelling notation can be adapted for inter-context communication:

- Bounded contexts themselves become the actors
- Commands, events, and queries become work objects
- Sequence numbers show message flow order
- Activities are replaced by the messages themselves

This bridges from Domain Storytelling to Context Mapping and ultimately to architecture design.

**Source:** [domainstorytelling.org/articles/domain-message-flow-modeling/](https://domainstorytelling.org/articles/domain-message-flow-modeling/)

---

## 6. Minimum Viable Story Set

### 6.1 What the Authors Say

Stefan Hofer states directly:
- **"Maybe two, three examples are enough to really understand a business process"**
- **"Usually a handful of examples is enough to understand even complex business processes"**
- Focus on: the primary happy path (80% of cases) and one to two important error scenarios

**Source:** [Tech Lead Journal #75](https://techleadjournal.dev/episodes/75/), [InfoQ Podcast](https://www.infoq.com/podcasts/domain-storytelling/)

### 6.2 Minimum for Bootstrapping a Domain Model

For alto's use case (turning a 4-5 sentence idea into DDD artifacts), the minimum viable story set is:

| Story | Type | Purpose | DDD Artifacts Produced |
|-------|------|---------|----------------------|
| **Story 1** | Primary happy path (coarse-grained) | Establish actors, work objects, main flow | Actor list, work object inventory, primary bounded contexts |
| **Story 2** | Primary failure case (coarse-grained) | Surface invariants and error handling | Invariant candidates, error stories, business rules |
| **Story 3** | Secondary workflow (coarse-grained) | Discover additional bounded contexts and context relationships | Context map relationships, additional contexts |

**Optional (for core subdomains):**
| Story | Type | Purpose |
|-------|------|---------|
| **Story 4** | Primary happy path (fine-grained, core context only) | Design aggregates, commands, value objects within core bounded context |
| **Story 5** | Primary failure case (fine-grained, core context only) | Refine invariants and aggregate boundaries |

### 6.3 What Makes a "Good Enough" First Story

A good first story should:
1. **Cover the primary use case** — the most common scenario (80% case)
2. **Include at least 2-3 actors** — reveals who participates
3. **Include at least 3-5 work objects** — reveals what the domain is about
4. **Have 5-10 numbered activities** — enough to see the flow without being overwhelming
5. **Be concrete** — "Customer orders a pizza and pays with credit card" not "User interacts with system"
6. **Use domain language** — terms the domain expert would use, not technical jargon

**Source:** [domainstorytelling.org/quick-start-guide](https://domainstorytelling.org/quick-start-guide), [InfoQ Podcast](https://www.infoq.com/podcasts/domain-storytelling/)

---

## 7. Suitability for AI-Guided Discovery

### 7.1 Why Domain Storytelling Fits alto

| alto Requirement | Domain Storytelling Fit |
|-----------------|------------------------|
| Conversational DDD discovery (PRD line 129) | Moderator-driven Q&A maps directly to CLI conversation |
| "Actor does Activity using Work Object" format (PRD line 129) | This IS the Domain Storytelling sentence structure |
| One story at a time to prevent overwhelm (PRD line 129) | Scenario-based: one story per pass by design |
| Maps directly to DDD concepts (PRD line 129) | Documented mapping from stories to aggregates, commands, entities, bounded contexts |
| Lighter than Event Storming (PRD line 129) | Lower cognitive load, no special notation training needed |
| Works for non-coders (PRD line 86) | Narrative format accessible to anyone |
| AI acts as moderator (PRD line 129) | Moderator role is well-defined with specific question patterns |

### 7.2 AI Moderator Capabilities

The Domain Storytelling moderator role can be decomposed into discrete, AI-executable actions:

1. **Propose an initial story** based on README input (Phase 1: Seed)
2. **Ask "what happens next?"** to extend the story
3. **Ask "who does that?"** to identify actors
4. **Ask "what do they use?"** to identify work objects
5. **Record each sentence** in structured format
6. **Number the activities** automatically
7. **Validate understanding** by replaying the story back
8. **Detect vocabulary** and build glossary entries
9. **Identify boundary signals** (language shifts, handoffs, different triggers)
10. **Propose bounded contexts** based on accumulated heuristics

All of these are text-based operations that do not require visual diagramming. An AI agent can perform all of them in a CLI conversation.

### 7.3 Text Representation of Domain Stories

For a CLI tool, domain stories can be represented in structured text:

```
Domain Story: "Customer Orders a Pizza"
Actors: Customer, Cashier, Kitchen
Work Objects: Menu, Order, Pizza, Receipt

1. Customer browses Menu
2. Customer places Order with Cashier
3. Cashier confirms Order
4. Cashier sends Order to Kitchen
5. Kitchen prepares Pizza using Order
6. Kitchen notifies Cashier that Pizza is ready
7. Cashier hands Pizza to Customer
8. Customer pays using Receipt
```

This format is:
- Human-readable (domain experts can validate)
- Machine-parseable (AI can extract DDD artifacts)
- Storable (can be serialized to TOML/JSON)
- Composable (multiple stories build a domain model)

### 7.4 AI-Assisted Story Proposal

For users who lack domain knowledge (alto PRD Scenario 5), the AI can:
1. Research the domain via web search
2. Propose initial stories based on industry patterns ("In most e-commerce systems, the primary flow is: Customer browses Catalog, adds Items to Cart, proceeds to Checkout...")
3. Ask the user to confirm, correct, or extend
4. Iterate until the story matches the user's reality

This fits the PRD's Knowledge Trust Hierarchy: USER_STATED > USER_CONFIRMED > AI_RESEARCHED > AI_INFERRED (PRD line 145).

---

## 8. Implications for alto

### 8.1 Recommended Discovery Flow for alto

```
Phase 1: Seed (automated)
  - AI reads README (4-5 sentences)
  - AI proposes initial actors, work objects
  - User confirms or corrects

Phase 2: Primary Story (Domain Storytelling — coarse-grained)
  - AI proposes initial story: "Based on your description, here's how I think the primary flow works..."
  - User refines sentence by sentence
  - AI records, numbers, builds glossary
  - Produces: Actor list, work object inventory, primary story

Phase 3: Failure Story (Domain Storytelling — coarse-grained)
  - AI asks: "What's the most important thing that can go wrong?"
  - User tells the failure story
  - Produces: Invariant candidates, error stories

Phase 4: Boundary Discovery
  - AI analyzes accumulated stories for boundary signals
  - AI proposes bounded context boundaries
  - User validates or adjusts
  - Produces: Bounded context map, context relationships

Phase 5: Subdomain Classification
  - AI proposes Core/Supporting/Generic classification based on story complexity
  - User confirms
  - Produces: Complexity budget

Phase 6 (optional, Core only): Fine-Grained Stories
  - For each Core bounded context, AI proposes detailed stories
  - User refines
  - Produces: Aggregate candidates, commands, value objects

Phase 7: Artifact Generation
  - AI generates PRD, DDD.md, ARCHITECTURE.md from accumulated stories
  - User reviews and approves
```

### 8.2 Alignment with Existing alto Research

This research aligns with and extends:
- **20260222_ddd_question_framework.md** — The 10-question framework already uses Domain Storytelling questions; this research provides the theoretical foundation and boundary detection heuristics
- **20260222_guided_ddd_question_framework_consolidated.md** — The dual-register approach works perfectly with Domain Storytelling (non-technical register uses pure business language)
- **20260305_ai_assisted_ddd_session_design.md** — The three-round iterative protocol can use Domain Storytelling as the primary technique in Round 1 (Express Discovery)
- **20260305_ddd_collaborative_modeling_2026.md** — Domain Storytelling is confirmed as "rising fast in adoption" and used before Event Storming

### 8.3 Tooling Note: Egon.io

Egon.io is the open-source tool for visualizing domain stories:
- **License:** GPLv3.0 (NOT permissive — cannot be embedded in alto)
- **Format:** .egn (JSON-based, formerly .dst)
- **Stars:** 825 on GitHub
- **Latest release:** v3.1.0 (March 2026)
- **Runs in browser** — no server needed

alto should NOT embed Egon.io (GPL license). However, alto could:
- Export domain stories in .egn format for visualization in Egon.io
- Generate domain stories in a text format that users can optionally paste into Egon.io
- Define its own text-based domain story format (as shown in section 7.3)

**Source:** [github.com/WPS/egon.io](https://github.com/WPS/egon.io)

---

## 9. Sources

### Primary (Authors)

- [domainstorytelling.org](https://domainstorytelling.org/) — Official site by Hofer & Schwentner
- [domainstorytelling.org/quick-start-guide](https://domainstorytelling.org/quick-start-guide) — Notation, scope dimensions, examples
- [domainstorytelling.org/domain-driven-design](https://domainstorytelling.org/domain-driven-design) — DDD mapping
- [domainstorytelling.org/articles/domain-message-flow-modeling/](https://domainstorytelling.org/articles/domain-message-flow-modeling/) — Inter-context communication
- Hofer, Stefan & Schwentner, Henning. "Domain Storytelling: A Collaborative, Visual, and Agile Way to Build Domain-Driven Software." Addison-Wesley, 2022. ISBN 9780137458912.

### Interviews and Talks

- [InfoQ Podcast — Domain Storytelling with Stefan Hofer and Henning Schwentner](https://www.infoq.com/podcasts/domain-storytelling/)
- [Tech Lead Journal #75 — Domain Storytelling with Stefan Hofer](https://techleadjournal.dev/episodes/75/)
- [InfoQ News — Finding Bounded Contexts Using Domain Storytelling (DDD Europe 2018)](https://www.infoq.com/news/2018/02/storytelling-domain-contexts/)

### Comparisons and Analysis

- [Kalele — Why EventStorming Practitioners Should Try Domain Storytelling](https://kalele.io/why-eventstorming-practitioners-should-try-domain-storytelling/)
- [Axxes — Event Storming & Domain Storytelling](https://www.axxes.com/en/insights/event-storming-domain-storytelling)
- [Thoughtworks Technology Radar — Domain Storytelling (Trial)](https://www.thoughtworks.com/radar/techniques/domain-storytelling)
- [Lambrych — EventStorming for DDD: Strengths and Limitations](https://medium.com/@lambrych/eventstorming-for-domain-driven-design-strengths-and-limitations-3f0b49009c38)

### Community and Practice

- [Open Practice Library — Domain Storytelling](https://openpracticelibrary.com/practice/domain-storytelling/)
- [emmanuelvalverderamos.substack.com — Introduction to Domain Storytelling](https://emmanuelvalverderamos.substack.com/p/introduction-to-domain-storytelling)
- [richard-seidl.com — Domain Storytelling](https://www.richard-seidl.com/en/blog/domain-storytelling)
- [DDD Academy — Domain Storytelling](https://ddd.academy/domain-storytelling/)

### Tooling

- [Egon.io — The Domain Story Modeler (GPLv3.0)](https://egon.io/)
- [github.com/WPS/egon.io](https://github.com/WPS/egon.io) — Source code
- [github.com/WPS/egon.io-examples](https://github.com/WPS/egon.io-examples) — Example domain stories

### Related alto Research

- [20260222_ddd_question_framework.md](/home/kusanagi/Alty/alty-cli/docs/research/20260222_ddd_question_framework.md)
- [20260222_guided_ddd_question_framework_consolidated.md](/home/kusanagi/Alty/alty-cli/docs/research/20260222_guided_ddd_question_framework_consolidated.md)
- [20260305_ai_assisted_ddd_session_design.md](/home/kusanagi/Alty/alty-cli/docs/research/20260305_ai_assisted_ddd_session_design.md)
- [20260305_ddd_collaborative_modeling_2026.md](/home/kusanagi/Alty/alty-cli/docs/research/20260305_ddd_collaborative_modeling_2026.md)
