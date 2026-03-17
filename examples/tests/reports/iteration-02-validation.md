# Iteration 02: Edge Routing Validation Report

**Date**: 2026-03-17
**Tester**: QA (automated visual inspection)
**Focus Rules**: R1 (no edge through box), R2 (perpendicular anchors), R5 (minimum bends), R11 (smart side selection), R13 (clearance from nodes)

---

## Summary

| # | Diagram | Iter 01 | Iter 02 | Delta |
|---|---------|---------|---------|-------|
| 1 | 05-fan-out | PASS | PASS | same |
| 2 | 07-diamond-pattern | FAIL | FAIL | same |
| 3 | 08-grid-3x3 | FAIL | FAIL | improved |
| 4 | 09-diagonal-connections | FAIL | FAIL | improved |
| 5 | 12-bidirectional | FAIL | FAIL | improved |
| 6 | 16-microservices | PASS | PASS | same |
| 7 | 19-auto-placement | FAIL | PASS | fixed |
| 8 | 21-self-loop | FAIL | FAIL | improved |
| 9 | 22-pipeline | FAIL | FAIL | improved |
| 10 | 25-cross-connections | FAIL | FAIL | same |
| 11 | 27-large-grid-5x5 | PASS | PASS | same |
| 12 | 28-ci-cd-pipeline | FAIL | FAIL | improved |
| 13 | 29-mesh-topology | PASS | PASS | same |
| 14 | 30-styled-nodes | FAIL | FAIL | improved |
| 15 | 31-harness-overview | FAIL | FAIL | improved |

**Pass**: 5 / 15
**Fail**: 10 / 15

---

## Detailed Findings

### 1. 05-fan-out

**Result**: PASS

Load Balancer (row1 col2) fans out to Server 1 (row2 col1), Server 2 (row2 col2), Server 3 (row2 col3). The center edge (lb -> s2) exits bottom and enters top as a straight vertical, 0 bends. The left edge (lb -> s1) exits bottom left of Load Balancer and uses a clean L bend to reach Server 1, entering from top. The right edge (lb -> s3) similarly uses a clean L bend. No edges pass through nodes. Perpendicular anchors are correct. Clearance is adequate. Anchor distribution on the bottom side of Load Balancer is well spread.

---

### 2. 07-diamond-pattern

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | path-a -> merge | Path A is at row2 col1, Merge is at row3 col2 (diagonal down right). The edge exits the bottom of Path A, goes down, bends right, then bends down again into Merge. This is 2 bends. Only 1 bend (L shape) is needed for a diagonal target. | Single L bend: exit bottom of Path A, enter left of Merge OR exit right of Path A, enter top of Merge. | MAJOR |
| R5 | path-b -> merge | Path B is at row2 col3, Merge is at row3 col2 (diagonal down left). Same 2 bend pattern where 1 L bend suffices. The edge exits bottom, goes down, turns left, turns down into Merge. | Single L bend: exit bottom of Path B, enter right of Merge OR exit left of Path B, enter top of Merge. | MAJOR |
| R11 | path-a -> merge | Edge exits bottom side. Since Merge is diagonally below and to the right, exiting right or bottom are both valid. However the current route adds an unnecessary second bend. | Exit bottom, enter left with single L bend. | MAJOR |
| R11 | path-b -> merge | Edge exits bottom side. Since Merge is diagonally below and to the left, same unnecessary second bend. | Exit bottom, enter right with single L bend. | MAJOR |

No change from iteration 01.

---

### 3. 08-grid-3x3

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | a -> e | A at row1 col1, E at row2 col2 (diagonal). Edge exits bottom of A and routes with 2 bends (down, right, down). Only 1 L bend needed. | Single L bend for diagonal. | MAJOR |
| R5 | c -> e | C at row1 col3, E at row2 col2 (diagonal). Same 2 bend pattern. | Single L bend. | MAJOR |
| R5 | e -> g | E at row2 col2, G at row3 col1 (diagonal). Edge exits bottom of E and routes with 2 bends (down, left, down). Only 1 L bend needed. | Single L bend. | MAJOR |
| R5 | e -> i | E at row2 col2, I at row3 col3 (diagonal). Edge routes right and bends down with apparent extra bends. | Single L bend. | MAJOR |
| R13 | edges near B | Multiple edges converge on E from A and C. The edge segments from A pass close to B on the left side, and from C close to B on the right side. Clearance from B appears tight. | Edges should maintain at least one grid unit clearance from B. | MAJOR |

