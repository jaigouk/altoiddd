---
name: rca
description: Drive a blameless root-cause analysis on a bug ticket — Five Whys, evidence, verification, action items — and write the artifact to docs/bugs/ or inline on the ticket per severity tier
kind: command
phase: implement
when_to_use: When investigating, fixing, or closing a bug ticket — runs the blameless Five Whys flow and produces the verification + action-items artifact
tools: Read, Write, Edit, Grep, Glob, Bash
bash_substitution_policy: quoted
license: Apache-2.0
---

# /rca

Drive a **blameless root-cause analysis** on a bug ticket. Walks the user
through reproduction, Five Whys, fix-verification, and action items, then
writes the artifact in the right place for the bug's severity tier.

This is the bug-fix counterpart to [`/groom`](./groom.md): groom prepares
implementation work; `/rca` prepares (and closes) bug work. Run it before
claiming a bug ticket AND again before closing it — the doc evolves from
"investigating" to "verified" as the fix lands.

## Why This Exists

A bug closed without recorded root cause and before/after verification has
nothing stopping the same regression from shipping again next quarter.
This command enforces three habits:

1. **Reproduce before fixing.** No fix lands without a documented RED
   regression test.
2. **Blameless Five Whys.** The first two whys describe the incident; the
   last three describe the *system* that allowed it. Action items come
   from whys 3–5, not 1.
3. **Verification with verbatim evidence.** Before-output and after-output
   pasted into the doc make the fix audit-able after the fact.

## Usage

```
/rca <bug-id>
/rca <bug-id> --phase=investigate
/rca <bug-id> --phase=verify
/rca <bug-id> --phase=close
```

- No `--phase`: auto-detects from ticket status + RCA doc existence.
- `investigate`: produces the initial doc with Five Whys skeleton + repro.
- `verify`: fills in the before/after verification block after the fix
  lands.
- `close`: final sign-off — checks every section, files action items as
  separate tickets, suggests `bd close <id>` with the right `--reason`.

## Process

### Phase 1 — Severity gate

Read the bug ticket and confirm the severity tier:

```bash
bd show "$BUG_ID" --json
```

Apply the routing table from
[`beads-bug-template.md`](../templates/beads-bug-template.md):

- **S0, S1**: produce a separate RCA doc at
  `docs/bugs/YYYYMMDD_<bug-id>_<slug>.md` using
  [`bug-rca-template.md`](../templates/bug-rca-template.md). Mandatory.
- **S2, S3**: do **NOT** create a separate doc by default. Fill in the
  inline `## Root Cause` and `## Verification` sections on the ticket
  itself. Offer the user the option to upgrade to a full doc if the bug
  turns out to be more systemic than the severity suggested.

If the ticket has no severity set, prompt for it before continuing.

### Phase 2 — Reproduce

Confirm the bug reproduces on the current `HEAD` before sinking
investigation time. Three outcomes:

1. **Reproduces:** proceed to Phase 3.
2. **Does NOT reproduce:** check whether an unrelated commit already
   fixed it. Run `git log --oneline -50` on the affected files. If a fix
   slipped in, present the candidate commit to the user and suggest
   `bd close <id> --reason "Resolved by <commit-hash>"` instead of
   continuing the RCA.
3. **Cannot reach a known state:** the Steps to Reproduce on the ticket
   are inadequate. Stop and ask the user for missing context — do NOT
   guess. Update the ticket's Steps to Reproduce before continuing.

### Phase 3 — Draft the artifact

#### For S0 / S1 (separate doc)

1. Generate the slug: lowercase-kebab of the bug ticket title (strip
   `fix(...):`, punctuation, articles).
2. Read today's date in `YYYYMMDD` format (match the convention used by
   `docs/research/` spike reports — same date shape, dash-free).
3. Compose the path: `docs/bugs/YYYYMMDD_<bug-id>_<slug>.md`.
4. Ensure the path is a **direct child** of `docs/bugs/` — no `..`
   segments, no slashes inside the slug. Reject otherwise.
5. Copy the body of `alto-scaffold/templates/bug-rca-template.md` into
   the new file. Replace placeholders with what is already known from
   the ticket (bug id, title, reporter, date, severity).
6. Set `Status: Investigating` in the frontmatter line of the doc.

#### For S2 / S3 (inline on the ticket)

1. Build the inline `## Root Cause` table (Five Whys, 5 rows) and the
   `## Verification` block from the bug template's matching sections.
2. Use `bd update <bug-id> --body-file -` to append the new sections
   to the ticket's description — do NOT replace the description; preserve
   existing content. If the ticket already has these sections, EDIT them
   in place, do not duplicate.

