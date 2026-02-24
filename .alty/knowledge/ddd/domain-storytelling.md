# Domain Storytelling

Quick reference for AI agents facilitating domain discovery and identifying bounded contexts.

---

## What It Is

A pictographic modeling technique where domain experts tell stories about how they do their work. The stories are recorded as diagrams using a simple notation, revealing domain boundaries and workflows.

---

## Key Elements

| Element | Symbol | Description | Example |
|---------|--------|-------------|---------|
| **Actor** | Stick figure | A person or system that does something | Customer, Admin, Payment Gateway |
| **Work Object** | Labeled icon | A thing that is created, used, or exchanged | Order, Invoice, Report |
| **Activity** | Arrow + verb | What the actor does with the work object | "submits", "reviews", "generates" |
| **Sequence Number** | Number on arrow | Order of activities in the story | 1, 2, 3... |
| **Annotation** | Text note | Clarification, business rule, or constraint | "Only managers can approve > $10k" |

---

## Process

### Step 1: Record Stories

1. Ask the domain expert: "Walk me through how [workflow] works"
2. Draw actors, work objects, and activities as they narrate
3. Number each activity in sequence
4. Note annotations for rules, exceptions, and edge cases

### Step 2: Identify Scope

Each story has a scope:
- **AS-IS stories:** How things work today
- **TO-BE stories:** How things should work in the new system
- Start with AS-IS to understand the domain, then create TO-BE for design

### Step 3: Find Boundaries

Bounded contexts emerge from stories:
- Different actors using the same term differently = context boundary
- Activities that involve a handoff between groups = context boundary
- Work objects that change meaning mid-story = context boundary

### Step 4: Extract Domain Model

From each story:
- Actors become external systems or user roles
- Work objects become entities or value objects
- Activities become commands or domain service methods
- Sequence reveals process flow and dependencies

---

## Mapping Stories to Code

| Story Element | Code Artifact | Layer |
|---------------|---------------|-------|
| Actor (person) | User role / auth concern | Infrastructure |
| Actor (system) | External system adapter | `infrastructure/external/` |
| Work Object | Entity or Value Object | `domain/models/` |
| Activity | Command handler or domain method | `application/commands/` or entity method |
| Business rule (annotation) | Domain invariant or specification | `domain/models/` or `domain/services/` |
| Handoff between actors | Domain event or integration point | `domain/events/` |

---

## How alty Uses Domain Storytelling

1. **DDD Story Template:** `docs/templates/DDD_STORY_TEMPLATE.md` guides users through story recording
   - Each story captures one workflow end-to-end
   - Template prompts for actors, work objects, activities, and rules

2. **During `alty init`:** Simplified story collection
   - "Who are the main actors in your system?"
   - "What do they do? Walk me through a typical workflow."
   - "What are the key things (objects) they work with?"

3. **Context boundary detection:**
   - Compare stories from different parts of the system
   - Where the same work object appears with different attributes = boundary
   - Where actors hand off work = potential boundary

4. **Ticket generation:**
   - Each story becomes a feature epic
   - Each activity becomes one or more implementation tickets
   - Business rules from annotations become acceptance criteria

---

## Domain Storytelling vs Event Storming

| Aspect | Domain Storytelling | Event Storming |
|--------|-------------------|----------------|
| Focus | Actor workflows | Events and reactions |
| Starting point | "How do you do X?" | "What happens in the system?" |
| Best for | Understanding user journeys | Discovering event flows |
| Output | Workflow diagrams | Event timelines |
| Use together | Stories first (understand domain), then event storm (discover events) |

---

## Tips for AI Agents

- Always use the domain expert's exact words for activities and work objects
- One story = one workflow = one scope. Do not combine multiple workflows
- If a story gets complex (>10 activities), it likely spans bounded contexts -- split it
- Annotations about rules are gold -- they become invariants and acceptance criteria
- Sequence numbers reveal temporal coupling and potential saga/process manager needs
- Stories from different actors about the same process reveal different perspectives and context boundaries