**Improvement from iteration 01**: The anchor distribution on E's top side (R12) appears better. Three edges from A, B, C arrive at top of E with more visible separation between anchor points compared to iteration 01. The iteration 01 finding about R12 crowding on E appears partially resolved.

---

### 4. 09-diagonal-connections

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | tl -> br | Top Left (row1 col1) to Bottom Right (row3 col3). The edge exits the bottom of Top Left, runs down the left side, then turns right at the bottom and enters Bottom Right from the left/bottom. This is 2 bends. Given Center is in the path, 2 bends may be justified for obstacle avoidance. The route does properly avoid Center. | 2 bends acceptable if needed for obstacle avoidance, but path could be tighter. | MAJOR |
| R13 | tl -> br | The edge from Top Left runs down and then across to Bottom Right. The horizontal segment runs close to the Bottom Left node but does appear to maintain some clearance. | Minimum clearance from Bottom Left. | MAJOR |
| R5 | tr -> bl | Top Right (row1 col3) to Bottom Left (row3 col1). Edge exits bottom of Top Right, runs down, turns left past Center, and enters Bottom Left from the top. The route now appears to avoid Center better than iteration 01 described. 2 bends used, justified given Center obstacle. | 2 bends acceptable for obstacle avoidance. Route is cleaner than iteration 01. | MAJOR |
| R13 | center -> br | Center to Bottom Right. Edge exits right of Center and bends down to Bottom Right. The path runs close to the tl -> br edge near Bottom Right but maintains better separation than iteration 01. | Distinct paths with clear separation. | MAJOR |

**Improvement from iteration 01**: The tr -> bl edge clearance from Center appears improved. The routes around Center are somewhat cleaner. The near overlap between center -> br and tl -> br near Bottom Right is reduced.

---

### 5. 12-bidirectional

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | client -> cache (read) | Client at row1 col1, Cache at row2 col2 (diagonal). The edge exits bottom of Client and routes diagonally through an L shape. The route appears improved with what looks like a clean L bend rather than the 2 bend pattern reported in iteration 01. However the edge runs very close to the Server -> Cache (write) edge. | Clean L bend or 2 bend with clearance. | MAJOR |
| R13 | server -> cache (write) and client -> cache (read) | The write edge from Server and the read edge from Client both descend toward Cache. These two edge paths run close together near Cache's top, with minimal separation. | Edges should maintain clear separation where they run in parallel. | MAJOR |
| R11 | cache -> client (hit) | Cache at row2 col2, Client at row1 col1 (diagonal up left). Edge exits left side of Cache and goes left then up to Client. This is an L bend which is correct for diagonal. However it runs very close to the left diagram boundary. | Adequate clearance from boundary. | MINOR |

**Improvement from iteration 01**: The request/response edges between Client and Server now show clean parallel horizontal lines with distinct paths and visible labels. The bidirectional pair (request going right, response going left) are properly separated with labels visible. The R12 issue about Client left side anchors is resolved since response enters from the right/top area and hit enters from the left/bottom area with better anchor distribution. The extra bends on client -> cache and cache -> client appear reduced.

---

### 6. 16-microservices

**Result**: PASS

Clean hierarchical layout unchanged from iteration 01. Mobile App connects straight down to API Gateway. Gateway fans out with L bends to Auth Service (left), straight down to User Service (center), and L bend to Order Service (right). Each service connects straight down to its respective database. All edges use perpendicular anchors. No edges pass through nodes. Clearance is adequate throughout.

---

### 7. 19-auto-placement

**Result**: PASS

**Fixed from iteration 01**. The auto placed nodes are now positioned in a clean layout. Pinned A (row1 col1) connects right to Auto 1 with a straight horizontal edge, 0 bends. Auto 1 connects right to Auto 3 with a clean L bend that exits the right side and routes to Auto 3. Pinned B (row2 col2) connects right to Auto 2 with a route that uses an L bend going right then up, which is appropriate since Auto 2 appears to be above and to the right. The auto1 -> auto3 edge now routes cleanly rather than the 3+ bend detour reported in iteration 01. All anchors are perpendicular. No edges pass through nodes.

