# Domain Research: Municipal Water Treatment Plant Operations

## Search Queries Used

| # | Query | Useful Sources |
|---|-------|---------------|
| 1 | "municipal water treatment plant operations workflow steps process 2025" | 7/10 |
| 2 | "water treatment plant management software SCADA features" | 8/10 |
| 3 | "water treatment plant operator roles responsibilities certifications" | 6/10 |
| 4 | "water treatment plant EPA compliance reporting requirements regulations" | 5/10 |
| 5 | "water treatment chemical dosing process control equipment maintenance CMMS" | 7/10 |

**Total useful sources:** 33 across 5 queries

## Extracted Structured Knowledge

```
Domain: Municipal Water Treatment Plant Operations

Actors:
  - Plant Operator (monitors processes, adjusts parameters, takes samples)
  - Lead Operator / Shift Supervisor (oversees shift operations, handles escalations)
  - Plant Manager / Superintendent (overall operations, budget, staffing)
  - Lab Technician (runs water quality tests, analyzes samples)
  - Maintenance Technician (preventive/corrective equipment maintenance)
  - Environmental Compliance Officer (regulatory reporting, permit management)
  - EPA / State Regulator (issues NPDES permits, conducts inspections)
  - SCADA System (automated monitoring, process control, alarm generation)
  - Chemical Supplier (delivers treatment chemicals)

Key Entities:
  - Treatment Process (coagulation, flocculation, sedimentation, filtration, disinfection)
  - Water Quality Sample (pH, turbidity, chlorine residual, ammonia, dissolved oxygen)
  - NPDES Permit (discharge limits, monitoring requirements, reporting schedule)
  - Chemical Dosing Record (chemical type, dosage rate, tank levels)
  - Equipment / Asset (pumps, valves, motors, filters, sensors)
  - Work Order (maintenance task, priority, assigned technician, status)
  - Compliance Report (Discharge Monitoring Report, monthly/quarterly submissions)
  - Alarm / Alert (parameter exceedance, equipment failure, threshold breach)
  - Shift Log (operator notes, readings, events during shift)
  - Chemical Inventory (chemical type, quantity on hand, reorder point)

Primary Workflow (Daily Operations Monitoring):
  1. Plant Operator begins shift, reviews Shift Log from previous shift
  2. Plant Operator checks SCADA System for current process parameters
  3. Plant Operator performs plant walk-through, inspects equipment visually
  4. Lab Technician collects Water Quality Samples at multiple process points
  5. Lab Technician analyzes Samples for pH, turbidity, chlorine residual, etc.
  6. Plant Operator reviews Lab Results against NPDES Permit limits
  7. Plant Operator adjusts Chemical Dosing rates if parameters are out of range
  8. SCADA System logs all parameter changes and operator adjustments
  9. Plant Operator records readings and observations in Shift Log
  10. Lead Operator reviews Shift Log and flags any anomalies
  11. Environmental Compliance Officer compiles data for Discharge Monitoring Report
  12. Environmental Compliance Officer submits Report to EPA/State Regulator

Failure Modes:
  - Water quality parameter exceeds NPDES permit limits (compliance violation)
  - Chemical dosing pump failure (under/over-dosing)
  - SCADA system communication failure (blind operations)
  - Equipment breakdown requiring emergency maintenance
  - Chemical supply shortage (dosing interruption)
  - Sensor calibration drift (inaccurate readings)
  - Power outage (backup generator activation)
  - Algae bloom or unusual source water conditions
  - Operator certification lapse

Regulatory:
  - Clean Water Act (CWA): primary federal regulatory framework
  - NPDES Permits: facility-specific discharge limits and monitoring schedules
  - Discharge Monitoring Reports (DMRs): regular submissions to EPA/state
  - EPA Effluent Guidelines: technology-based limits by industry category
  - State-level operator certification: multiple levels, periodic renewal
  - Unannounced compliance inspections by EPA or state agency
  - Safe Drinking Water Act (for drinking water treatment plants)
  - OSHA requirements: confined space, chemical handling, hazmat
  - 40 CFR Part 141: National Primary Drinking Water Regulations

Existing Software:
  - VTScada (water/wastewater SCADA platform)
  - AVEVA SCADA (industrial monitoring)
  - Ignition by Inductive Automation (SCADA + HMI)
  - eWorkOrders (CMMS for water/wastewater plants)
  - OxMaint (maintenance management, operations checklists)
  - Almiren CMMS (water treatment contractor focused)
  - osapiens CMMS (compliance-focused maintenance)
  - Hach WIMS (Water Information Management System)

Sources:
  - https://www.cdc.gov/drinking-water/about/how-water-treatment-works.html
  - https://www.bls.gov/ooh/production/water-and-wastewater-treatment-plant-and-system-operators.htm
  - https://alliancewater.com/how-does-scada-help-water-and-wastewater-management/
  - https://www.vtscada.com/water-and-wastewater-scada/
  - https://www.epa.gov/compliance/clean-water-act-cwa-compliance-monitoring
  - https://www.epa.gov/eg/effluent-guidelines-implementation-compliance
  - https://oxmaint.com/industries/government/water-treatment-plant-daily-operations-checklist
  - https://eworkorders.com/water-wastewater-treatment-plants-cmms/
  - https://www.waterandwastewater.com/chemical-dosing-essential-practices-for-industry-success/
```

