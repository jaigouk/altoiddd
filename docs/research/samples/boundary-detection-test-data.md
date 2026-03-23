# Boundary Detection Test Data Tables

**Date:** 2026-03-23
**Purpose:** Raw data for precision/recall calculations across 3 domains + 2 edge cases.

---

## 1. Known-Correct Bounded Contexts (Ground Truth)

### E-commerce Marketplace

| Context | Key Actors | Key Work Objects | Source |
|---------|-----------|-----------------|--------|
| Catalog | Seller, Customer | Product Catalog, Product Listing, Price | Domain expertise |
| Cart/Checkout | Customer | Cart, Order | Domain expertise |
| Payment | Customer, Payment Provider | Payment | Domain expertise |
| Fulfillment/Shipping | Seller, Delivery Service | Delivery, Shipment | Domain expertise |
| Commission/Platform | Platform | Commission | Domain expertise |
| Inventory | Seller | Inventory Record, Stock | Domain expertise |

**Total known contexts: 6** (could reasonably be argued as 5 if Inventory is merged with Catalog)

### Vet Clinic

| Context | Key Actors | Key Work Objects | Source |
|---------|-----------|-----------------|--------|
| Scheduling | Pet Owner, Receptionist | Appointment, Vet Schedule | Domain expertise |
| Clinical/Examination | Veterinarian, Vet Tech | Medical Record, Diagnosis, Treatment, Triage Assessment | Domain expertise |
| Billing | Billing System, Receptionist | Invoice | Domain expertise |
| Pharmacy/Dispensary | Pharmacy, Veterinarian | Prescription, Medication | Domain expertise |

**Total known contexts: 4**

### alto

| Context | Key Actors | Key Work Objects | Source |
|---------|-----------|-----------------|--------|
| Bootstrap | Developer, alto CLI | README, Project Config | docs/DDD.md |
| Guided Discovery | Developer, alto CLI, AI Domain Researcher | Domain Stories, Bounded Context Map | docs/DDD.md |
| Ticket Pipeline | alto CLI, Beads Issue Tracker | Epic, Task, Dependency Graph | docs/DDD.md |
| Architecture Testing | alto CLI | Fitness Tests | docs/DDD.md |
| Tool Translation | alto CLI, AI Coding Tool | Tool Configs | docs/DDD.md |
| Knowledge Base | AI Domain Researcher | Domain Research | docs/DDD.md |

**Total known contexts: 6** (at minimum; DDD.md lists 10 including shared/generic)

### Edge Case 1: TODO App

| Context | Key Actors | Key Work Objects | Source |
|---------|-----------|-----------------|--------|
| (Single context) | User | TODO Item, TODO List | Domain expertise |

**Total known contexts: 1**

### Edge Case 2: Recipe Sharing (Ambiguous)

Reasonable interpretations:
- **A) 1 context:** Simple CRUD — everything is Recipe Management
- **B) 2 contexts:** Recipe Management + Community/Social
- **C) 3 contexts:** Recipe Management + Community/Social + Moderation

**Ground truth: ambiguous (1-3)**

---

## 2. Manual Heuristic Analysis

### E-commerce (3 stories)

| Signal Type | Signal Found | Stories | Sentences | Proposed Boundary |
|------------|-------------|---------|-----------|-------------------|
| One-way flow | Order flows from Cart/Checkout to Seller, never back | 01, 01b | 01:6-8, 01b:7-9 | Cart/Checkout -> Fulfillment |
| One-way flow | Payment info flows from Customer to Payment Provider, result flows to Platform | 01, 01b | 01:5-6, 01b:2-3 | Payment is separate context |
| One-way flow | Product flows from Seller to Delivery Service to Customer | 01 | 01:9-11 | Fulfillment/Shipping is separate |
| One-way flow | Commission calculated by Platform, never modified by Seller | 01, 01c | 01:7, 01c:6 | Platform/Commission is separate |
| Language diff | "Product" in Catalog (listing attributes) vs "Product" in Fulfillment (physical item) | 01, 01c | 01:2 vs 01:9 | Catalog vs Fulfillment |
| Language diff | "Order" in Cart (items+total) vs "Order" in Seller context (notification to fulfill) | 01, 01b | 01:6 vs 01:8 | Cart/Checkout vs Fulfillment |
| Diff trigger | Browsing/purchasing is customer-initiated; Inventory mgmt is seller-initiated | 01 vs 01c | all | Catalog vs Inventory |
| Org boundary | Delivery Service never appears in inventory/pricing stories | 01 vs 01c | n/a | Fulfillment is separate org |
| Same obj diff ctx | "Payment" processed by Provider (approval) vs calculated by Platform (commission) | 01, 01b | 01:6-7, 01b:3,7 | Payment vs Platform are separate |