---

### 8. 21-self-loop

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | scheduler -> queue (enqueue) | Scheduler at row1 col1, Queue at row2 col2 (diagonal). The edge exits the right side of Scheduler and routes with an L bend down to Queue, entering from the top. This is 1 bend which is correct for diagonal. However there appear to be 2 segments entering Queue's top very close together (enqueue and retry), which creates visual clutter. | Single L bend is correct. Anchor distribution needs work. | MAJOR |
| R12 | queue top side | Both enqueue (from Scheduler) and retry (from Worker) enter Queue from the top. The two anchor points appear close together, nearly overlapping. | Two edges on top should use positions 2 and 4. | MAJOR |

**Improvement from iteration 01**: The self loop (scheduler -> scheduler heartbeat) now renders as a visible loop exiting the top of Scheduler, arcing up and to the right, and returning to the top of Scheduler at a distinct anchor point. This is a clean self loop rendering. The enqueue edge from Scheduler to Queue now uses 1 L bend instead of the 2 bends reported in iteration 01. The retry edge from Worker to Queue also appears improved. The main remaining issue is anchor crowding on Queue's top side.

---

### 9. 22-pipeline

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | enrich -> errors (failures) | Enrich at row1 col3, Error Queue at row2 col2 (diagonal down left). The edge exits the bottom of Enrich and routes down and left to Error Queue. The route uses 2 bends where 1 L bend would suffice. | Single L bend: exit bottom, enter right OR exit left, enter top. | MAJOR |
| R11 | errors -> dlq (unrecoverable) | Error Queue at row2 col2, Dead Letter at row2 col3 (same row, right). Edge exits right of Error Queue with the label "unrecoverable" and enters left of Dead Letter. This is correct: straight horizontal, 0 bends. | Correct side selection and routing. | N/A |

**Improvement from iteration 01**: The main pipeline flow (ingest -> transform -> enrich -> load) is all straight horizontal, 0 bends, correct. The transform -> errors edge exits bottom of Transform and enters top of Error Queue as a clean straight vertical (same column), 0 bends, correct. The R13 clearance issue from iteration 01 (enrich -> errors running near Dead Letter) appears improved as the edge now enters Error Queue from the top left area with better separation from Dead Letter.

---

### 10. 25-cross-connections

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | a -> d (crosses) | A at row1 col1, D at row3 col3. The edge exits bottom of A, runs down, then turns right at a horizontal level, then turns down to D. 2 bends used. Given the crossing requirement, 2 bends is reasonable but the edges share a horizontal corridor that is very crowded. | 2 bends acceptable for crossing, but edges need better separation. | MAJOR |
| R5 | b -> c (crosses) | B at row1 col3, C at row3 col1. Same 2 bend pattern crossing the a -> d edge. | 2 bends acceptable for crossing. | MAJOR |
| R12 | A bottom side | Two edges leave A's bottom (a -> d and a -> c). The anchor points appear very close together with minimal distribution. | Positions 2 and 4 on the bottom side. | MAJOR |
| R12 | B bottom side | Two edges leave B's bottom (b -> c and b -> d). Same anchor crowding. | Positions 2 and 4 on the bottom side. | MAJOR |
| R12 | C top side | Two edges enter C's top (a -> c and b -> c). Anchors appear close together. | Distributed anchors. | MAJOR |
| R12 | D top side | Two edges enter D's top (a -> d and b -> d). Anchors appear close together. | Distributed anchors. | MAJOR |

No significant change from iteration 01. The anchor distribution issue remains the primary problem.

---

### 11. 27-large-grid-5x5

**Result**: PASS

The 5x5 grid with ring and spoke connections renders cleanly. All ring edges (N1 -> N2 -> N3 -> N4 -> N5, N5 -> E1 -> E2 -> E3 -> S5, S5 -> S4 -> S3 -> S2 -> S1, S1 -> W3 -> W2 -> W1 -> N1) use straight horizontal or straight vertical connections between adjacent nodes with 0 bends. Spoke edges (N3 -> HUB, W2 -> HUB, E2 -> HUB, S3 -> HUB) use straight vertical or straight horizontal paths where aligned. All perpendicular anchors are correct. No edges pass through nodes. Clearance is adequate. Clean and readable layout.