## Proposed 3-Story Set

### Story 1: Operator Monitors and Adjusts Treatment Process (Primary Happy Path)

```
Story: "Plant Operator Manages Daily Water Treatment Operations"
1. Plant Operator begins shift, reviews previous Shift Log
2. Plant Operator logs into SCADA System, reviews process dashboard
3. Plant Operator performs physical Walk-Through of treatment stages
4. Lab Technician collects Water Quality Samples from intake, mid-process, and effluent
5. Lab Technician runs tests for pH, turbidity, chlorine residual, dissolved oxygen
6. Lab Technician records Results in Lab Information System
7. Plant Operator reviews Lab Results against NPDES Permit limits
8. Plant Operator adjusts Chemical Dosing rate for chlorine disinfection
9. SCADA System confirms new dosing parameters and logs change
10. Plant Operator records all readings and adjustments in Shift Log
11. Lead Operator reviews Shift Log at shift handover
```

### Story 2: Parameter Exceedance Triggers Compliance Response (Primary Failure Case)

```
Story: "Turbidity Exceeds Permit Limit"
1. SCADA System detects turbidity reading above NPDES Permit threshold
2. SCADA System generates Alarm and notifies Plant Operator
3. Plant Operator acknowledges Alarm in SCADA System
4. Plant Operator investigates source: checks coagulant feed, filter status
5. Plant Operator increases Chemical Dosing for coagulant
6. Plant Operator orders Lab Technician to run confirmatory Sample
7. Lab Technician collects emergency Sample and runs turbidity test
8. Lab Technician confirms elevated reading
9. Plant Operator adjusts backwash cycle on affected Filter
10. Environmental Compliance Officer logs exceedance in Compliance Record
11. Environmental Compliance Officer determines if reportable to State Regulator
12. Lead Operator documents corrective actions in Incident Report
13. Plant Operator monitors until parameters return to permit limits
```

### Story 3: Scheduled Equipment Maintenance (Secondary Workflow)

```
Story: "Maintenance Technician Performs Preventive Maintenance on Chemical Feed Pump"
1. CMMS generates scheduled Work Order for chemical feed pump calibration
2. Maintenance Technician receives Work Order assignment
3. Maintenance Technician reviews pump maintenance history in CMMS
4. Maintenance Technician notifies Plant Operator of pump downtime
5. Plant Operator switches to backup Chemical Feed Pump via SCADA
6. Maintenance Technician performs calibration using volumetric test
7. Maintenance Technician inspects pump seals, valves, and tubing
8. Maintenance Technician replaces worn parts from Spare Parts Inventory
9. Maintenance Technician records all work in Work Order
10. Plant Operator switches back to primary pump, confirms SCADA readings
11. Maintenance Technician closes Work Order in CMMS with parts used and time
```

## Quality Rating: 3 (Usable)

**What is right:**
- All 9 actors are accurate: the distinction between Operator, Lead Operator, Lab Technician, and Maintenance Technician matches BLS job descriptions and AWWA operator roles
- The treatment process sequence (coagulation -> flocculation -> sedimentation -> filtration -> disinfection) is textbook CDC-verified
- NPDES permit compliance workflow is accurate (EPA/CWA framework)
- SCADA as central monitoring system is industry standard
- Chemical dosing adjustment is a real daily operational task
- DMR (Discharge Monitoring Report) submission is the real regulatory artifact
- CMMS integration for maintenance is industry best practice
- Operator certification requirements are accurately described

**What a domain expert would correct:**
- May differentiate between drinking water treatment and wastewater treatment more explicitly (the seed mentions "water treatment" broadly)
- Specific chemical names (alum, ferric chloride, polymer) would be added
- Laboratory procedures have specific QA/QC protocols (chain of custody, duplicates)
- Emergency response plans (chemical spills, pipe breaks) are a major workflow
- Sludge handling and disposal is a significant operational concern
- The SCADA alarm management might be more nuanced (priority levels, escalation chains)

**Conclusion:** A water treatment plant operator would recognize these workflows immediately. The regulatory framework (NPDES, CWA, DMR), the SCADA monitoring pattern, and the daily operations cycle are all accurate. An expert would add industry-specific detail but would not need to restructure the stories.

## Timing

- Start: 23:16:00
- End: 23:16:17
- **Duration: ~17 seconds** (search only)
