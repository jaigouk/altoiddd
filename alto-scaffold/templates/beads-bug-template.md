# Beads Bug Template

Use this template for **defect tickets** (something works incorrectly,
crashes, leaks, regresses, or violates a documented contract). Bug tickets
follow a stricter flow than tasks: **reproduce → write failing
regression test → fix → verify → record root cause**. The regression test
and the verification evidence are **mandatory acceptance criteria**, not
optional polish.

```bash
bd create --type=bug --title="fix(<scope>): <one-line symptom>" --priority=<0-3>
# Or as a child of an epic:
bd create --type=bug --title="..." --parent <epic-id>
```

---

> **Before Starting:** Always groom the ticket first. Confirm the bug
> reproduces from the steps below in a clean checkout. If you cannot
> reproduce, the ticket is not ready — gather more evidence first.

> **Freshness:** If this ticket has a `review_needed` label, read the ripple
> comments (`bd comments <id>`) before claiming. Bugs filed weeks ago may
> already be fixed by an unrelated commit — verify the bug still reproduces
> on `main` before sinking time into a fix.

## Severity & Priority

These are **independent dimensions**. Set both.

| Dimension | Scale | Definition |
|-----------|-------|------------|
| **Severity** (technical impact) | S0 / S1 / S2 / S3 | How badly the system is broken |
| **Priority** (business urgency) | P0 / P1 / P2 / P3 | How soon the fix needs to land |

| Severity | Meaning | Example |
|----------|---------|---------|
| **S0** | Data loss, security breach, total outage, build broken on `main` | Lost user data; published binary corrupts state; `make ci` red on main |
| **S1** | Major feature broken; no workaround | Doc-health silently passes broken files; CLI subcommand crashes on every input |
| **S2** | Feature degraded; workaround exists | Hook fires but silently no-ops; rule misclassifies one valid input shape |
| **S3** | Cosmetic; rare edge case; developer-only nit | Typo in error message; missing template section in lint |

| Severity | RCA requirement |
|----------|-----------------|
| **S0, S1** | Separate RCA document at `docs/bugs/YYYYMMDD_<bug-id>_<slug>.md` is **MANDATORY**. Use `/rca <bug-id>` to draft it. |
| **S2, S3** | Inline `## Root Cause` and `## Verification` sections in this ticket are **MANDATORY**. Separate doc is optional. |

**Set here:**

- **Severity:** S\_\_\_
- **Priority:** P\_\_\_
- **RCA artifact:** `docs/bugs/YYYYMMDD_<bug-id>_<slug>.md` (S0/S1) | inline below (S2/S3)

## Goal / Problem

One sentence. What does the system do that it should not, OR what does it
fail to do that it should?

## Environment

| Field | Value |
|-------|-------|
| **OS / runtime** | e.g. Linux 6.8.0-111-generic / `<language-runtime>` (Go 1.26, Node 22, Python 3.12, Ruby 3.3, …) |
| **alto version** | `git describe --tags --always --dirty` or commit hash |
| **Tool surface** | CLI / MCP / dochealth rule / scaffold writer / etc. |
| **First-seen commit** | The commit on which this bug was first observed (or "unknown") |
| **Last-known-good commit** | The most recent commit on which the bug did NOT reproduce (or "unknown") |

If `git bisect` was used to narrow the introduction range, paste the
narrowed range here.

## Steps to Reproduce

Numbered, exact, copy-pasteable. Start from a known state (`git checkout
main && make clean && make build`). Anyone unfamiliar with the bug should
be able to run these and observe the failure on the first try.

1. ...
2. ...
3. ...

## Expected Result

What the system should do at step N if behaving correctly.

## Actual Result

What the system actually does at step N. **Behaviour, not hypothesis.** "It
returns 500" — not "the database connection must be leaking". Speculation
belongs in Root Cause, below.

## Evidence

Logs, stack traces, screenshots, `file:line` citations, `bd show`
snapshots — anything that disambiguates "what happened". Paste verbatim,
not paraphrased.

```
<paste evidence here>
```

## Reproducibility

| Field | Value |
|-------|-------|
| **Rate** | Always / Often (≥50%) / Sometimes (10–50%) / Rare (<10%) |
| **Across machines?** | Yes / No / Unknown |
| **Concurrency-dependent?** | Yes / No / Unknown |
| **Time-dependent?** (DST, UTC midnight, etc.) | Yes / No / Unknown |

## Detection

How was this bug discovered? Pick one:

- [ ] User report
- [ ] CI failure on `main`
- [ ] CI failure on a branch
- [ ] Internal smoke / dogfood
- [ ] Code review
- [ ] Static analysis / lint
- [ ] Adjacent ticket grooming surfaced it

## DDD Alignment

| Aspect | Detail |
|--------|--------|
| Bounded Context | Which BC owns the buggy code? |
| Layer | Domain / Application / Infrastructure |
| Aggregate (if applicable) | Which aggregate enforces the invariant the bug violates? |

If the bug crosses bounded contexts, list both and call out the boundary
where the contract breaks.

## Suspected Root Cause

One-paragraph hypothesis with file:line citations. State **what you
believe is broken**, **why you believe it**, and **what evidence would
prove or refute the hypothesis**. The post-fix Five Whys in the **Root
Cause** section below validates or refutes this hypothesis once the
investigation is complete.