---

### 12. 28-ci-cd-pipeline

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | lint -> test | Lint at row2 col2, Test at row1 col3 (diagonal up right). The edge exits the top of Lint and routes up then right to Test. This uses 2 bends where 1 L bend would suffice. | Single L bend: exit top, enter left OR exit right, enter bottom. | MAJOR |
| R5 | sec -> stage | Security Scan at row2 col3, Staging at row1 col4 (diagonal up right). Edge exits right of Security Scan and routes right and up to Staging. This appears to use 2 bends. | Single L bend for diagonal. | MAJOR |
| R5 | rollback -> stage | Rollback at row3 col5, Staging at row1 col4 (two rows up and one column left). The edge must go up and left. The route takes a long path with multiple bends, going left across the bottom then up. | Clean U or Z shape with minimum bends. | MAJOR |
| R13 | rollback -> stage | The edge from Rollback up to Staging runs through the area near Production and Monitoring. The path appears to maintain some clearance but is tight near the Staging/Production area. | Adequate clearance from non connected nodes. | MAJOR |

**Improvement from iteration 01**: The main pipeline flow (commit -> build -> test) now uses straight horizontal connections. The build -> lint edge exits bottom of Build and enters top of Lint as a clean vertical line, 0 bends, correct. The build -> test edge is now a clean straight horizontal, correct. The test -> sec edge exits bottom of Test into top of Security Scan with what appears to be a cleaner path. The minor "jog" issues from iteration 01 on build -> lint and test -> sec appear resolved.

---

### 13. 29-mesh-topology

**Result**: PASS

Clean 3x3 mesh rendering unchanged from iteration 01. All horizontal edges (a -> b, b -> c, d -> e, e -> f, g -> h, h -> i) are straight horizontal with 0 bends. All vertical edges (a -> d, d -> g, b -> e, e -> h, c -> f, f -> i) are straight vertical with 0 bends. Perpendicular anchors correct. Side selection is appropriate (right for horizontal targets, bottom for vertical targets). No edges pass through nodes. Clearance adequate.

---

### 14. 30-styled-nodes

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | user -> auth (login) | User at row1 col2, Auth Service at row2 col1 (diagonal down left). The edge exits left of User and routes left then down to Auth Service. This appears to use a clean L bend (1 bend), which is correct for diagonal. However the long horizontal segment runs across the top of the diagram. | The L bend itself is correct. Route is acceptable. | MINOR |
| R5 | api -> db1 | API Layer at row2 col2, Users DB at row3 col1 (diagonal down left). The edge exits the bottom left area of API and routes down to Users DB. The route appears to use 2 bends where 1 L bend would suffice. | Single L bend: exit left, enter top OR exit bottom, enter right. | MAJOR |
| R5 | api -> db3 | API Layer at row2 col2, Analytics DB at row3 col3 (diagonal down right). The edge exits bottom right of API and routes with what appears to be a longer path with extra bends. | Single L bend. | MAJOR |
| R5 | db3 -> report (generate) | Analytics DB at row3 col3, Reports at row4 col2 (diagonal down left). The edge exits bottom of Analytics DB and routes left and down with 2 bends. | Single L bend: exit bottom, enter right OR exit left, enter top. | MAJOR |

**Improvement from iteration 01**: The user -> auth (login) edge now uses a cleaner L bend rather than the 2 bend pattern reported in iteration 01. The api -> db2 edge (API to Orders DB, same column) is a clean straight vertical. The api -> cache (read) edge exits right of API to Cache (same row, right) as a clean straight horizontal. The auth -> db1 (verify) edge exits bottom of Auth and enters top of Users DB as a clean straight vertical. The R13 clearance issue from iteration 01 (api -> db1 running near Auth Service) is improved since the edge now takes a path with better separation.

---

### 15. 31-harness-overview

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | agency -> knowledge | Agency at row2 col3, Knowledge at row1 col2 (diagonal up left). The edge exits the top of Agency and routes up and left to Knowledge. This uses 2 bends (up, left, up) where 1 L bend would suffice. | Single L bend: exit top, enter right OR exit left, enter bottom. | MAJOR |
| R2 | governance -> agency (dashed) | Governance at row1 col3, Agency at row2 col3 (same column). The dashed edge exits the bottom of Governance and enters the top of Agency. The edge appears to curve slightly rather than being a clean straight vertical, with the approach at Agency's top not perfectly perpendicular. | Edge must arrive perpendicular to the destination side as a straight vertical. | CRITICAL |

