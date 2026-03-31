---
marp: true
theme: default
paginate: false
size: 16:9
style: |
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;600;800;900&family=JetBrains+Mono:wght@400;700&display=swap');

  :root {
    --teal: #2a9d8f;
    --deep: #1a2332;
    --mid: #243447;
    --bright: #5eead4;
    --muted: #94a3b8;
    --warn: #f59e0b;
    --bg: #f8f9fa;
    --card: #ffffff;
  }

  section {
    font-family: 'Inter', system-ui, sans-serif;
    background: var(--bg);
    color: #1a1a2e;
    padding: 56px 80px;
  }

  h1 {
    font-weight: 900;
    letter-spacing: -0.04em;
    line-height: 1.05;
    color: var(--deep);
  }

  h2 {
    color: var(--teal);
    font-weight: 600;
  }

  code {
    font-family: 'JetBrains Mono', monospace;
    background: #e9ecef;
    color: var(--deep);
    padding: 3px 10px;
    border-radius: 6px;
    font-size: 0.88em;
  }

  pre {
    background: var(--deep);
    border-radius: 14px;
    padding: 24px 28px;
    box-shadow: 0 8px 32px rgba(26,35,50,0.15);
  }

  pre code {
    background: none;
    padding: 0;
    font-size: 1.15em;
    color: #e2e8f0;
  }

  em { color: var(--teal); font-style: normal; font-weight: 700; }
  strong { color: var(--deep); font-weight: 700; }

  table {
    font-size: 0.88em;
    border-collapse: separate;
    border-spacing: 0;
    width: 100%;
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 2px 12px rgba(0,0,0,0.06);
  }

  th {
    background: var(--deep);
    color: white;
    padding: 14px 20px;
    text-align: left;
    font-weight: 600;
  }

  td {
    padding: 12px 20px;
    border-bottom: 1px solid #e9ecef;
    background: white;
  }

  blockquote {
    border: none;
    background: linear-gradient(135deg, #f0fdfa, #e9ecef);
    padding: 16px 28px;
    margin: 20px 0;
    border-radius: 12px;
    font-style: italic;
    color: #495057;
    box-shadow: 0 2px 8px rgba(0,0,0,0.04);
    border-left: 5px solid var(--teal);
  }

  ul { list-style: none; padding-left: 0; }

  ul li {
    padding: 10px 0 10px 40px;
    position: relative;
    font-size: 1.05em;
    line-height: 1.4;
  }

  ul li::before {
    content: '>';
    position: absolute;
    left: 8px;
    color: white;
    font-family: 'JetBrains Mono', monospace;
    font-weight: 700;
    font-size: 0.75em;
    background: var(--teal);
    width: 22px;
    height: 22px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    top: 12px;
  }

  /* ===== TITLE ===== */
  section.title {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    text-align: center;
    background: linear-gradient(135deg, #f0fdfa 0%, #ccfbf1 50%, #e9ecef 100%);
  }

  section.title h1 {
    font-size: 5.5em;
    color: var(--deep);
    margin-bottom: 0;
  }

  section.title p {
    color: var(--teal);
    font-size: 1.4em;
    font-family: 'JetBrains Mono', monospace;
    margin-top: 12px;
    font-weight: 400;
  }

  /* ===== PROBLEM ===== */
  section.problem {
    background: linear-gradient(160deg, #fef3c7 0%, var(--bg) 50%);
  }

  section.problem h1 {
    color: var(--warn);
    font-size: 2.8em;
  }

  section.problem p:last-of-type {
    color: #868e96;
    font-size: 1.1em;
    margin-top: 20px;
    padding-top: 16px;
    border-top: 2px dashed #dee2e6;
  }

  /* ===== STORY ===== */
  section.story {
    background: linear-gradient(160deg, #f0fdfa 0%, var(--bg) 50%);
  }

  section.story h1 {
    font-size: 2.4em;
    margin-bottom: 4px;
  }

  /* ===== SOLUTION ===== */
  section.solution {
    background: linear-gradient(160deg, #ccfbf1 0%, var(--bg) 50%);
  }

  section.solution h1 {
    font-size: 2.4em;
    margin-bottom: 4px;
  }

  section.solution em {
    font-size: 1.2em;
    display: block;
    margin-top: 20px;
    text-align: center;
  }

  /* ===== DEMO ===== */
  section.demo {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    text-align: center;
    background: var(--deep);
    color: #e2e8f0;
  }

  section.demo h1 {
    color: var(--bright);
    font-size: 2em;
    margin-bottom: 8px;
  }

  section.demo h2 {
    font-size: 1.2em;
    font-weight: 400;
    color: var(--muted);
    font-style: italic;
    margin-top: 0;
  }

  /* ===== CHEAT ===== */
  section.cheat {
    background: #ffffff;
    padding: 40px 72px;
  }

  section.cheat h1 {
    font-size: 1.8em;
    margin-bottom: 8px;
  }

  section.cheat pre {
    margin: 5px 0;
    padding: 10px 20px;
    background: var(--deep);
    border-radius: 10px;
  }

  section.cheat pre code {
    font-size: 0.85em;
  }

  section.cheat p:first-of-type {
    color: #868e96;
    margin-bottom: 8px;
    font-size: 0.95em;
  }

  /* ===== LANDSCAPE ===== */
  section.landscape {
    background: linear-gradient(160deg, #f0fdfa 0%, var(--bg) 50%);
    padding: 40px 72px;
  }

  section.landscape h1 {
    font-size: 2em;
    margin-bottom: 4px;
  }

  section.landscape td:first-child {
    font-weight: 600;
  }

  /* ===== CTA ===== */
  section.cta {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    text-align: center;
    background: linear-gradient(135deg, #ccfbf1 0%, #f0fdfa 50%, #ccfbf1 100%);
  }

  section.cta h1 {
    font-size: 3.2em;
    color: var(--deep);
    margin-bottom: 20px;
  }

  section.cta p {
    color: var(--teal);
    font-size: 1.2em;
  }

  section.cta pre {
    margin: 24px 0;
    text-align: left;
  }
---

<!-- _class: title -->

# alto

Your AI builds apps fast. alto makes sure they don't fall apart.

<!--
SPEAKER NOTES — SLIDE 1 (10 seconds)

Just the title. Let it breathe.

SAY:
"Hi, I'm Jaigouk. I've been building tools for AI-assisted development.
Today I want to share a problem I ran into — and how I solved it."
-->

---

<!-- _class: problem -->

# The Problem

You tell an AI tool "build me an invoice app." You get working code in minutes.

- *Change one thing, break three others* — no clear boundaries between features
- *Business rules are scattered* — nobody knows where "the logic" actually lives
- *Dev says "order", PM says "request"* — no shared vocabulary, constant miscommunication
- *Finish a task, others go stale* — work starts on **outdated assumptions**

Spec-driven tools (Kiro, Spec Kit) generate tasks — but skip *domain discovery*.

<!--
SPEAKER NOTES — SLIDE 2 (30 seconds)

SAY:
"Raise your hand if you've used Claude Code, Cursor, or Copilot
to build something. Yeah — they're amazing.

But here's what happens. You get a working prototype in minutes.
Then two weeks later, you change the pricing logic and the
notification system breaks. Nobody can find where the business
rules actually live. Your dev says 'order', your PM says 'request' —
they're talking about the same thing but nobody knows it.

The AI doesn't know your domain. And nobody told it."
-->

---

<!-- _class: story -->

# My Journey

The ideal flow: README → PRD → DDD → Architecture → Epics → Tickets

| What I had | What went wrong |
|-----------|----------------|
| README.md | Idea was clear enough |
| PRD.md | *AI-generated, shallow* — no real domain understanding |
| ~~DDD.md~~ | **Didn't exist.** This was the missing step. |
| ARCHITECTURE.md | Built on a weak PRD — wrong boundaries from the start |
| Epics & tickets | Looked structured, but the *foundation was wrong* |

> Without proper domain discovery, everything downstream inherits the mistakes.

<!--
SPEAKER NOTES — SLIDE 3 (60 seconds)

SAY:
"Here's what happened to me.

I knew the right flow — README, PRD, DDD, architecture, then tickets.
But my PRD was AI-generated. It looked professional but it was shallow.
It didn't capture what the business actually does.

And I had no DDD.md at all. No domain model, no shared vocabulary,
no bounded contexts. I jumped straight from a weak PRD to architecture.

So the architecture had the wrong boundaries. The tickets looked structured
but were built on a shaky foundation. The AI tools followed the tickets
perfectly — and built the wrong thing, perfectly.

The problem wasn't the tickets. It wasn't even the architecture.
It was that nobody did proper domain discovery.
Everything downstream inherits that mistake.

That's why I built alto — to make domain discovery
the mandatory first step, not an afterthought."
-->

---

<!-- _class: solution -->

# What alto Does

The planning step *before* coding starts.

| | What | Why it matters |
|-|------|---------------|
| 1 | *Domain Storytelling* — guided conversation discovers your domain | AI learns your business, not the other way around |
| 2 | *Bounded contexts + ubiquitous language* → DDD.md | Code organized by business domain, not technical layers |
| 3 | *Architecture fitness tests* generated from domain model | Boundaries enforce themselves — wrong imports fail CI |
| 4 | *Dependency-ordered tickets* with TDD phases | AI tools know what to build next, in the right order |

*20 minutes of storytelling saves 20 hours of rewrites.*

<!--
SPEAKER NOTES — SLIDE 4 (30 seconds)

SAY:
"alto is the architect that runs before the builders start.

You describe your idea in a few sentences. alto runs a Domain Storytelling
conversation — it proposes concrete stories about how your system works,
and you refine them.

From those stories, it generates four things that no other tool produces
together: a domain model with ubiquitous language, executable architecture
tests that enforce boundaries in CI, dependency-ordered tickets with TDD
phases, and tool-native configs for Claude Code, Cursor, or any AI tool.

The key insight: fitness functions mean your architecture enforces itself.
If an AI tool writes code that crosses a boundary, the test fails.
No code review needed — the guardrail is automated.

Let me show you."

TRANSITION: Next slide introduces the demo project.
-->

---

<!-- _class: solution -->

# Let's Try It

We have a README with 4 sentences:

```
A CLI tool that helps restaurant owners manage daily specials.
Owners enter dishes with prices and dietary tags.
The tool generates a formatted menu board for a shared display.
It tracks which specials sell out and suggests reorders.
```

We run `alto guide` in Claude Code. alto reads the README, asks questions, and proposes domain stories.

<!--
SPEAKER NOTES — SLIDE 5 (15 seconds)

SAY:
"Let me show you with a concrete example.

Here's a simple README — a restaurant daily specials manager.
We run 'alto guide' and alto reads this, then starts a
Domain Storytelling conversation to discover the domain."

TRANSITION: Next slide is the demo GIF.
-->

---

<!-- _class: demo -->

# alto in action

## Domain Storytelling — the ping-pong conversation

![w:1000](./alto-demo.gif)

<!--
SPEAKER NOTES — SLIDE 5 (2-3 minutes)

The GIF shows the REAL alto guide flow (prompts match actual CLI code):

1. Mode selection: Rapid vs Thorough — user picks Rapid
2. Persona: Domain Expert ("I know the business")
3. Opening: "What makes this process start?" → "Restaurant opens"
4. "Who starts this process?" → "Owner"
5. Narration loop: "What happens next?" → activity → "Who does this?" → "What thing is involved?" → confirm sentence
6. User EDITS sentence 2: "Dietary Info" → "Allergen Tag" (ubiquitous language)
7. Business rule annotation: "Must have allergen tag before publishing"
8. Full story replay with synthesis checkpoint
9. Boundary detection: MenuManagement + Inventory

SAY (narrate over the GIF):
"alto reads the README and starts the guide flow.
First — discovery mode and persona. The user picks
Domain Expert. This adjusts the question language.

Then the opening questions — 'What makes this process start?'
and 'Who starts it?' These come from alto's moderator
question bank, not hardcoded scripts.

Now the narration loop. For each step in the workflow,
alto asks three questions: what happens, who does it,
and what work object is involved. Then it shows the
structured sentence: Actor does Activity on Work Object.

Watch the edit — 'Dietary Info' becomes 'Allergen Tag'.
That's the domain expert's word for it. It goes straight
into the ubiquitous language — and that's what the code uses.

Business rules become annotations — 'must have allergen tag
before publishing' becomes a domain invariant.

Then boundary detection — two bounded contexts:
Menu Management and Inventory. From one conversation,
alto generates DDD.md, architecture, fitness tests, tickets."

IF GIF DOESN'T LOAD:
Run `bash docs/demo-guide-sim.sh` in terminal.
-->

---

<!-- _class: cheat -->

# Demo Cheat Sheet

Live fallback — run the simulated guide session:

```
bash docs/demo-guide-sim.sh
```

Or run the real thing:

```
mkdir demo && cd demo && git init
cat > README.md << 'EOF'
A CLI tool that helps restaurant owners manage daily specials.
Owners enter dishes with prices and dietary tags.
EOF
alto init -y && alto guide --no-tui
```

<!--
SPEAKER NOTES — CHEAT SHEET (presenter view only)

Keep this on your second monitor. The audience sees the demo slide.

The GIF shows the simulated guide session (30s, deterministic).
If it fails, run the sim script directly in terminal.

The real `alto guide --no-tui` works too but needs LLM connectivity
and takes longer. Only use as backup if you have time.

RECOVERY: If anything breaks, just move to the next slide.
-->

---

<!-- _class: landscape -->

# Spec-Driven Dev Works. DDD Makes It *Right*.

| Without DDD | With DDD |
|-------------|----------|
| AI generates epics from a vague spec | Epics map to *bounded contexts* — clear ownership |
| Tickets look structured but have wrong boundaries | Tickets follow *domain stories* — real business workflows |
| Architecture is guessed from requirements | Architecture is *derived from the domain model* |
| Fitness tests? What fitness tests? | Boundaries *enforce themselves* in CI |

> Spec-driven development automates the plan. DDD makes sure it's the *right* plan.

<!--
SPEAKER NOTES — SLIDE 7 (20 seconds)

SAY:
"Spec-driven development is real. It works. Tools like gstack and Kiro
prove that generating specs before code is the right idea.

But the spec is only as good as the domain understanding behind it.
Without DDD, you're automating the wrong plan — fast.

With domain discovery, your epics map to bounded contexts.
Your tickets follow real business workflows.
Your architecture is derived from the domain, not guessed.
And your boundaries enforce themselves with fitness tests.

alto adds the domain discovery step that makes spec-driven dev
actually produce the right structure."
-->

---

<!-- _class: cta -->

# Structure Before Speed

```
$ go install github.com/jaigouk/altoiddd/cmd/alto@latest
$ alto init
```

**Open Source. Go. Apache 2.0.**

github.com/jaigouk/altoiddd

<!--
SPEAKER NOTES — SLIDE 8 (15 seconds)

SAY:
"That's alto. Domain discovery before coding starts.

One Go install, one command to begin. Open source, Apache 2.0.

Link is on the screen. Questions?"
-->