### Phase 4 — Walk the Five Whys

Interactively, one why at a time. For each:

1. State the question. Q1 quotes the **Actual Result** from the ticket
   verbatim, so the analysis stays grounded.
2. Ask the user for the answer. Reject answers that:
   - Name a person ("Bob forgot", "the previous dev didn't"). Push back:
     "Why was the system arranged so that was possible?"
   - Restate the symptom rather than explain it. Push back: "What made
     <symptom> happen at this level?"
3. Require evidence for the answer: a `file:line` citation, a commit
   hash, or a log excerpt. Refuse to advance without it.
4. After Q2, announce the transition to "system layer" so Q3–Q5 land on
   missing tests / missing types / missing rules / unclear ownership.
5. After Q5, summarise the root cause in one sentence.

Write each answer + evidence into the doc (or ticket) as you go — do not
batch.

### Phase 5 — TDD regression test

Confirm a RED regression test exists in the diff (or is about to be added)
that:

- Reproduces the **Actual Result** before the fix
- Lives next to the production code being fixed (same `_test.go` file
  family)
- Has a name matching `Test<Subject>_When<BugCondition>_Expect<CorrectBehaviour>`
- Will stay in the test suite after the fix (regression guard, not
  scaffolding)

If no such test is planned, stop. The user must add it before the fix
lands. This is the non-negotiable AC from the bug template.

### Phase 6 — Verification

Once the fix is in the working tree:

1. Capture the BEFORE evidence: `git log -1 --format='%H' HEAD~1` (or
   whatever commit is "just before the fix") + the verbatim output of
   the failing test on that commit.
2. Capture the AFTER evidence: `git log -1 --format='%H' HEAD` +
   the verbatim output of the now-passing test.
3. Paste both into the doc's (or ticket's) Verification section, in
   fenced code blocks, unedited.

If the user is unwilling to run the BEFORE check (e.g. because they
already committed), use `git stash` + `git checkout HEAD~1` to reproduce
the failure briefly, then restore. Do NOT skip this step.

### Phase 7 — Action items

From the Five Whys' Q3–Q5 answers, identify concrete follow-ups. For each:

- Assign an owner (the user, by default).
- Decide whether it is a separate beads ticket OR a check-off on the doc:
  - **Separate ticket** if it touches code, requires its own gates, or
    blocks future work.
  - **Doc-only** if it is a small note (e.g. "added a comment").
- For separate tickets, draft via `bd create --title="..." --type=task
  --priority=<n>` and link with `bd dep add <new-id> <bug-id>` so the
  dependent shows up in `bd ready` after the bug closes.

### Phase 8 — Close

Final checks before suggesting `bd close <bug-id>`:

- Every AC checkbox on the bug ticket is ticked.
- For S0/S1: the doc's `Status:` line is set to `Verified` (or `Closed`
  after sign-off).
- For S2/S3: the inline Root Cause + Verification sections are complete.
- All action items either have a ticket ID or are explicitly deferred in
  writing.
- The regression test is in the suite (run `make preflight` to confirm).

Suggest the close command with a reason that summarises what shipped:

```
bd close <bug-id> --reason "Fixed in <commit-hash>; regression test <name> guards. RCA: docs/bugs/<filename>"
```

After close, the user should:

- Run the project's ripple script if applicable (see CLAUDE.md After-Close
  Protocol).
- Push the fix + the RCA doc together so the audit trail is one commit
  range, not two.

## Rules

- **Never invent evidence.** Every Five Whys answer must cite a real
  `file:line`, commit, or log excerpt. If the evidence isn't there yet,
  stop and gather it; do not advance the analysis.
- **Never name a person as root cause.** Push back on the user if they
  try. The honest blameless answer is always a system answer.
- **Never skip the RED test.** Without a failing test, the fix has no
  proof and no future guard. This is the non-negotiable AC.
- **Never close without verification.** Before- and after-output must be
  in the doc or ticket in verbatim fenced blocks. "Looks fine" is not
  verification.
- **One bug per RCA.** If investigation reveals two independent root
  causes, split the ticket and run `/rca` on each. Joint RCAs let one
  fix slip through unverified.
- **S2/S3 stay inline.** Resist the urge to generate a full doc for
  cosmetic bugs. The doc is for incidents, not typos.
- **Action items without owners do not exist.** Strip them or assign
  them; never leave a checkbox dangling.
- **Date format matches the project's spike convention.** `YYYYMMDD`
  (no separators) so a chronological `ls docs/bugs/` is the timeline.
