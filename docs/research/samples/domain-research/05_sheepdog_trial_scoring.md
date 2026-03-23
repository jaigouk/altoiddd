# Domain Research: Competitive Sheep Dog Trial Scoring

## Search Queries Used

| # | Query | Useful Sources |
|---|-------|---------------|
| 1 | "competitive sheepdog trial scoring phases outrun lift fetch drive shed pen points" | 8/10 |
| 2 | "sheepdog trial rules regulations USBCHA ISDS scoring system" | 7/10 |
| 3 | "sheepdog trial management software scoring tracking results" | 4/10 |
| 4 | "sheepdog trial event organization roles handler judge course director" | 6/10 |
| 5 | "sheepdog trial national rankings cumulative scoring championship finals" | 5/10 |

**Total useful sources:** 30 across 5 queries

## Extracted Structured Knowledge

```
Domain: Competitive Sheep Dog Trial Scoring

Actors:
  - Handler (commands the dog via whistles and gestures, stays at handler's post)
  - Dog (executes herding phases, controlled by handler)
  - Judge (scores each phase, records deductions on score sheet)
  - Scribe (official timekeeper and scorer, records judge's deductions)
  - Course Director (manages running order, available for questions, handles complaints)
  - Trial Secretary (processes entries, manages logistics, DogTrialEntry.com)
  - Set-Out Crew / Exhaust Crew (places sheep at start, removes them after run)
  - Competitor / Entrant (handler-dog team as unit of competition)

Key Entities:
  - Run (one handler-dog team's complete trial attempt)
  - Score Sheet (per-run record of deductions by phase)
  - Trial Event (sanctioned competition at a venue, with date and class)
  - Course (physical layout: handler's post, fetch panels, drive panels, shedding ring, pen)
  - Phase Score (outrun: 20pts, lift: 10pts, fetch: 20pts, drive: 30pts, shed: 10pts, pen: 10pts)
  - Time Limit (maximum allowed time for complete run)
  - Running Order (sequence of handler-dog entries for trial)
  - Season Record (handler-dog cumulative results across trials)
  - National Ranking (top 150 qualify for USBCHA National Finals)
  - Trial Class (Open, Nursery, Pro-Novice, Ranch, etc.)

Primary Workflow (Single Trial Run Scoring):
  1. Course Director announces next Handler-Dog team from Running Order
  2. Handler takes position at Handler's Post
  3. Set-Out Crew places sheep at far end of course
  4. Judge signals ready; Scribe starts timer
  5. Handler sends Dog on Outrun (pear-shaped path behind sheep) [20 pts]
  6. Judge scores Outrun: deducts for crossing, running too tight/wide, stopping
  7. Dog reaches sheep; performs Lift (first contact, moves sheep) [10 pts]
  8. Judge scores Lift: deducts for rushing, scattering, rough contact
  9. Dog brings sheep on Fetch (straight line through fetch panels to handler) [20 pts]
  10. Judge scores Fetch: deducts for missed panels, offline movement
  11. Handler directs Dog to Drive sheep away through drive panels [30 pts]
  12. Judge scores Drive: deducts for missed panels, crooked lines, poor turns
  13. Handler and Dog perform Shed (separate marked sheep in shedding ring) [10 pts]
  14. Judge scores Shed: deducts for incomplete separation, dog not in control
  15. Handler and Dog Pen sheep (guide into pen, close gate) [10 pts]
  16. Judge scores Pen: deducts for excessive force, sheep escaping, failure to close
  17. Scribe records final time and total score (100 minus total deductions)
  18. Course Director posts scores to results board

Failure Modes:
  - Dog fails to complete outrun (retires or disqualified)
  - Time limit exceeded before completing all phases
  - Dog bites sheep (immediate disqualification)
  - Sheep escape the course (loss of points, possible retirement)
  - Dog does not respond to handler commands (loss of control)
  - Incorrect shed (wrong sheep separated)
  - Handler leaves post during fetch/drive (penalty)
  - Weather disrupts trial (rain, wind affecting sheep behavior)
  - Sheep quality inconsistency (some groups harder than others)

Regulatory:
  - USBCHA rules govern sanctioned trials in the US
  - ISDS (International Sheep Dog Society) provides foundational rules worldwide
  - USBCHA judging guidelines standardize scoring specifics
  - Trial must be USBCHA-sanctioned for results to count toward national rankings
  - Top 150 Open dogs qualify for USBCHA National Sheepdog Finals
  - Course Director must be designated and available during all runs
  - Judge must be experienced handler approved by USBCHA
  - Complaints must be filed with Course Director
  - State/county-specific livestock handling regulations may apply

Existing Software:
  - DogTrialEntry.com (USBCHA-compliant running orders, online entries, posted scores)
  - USBCHA provided score sheets, scribe sheets, results forms (paper-based)
  - Mostly manual: paper score sheets, scribes, manual tallying
  - No comprehensive digital scoring platform found

Sources:
  - https://usbcha.com/resources/trial-diagrams/
  - https://usbcha.com/resources/sheepdog-trials-explained/
  - https://usbcha.com/resources/sheepdog-and-cattledog-rules/
  - https://usbcha.com/resources/judging-guidelines-part-1/
  - https://usbcha.com/resources/judging-sheepdog-trials/
  - https://www.littlehats.net/apprentice/trialing/clerking/
  - https://www.texassheepdogassoc.org/information/trial-resources
  - https://soldierhollowclassic.com/the-competition-course/
  - https://www.longshawsheepdog.com/judging-guide
  - https://sheepdogfinals.usbcha.com/
  - https://en.wikipedia.org/wiki/Sheepdog_trial
```