**Manual detection result:** 5-6 contexts identified (Catalog, Cart/Checkout, Payment, Fulfillment, Platform/Commission, [Inventory])

### Vet Clinic (3 stories)

| Signal Type | Signal Found | Stories | Sentences | Proposed Boundary |
|------------|-------------|---------|-----------|-------------------|
| One-way flow | Appointment flows from Scheduling to Clinical, never back | 02, 02c | 02:3->6, 02c:3->5 | Scheduling -> Clinical |
| One-way flow | Treatment info flows from Clinical to Billing | 02 | 02:8->9 | Clinical -> Billing |
| One-way flow | Prescription flows from Veterinarian to Pharmacy | 02c | 02c:7->8 | Clinical -> Pharmacy |
| Language diff | "Pet" in Scheduling (name, owner, appointment time) vs "Pet" in Clinical (patient, medical history) | 02, 02b | 02:3 vs 02:6 | Scheduling vs Clinical |
| Language diff | "Invoice" generated by Billing System (02) vs Receptionist (02b, 02c) | 02 vs 02b,02c | 02:9 vs 02b:9 | Billing actor ambiguity |
| Diff trigger | Scheduled visit (time-based booking) vs Emergency walk-in (event-based arrival) | 02 vs 02b | 02:1 vs 02b:1 | Scheduling vs Emergency/Triage |
| Org boundary | Billing System only appears in story 02; Pharmacy only in 02c; Vet Tech only in 02b | all | n/a | Separate departments |
| Same obj diff ctx | "Medical Record" in Scheduling (reference for booking) vs Clinical (active documentation) | 02, 02b, 02c | 02c:2 vs 02:6-7 | Scheduling vs Clinical |

**Manual detection result:** 4 contexts identified (Scheduling, Clinical, Billing, Pharmacy)

### alto (3 stories)

| Signal Type | Signal Found | Stories | Sentences | Proposed Boundary |
|------------|-------------|---------|-----------|-------------------|
| One-way flow | Domain Stories flow from Discovery to Ticket Pipeline, never back | 03 vs 03c | 03:5->12, 03c:1 | Discovery -> Ticket Pipeline |
| One-way flow | Research findings flow from AI Domain Researcher to alto CLI | 03b | 03b:6->7 | Knowledge -> Discovery |
| One-way flow | Generated artifacts flow from alto CLI to AI Coding Tool | 03 | 03:14->15 | Generation -> Tool Translation |
| Language diff | "DDD Artifacts" in Discovery (produced) vs Ticket Pipeline (consumed) | 03, 03c | 03:9, 03c:1 | Discovery vs Ticket Pipeline |
| Diff trigger | Bootstrap/Discovery is developer-initiated; Ticket generation is artifact-triggered | 03 vs 03c | 03:1 vs 03c:1 | Bootstrap/Discovery vs Ticket Pipeline |
| Org boundary | AI Domain Researcher only in story 03b; Beads Issue Tracker only in 03c | 03b, 03c | n/a | Knowledge vs Ticket Pipeline |
| Same obj diff ctx | "Bounded Context Map" identified (03:6) vs read (03c:2) — different lifecycle stages | 03, 03c | 03:6, 03c:2 | Discovery vs Ticket Pipeline |

**Manual detection result:** 4-5 contexts identified (Bootstrap/Discovery, Ticket Pipeline, Knowledge/Research, Tool Translation, [Architecture Testing])

### Edge Case 1: TODO App (1 story)

| Signal Type | Signal Found | Stories | Sentences | Proposed Boundary |
|------------|-------------|---------|-----------|-------------------|
| One-way flow | None — single actor, no handoffs | edge-01 | n/a | No boundary |
| Language diff | None — consistent vocabulary | edge-01 | n/a | No boundary |
| Diff trigger | None — all user-initiated | edge-01 | n/a | No boundary |
| Org boundary | None — single actor | edge-01 | n/a | No boundary |
| Same obj diff ctx | None — TODO Item used consistently | edge-01 | n/a | No boundary |

**Manual detection result:** 1 context (correct: no split needed)

### Edge Case 2: Recipe Sharing (1 story)

