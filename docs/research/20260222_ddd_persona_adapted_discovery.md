---
last_reviewed: 2026-02-22
type: spike
status: complete
---

# DDD Discovery Questions: Persona-Adapted Language

## Research Question

How should alty adapt DDD discovery questions for different user personas
(technical vs non-technical), and what plain language equivalents exist for core
DDD terminology?

## Context

alty's PRD defines five personas (PRD Section 3):

| Persona | Technical? | DDD Literacy |
|---------|-----------|--------------|
| Solo Developer | Yes | Moderate-High |
| Team Lead | Yes | Moderate-High |
| AI Tool Switcher | Yes | Low-Moderate |
| Product Owner | No | None |
| Domain Expert (HR/Sales/Ops) | No | None |

The guided question framework (PRD Section 9, open spike) must work for ALL
five. This research establishes the plain language mapping and question
rewriting patterns needed to implement two-register question delivery.

---

## Part A: Plain Language Equivalents for DDD Concepts

### Terminology Mapping Table

Sources: Philippe Bourgau's EventStorming-to-DDD bridge
([source](https://philippe.bourgau.net/how-to-use-event-storming-to-introduce-domain-driven-design/)),
DDD-Crew EventStorming Glossary
([source](https://github.com/ddd-crew/eventstorming-glossary-cheat-sheet)),
DDD Starter Modelling Process
([source](https://ddd-crew.github.io/ddd-starter-modelling-process/)).

| DDD Term | Plain Language (recommended) | Alternative phrasings | Why this works | Source |
|----------|-----------------------------|-----------------------|----------------|--------|
| **Bounded Context** | "Area of your business" | "Functional area", "business area", "department" | Bourgau: "nobody understands Bounded Context from the start, everyone gets 'Functional Area'" | [Bourgau](https://philippe.bourgau.net/how-to-use-event-storming-to-introduce-domain-driven-design/) |
| **Aggregate** | "Key business object" | "The main thing you manage", "central record" | EventStorming officially deprecated "aggregate" for business audiences; replaced with "Constraint" (business rules that protect consistency) | [DDD-Crew](https://github.com/ddd-crew/eventstorming-glossary-cheat-sheet) |
| **Aggregate Root** | "The main thing you interact with" | "Entry point", "the thing you manage as one unit" | Car analogy: "you drive the car, not each wheel individually" | [Baeldung](https://www.baeldung.com/cs/aggregate-root-ddd) |
| **Entity** | "Something you track by its identity" | "A record with an ID", "something that changes over time but stays the same thing" | Business people think in "records" (customer record, order record) | [Martin Fowler](https://martinfowler.com/bliki/DDD_Aggregate.html) |
| **Value Object** | "A description or measurement" | "A label", "details that describe something", "data without an ID" | e.g., "an address is a description of a location, not a tracked thing" | Evans, DDD (2003) |
| **Domain Event** | "Something important that happened" | "A business event", "a trigger" | EventStorming's entire premise: events are universally understood | [DDD-Crew](https://github.com/ddd-crew/eventstorming-glossary-cheat-sheet) |
| **Command** | "An action someone takes" | "A decision", "a request to do something" | EventStorming uses "Action" for business audiences (vs "Command" for devs) | [DDD-Crew](https://github.com/ddd-crew/eventstorming-glossary-cheat-sheet) |
| **Policy** | "A business rule that triggers automatically" | "Whenever X happens, we do Y" | This phrasing maps directly to EventStorming lilac stickies | [DDD-Crew](https://github.com/ddd-crew/eventstorming-glossary-cheat-sheet) |
| **Ubiquitous Language** | "The words your team actually uses" | "Shared vocabulary", "common language" | Bourgau: "'Ubiquitous Language' is a great name... once you get it!" -- use "shared vocabulary" until then | [Bourgau](https://philippe.bourgau.net/how-to-use-event-storming-to-introduce-domain-driven-design/) |
| **Subdomain** | "Part of your business" | "Business capability", "area of responsibility" | DDD Starter process step 3 (Decompose) uses "loosely-coupled parts of the domain" | [DDD-Crew Starter](https://ddd-crew.github.io/ddd-starter-modelling-process/) |
| **Core Domain** | "What makes your business different" | "Your competitive advantage", "what you do that nobody else does" | DDD Starter process step 4 (Strategize) frames this as differentiation | [DDD-Crew Starter](https://ddd-crew.github.io/ddd-starter-modelling-process/) |
| **Supporting Subdomain** | "Stuff you need but it's not your edge" | "Necessary but not unique" | Complexity budget: less engineering investment here | Evans, DDD (2003) |
| **Generic Subdomain** | "Off-the-shelf stuff" | "Standard business functions", "buy don't build" | e.g., authentication, email sending, payment processing | Evans, DDD (2003) |
| **Domain Service** | "An operation that doesn't belong to one thing" | "A business action that involves multiple things" | e.g., "transferring money involves two accounts" | Evans, DDD (2003) |
| **Repository** | "Where you store and look up things" | "Your filing system", "the database" | Business people already think of "where we keep records" | Evans, DDD (2003) |
| **Invariant** | "A rule that must always be true" | "A constraint", "something that can never be violated" | e.g., "an order can never have a negative total" | Evans, DDD (2003) |
| **Context Map** | "How your business areas talk to each other" | "Integration diagram", "how departments connect" | Visual + relational framing works for non-technical users | [DDD-Crew Starter](https://ddd-crew.github.io/ddd-starter-modelling-process/) |
| **Upstream/Downstream** | "Who has the upper hand" | "Who provides vs who consumes", "who depends on whom" | Bourgau: "Who has the upper hand?" replaces upstream/downstream | [Bourgau](https://philippe.bourgau.net/how-to-use-event-storming-to-introduce-domain-driven-design/) |

### Key Principle

From Bourgau: **"Don't be a Smug DDD Weenie."** Start with business language,
let technical concepts emerge from the conversation. Never lead with jargon.

---

## Part B: Question Rewriting Patterns (Technical vs Non-Technical)

### Pattern 1: Bounded Context Discovery

**Technical register:**
> "What are the bounded contexts in your system? Where do the same terms mean
> different things?"

**Non-technical register:**
> "Walk me through the different areas of your business. Are there places where
> the same word means something different? For example, does 'account' mean
> something different to Sales vs Finance?"

---

### Pattern 2: Aggregate / Key Business Object Discovery

**Technical register:**
> "What are the aggregate roots? What entities cluster together under a single
> consistency boundary?"

**Non-technical register:**
> "What are the main things your business keeps track of? For each one, what
> related details always change together? For example, when you update an order,
> do the line items always change with it?"

---

### Pattern 3: Domain Event Discovery

**Technical register:**
> "What domain events does this context publish? What state transitions matter?"

**Non-technical register:**
> "What important things happen in this process? When something happens, who
> needs to know about it? For example, when an order is shipped, does the
> warehouse need to know? Does the customer get a notification?"

---

### Pattern 4: Invariant / Business Rule Discovery

**Technical register:**
> "What invariants must the aggregate enforce? What are the consistency rules?"

**Non-technical register:**
> "What rules can never be broken? What would cause a real problem if it
> happened? For example, can you sell more items than you have in stock? Can a
> customer place an order without a delivery address?"

---

### Pattern 5: Entity vs Value Object

**Technical register:**
> "Which concepts need identity tracking (entities) vs which are just
> descriptions (value objects)?"

**Non-technical register:**
> "Which of these things do you track individually over time, and which are just
> descriptions? For example, do you care about *this specific* address, or just
> what the address says? Does a customer have an ID you look them up by?"

---

### Pattern 6: Ubiquitous Language Discovery

**Technical register:**
> "What is the ubiquitous language in this context? Are there term
> inconsistencies between code and domain experts?"

**Non-technical register:**
> "What words does your team actually use when talking about this? If I
> overheard your team in a meeting, what terms would I hear? Are there any words
> that different people use differently?"

---

### Pattern 7: Subdomain Classification (Complexity Budget)

**Technical register:**
> "Classify each subdomain as Core, Supporting, or Generic. Which get full DDD
> treatment?"

**Non-technical register:**
> "Which parts of your business are unique to you -- what you do that
> competitors don't? Which parts are necessary but standard? Which parts could
> you buy off the shelf? For example, is your special sauce in how you price
> things, or how you ship things?"

---

### Pattern 8: Context Map / Integration Discovery

**Technical register:**
> "What are the context relationships? Which are upstream/downstream? Any
> shared kernels or anti-corruption layers?"

**Non-technical register:**
> "How do these different areas of your business talk to each other? When Sales
> closes a deal, what does Fulfillment need to know? Who depends on whom? If
> one area changes how they work, which other areas are affected?"

---

### Pattern 9: Actor/Role Discovery

**Technical register:**
> "What actors interact with this bounded context? What are their
> responsibilities?"

**Non-technical register:**
> "Who is involved in this process? What role does each person play? Can someone
> be involved in more than one step?"

---

### Pattern 10: Process / Workflow Discovery

**Technical register:**
> "Describe the command flow and event chain for this use case."

**Non-technical register:**
> "Walk me through what happens step by step. When a customer places an order,
> what happens first? Then what? Who does what at each step? What decisions need
> to be made along the way?"

---

## Part C: Existing Research on Non-Technical DDD Discovery

### 1. Domain Storytelling (Hofer & Schwentner)

**Source:** [domainstorytelling.org](https://domainstorytelling.org/),
[Open Practice Library](https://openpracticelibrary.com/practice/domain-storytelling/),
[Tech Lead Journal #75](https://techleadjournal.dev/episodes/75/)

**How it works:**
- Domain experts narrate real business scenarios step by step
- A moderator draws the story using a pictographic language (actors, work
  objects, activities)
- The core facilitation question: **"Who does what, with what, why?"**
- Each story covers ONE concrete example, keeping discussions manageable

**Key technique for alty:** The "who-does-what-with-what-why" framework
maps directly to a conversational question sequence:
1. "Who starts this process?" (Actor)
2. "What do they do?" (Activity)
3. "What thing do they work with?" (Work Object -- maps to Entity/Aggregate)
4. "Why do they do it?" (Business motivation)
5. "What happens next?" (Flow / Event chain)

**Adaptation for text-based CLI:** Since alty is text-only (no visual
sticky notes), the question flow replaces the pictographic language. The answers
build a textual domain story that can later be visualized in docs.

### 2. EventStorming (Brandolini)

**Source:** [DDD-Crew Glossary](https://github.com/ddd-crew/eventstorming-glossary-cheat-sheet),
[Lucidchart EventStorming Guide](https://www.lucidchart.com/blog/ddd-event-storming)

**How it works:**
- Colored sticky notes replace jargon: orange = event, blue = action,
  lilac = policy, yellow = actor, pink = hotspot
- Events are written in past tense ("Order Placed", "Payment Received")
- No technical vocabulary required -- "things that happened" is universal

**Text-based equivalent for alty:**
- Instead of colored stickies, use labeled categories in the conversation:
  - "What happened?" -> Domain Event
  - "Who did it?" -> Actor
  - "What action triggered it?" -> Command
  - "Is there a rule that says 'whenever X happens, do Y'?" -> Policy
  - "What information was needed to make that decision?" -> Read Model
  - "What could go wrong here?" -> Hotspot

**Critical vocabulary note from EventStorming community:**
- "Action" is preferred over "Command" for business audiences
- "Information" is preferred over "Query Model"
- "Aggregate" is officially legacy terminology in EventStorming -- replaced
  by "Constraint" (business rules protecting consistency)

### 3. Example Mapping (BDD)

**Source:** [Wikipedia BDD](https://en.wikipedia.org/wiki/Behavior-driven_development),
Three Amigos / Specification Workshop pattern

**How it works:**
- Uses concrete examples to discover rules: "Give me an example of when X
  happens"
- Three Amigos (business, dev, QA) discuss one story at a time
- Discovers edge cases through "What about when...?" questions

**Text-based equivalent for alty:**
- After identifying a key business object, ask: "Give me an example of a
  typical [order/request/application]"
- Then probe: "What about when something goes wrong? What if the payment
  fails? What if the item is out of stock?"
- Each example surfaces invariants and business rules naturally

### 4. DDD Starter Modelling Process (DDD-Crew)

**Source:** [DDD-Crew Starter Process](https://ddd-crew.github.io/ddd-starter-modelling-process/)

Eight-step iterative process:
1. **Understand** -- Business model, user needs, goals
2. **Discover** -- Collaborative exploration (EventStorming)
3. **Decompose** -- Break into subdomains
4. **Strategize** -- Core vs Supporting vs Generic
5. **Connect** -- How subdomains interact
6. **Organise** -- Team alignment
7. **Define** -- Bounded Context Canvas
8. **Code** -- Implementation

**Relevance to alty:** Steps 1-5 map to the guided question flow. Steps
6-8 map to artifact generation and ticket creation. The process is explicitly
"not linear" -- alty should allow jumping between phases.

### 5. SingleStone's Domain-Driven Discovery

**Source:** [SingleStone](https://www.singlestoneconsulting.com/domain-driven-discovery)

Four-phase workshop methodology:
1. **Frame the Problem** -- "What problem are we solving? Who is affected?
   What does success look like? What constraints exist?"
2. **Analyze Current State** -- EventStorming + C4 diagrams
3. **Explore Future State** -- Bounded contexts + Core Domain Charts
4. **Create Roadmap** -- Now/Next/Later prioritization

**Relevance to alty:** Phase 1 questions ("What problem? Who? What
success? What constraints?") are ideal opening questions for ALL personas.

### 6. Qlerify (Commercial SaaS)

**Source:** [Qlerify](https://www.qlerify.com/domain-driven-design-tool)

- AI-powered: generates domain events, bounded contexts, and API code from
  text prompts
- Guides teams through refining AI-generated models with discovery questions
- Asks: "Does this terminology reflect the language used in your business?"
- Cloud-only, no local/CLI equivalent

**Gap Qlerify fills vs alty:** Qlerify uses visual EventStorming boards;
alty must achieve the same discovery through text-only conversation.

---

## Part D: Recommended Question Flow for Non-Technical Users

### Design Principles

1. **Start from business problems, not domain models** -- SingleStone's
   "Frame the Problem" phase first
2. **Use concrete examples, not abstract categories** -- "Walk me through
   what happens when..." not "What are your entities?"
3. **Build vocabulary from the user's words** -- capture their terms, then
   map to DDD concepts internally
4. **Progressive disclosure** -- start broad, narrow down; never front-load
   complexity
5. **One story at a time** -- Domain Storytelling principle: one concrete
   scenario per pass

### Recommended Question Flow (Non-Technical Personas)

#### Phase 1: Frame the Problem (2-3 questions)

```
Q1: "In a few sentences, what problem are you trying to solve?"
    -> Captures: Problem domain, initial language

Q2: "Who has this problem? Who else is involved?"
    -> Captures: Actors, roles, stakeholders

Q3: "What does success look like? How will you know this is working?"
    -> Captures: Success criteria, implicit business rules
```

**Maps to DDD:** Problem space framing, initial actor identification

#### Phase 2: Walk Through the Happy Path (3-5 questions)

```
Q4: "Walk me through the most common scenario, step by step.
     Start with: 'A [person] wants to [do something]...'"
    -> Captures: Primary use case, main flow, key domain events

Q5: "At each step, what thing are they working with?
     For example, an order, an application, a request..."
    -> Captures: Key business objects (future aggregates/entities)

Q6: "Who needs to know when each step is done?
     Does anyone else get notified or need to take action?"
    -> Captures: Events, downstream processes, integration points

Q7: "What information does each person need to do their part?"
    -> Captures: Read models, data requirements

Q8: "Are there any decisions or approvals needed along the way?"
    -> Captures: Commands, policies, authorization rules
```

**Maps to DDD:** Domain events, aggregates, commands, policies, read models

#### Phase 3: Discover Rules and Edge Cases (3-4 questions)

```
Q9:  "What could go wrong in this process? What are the common problems?"
     -> Captures: Hotspots, error scenarios, compensating actions

Q10: "What rules can never be broken? What would cause a real problem?"
     -> Captures: Invariants, business constraints

Q11: "Give me an example of an unusual case. What happens when
      [the standard assumption] is NOT true?"
     -> Captures: Edge cases, alternative flows

Q12: "Are there time-sensitive steps? Deadlines? Things that expire?"
     -> Captures: Temporal constraints, saga patterns
```

**Maps to DDD:** Invariants, domain events (failure), policies, temporal rules

#### Phase 4: Identify Business Areas (2-3 questions)

```
Q13: "Are there different groups or departments involved in this?
      Could different teams own different parts?"
     -> Captures: Potential bounded contexts

Q14: "Are there places where the same word means different things
      to different groups? For example, does 'customer' mean the same
      thing to Sales as it does to Support?"
     -> Captures: Context boundaries, ubiquitous language conflicts

Q15: "Which part of this is unique to your business vs something
      anyone could buy off the shelf?"
     -> Captures: Core vs Supporting vs Generic subdomains
```

**Maps to DDD:** Bounded contexts, context map, complexity budget

#### Phase 5: Confirm and Name (1-2 questions)

```
Q16: "Let me play back what I heard. [Summary]. Did I get that right?
      What did I miss?"
     -> Captures: Validation, corrections, missed concepts

Q17: "You used the word '[term]' -- is that the word everyone on your
      team uses, or do some people call it something else?"
     -> Captures: Ubiquitous language refinement
```

**Maps to DDD:** Ubiquitous language glossary, model validation

### Recommended Question Flow (Technical Personas)

Technical users get the same logical phases but with DDD terminology and
fewer explanatory prompts:

#### Phase 1: Scope (1 question)

```
Q1: "Describe your project and the problem domain in a few sentences."
```

#### Phase 2: Domain Model (3-4 questions)

```
Q2: "What are the key domain events? List the important things that happen."

Q3: "What are the main aggregates? What entities and value objects cluster
     together under consistency boundaries?"

Q4: "What invariants must each aggregate enforce?"

Q5: "What commands trigger these events? Who (which actor) issues them?"
```

#### Phase 3: Strategic Design (2-3 questions)

```
Q6: "What bounded contexts do you see? Where does terminology diverge?"

Q7: "Classify each subdomain: Core (build + invest), Supporting (build
     simple), Generic (buy/use library)."

Q8: "How do contexts communicate? What data flows between them?"
```

#### Phase 4: Validate (1 question)

```
Q9: "Here is the domain model I extracted: [summary]. Corrections?"
```

---

## Implementation Recommendations for alty

### 1. Persona Detection Strategy

Detect persona early in the flow (during `alty init` or `alty guide`):

```
"Before we start, which best describes you?
 1. Developer or technical lead (I know what bounded contexts are)
 2. Product owner or business person (I know the business, not the code)
 3. Domain expert (I know the specific problem area deeply)
 4. Not sure / mixed"
```

Options 1 -> technical register. Options 2-4 -> non-technical register.

### 2. Dual-Register Question Storage

Store questions in a structured format with both registers:

```yaml
- id: aggregate_discovery
  phase: domain_model
  technical:
    question: "What are the aggregate roots in this context?"
    follow_up: "What consistency boundaries do they enforce?"
  non_technical:
    question: "What are the main things your business keeps track of?"
    follow_up: "When you update one of these, what related details always change with it?"
  maps_to:
    - aggregate
    - entity
    - value_object
```

### 3. Internal Terminology Mapping

Regardless of which register the questions use, alty should internally
map answers to DDD concepts. The generated artifacts (DDD.md, architecture
tests, tickets) always use proper DDD terminology, but the glossary includes
the user's original words.

### 4. Progressive Depth

For non-technical users, start shallow and go deeper only where the user
shows engagement. Not every project needs all 17 questions. Minimum viable
discovery:

- Q1 (problem), Q2 (who), Q4 (happy path), Q10 (rules), Q13 (areas)
- Five questions can produce a basic bounded context map

### 5. Playback Pattern

After every 3-4 questions, summarize what was captured using the user's own
words. This is the Domain Storytelling "validation loop" adapted for text.

---

## Sources

1. Bourgau, Philippe. "How to use Event Storming to introduce Domain Driven Design." [https://philippe.bourgau.net/how-to-use-event-storming-to-introduce-domain-driven-design/](https://philippe.bourgau.net/how-to-use-event-storming-to-introduce-domain-driven-design/)
2. DDD-Crew. "EventStorming Glossary & Cheat Sheet." [https://github.com/ddd-crew/eventstorming-glossary-cheat-sheet](https://github.com/ddd-crew/eventstorming-glossary-cheat-sheet)
3. DDD-Crew. "DDD Starter Modelling Process." [https://ddd-crew.github.io/ddd-starter-modelling-process/](https://ddd-crew.github.io/ddd-starter-modelling-process/)
4. Hofer & Schwentner. "Domain Storytelling." [https://domainstorytelling.org/](https://domainstorytelling.org/)
5. Open Practice Library. "Domain Storytelling." [https://openpracticelibrary.com/practice/domain-storytelling/](https://openpracticelibrary.com/practice/domain-storytelling/)
6. SingleStone. "Domain-Driven Discovery." [https://www.singlestoneconsulting.com/domain-driven-discovery](https://www.singlestoneconsulting.com/domain-driven-discovery)
7. Qlerify. "AI-Powered DDD Modeling Tool." [https://www.qlerify.com/domain-driven-design-tool](https://www.qlerify.com/domain-driven-design-tool)
8. Evans, Eric. "Domain-Driven Design: Tackling Complexity in the Heart of Software." Addison-Wesley, 2003.
9. Brandolini, Alberto. "Introducing EventStorming." Leanpub, 2021.
10. Lucidchart. "Event Storming 101." [https://www.lucidchart.com/blog/ddd-event-storming](https://www.lucidchart.com/blog/ddd-event-storming)
11. Baeldung. "What Is an Aggregate Root?" [https://www.baeldung.com/cs/aggregate-root-ddd](https://www.baeldung.com/cs/aggregate-root-ddd)
12. Martin Fowler. "Bounded Context." [https://martinfowler.com/bliki/BoundedContext.html](https://martinfowler.com/bliki/BoundedContext.html)