If the hypothesis is "we don't know yet", write that explicitly with a
plan for narrowing it down (logging, bisection, instrumentation).

## Files in Scope

Source of truth for what files this fix will touch. A single-agent
`developer` spawn refuses to touch any file not listed here.

| Path | Action | Owner / Notes |
|------|--------|---------------|
| `<path>` | MODIFY \| NEW (test) \| DELETE | role in the fix |

## Root Cause

**S0/S1: leave this section as a one-line summary; the Five Whys live in
the RCA doc at `docs/bugs/YYYYMMDD_<bug-id>_<slug>.md`.**

**S2/S3: complete the inline Five Whys here.** Each "why" must be supported
by code references (`file:line`) or logs — no speculation. Whys 1–2
describe the incident; whys 3–5 describe the **system that allowed it**
(missing guard, missing test, mis-configured tool, unclear ownership).

> Five Whys discipline: never name a person as a root cause. "Bob forgot"
> is not an answer — "the lint rule did not catch the missing field" is.

| # | Why? | Evidence (file:line, commit, log) |
|---|------|-----------------------------------|
| Q1 | Why did <X observed in Actual Result> happen? | |
| Q2 | Why did <A1>? | |
| Q3 | Why did <A2>? *(system layer begins here)* | |
| Q4 | Why did <A3>? | |
| Q5 | Why did <A4>? | |

**Root cause (one sentence):** ...

**Contributing factors** (conditions that made it worse but weren't the primary cause):

- ...

## TDD Workflow

### RED Phase — failing regression test

Write the test that reproduces the bug BEFORE touching production code.
The test must fail on `main` with the current bug present.

```
<test-runner> <new-test-name> -count=1
# expected: FAIL
```

Test name(s) to add:

- `Test<Subject>_When<BugCondition>_Expect<CorrectBehaviour>` in
  `<path/to/test_file>`

### GREEN Phase — minimal fix

Make the smallest change that turns the RED test GREEN. Resist the urge to
refactor.

```
<test-runner> <new-test-name> -count=1
# expected: PASS
```

### REFACTOR Phase — clean + verify nothing else broke

- Run the full test suite — no regressions.
- Run lint, vet, formatting.
- Verify the fix did not introduce new edge cases (re-read the changed
  function with the RED test's edge-case lens).

```
make check       # or: make preflight && make ci-local
```

## Verification

**Mandatory for every bug fix.** This section is what proves the bug is
gone AND will stay gone.

### Before fix (REPRODUCE)

Commit/branch where the bug reproduces (`git rev-parse HEAD` BEFORE the fix):

```
commit: <hash>
$ <command that triggers the bug>
<observed failure output — paste verbatim>
```

### After fix (RESOLVED)

Commit where the regression test passes (`git rev-parse HEAD` AFTER the fix):

```
commit: <hash>
$ <same command as above>
<observed success output — paste verbatim>
```

### Regression guard (FUTURE-PROOF)

- New test: `<test name>` in `<path/to/test_file>` — keeps the fix verified
  on every CI run. **This test must stay in the suite.** Deleting it
  later requires a written justification on the PR.

## Acceptance Criteria

Every box must be ticked before close. The first three are **non-negotiable**.

- [ ] **Failing regression test added** in the RED phase (cite test name +
      file)
- [ ] **Fix passes the new test** AND the full suite — no other failures
      introduced
- [ ] **Verification section** above filled in with before/after commit
      hashes and verbatim command output
- [ ] (S0/S1 only) **RCA document** at `docs/bugs/YYYYMMDD_<bug-id>_<slug>.md`
      exists with completed Five Whys, timeline, and action items
- [ ] (S2/S3) **Inline Root Cause section** above completed with all five
      whys and `file:line` evidence
- [ ] **Action items** from RCA (where applicable) filed as separate tickets
      via `bd create`, with `bd dep add` set if they block other work
- [ ] All quality gates pass (`make ci-local` or equivalent)
- [ ] Doc-health passes if any scaffold asset was touched
- [ ] Out-of-scope items (below) are explicitly NOT addressed

## Out of Scope

What this fix deliberately does NOT address. List adjacent issues you spotted
during investigation but chose not to fix here — and create separate tickets
for them.

- ...

## Risks / Dependencies

- Risk: ...
- Dependency: ... (also set `bd dep add <this-id> <blocking-id>` — text-only
  deps are invisible to `bd ready` / ripple review)

## Labels

`bug`, `severity:S<n>`, `<area>` (e.g. `dochealth`, `bootstrap`, `scaffold`),
plus optional flow labels (`dogfooding`, `regression`, `flaky`).

---

## Notes on this template

- **Test before fix.** A bug closed without a RED-then-GREEN test in the
  diff has nothing stopping the same regression from shipping again. The
  pre-push hook + CI lint will not catch this — reviewers must.
- **Verification is not optional.** "Looks good on my machine" is not a fix;
  pasted before/after evidence is.
- **Blameless RCA.** Whys 3–5 describe systems (missing tests, missing
  types, missing rules), not people. If a "why" lands on a person, ask
  "why was the system arranged so they could ship that?" and continue.
- **One RCA per bug.** If investigation reveals two distinct root causes,
  split into two bug tickets so each can be verified independently.