| Signal Type | Signal Found | Stories | Sentences | Proposed Boundary |
|------------|-------------|---------|-----------|-------------------|
| One-way flow | Recipe flows Cook -> Community Member, never back | edge-02 | 1->3 | Recipe Mgmt vs Social |
| One-way flow | Flagged content flows Community -> Moderator | edge-02 | 6->7 | Social vs Moderation |
| Language diff | "Recipe" in Cook's context (creation, editing) vs Community context (browsing, rating) | edge-02 | 1-2 vs 3-6 | Weak — same meaning |
| Org boundary | Moderator acts only on flagged content, never on recipes directly | edge-02 | 7 | Moderation separate |

**Manual detection result:** 1-2 contexts (Recipe Management + maybe Community, Moderation is too thin)

---

## 3. Algorithmic Detection Results

### E-commerce

| Signal Type | Algo Found | Correct? | Notes |
|------------|-----------|----------|-------|
| one_way_flow: Payment Provider -> Platform | Yes | Partial | Correct direction but names a sub-relationship |
| one_way_flow: Customer -> Payment Provider | Yes | Yes | Payment boundary |
| one_way_flow: Seller -> Delivery Service | Yes | Yes | Fulfillment boundary |
| one_way_flow: Platform -> Customer | Yes | Partial | This is notification, not a BC signal per se |
| same_object_diff_context: Payment (declines, processes) | Yes | Yes | Payment used differently in success/failure |
| different_trigger: 2 types | Yes | Yes | Seller-initiated vs customer-initiated |
| MISSED: Catalog vs Cart distinction | - | - | Algorithm does not detect this |
| MISSED: Inventory as separate | - | - | Too few signals from 3 stories |

**Algorithm context proposals:** 3 (Payment, Platform/Commission, Customer+Seller cluster)
**Correct contexts identified:** 2-3 of 6 known

### Vet Clinic

| Signal Type | Algo Found | Correct? | Notes |
|------------|-----------|----------|-------|
| one_way_flow: Receptionist -> Vet Tech | Yes | Yes | Scheduling -> Triage handoff |
| one_way_flow: Vet Tech -> Veterinarian | Yes | Yes | Triage -> Clinical handoff |
| one_way_flow: Veterinarian -> Pharmacy | Yes | Yes | Clinical -> Pharmacy |
| one_way_flow: Pet Owner -> Veterinarian | Yes | Partial | Not really a BC signal (patient arrival) |
| one_way_flow: Pharmacy -> Receptionist | Yes | Partial | This flow is indirect |
| one_way_flow: Vet -> Billing System | Yes | Yes | Clinical -> Billing |
| one_way_flow: Billing System -> Pet Owner | Yes | Yes | Billing -> outside |
| same_object_diff_context: Invoice | Yes | Yes | Generated vs paid |
| same_object_diff_context: Medical Record | Yes | Yes | Checked (scheduling) vs reviewed/updated (clinical) |
| org_boundary: Billing System / Pharmacy | Yes | Yes | Never co-appear |
| org_boundary: Billing System / Vet Tech | Yes | Yes | Never co-appear |
| org_boundary: Pharmacy / Vet Tech | Yes | Partial | Coincidence (both only in 1 story each) |
| different_trigger: 2 types | Yes | Partial | Emergency classified as "system" (wrong — should be event) |

**Algorithm context proposals:** 3 (Receptionist+Billing+Veterinarian cluster, Pet Owner+Pharmacy cluster, Vet Tech cluster)
**Correct contexts identified:** 3 of 4 known (Scheduling/Clinical merged; Billing partially; Pharmacy partially; Triage identified)

### alto

| Signal Type | Algo Found | Correct? | Notes |
|------------|-----------|----------|-------|
| one_way_flow: AI Domain Researcher -> alto CLI | Yes | Yes | Knowledge -> Discovery |
| one_way_flow: Developer -> AI Domain Researcher | Yes | Partial | This is user interaction, not a BC signal |
| same_object_diff_context: Bounded Context Map | Yes | Yes | Created vs consumed |
| different_trigger: 2 types | Yes | Yes | Developer-initiated vs artifact-triggered |

**Algorithm context proposals:** 2 (Developer+alto CLI cluster, AI Domain Researcher cluster)
**Correct contexts identified:** 2 of 6 known (Discovery and Knowledge partially identified)

### Edge Case 1: TODO App

| Signal Type | Algo Found | Correct? | Notes |
|------------|-----------|----------|-------|
| (none) | Correct | Yes | No signals = no split = 1 context |

**Algorithm context proposals:** 1
**Correct contexts identified:** 1 of 1 known

### Edge Case 2: Recipe Sharing

| Signal Type | Algo Found | Correct? | Notes |
|------------|-----------|----------|-------|
| one_way_flow: Cook -> Community Member | Yes | Debatable | Could indicate boundary or just sequence |
| one_way_flow: Community Member -> Moderator | Yes | Debatable | Could indicate moderation boundary |
| one_way_flow: Moderator -> Cook | Yes | Debatable | Feedback loop, not necessarily a BC |