## Proposed 3-Story Set

### Story 1: Judge Scores a Complete Trial Run (Primary Happy Path)

```
Story: "Judge Scores Handler-Dog Team Through Complete Run"
1. Course Director calls next Handler from Running Order
2. Handler takes position at Handler's Post with Dog
3. Set-Out Crew places five sheep at far end of Course
4. Judge signals ready; Scribe starts timer on Run
5. Handler sends Dog on Outrun around sheep
6. Judge evaluates Outrun and tells Scribe deduction points [20 pts max]
7. Dog reaches sheep and performs Lift
8. Judge evaluates Lift quality and tells Scribe deduction points [10 pts max]
9. Dog brings sheep through Fetch Panels toward Handler
10. Judge evaluates Fetch line and panel navigation, tells Scribe deductions [20 pts max]
11. Handler turns sheep and Dog drives through Drive Panels
12. Judge evaluates Drive lines and turns, tells Scribe deductions [30 pts max]
13. Handler enters Shedding Ring, Handler and Dog separate marked sheep
14. Judge evaluates Shed and tells Scribe deduction [10 pts max]
15. Handler opens Pen gate, Dog guides sheep into Pen
16. Judge evaluates Pen and tells Scribe final deduction [10 pts max]
17. Scribe records total time and calculates final Score (100 minus deductions)
18. Course Director posts Score to results board
```

### Story 2: Dog Disqualified During Run (Primary Failure Case)

```
Story: "Dog Bites Sheep and Is Disqualified"
1. Handler sends Dog on Outrun
2. Dog completes Outrun and performs Lift
3. Sheep scatter during Fetch; Dog struggles to gather
4. Dog grips (bites) a sheep while attempting to control group
5. Judge immediately signals disqualification to Scribe
6. Scribe records DQ on Score Sheet with reason
7. Course Director instructs Handler to retire from Course
8. Handler recalls Dog and leaves field
9. Set-Out Crew collects sheep from Course
10. Course Director logs incident in Trial Record
11. Scribe posts DQ result to results board
```

### Story 3: Season Qualification for National Finals (Secondary Workflow)

```
Story: "Handler Accumulates Season Points for National Finals Qualification"
1. Handler enters Dog in USBCHA-Sanctioned Trial Events throughout season
2. Trial Secretary processes Entry for each event via DogTrialEntry.com
3. Handler and Dog compete in Open Class at each Trial Event
4. Judge scores each Run; Scribe records Results
5. Trial Secretary submits final Results to USBCHA after each event
6. USBCHA aggregates Results across all sanctioned events for season
7. USBCHA updates National Rankings based on cumulative performance
8. Handler checks Season Record against top-150 qualification threshold
9. USBCHA publishes qualified Handler-Dog teams for National Finals
10. Handler registers for National Sheepdog Finals (Alturas CA, Sep-Oct)
```

## Quality Rating: 3 (Usable)

**What is right:**
- The scoring phases and point allocations (20/10/20/30/10/10 = 100) are accurate per USBCHA (usbcha.com)
- All 8 actors are real: Handler, Judge, Scribe, Course Director, Set-Out Crew, Trial Secretary match USBCHA resources and TSDA rules
- The scoring method (100 minus deductions) is correct per USBCHA judging guidelines
- Phase descriptions (pear-shaped outrun, shedding ring, pen gate) match official course descriptions
- Disqualification for biting sheep is a real rule
- Top 150 qualifying for National Finals is accurate
- DogTrialEntry.com is the real software used for entries/running orders
- The paper-based scoring reality is correctly captured

**What a domain expert would correct:**
- May add "single" as a separate phase from "shed" in some trial formats
- Cross-drive is sometimes scored separately from drive
- The National Finals format has preliminary, semi-final, and final rounds with different courses
- Double-lift format for championship classes (two groups of sheep, longer outrun)
- Re-runs may be granted for issues outside handler/dog control
- "Nursery" class has different rules for young dogs
- Specific sheep breeds affect difficulty (hair sheep vs. wool sheep)
- Post-trial "drawing" to determine next day's running order

**Conclusion:** This is the most surprising result. Despite being a "very obscure" domain, web search produced detailed, accurate information because USBCHA maintains excellent online documentation. A sheepdog trial judge or course director would recognize these stories as accurate and add nuances about specific trial formats and classes, but the core scoring workflow, roles, and rules are correct.

## Timing

- Start: 23:18:44
- End: 23:19:00
- **Duration: ~16 seconds** (search only)
