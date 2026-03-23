# Domain Research: Artisan Cheese Aging and Inventory

## Search Queries Used

| # | Query | Useful Sources |
|---|-------|---------------|
| 1 | "artisan cheese aging process workflow steps cave cellar affinage" | 7/10 |
| 2 | "artisan cheese making business inventory management software" | 6/10 |
| 3 | "cheese aging cave temperature humidity monitoring batch tracking" | 5/10 |
| 4 | "artisan cheese wholesale orders restaurants distribution roles" | 5/10 |
| 5 | "cheese making FDA regulations food safety HACCP requirements artisan" | 5/10 |

**Total useful sources:** 28 across 5 queries

## Extracted Structured Knowledge

```
Domain: Artisan Cheese Aging and Inventory

Actors:
  - Cheese Maker / Head Cheese Maker (makes cheese, manages production)
  - Affineur (manages aging process, turns/washes/brushes wheels)
  - Production Assistant (assists with make day, handles curds)
  - Cave/Cellar Manager (monitors cave conditions, organizes wheels)
  - Sales/Wholesale Manager (manages restaurant/shop accounts, takes orders)
  - Delivery Driver (distributes cheese to wholesale customers)
  - Restaurant Buyer / Chef (places wholesale orders, receives deliveries)
  - Retail Customer (buys directly at farm shop or farmers market)
  - FDA Inspector / Health Inspector (audits food safety compliance)

Key Entities:
  - Cheese Wheel (style, batch, make date, weight, current age, location)
  - Batch / Make (milk source, recipe, date, yield, notes)
  - Aging Cave / Cellar (name, temperature, humidity, capacity)
  - Environmental Reading (temperature, humidity, timestamp, cave ID)
  - Cheese Style / Recipe (name, milk type, target age range, process steps)
  - Wholesale Order (customer, items, quantities, delivery date)
  - Customer Account (restaurant/shop, contact, delivery zone, order history)
  - Inventory Count (style, age range, quantity available, quantity aging)
  - HACCP Plan (critical control points, monitoring procedures, corrective actions)
  - Food Safety Log (date, CCP checked, reading, corrective action if any)

Primary Workflow (Cheese Production Through Aging):
  1. Cheese Maker receives Milk from dairy farm (logs source, temperature, quality)
  2. Cheese Maker pasteurizes Milk (or uses raw, noting 60-day aging requirement)
  3. Cheese Maker adds cultures, rennet; manages curd formation per Recipe
  4. Cheese Maker presses curds into Cheese Wheel molds
  5. Cheese Maker salts wheels (brine or dry salt)
  6. Cheese Maker records Batch details (date, milk source, recipe, yield)
  7. Cheese Maker transfers Wheels to Aging Cave
  8. Affineur places Wheels on shelves, records Cave location and position
  9. Affineur monitors Environmental Readings (temperature 50-55F, humidity 85-95%)
  10. Affineur performs regular care: turning, brushing, or washing Wheels on schedule
  11. Affineur evaluates Wheel maturity (texture, aroma, taste sampling)
  12. Affineur marks Wheels as ready for sale when target age reached

Failure Modes:
  - Temperature or humidity drift in aging cave (mold issues, cracking)
  - Unwanted mold contamination on wheel surface
  - Milk quality issue at reception (antibiotic residue, high bacteria count)
  - Wheel cracks during aging (moisture loss too rapid)
  - Batch fails quality check (off-flavors, texture defects)
  - Cave equipment failure (refrigeration, humidifier)
  - Inventory mismatch (wheels sold that are not yet ready)
  - FDA inspection finding (HACCP non-compliance)
  - Wholesale order for quantity not available in correct age range
  - Delivery logistics failure (temperature control during transport)

Regulatory:
  - FDA Food Safety Modernization Act (FSMA): preventive controls required
  - HACCP plans required: critical control points from milk reception to packaging
  - Raw milk cheese: must be aged minimum 60 days at >35F (21 CFR 133)
  - Pasteurization requirements for non-60-day-aged cheese
  - State dairy licensing for commercial cheese production
  - American Cheese Society audit checklists and compliance guides
  - GFSI (Global Food Safety Initiative) certification optional but valued by retailers
  - Labeling requirements: ingredients, allergens, net weight, facility info
  - Cold chain requirements for distribution (temperature-controlled transport)

Existing Software:
  - CheeseCrafter (production + quality management, batch tracking)
  - DairyCraftPro (batch tracking, inventory, compliance for small creameries)
  - Acctivate (dairy/cheese process manufacturing, lot traceability)
  - BatchMaster (dairy ERP, FDA compliance, BRC/SQF lot traceability)
  - Strinos ERP (cheese production, quality, distribution)
  - Minotaur ERP (food manufacturing, ingredient to finished goods traceability)
  - WiFi environmental sensors + data loggers (temperature/humidity monitoring)

Sources:
  - https://www.cheeseprofessor.com/blog/affinage-101-aging-cheese
  - https://www.cheeseconnoisseur.com/the-process-of-aging-cheese/
  - https://clovercreekcheese.com/2025/08/07/affinage-and-aging-how-we-turn-time-into-taste/
  - https://cheeseforthought.com/cheese-temperature-control/
  - https://dairycraftpro.com/cheese-production-app/
  - https://pagepedersen.com/products/software/cheesecrafter-total-production-and-quality-management-software
  - https://guides.cheesesociety.org/safecheesemakinghub/faq
  - https://extension.psu.edu/food-safety-plans-for-small-scale-cheesemakers
  - https://agriculture.vermont.gov/sites/agriculture/files/documents/AgDevReports/Specialty%20Cheese%20Market%20Research%20Report.pdf
  - https://academyofcheese.org/subjects/cheese-buying-and-distribution/
```