**Algorithm context proposals:** 3
**Correct contexts identified:** 1-3 of 1-3 known (ambiguous ground truth)

---

## 4. Precision and Recall Calculations

### Definition
- **True Positive (TP):** Detected boundary that matches a known-correct boundary
- **False Positive (FP):** Detected boundary that does not match any known-correct boundary
- **False Negative (FN):** Known-correct boundary that was not detected
- **Precision = TP / (TP + FP)**
- **Recall = TP / (TP + FN)**

### Manual Detection

| Domain | Known | Detected | TP | FP | FN | Precision | Recall |
|--------|-------|----------|----|----|----|-----------|--------|
| E-commerce | 6 | 5-6 | 5 | 0-1 | 1 | 0.83-1.00 | 0.83 |
| Vet Clinic | 4 | 4 | 4 | 0 | 0 | 1.00 | 1.00 |
| alto | 6 | 4-5 | 4 | 0-1 | 2 | 0.80-1.00 | 0.67 |
| TODO App | 1 | 1 | 1 | 0 | 0 | 1.00 | 1.00 |
| Recipe | 1-3 | 1-2 | 1 | 0-1 | 0-2 | 0.50-1.00 | 0.50-1.00 |
| **Average** | | | | | | **0.87** | **0.85** |

### Algorithmic Detection

| Domain | Known | Detected | TP | FP | FN | Precision | Recall |
|--------|-------|----------|----|----|----|-----------|--------|
| E-commerce | 6 | 3 | 2 | 1 | 4 | 0.67 | 0.33 |
| Vet Clinic | 4 | 3 | 2.5 | 0.5 | 1.5 | 0.83 | 0.63 |
| alto | 6 | 2 | 1.5 | 0.5 | 4.5 | 0.75 | 0.25 |
| TODO App | 1 | 1 | 1 | 0 | 0 | 1.00 | 1.00 |
| Recipe | 1-3 | 3 | 1-3 | 0-2 | 0 | 0.33-1.00 | 1.00 |
| **Average** | | | | | | **0.72** | **0.52** |

---

## 5. Confidence Score Calibration Data

### Signals that corresponded to TRUE boundaries

| Signal | Domain | Confidence | True Boundary? |
|--------|--------|-----------|----------------|
| one_way_flow: Customer -> Payment Provider | E-comm | 0.50 | YES |
| one_way_flow: Seller -> Delivery Service | E-comm | 0.50 | YES |
| same_object_diff_context: Payment | E-comm | 0.70 | YES |
| different_trigger: 2 types | E-comm | 0.60 | YES |
| one_way_flow: Vet Tech -> Veterinarian | Vet | 0.50 | YES |
| one_way_flow: Vet -> Pharmacy | Vet | 0.50 | YES |
| one_way_flow: Vet -> Billing System | Vet | 0.50 | YES |
| same_object_diff_context: Medical Record | Vet | 0.85 | YES |
| same_object_diff_context: Invoice | Vet | 0.80 | YES |
| org_boundary: Billing/Pharmacy | Vet | 0.40 | YES |
| one_way_flow: AI Researcher -> alto CLI | alto | 0.50 | YES |
| same_object_diff_context: BCM | alto | 0.70 | YES |
| different_trigger: 2 types | alto | 0.60 | YES |

### Signals that corresponded to FALSE boundaries

| Signal | Domain | Confidence | True Boundary? |
|--------|--------|-----------|----------------|
| one_way_flow: Platform -> Customer | E-comm | 0.50 | NO (notification, not BC) |
| one_way_flow: Pet Owner -> Vet | Vet | 0.50 | NO (patient arrival, not BC) |
| one_way_flow: Pharmacy -> Receptionist | Vet | 0.50 | NO (indirect flow artifact) |
| org_boundary: Pharmacy/Vet Tech | Vet | 0.40 | PARTIAL (coincidence from few stories) |
| one_way_flow: Dev -> AI Researcher | alto | 0.50 | NO (user interaction, not BC) |

### Calibration Summary

| Confidence Range | True Boundary Count | False Boundary Count | Accuracy |
|-----------------|--------------------|--------------------|----------|
| 0.80 - 1.00 | 2 | 0 | 100% |
| 0.60 - 0.79 | 4 | 0 | 100% |
| 0.40 - 0.59 | 7 | 5 | 58% |
| 0.00 - 0.39 | 0 | 0 | n/a |