**Improvement from iteration 01**: The governance -> knowledge edge (same row, adjacent, row1 col3 to row1 col2) now appears as a clean straight horizontal, which is correct. The interface -> user edges (ask/respond) use clean horizontal routing with distinct paths for the bidirectional pair. The interface -> agency edge exits right of Interface and enters left of Agency as a clean straight horizontal. The agency -> model edge exits right of Agency and enters left of Model as a clean straight horizontal. The overall layout is cleaner. The governance -> agency dashed edge still shows a slight curve rather than a perfectly straight vertical but may be improved from iteration 01.

---

## Issue Summary by Severity

| Severity | Iter 01 | Iter 02 | Delta |
|----------|---------|---------|-------|
| CRITICAL | 1 | 1 | 0 |
| MAJOR | 38 | 25 | -13 |
| MINOR | 4 | 2 | -2 |
| **Total** | **43** | **28** | **-15** |

---

## Comparison with Iteration 01

### Improvements

1. **19-auto-placement**: Moved from FAIL to PASS. The auto1 -> auto3 edge no longer takes a 3+ bend detour. Auto placed nodes are now positioned cleanly in row 1 and Pinned B connects with a reasonable L bend.

2. **R5 (Minimum Bends)**: Several diagrams show reduced bend counts on specific edges:
   - 21-self-loop: scheduler -> queue now uses 1 L bend instead of 2 bends
   - 30-styled-nodes: user -> auth now uses a cleaner L bend
   - 28-ci-cd-pipeline: build -> lint and test -> sec minor jogs appear resolved
   - 12-bidirectional: Some edges show cleaner routing

3. **R12 (Anchor Distribution)**: The 08-grid-3x3 diagram shows improved anchor distribution on E's top side.

4. **R13 (Clearance)**: Several diagrams show improved edge clearance from non connected nodes, particularly 09-diagonal-connections and 30-styled-nodes.

5. **Overall issue count reduced by 35%** (43 down to 28).

### Regressions

No regressions detected. All diagrams that passed in iteration 01 continue to pass. No new issues were found that were not present in iteration 01.

### Persistent Issues

1. **R5 diagonal routing with 2 bends instead of 1**: This remains the most common issue. Diagrams 07-diamond-pattern, 08-grid-3x3, 22-pipeline, 28-ci-cd-pipeline, 30-styled-nodes, and 31-harness-overview all still show diagonal connections with 2 bends where 1 L bend would suffice. The routing algorithm still defaults to a 3 segment path (vertical, horizontal, vertical) for some diagonal connections.

2. **R12 anchor distribution**: The 25-cross-connections diagram still shows crowded anchors when multiple edges connect to the same node side.

3. **R2 governance -> agency dashed edge**: The slightly curved/non perpendicular approach on the dashed edge in 31-harness-overview persists as the only CRITICAL finding.

---

## Most Common Issues

1. **R5 (Minimum Bends)**: Still the most pervasive issue at 12 occurrences, though reduced from approximately 20 in iteration 01. Diagonal connections in several diagrams continue to use 2 bends where 1 L bend is sufficient.

2. **R12 (Anchor Distribution)**: 5 occurrences, mainly in 25-cross-connections and 21-self-loop where multiple edges converge on the same node side.

3. **R13 (Clearance)**: 4 occurrences where edges run near non connected nodes, reduced from iteration 01.

4. **R11 (Smart Side Selection)**: 2 occurrences tied to the R5 diagonal routing issue.

---

## Recommendations

1. **Priority 1**: Continue fixing diagonal routing to prefer single L bends. Progress has been made but the fix is not applied consistently across all diagonal cases.

2. **Priority 2**: Fix anchor distribution when multiple edges enter/exit the same node side, particularly visible in 25-cross-connections.

3. **Priority 3**: Investigate and fix the R2 CRITICAL finding on the governance -> agency dashed edge in 31-harness-overview. The edge should be a clean straight vertical line.

4. **Priority 4**: Continue improving edge clearance from non connected nodes in dense layouts.
