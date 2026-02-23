---
last_reviewed: 2026-02-22
owner: researcher
status: complete
spike_ticket: k7m.2
---

# DDD Question Framework: Domain Storytelling + Event Storming for AI-Guided Discovery

## Research Questions

1. What questions does Domain Storytelling use to elicit domain knowledge? What is the typical flow/order?
2. How do domain stories map to bounded contexts, ubiquitous language, aggregates, and domain events?
3. How can we adapt Domain Storytelling for a solo developer working with an AI assistant?
4. What is the minimum effective question set that still produces useful domain stories?
5. What questions drive Event Storming sessions, and how does it identify DDD artifacts?
6. Can Event Storming be done asynchronously with text prompts?
7. What existing tools adapt these workshop techniques for solo/AI use?

---

## 1. Domain Storytelling Methodology (Hofer & Schwentner)

### 1.1 Core Concept

Domain Storytelling is a collaborative modeling method where domain experts tell concrete stories about how they work, and a moderator records them using a pictographic language. The output is a visual narrative showing "who (actor) does what (activity) with what (work objects) with whom (other actors)."

**Source:** [domainstorytelling.org](https://domainstorytelling.org/), Stefan Hofer & Henning Schwentner, Addison-Wesley 2022

### 1.2 Pictographic Language Elements

| Element | Representation | Example |
|---------|---------------|---------|
| **Actors** | Stick figures / system icons | Cashier, Customer, Billing System |
| **Activities** | Labeled arrows (domain-language verbs) | "suggests", "searches", "marks" |
| **Work Objects** | Document/item icons | Reservation, Seating Plan, Ticket |
| **Sequence Numbers** | Numbers at arrow origins | 1, 2, 3... (chronological order) |
| **Annotations** | Text callouts | Variations, assumptions, domain terms |
| **Groups** | Outlined clusters | Repeated actions, locations, org boundaries |

**Sentence structure:** Subject (Actor) -- Predicate (Activity) --> Object (Work Object)

Example: "Cashier [1] searches seating plan for available seats"

**Source:** [domainstorytelling.org/quick-start-guide](https://domainstorytelling.org/quick-start-guide)

### 1.3 Workshop Questions (Facilitator Flow)

The Domain Storytelling moderator uses these core questions to drive story elicitation:

**Opening questions:**
1. "Can you walk me through a typical [process/task] from start to finish?"
2. "Who starts this process? What triggers it?"

**Progression questions (asked repeatedly):**
3. "What happens next?"
4. "Who does that?"
5. "What do they use to do it?" / "Where do they get this information?"
6. "What do they produce / hand off?"
7. "How do they determine what to do next?"

**Deepening questions:**
8. "What vocabulary do you use for this?" (captures ubiquitous language)
9. "What surprised you about this?" (surfaces assumptions)
10. "What is ambiguous here?" (reveals boundary conflicts)

**Scope control:**
- Start with the "80% case" -- the most common, happy-path scenario
- "First establish a sound understanding of typical cases -- tell stories. Only then discuss what else could happen -- collect rules."
- Model significant variations as separate domain stories; annotate minor variations

**Source:** [domainstorytelling.org/quick-start-guide](https://domainstorytelling.org/quick-start-guide), [Tech Lead Journal #75 - Stefan Hofer interview](https://techleadjournal.dev/episodes/75/)

### 1.4 Scope Dimensions

Domain stories vary across three dimensions:

| Dimension | Options | When to Use |
|-----------|---------|-------------|
| **Granularity** | Coarse-grained (overview) to fine-grained (detailed) | Start coarse, drill into core subdomains |
| **Timing** | As-Is (current process) vs To-Be (future process) | As-Is for discovery, To-Be for design |
| **Domain Purity** | Pure (no software) vs Digitalized (includes systems) | Pure first, then digitalized for implementation |

**Source:** [domainstorytelling.org/quick-start-guide](https://domainstorytelling.org/quick-start-guide)

### 1.5 Mapping Domain Stories to DDD Artifacts

Stefan Hofer describes the translation pathway explicitly:

| Domain Story Element | DDD Artifact | How to Identify |
|---------------------|-------------|-----------------|
| **Actors** | User roles, external systems | Who appears in stories |
| **Work Objects** | Entities, Aggregates | Objects that are created, modified, or exchanged |
| **Activities on Work Objects** | Commands / Methods | "they can make a reservation, they can cancel it" |
| **Work Object properties** | Value Objects | "movie or time could become value objects" |
| **Sequence of activities** | Domain Events | State transitions between steps |
| **Groups / spatial clusters** | Bounded Contexts | "group parts of the story together that belong in the same subdomain" |
| **Handoffs between groups** | Context Map relationships | Where actors pass work objects across boundaries |
| **Vocabulary differences** | Context boundaries | Same term used differently = different bounded context |

**Key heuristic:** "If two groups cannot agree on the meaning of some events or sequence of steps, it's a good indicator they're talking about different contexts."

**Source:** [Tech Lead Journal #75](https://techleadjournal.dev/episodes/75/), [domainstorytelling.org/domain-driven-design](https://domainstorytelling.org/domain-driven-design)

### 1.6 Bounded Context Identification from Stories

Heuristics from Domain Storytelling that signal context boundaries:

1. **Vocabulary shift** -- The same word means different things to different actors
2. **Organizational boundary** -- Different departments or teams own different parts of the story
3. **Handoff points** -- Where work objects are passed between groups
4. **Different granularity needs** -- Parts of the story that need much more detail than others
5. **Different life-cycle stages** -- What happens before a pivotal event vs after

**Source:** [domainstorytelling.org/domain-driven-design](https://domainstorytelling.org/domain-driven-design), [LinkedIn DDD discussion](https://www.linkedin.com/advice/0/how-do-you-identify-bounded-contexts-legacy)

---

## 2. Event Storming Methodology (Brandolini)

### 2.1 Core Concept

Event Storming is a workshop format where participants collaboratively map domain events on a timeline using colored sticky notes. It operates at three levels: Big Picture, Process Modeling, and Software Design.

**Source:** [eventstorming.com](https://www.eventstorming.com/), Alberto Brandolini, Leanpub 2013+

### 2.2 Three Levels and Their Questions

#### Level 1: Big Picture Event Storming

**Purpose:** Explore the entire domain, discover unknown unknowns

**Driving questions:**

| Phase | Questions Asked |
|-------|----------------|
| Chaotic Exploration | "What events happen in your business?" / "Do you receive payments?" / "Are contracts signed?" |
| Timeline Enforcement | "Which event comes first?" / "Do these follow in sequence?" |
| People & Systems | "Who are the key roles?" / "What external systems are involved?" |
| Explicit Walkthrough | "Is this unclear?" / "What's missing?" / "Why does this follow that?" |
| Reverse Narrative | "For X to happen, what must happen first?" |
| Problems & Opportunities | "What 2 problems matter most right now?" |

**Source:** [Qlerify Event Storming Guide](https://www.qlerify.com/post/event-storming-the-complete-guide), [IBM Event Storming Reference](https://ibm-cloud-architecture.github.io/refarch-eda/methodology/event-storming/)

#### Level 2: Process Modeling

**Purpose:** Detail a specific process end-to-end with commands, policies, and read models

**Driving questions:**

| Phase | Questions Asked |
|-------|----------------|
| Frame the Process | "What is the scope? What outcome are we looking for?" |
| Happy Path | "What is the most common successful path?" |
| Command Identification | "Why did this event occur?" / "What triggered it?" |
| Policy Discovery | "Whenever [event], then what happens automatically?" |
| Read Model | "What information does the person need to make this decision?" |
| Alternative Paths | "What else could happen here?" / "What goes wrong?" |

**Grammar:** Event --> Policy --> Command --> System --> Event

When human intervention is needed: Policy --> Human --> Read Model --> Command

**Source:** [Qlerify Guide](https://www.qlerify.com/post/event-storming-the-complete-guide), [IBM Reference](https://ibm-cloud-architecture.github.io/refarch-eda/methodology/event-storming/)

#### Level 3: Software Design

**Purpose:** Design aggregates and bounded contexts for implementation

**Driving questions:**

| Phase | Questions Asked |
|-------|----------------|
| Bounded Context | "Where do different teams use different language for the same concept?" |
| Aggregate Discovery | "What data must be consistent together in one transaction?" |
| Aggregate API | "What commands does this aggregate accept? What events does it emit?" |
| Completeness | "Is the lifecycle complete? Are we missing create/update/delete steps?" |

**Source:** [Qlerify Guide](https://www.qlerify.com/post/event-storming-the-complete-guide), [Leonardo Max Almeida - Brandolini book notes](https://leomax.fyi/blog/book-notes-and-main-takeaways-event-storming-by-alberto-brandolini/)

### 2.3 Sticky Note Color Grammar

| Color | Represents | Phrasing |
|-------|-----------|----------|
| Orange | Domain Events | Past tense: "Order Placed", "Payment Received" |
| Blue | Commands | Imperative: "Place Order", "Submit Payment" |
| Yellow (small) | Actors / People | Role name: "Customer", "Admin" |
| Yellow (large) | Aggregates | Noun: "Order", "Account" |
| Purple/Lilac | Policies | "Whenever [event], then [command]" |
| Green | Read Models | "Customer sees: price, reviews, stock" |
| Pink (large) | External Systems | "Payment Gateway", "Email Service" |
| Red/Magenta | Hot Spots | Problems, conflicts, unknowns |
| Light Green | Opportunities | Improvement ideas |

**Source:** [Qlerify Guide](https://www.qlerify.com/post/event-storming-the-complete-guide)

### 2.4 Mapping Event Storming to DDD Artifacts

| Event Storming Element | DDD Artifact | How to Identify |
|-----------------------|-------------|-----------------|
| Domain Events (orange) | Domain Events | Past-tense business occurrences |
| Commands (blue) | Command objects / methods | Imperative actions that cause events |
| Actors (yellow) | User roles | Who initiates commands |
| Aggregates (yellow large) | Aggregate roots | Groups of events + commands with shared consistency |
| Policies (purple) | Domain Services / Event Handlers | Automated reactions to events |
| Read Models (green) | Query models / Projections | Data views for decision-making |
| Pivotal Events | Bounded Context boundaries | Major phase transitions |
| Swimlanes | Bounded Context boundaries | Parallel processes owned by different roles |
| Hot Spots | Open questions, invariants | Unresolved conflicts needing design decisions |

**Bounded context heuristics from Event Storming:**

1. **Pivotal events** -- "Different phases usually mean different problems, which usually leads to different models" (Brandolini)
2. **Language disagreements** -- "Different wording for the same events is a clear indication of different bounded contexts"
3. **Swimlane boundaries** -- Visual separation of parallel processes
4. **Actor clusters** -- Different user types suggest different models
5. **Verb patterns** -- "Looking at verbs provides much more consistency around one specific purpose"

**Source:** [Leonardo Max Almeida - Brandolini notes](https://leomax.fyi/blog/book-notes-and-main-takeaways-event-storming-by-alberto-brandolini/), [IBM Reference](https://ibm-cloud-architecture.github.io/refarch-eda/methodology/event-storming/)

### 2.5 IBM's 8-Step Event Storming Process

IBM documented a practical 8-step process:

1. **Domain Events Discovery** -- Write events on orange stickies using past-tense verbs
2. **Tell the Story** -- Replay from persona viewpoints, ask "what happens next?"
3. **Find the Boundaries** -- Identify pivotal events, detect subject/swimlane boundaries
4. **Locate the Commands** -- Ask "Why did this event occur?" for each event
5. **Describe the Data** -- Define data needed for each command-event pair
6. **Identify the Aggregates** -- Group related events, commands, data by lifecycle
7. **Define Bounded Context** -- Establish where terms have specific meanings
8. **Insight Storming** -- "What if we could predict this event in advance?"

**Source:** [IBM Cloud Architecture - Event Storming](https://ibm-cloud-architecture.github.io/refarch-eda/methodology/event-storming/)

---

## 3. DDD-Crew Starter Modelling Process

The DDD-Crew (Nick Tune et al.) defined an 8-step discovery-to-code process that integrates both methodologies:

| Step | Name | Purpose | Tools Used |
|------|------|---------|------------|
| 1 | **Understand** | Align with business model and goals | Business Model Canvas, Impact Mapping |
| 2 | **Discover** | Visually discover the domain | EventStorming, Domain Storytelling |
| 3 | **Decompose** | Break domain into loosely-coupled parts | Context Maps, Independent Service Heuristics |
| 4 | **Strategize** | Identify core domains (competitive advantage) | Core Domain Charts, Purpose Alignment Model |
| 5 | **Connect** | Design interactions between subdomains | Domain Message Flow Modelling |
| 6 | **Organise** | Form teams aligned with context boundaries | Team Topologies |
| 7 | **Define** | Define bounded context roles and responsibilities | Bounded Context Canvas |
| 8 | **Code** | Implement with alignment to domain concepts | Aggregate Design Canvas, Hexagonal Architecture |

**Key insight:** "This process is for beginners. It is not a linear sequence of steps that you should standardise as a best practice. Domain-Driven Design is an evolutionary design process which necessitates continuous iteration."

**Source:** [DDD-Crew Starter Modelling Process](https://ddd-crew.github.io/ddd-starter-modelling-process/)

---

## 4. Solo Developer + AI Adaptation

### 4.1 Problems with Direct Translation

Both Domain Storytelling and Event Storming were designed for multi-person workshops:

| Workshop Assumption | Solo Developer Reality |
|--------------------|----------------------|
| Multiple stakeholders with different perspectives | One person, one perspective |
| Facilitator controls pacing | AI must pace the conversation |
| Whiteboard with sticky notes | Text-based conversation |
| Real-time visual feedback | Sequential Q&A |
| 2-8 hour dedicated workshop | 15-30 minute guided session |
| Domain expert present | Developer may also be domain expert |
| Body language signals confusion | No non-verbal cues |

### 4.2 What Carries Over (Text-Compatible)

These elements translate well to text-based AI interaction:

1. **The sentence structure**: Actor -- Activity --> Work Object (becomes fill-in-the-blank)
2. **The question flow**: "What happens next?" / "Who does that?" / "What do they use?"
3. **The story focus**: Concrete scenarios, not abstract requirements
4. **The vocabulary capture**: Recording exact terms the user provides
5. **The 80% rule**: Start with the happy path, variations later
6. **The scope dimensions**: Coarse-to-fine, as-is before to-be
7. **The boundary heuristics**: Vocabulary shifts, handoff points, pivotal events

### 4.3 What Must Change

1. **No visual board** -- Replace with structured text output (YAML/Markdown domain story)
2. **No group dynamics** -- AI plays devil's advocate, asks "What about...?" prompts
3. **No facilitator** -- AI structures the conversation with a fixed question flow
4. **No domain expert in the room** -- User IS the domain expert; AI asks clarifying questions
5. **No chaotic exploration** -- Replace with guided sequential elicitation
6. **No sticky-note sorting** -- AI organizes artifacts automatically from answers

### 4.4 Recommended Hybrid Approach for alty

Combine Domain Storytelling's concrete-scenario focus with Event Storming's artifact taxonomy:

**Phase 1: Seed (from README, 1 minute)**
- AI reads the 4-5 sentence idea
- AI identifies candidate actors, work objects, and events from the text
- AI presents its understanding and asks "Did I get this right?"

**Phase 2: Story Elicitation (Domain Storytelling style, 5-10 minutes)**
- AI asks the user to walk through the primary workflow as a story
- Uses the sentence pattern: "[Who] [does what] [with what] [to produce what]"
- Captures 2-3 stories (happy path + 1-2 key variations)
- Records exact vocabulary

**Phase 3: Event Discovery (Event Storming style, 5-10 minutes)**
- AI extracts events from the stories and asks "What else happens?"
- Asks about triggers: "What causes [event]? A person, a system, or a rule?"
- Asks about policies: "Whenever [event] happens, what must happen automatically?"
- Asks about read models: "What information does [actor] need to decide?"

**Phase 4: Boundary Detection (5 minutes)**
- AI identifies vocabulary conflicts and handoff points from stories
- Asks about organizational boundaries: "Are these handled by different teams or roles?"
- Proposes bounded context boundaries and asks for confirmation
- Classifies subdomains as Core/Supporting/Generic

**Phase 5: Artifact Synthesis (automated, 1-2 minutes)**
- AI generates: Ubiquitous Language glossary, Bounded Context map, Aggregate candidates, Domain Events list, Subdomain classification
- User reviews and corrects

### 4.5 Persona-Adapted Question Paths

The same questions, worded for different personas:

| Question Purpose | Technical (Developer) | Non-Technical (PO / Domain Expert) |
|-----------------|----------------------|-----------------------------------|
| Identify actors | "Who are the actors in this domain?" | "Who are the people involved in this process?" |
| Identify work objects | "What are the key domain entities?" | "What things do people work with? (documents, orders, reports)" |
| Identify activities | "What commands/operations exist?" | "What do people do with those things?" |
| Identify events | "What domain events occur?" | "What important things happen during this process?" |
| Identify boundaries | "Where might bounded contexts separate?" | "Are there parts of this process owned by different teams or departments?" |
| Identify invariants | "What business rules must always be true?" | "What rules can never be broken? What would be a mistake?" |
| Classify subdomains | "Is this a core, supporting, or generic subdomain?" | "Is this the thing that makes your business special, or is it something every business does?" |

**Source (for adaptation approach):** Derived from PRD.md persona requirements and Domain Storytelling's principle of "using participants' domain language, not technical jargon"

---

## 5. Minimum Effective Question Set

Based on the research, the minimum question set that produces useful DDD artifacts is **10 questions** organized in 5 phases:

### Phase 1: The Idea (seed from README)
No questions needed -- AI parses the initial description.

### Phase 2: Actors and Work Objects (2 questions)
**Q1:** "Who are the people or systems involved in [domain]? List everyone who does something or receives something."
- **Produces:** Actor list -> User roles, external systems

**Q2:** "What things do these people work with? (documents, orders, items, records, messages)"
- **Produces:** Work Object list -> Entity/Aggregate candidates

### Phase 3: The Primary Story (3 questions)
**Q3:** "Walk me through the most common workflow from start to finish. Use this pattern: [Who] [does what] [with what]."
- **Produces:** Primary domain story, activity sequence, command candidates

**Q4:** "What is the most important thing that can go wrong in this workflow?"
- **Produces:** Error/alternative story, invariant candidates

**Q5:** "Are there any other important workflows? (just name them, we can detail later)"
- **Produces:** Story inventory, scope assessment

### Phase 4: Events and Rules (3 questions)
**Q6:** "Looking at these workflows, what are the key moments where the state of something changes? (e.g., 'order was placed', 'payment was received')"
- **Produces:** Domain Events list

**Q7:** "Are there any automatic rules? Whenever [X happens], then [Y must happen]?"
- **Produces:** Policies, event handlers, domain service candidates

**Q8:** "What information does [actor] need to see before making a decision at step [N]?"
- **Produces:** Read Models, query requirements

### Phase 5: Boundaries and Classification (2 questions)
**Q9:** "Are there parts of this system that could work completely independently? Parts where different people are responsible, or where the same word means something different?"
- **Produces:** Bounded Context candidates

**Q10:** "Which part of this is the core of your idea -- the thing that makes it unique? Which parts are standard stuff every system needs (login, notifications, etc.)?"
- **Produces:** Subdomain classification (Core/Supporting/Generic)

### Artifact Output from 10 Questions

| Artifact | Produced By Questions |
|----------|----------------------|
| Ubiquitous Language Glossary | Q1, Q2, Q3 (exact terms captured) |
| Domain Stories (2-3) | Q3, Q4 |
| Actor List | Q1 |
| Work Object / Entity List | Q2 |
| Command List | Q3 (verbs from stories) |
| Domain Events | Q6 |
| Policies | Q7 |
| Read Models | Q8 |
| Bounded Context Map | Q9 |
| Subdomain Classification | Q10 |
| Aggregate Candidates | Q2 + Q6 (work objects grouped by events) |
| Invariant Candidates | Q4, Q7 |

---

## 6. Existing Tools That Do Something Similar

### 6.1 Tools with DDD-Specific AI Assistance

| Tool | What It Does | DDD Output | Limitations |
|------|-------------|-----------|-------------|
| **Qlerify** | AI-powered process modeling with Event Storming support | Bounded contexts, aggregates, domain events, API boilerplate, unit tests | Cloud-only SaaS, commercial, no local-first, no project scaffold |
| **ChatDiagram Event Storming Maker** | Text-to-event-storming-diagram via AI | Visual event storming diagrams | Diagram generation only, no DDD artifact extraction |
| **ContextMapper** v6.12.0 | DDD DSL for context mapping and service decomposition | Context maps, bounded contexts, UML diagrams, service contracts | Java/Eclipse-based, no conversational interface, no AI, Apache 2.0 |
| **Egon.io** | Browser-based domain story visualization | Domain story diagrams with replay | Visualization only, no artifact extraction, no AI, GPLv3 |

**Source:** [Qlerify](https://www.qlerify.com/), [ChatDiagram](https://www.chatdiagram.com/tool/event-storming-diagram-maker), [ContextMapper](https://contextmapper.org/), [Egon.io](https://egon.io/)

### 6.2 Tools in the Vibe Coding Space

No vibe coding tool (Lovable, Bolt, v0, GPT-Engineer, Kiro, Spec-Kit, BMAD) incorporates DDD discovery questions. See `docs/research/20260222_vibe_coding_landscape.md` for full analysis.

The closest is **Amazon Kiro** which generates Requirements.md / Design.md / Tasks.md, but it does not ask DDD questions, identify bounded contexts, or produce ubiquitous language glossaries.

### 6.3 Academic Work (2025)

A 2025 paper in IJCSE ("Designing Scalable Multi-Agent AI Systems: Leveraging Domain-Driven Design and Event Storming") demonstrates using Event Storming + DDD to design multi-agent AI systems, but does not provide a text-based question flow for solo developers.

**Source:** [IJCSE Paper](https://www.internationaljournalssrg.org/IJCSE/paper-details?Id=588)

### 6.4 Gap Analysis

**No existing tool combines all of:**
1. Text-based conversational DDD question flow
2. Artifact generation (ubiquitous language, bounded contexts, aggregates, events)
3. Subdomain classification (Core/Supporting/Generic)
4. Project scaffold integration
5. Local-first operation

This is the white space alty fills. The question framework designed in Section 5 would be novel.

---

## 7. Event Storming: Text-Based Feasibility

### Can Event Storming be done asynchronously with text prompts?

**Yes, with adaptations.** Leonardo Max Almeida (who facilitated remote Event Storming sessions) notes: "in online Event Storming sessions, just adding events and enforcing a timeline already generates significant value."

**What works in text:**
- Event brainstorming (list events in past tense)
- Timeline ordering (numbered list)
- Command identification ("Why did this event occur?")
- Policy discovery ("Whenever X, then Y")
- Aggregate grouping (group events by consistency boundary)

**What is lost:**
- Chaotic simultaneous brainstorming (replaced by sequential prompting)
- Spatial layout revealing patterns (replaced by AI-generated groupings)
- Physical body language signals (replaced by explicit questions about confusion/disagreement)
- Group energy/momentum (replaced by structured question flow)

**Source:** [Leonardo Max Almeida - Remote Event Storming](https://leomax.fyi/blog/remote-big-picture-event-storming-template-instructions-and-learnings-from-facilitating-one/)

---

## 8. Recommendations for alty

### 8.1 Adopt the Hybrid Approach

Use Domain Storytelling as the primary discovery method (story-first, concrete scenarios) and Event Storming's artifact taxonomy (events, commands, policies, read models, aggregates) for structuring the output.

**Rationale:**
- Domain Storytelling's "walk me through a scenario" is more natural for solo text conversation than Event Storming's "brainstorm all events at once"
- Event Storming's color-coded grammar provides clear artifact categories
- The combination produces all DDD artifacts needed for architecture

### 8.2 Implement the 10-Question Framework

The minimum 10 questions (Section 5) produce all required DDD artifacts in approximately 15-20 minutes of conversation. This maps to the PRD's "< 30 minutes from README to first ticket" constraint.

### 8.3 Persona-Adapted Question Paths

Implement two question modes:
- **Technical mode:** Uses DDD terminology (for Solo Developer, Team Lead, AI Tool Switcher)
- **Business mode:** Uses plain business language (for Product Owner, Domain Expert)

Both modes produce identical artifacts. The difference is only in question wording.

### 8.4 Complexity Budget Integration

Question Q10 ("Which part is the core of your idea?") feeds directly into the complexity budget (P0 feature from k7m.9). Classification drives:
- Core: Full DDD tactical treatment, strict fitness functions, highest test coverage
- Supporting: Simple services, moderate testing
- Generic: CRUD / library recommendations, minimal testing

### 8.5 Output Format

Generate artifacts in the structure defined by `docs/templates/DDD_STORY_TEMPLATE.md`:
1. Domain Stories (Section 1)
2. Ubiquitous Language Glossary (Section 2)
3. Subdomain Classification (Section 3)
4. Bounded Contexts + Context Map (Section 4)
5. Aggregate Design (Section 5)
6. Event Storming Summary (Section 6)

### 8.6 Iterative Refinement

Following DDD-Crew's principle that "discovery is continuous," the question framework should support re-entry at any phase. After initial bootstrap, users can:
- Add new stories (re-enter Phase 2)
- Discover new events (re-enter Phase 3)
- Refine boundaries (re-enter Phase 4)

---

## 9. Follow-Up Tickets Needed

1. **Implementation: Question flow engine** -- Build the conversational flow that walks users through the 10 questions, captures answers, and generates DDD artifacts. Must support both technical and business language modes.

2. **Implementation: Artifact generator** -- Transform question answers into DDD_STORY_TEMPLATE.md format (ubiquitous language glossary, bounded context map, aggregate candidates, domain events, subdomain classification).

3. **Implementation: Complexity budget classifier** -- From Q10 answers, classify subdomains and apply appropriate treatment levels to generated tickets and fitness functions.

4. **Design: Question branching logic** -- Determine when follow-up questions are needed (e.g., if user lists > 5 actors, probe for bounded context boundaries earlier; if user cannot identify events, switch to story-first approach).

5. **Integration: Connect to ticket pipeline (k7m.11)** -- Use generated DDD artifacts as input to auto-generate dependency-ordered beads tickets.

6. **Integration: Connect to fitness function generation (k7m.10)** -- Use bounded context map to generate import-linter TOML / pytestarch test files.

---

## 10. Sources

### Primary Sources (Methodology Authors)
- Hofer, S. & Schwentner, H. (2022). *Domain Storytelling: A Collaborative, Visual, and Agile Way to Build Domain-Driven Software*. Addison-Wesley. [Book site](https://domainstorytelling.org/book)
- Brandolini, A. (2013+). *Introducing EventStorming*. Leanpub. [Book site](https://leanpub.com/introducing_eventstorming)
- [domainstorytelling.org](https://domainstorytelling.org/) -- Official Domain Storytelling site
- [eventstorming.com](https://www.eventstorming.com/) -- Official EventStorming site

### Community Resources
- [DDD-Crew Starter Modelling Process](https://ddd-crew.github.io/ddd-starter-modelling-process/) -- 8-step discovery-to-code process
- [DDD-Crew Bounded Context Canvas](https://github.com/ddd-crew/bounded-context-canvas) -- Structured BC documentation
- [DDD-Crew Aggregate Design Canvas](https://github.com/ddd-crew/aggregate-design-canvas) -- Structured aggregate documentation
- [Open Practice Library - Domain Storytelling](https://openpracticelibrary.com/practice/domain-storytelling/) -- Practice guide
- [Tech Lead Journal #75 - Stefan Hofer](https://techleadjournal.dev/episodes/75/) -- Interview with DS author

### Practitioner Guides
- [Qlerify Event Storming Complete Guide](https://www.qlerify.com/post/event-storming-the-complete-guide) -- Detailed ES facilitation
- [IBM Cloud Architecture - Event Storming](https://ibm-cloud-architecture.github.io/refarch-eda/methodology/event-storming/) -- IBM's 8-step process
- [Leonardo Max Almeida - Brandolini Book Notes](https://leomax.fyi/blog/book-notes-and-main-takeaways-event-storming-by-alberto-brandolini/) -- Key heuristics

### Tools
- [Egon.io](https://egon.io/) -- Open-source Domain Story Modeler (GPLv3)
- [ContextMapper](https://contextmapper.org/) -- DDD DSL for context mapping (Apache 2.0)
- [Qlerify](https://www.qlerify.com/) -- AI-powered DDD modeling (commercial SaaS)
- [ChatDiagram Event Storming Maker](https://www.chatdiagram.com/tool/event-storming-diagram-maker) -- AI diagram generator (free tier)

### Academic
- IJCSE (2025). "Designing Scalable Multi-Agent AI Systems: Leveraging Domain-Driven Design and Event Storming." [Paper](https://www.internationaljournalssrg.org/IJCSE/paper-details?Id=588)