## Proposed 3-Story Set

### Story 1: Cheese Maker Produces Batch and Begins Aging (Primary Happy Path)

```
Story: "Cheese Maker Creates New Batch and Places Wheels in Aging Cave"
1. Cheese Maker receives Milk Delivery from dairy farm
2. Cheese Maker checks Milk temperature and records quality in Food Safety Log
3. Cheese Maker pasteurizes Milk (or notes raw milk for 60-day aging rule)
4. Cheese Maker follows Recipe: adds cultures, rennet, manages curd formation
5. Production Assistant presses Curds into Wheel molds
6. Cheese Maker salts Wheels in brine tank
7. Cheese Maker creates Batch Record (milk source, recipe, date, yield, wheel count)
8. Cheese Maker labels each Wheel with Batch number and make date
9. Affineur transfers Wheels to designated Aging Cave
10. Affineur places Wheels on shelves, records location in Inventory
11. Affineur checks Cave Environmental Readings (temperature 50-55F, humidity 85-95%)
12. Affineur sets aging schedule: turning frequency, wash schedule per Cheese Style
```

### Story 2: Environmental Drift Threatens Aging Wheels (Primary Failure Case)

```
Story: "Cave Humidity Drops Below Threshold"
1. Environmental Sensor detects humidity drop below 80% in Aging Cave
2. System sends Alert to Cave Manager's phone
3. Cave Manager checks Environmental Reading history for trend
4. Cave Manager inspects humidifier equipment in Aging Cave
5. Cave Manager discovers humidifier malfunction
6. Cave Manager adjusts manual humidity control as temporary fix
7. Cave Manager submits maintenance request for humidifier repair
8. Affineur inspects Wheels for cracking or excessive moisture loss
9. Affineur moves most vulnerable Wheels to alternate Cave with proper humidity
10. Affineur records condition assessment in Batch Record for affected Wheels
11. Cave Manager confirms humidity restored after repair
12. Affineur continues monitoring affected Wheels on accelerated check schedule
```

### Story 3: Wholesale Order Fulfillment (Secondary Workflow)

```
Story: "Sales Manager Fulfills Restaurant Wholesale Order"
1. Restaurant Buyer contacts Sales Manager to place Wholesale Order
2. Sales Manager checks Inventory for requested Cheese Styles and ages
3. Sales Manager confirms availability and delivery date with Restaurant Buyer
4. Sales Manager creates Wholesale Order with quantities and pricing
5. Affineur selects Wheels from Aging Cave that meet age and quality criteria
6. Affineur performs final quality check on selected Wheels (taste, texture, appearance)
7. Affineur records Wheels pulled from inventory, updates Inventory counts
8. Production Assistant cuts and packages Wheels per order specifications
9. Production Assistant labels packages with required info (ingredients, weight, batch, date)
10. Delivery Driver loads packaged Cheese into temperature-controlled vehicle
11. Delivery Driver delivers order to Restaurant Buyer
12. Sales Manager updates Customer Account with order history
```

## Quality Rating: 3 (Usable)

**What is right:**
- The Affineur role is accurate and domain-specific (from Cheese Professor, Clover Creek Cheese Cellar)
- Environmental monitoring (50-55F, 85-95% humidity) matches industry sources
- The batch-to-wheel tracking workflow reflects CheeseCrafter and DairyCraftPro capabilities
- FDA/FSMA regulatory requirements and HACCP plans are correctly identified
- 60-day raw milk aging rule is a real FDA regulation (21 CFR 133)
- Wholesale distribution to restaurants matches Vermont specialty cheese market research (75% sell to restaurants)
- Environmental sensor monitoring with phone alerts is a real practice

**What a domain expert would correct:**
- May differentiate between hard/soft cheese aging protocols (very different care regimens)
- Specific turning schedules vary dramatically (daily for washed rinds, weekly for hard cheeses)
- "Cave Manager" may not be a separate role in a small operation (Cheese Maker does this)
- Milk quality testing is more specific: somatic cell count, antibiotic tests, fat percentage
- Seasonal milk variation affects cheese quality and recipe adjustments
- Farmers market direct sales is a significant revenue channel not covered in stories
- Pricing is often by-the-pound after aging weight loss, which is complex to calculate

**Conclusion:** Despite being a small niche domain, web search produced surprisingly rich results. The aging process, environmental monitoring, regulatory requirements, and wholesale distribution are all accurately captured. An artisan cheese maker would recognize these workflows and need only to add specifics about their cheese styles and operation size.

## Timing

- Start: 23:17:17
- End: 23:17:33
- **Duration: ~16 seconds** (search only)
