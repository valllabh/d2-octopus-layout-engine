# Iteration 04: Edge Routing Validation Report

**Date**: 2026-03-17
**Tester**: QA (automated visual inspection)
**Focus Rules**: R1 (no edge through node), R2 (perpendicular anchors), R5 (minimum bends), R11 (smart side selection), R12 (anchor distribution), R13 (clearance from nodes)

---

## Summary

| # | Diagram | Iter 03 | Iter 04 | Delta |
|---|---------|---------|---------|-------|
| 1 | 05-fan-out | PASS | PASS | same |
| 2 | 07-diamond-pattern | FAIL | FAIL | same |
| 3 | 08-grid-3x3 | FAIL | FAIL | same |
| 4 | 09-diagonal-connections | FAIL | FAIL | same |
| 5 | 12-bidirectional | FAIL | FAIL | same |
| 6 | 16-microservices | PASS | PASS | same |
| 7 | 19-auto-placement | PASS | PASS | same |
| 8 | 21-self-loop | FAIL | FAIL | same |
| 9 | 22-pipeline | FAIL | FAIL | same |
| 10 | 25-cross-connections | FAIL | FAIL | same |
| 11 | 27-large-grid-5x5 | PASS | PASS | same |
| 12 | 28-ci-cd-pipeline | FAIL | FAIL | same |
| 13 | 29-mesh-topology | PASS | PASS | same |
| 14 | 30-styled-nodes | FAIL | FAIL | same |
| 15 | 31-harness-overview | FAIL | FAIL | same |

**Pass**: 5 / 15
**Fail**: 10 / 15

---

## Detailed Findings

### 1. 05-fan-out

**Result**: PASS

Load Balancer (row1 col2) fans out to Server 1 (row2 col1), Server 2 (row2 col2), Server 3 (row2 col3). The center edge (lb -> s2) exits bottom of Load Balancer and enters top of Server 2 as a straight vertical line, 0 bends. The left edge (lb -> s1) exits the left portion of Load Balancer bottom, bends left and down to Server 1 top, 1 bend (L shape, correct for diagonal). The right edge (lb -> s3) mirrors with 1 bend to the right. Anchor distribution on Load Balancer bottom is well spread across three distinct positions. All arrowheads at correct ends. Perpendicular contact at all anchors.

- R1: PASS
- R2: PASS
- R5: PASS
- R11: PASS
- R12: PASS
- R13: PASS

---

### 2. 07-diamond-pattern

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | path-a -> merge | Path A is at row2 col1, Merge is at row3 col2 (diagonal down right). The edge exits the bottom of Path A, goes down, bends right, then bends down again into Merge top. This is 2 direction changes (3 segments: vertical, horizontal, vertical). No obstacle exists between Path A and Merge on an L shaped path. Only 1 bend (L shape) is needed. | Single L bend: exit bottom of Path A, bend right, enter left side of Merge. OR exit right of Path A, bend down, enter top of Merge. | MAJOR |
| R5 | path-b -> merge | Path B is at row2 col3, Merge is at row3 col2 (diagonal down left). Same 2 direction change pattern mirrored. Edge exits bottom, goes down, turns left, turns down into Merge top. 2 bends where 1 suffices with no obstacle. | Single L bend: exit bottom of Path B, bend left, enter right side of Merge. OR exit left of Path B, bend down, enter top of Merge. | MAJOR |
| R11 | path-a -> merge | Both path-a and path-b enter Merge from the top. For path-a (diagonally up left from Merge), entering the left side of Merge with a single L bend would be cleaner. The current routing forces both edges into the top side, requiring extra bends. | Exit bottom of Path A, enter left of Merge with 1 bend. | MAJOR |
| R11 | path-b -> merge | Mirror of path-a issue. Enter right side of Merge instead of top. | Exit bottom of Path B, enter right of Merge with 1 bend. | MAJOR |

- R1: PASS
- R2: PASS
- R5: FAIL (2 edges with extra bends)
- R11: FAIL (2 edges with suboptimal side entry)
- R12: PASS (two edges enter Merge top with adequate separation)
- R13: PASS

No change from iteration 03.

---

### 3. 08-grid-3x3

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | a -> e | A at row1 col1, E at row2 col2 (diagonal). Edge exits bottom of A, routes down, bends right, bends down into E top. This is 2 direction changes. B (row1 col2) is directly above E but is NOT on the L path from A to E. The L path from A bottom would go down then right, and B is above that path at row1. No obstacle on the L path. Only 1 bend needed. | Single L bend: exit bottom of A, go down, bend right, enter left side of E. Or exit right of A, go right, bend down, enter top of E. | MAJOR |
| R5 | c -> e | C at row1 col3, E at row2 col2 (diagonal). Same 2 direction change pattern mirrored (down, left, down). F (row2 col3) is to the right of E but not on the L path from C. No obstacle justifies 2 bends. | Single L bend. | MAJOR |
| R5 | e -> g | E at row2 col2, G at row3 col1 (diagonal down left). Edge exits bottom of E, routes with 2 bends (down, left, down). D (row2 col1) is directly left of E and above G, but is not on the L path from E bottom going down then left. No obstacle on the L path. 1 bend sufficient. | Single L bend. | MAJOR |
| R5 | e -> i | E at row2 col2, I at row3 col3 (diagonal down right). Edge exits bottom of E, routes with 2 bends (down, right, down). F (row2 col3) is to the right of E and above I, but not on the L path from E going down. 1 bend sufficient. | Single L bend. | MAJOR |
| R12 | E top side | Three edges (from A, B, C) and two edges (from D left, F right) converge on E. The top side receives three incoming arrows (A, B, C). The arrow tips appear tightly packed on E's top edge, with minimal visual spacing. | Three edges on top should use positions 2, 3, and 4 with clear visual spacing. | MAJOR |

- R1: PASS (no edges cross through B, D, F, H nodes)
- R2: PASS
- R5: FAIL (4 diagonal edges use 2 bends instead of 1)
- R11: PASS (exit sides are reasonable given the routing)
- R12: FAIL (E top side anchors are crowded)
- R13: PASS (edges from A and C route around B with visible clearance)

No change from iteration 03.

---

### 4. 09-diagonal-connections

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R13 | tl -> br | The long diagonal from Top Left (row1 col1) to Bottom Right (row3 col3) exits the left side of Top Left, routes down the left side of the diagram, then bends right at the bottom passing close to Bottom Left. The horizontal segment runs near the bottom of the diagram, passing close to the bottom edge of the Bottom Left node. Clearance from Bottom Left appears tight. | Adequate clearance from Bottom Left node. | MAJOR |
| R13 | tr -> bl | The long diagonal from Top Right (row1 col3) to Bottom Left (row3 col1) exits the right side of Top Right, routes down the right side past Center, then bends left at the bottom. The horizontal segment passes near Bottom Right. Clearance appears tight near Bottom Right. | Clearance from non connected nodes should be at least one grid unit. | MAJOR |

- R1: PASS
- R2: PASS
- R5: PASS (tl -> br: 2 bends justified by Center obstacle on the diagonal path between row1 col1 and row3 col3; tr -> bl: 2 bends justified by Center obstacle; tl -> center: 1 L bend correct for adjacent diagonal; center -> br: exits right and bends down, 1 L bend correct)
- R11: PASS (side selections are reasonable for obstacle avoidance routing)
- R12: PASS (Bottom Right receives edges on different sides with adequate distribution)
- R13: FAIL (tight clearance near Bottom Left and Bottom Right on the long diagonal routes)

No change from iteration 03.

---

### 5. 12-bidirectional

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R12 | cache top side | Both the client -> cache (read) edge and the server -> cache (write) edge enter Cache from the top. Looking at the rendering, the two anchor points where read and write arrive on Cache's top side are close together. With only 2 edges on this side, positions 2 and 4 should be used for clear separation. The anchors appear closer to center than they should be. | Two edges on top should use positions 2 and 4 (roughly 1/3 and 2/3 of the side width) for clear visual separation. | MAJOR |
| R13 | read and write edges near Cache | The client -> cache (read) and server -> cache (write) edges both descend toward Cache top. Their vertical segments run in close proximity above Cache, with minimal visual separation between the two incoming paths as they converge. | Edges should maintain clear separation where running in parallel near Cache. | MAJOR |

- R1: PASS
- R2: PASS
- R5: PASS (client -> server request: straight horizontal 0 bends; server -> client response: straight horizontal below request, 0 bends; client -> cache read: exits bottom of Client, L bend down to Cache top, 1 bend correct for diagonal; cache -> client hit: exits left of Cache, goes left and up to Client, 1 L bend correct; server -> cache write: exits bottom area of Server, L bend down to Cache top, 1 bend correct for diagonal)
- R11: PASS
- R12: FAIL (Cache top side anchors are crowded between read and write edges)
- R13: FAIL (read and write edges run too close together near Cache top)

No change from iteration 03.

---

### 6. 16-microservices

**Result**: PASS

Clean hierarchical layout. Mobile App (row1 col2) connects straight down to API Gateway (row2 col2), 0 bends. Gateway fans out: left L bend to Auth Service (row3 col1, 1 bend), straight down to User Service (row3 col2, 0 bends), right L bend to Order Service (row3 col3, 1 bend). Each service connects straight down to its respective database (row4), 0 bends each. All edges use perpendicular anchors. No edges pass through nodes. Clearance adequate. Anchor distribution on Gateway bottom side is well spread across three positions.

- R1: PASS
- R2: PASS
- R5: PASS
- R11: PASS
- R12: PASS
- R13: PASS

---

### 7. 19-auto-placement

**Result**: PASS

Pinned A (row1 col1) connects right to Auto 1 (row1 col2) with a straight horizontal edge, 0 bends. Auto 1 connects to Auto 3 (row1 col4) with an L bend routing down and right, going under Auto 2, then up to Auto 3. Pinned B (row2 col2) connects right to Auto 2 with an L bend. The auto1 -> auto3 edge routes cleanly with perpendicular anchors. No edges pass through nodes. Clearance adequate.

- R1: PASS
- R2: PASS
- R5: PASS
- R11: PASS
- R12: PASS
- R13: PASS

---

### 8. 21-self-loop

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R12 | queue top side | Both the scheduler -> queue (enqueue) and worker -> queue (retry) edges enter Queue from the top. Looking at the rendering, the two arriving arrow tips on Queue's top side are close together. The labels "enqueue" and "retry" are visible above Queue. The two anchor points appear to be near center positions rather than spread at positions 2 and 4. | Two edges on top should use positions 2 and 4 with clear visual separation. | MAJOR |

- R1: PASS
- R2: PASS
- R5: PASS (scheduler -> queue enqueue: exits right of Scheduler, routes right and bends down to Queue top, 1 bend correct for diagonal; queue -> worker dequeue: exits right of Queue, routes right and bends up to Worker, 1 L bend correct; worker -> queue retry: exits left of Worker, routes left and bends down to Queue top, 1 L bend correct; scheduler -> scheduler heartbeat: self loop exits top, arcs up and returns to top at a distinct anchor, visible and clean)
- R11: PASS
- R12: FAIL (Queue top side has two incoming edges with close anchor points)
- R13: PASS
- R23: PASS (self loop is visible, exits and returns at distinct anchor points on top of Scheduler)

No change from iteration 03.

---

### 9. 22-pipeline

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | enrich -> errors (failures) | Enrich at row1 col3, Error Queue at row2 col2 (diagonal down left). The edge exits the bottom left area of Enrich and routes down and left to Error Queue. The route uses 2 direction changes (down, left, down) where 1 L bend would suffice. Transform (row1 col2) is directly above Error Queue and to the left of Enrich, but is not on the L path from Enrich going down then left. No obstacle justifies the extra bend. | Single L bend: exit bottom of Enrich, bend left, enter right side of Error Queue. OR exit left of Enrich, bend down, enter top of Error Queue. | MAJOR |

- R1: PASS
- R2: PASS
- R5: FAIL (enrich -> errors uses 2 bends where 1 suffices; all other edges correct: main pipeline straight horizontal 0 bends, transform -> errors straight vertical 0 bends, errors -> dlq straight horizontal 0 bends)
- R11: PASS
- R12: PASS (Error Queue top side receives transform -> errors and enrich -> errors at distinct anchor points)
- R13: PASS

No change from iteration 03.

---

### 10. 25-cross-connections

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R6 | a -> d and b -> c | Looking at the rendered diagram, the a -> d (crosses) and b -> c (crosses) edges route in a way that creates an unexpected pattern. Both edges exit from the bottom of their source nodes and run vertically downward. The edge a -> d goes down from A to a horizontal level, runs right across to D's column, then goes down to D. The edge b -> c mirrors this. The horizontal segments of these two crossing edges overlap or run very close to each other at the same Y coordinate, labeled "crosses". The crossing occurs in open space which is correct, but the crossing edges run nearly on top of each other for the horizontal portion rather than having distinct separate horizontal segments. | The horizontal portions of the crossing edges should visually separate so both paths are individually traceable. | MAJOR |
| R12 | A bottom side | Two edges leave A's bottom side (a -> c straight down, a -> d going down then right). The anchor points on A's bottom appear very close together with minimal distribution. | Two edges should use positions 2 and 4 on the bottom side. | MAJOR |
| R12 | B bottom side | Two edges leave B's bottom side (b -> d straight down, b -> c going down then left). Same close anchor point issue. | Two edges should use positions 2 and 4. | MAJOR |
| R12 | C top side | Two edges enter C's top (a -> c straight down and b -> c from the right). Anchors appear close together. | Distributed anchors at positions 2 and 4. | MAJOR |
| R12 | D top side | Two edges enter D's top (b -> d straight down and a -> d from the left). Anchors appear close together. | Distributed anchors at positions 2 and 4. | MAJOR |

- R1: PASS (no edges pass through nodes; crossings occur in open space)
- R2: PASS
- R5: PASS (a -> c: straight vertical 0 bends; b -> d: straight vertical 0 bends; a -> d and b -> c: 2 bends each, acceptable for orthogonal cross routing)
- R6: FAIL (crossing edges share nearly the same horizontal corridor, making them hard to distinguish)
- R11: PASS
- R12: FAIL (all four nodes show crowded anchors where two edges share the same side)
- R13: PASS

Same as iteration 03. The R6 finding is newly noted in this iteration: the horizontal segments of the crossing edges are so close they are hard to distinguish visually.

---

### 11. 27-large-grid-5x5

**Result**: PASS

The 5x5 grid with ring and spoke connections renders cleanly. All ring edges between adjacent nodes use straight horizontal or straight vertical connections with 0 bends. Top ring (N1 -> N2 -> N3 -> N4 -> N5) straight horizontal. Right side (N5 -> E1 -> E2 -> E3 -> S5) straight vertical. Bottom ring (S5 -> S4 -> S3 -> S2 -> S1) straight horizontal going left. Left side (S1 -> W3 -> W2 -> W1 -> N1) straight vertical going up. Spoke edges to HUB (N3 -> HUB vertical, W2 -> HUB horizontal, E2 -> HUB horizontal, S3 -> HUB vertical) all use straight paths where aligned. All perpendicular anchors correct. No edges pass through nodes. Clearance adequate. HUB receives four spoke edges on four different sides with excellent distribution.

- R1: PASS
- R2: PASS
- R5: PASS
- R11: PASS
- R12: PASS
- R13: PASS

---

### 12. 28-ci-cd-pipeline

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | lint -> test | Lint at row2 col2, Test at row1 col3 (diagonal up right). The edge exits the top right area of Lint and routes up with 2 direction changes (up, right, up or a curving path). No obstacle exists between Lint and Test on an L path. 1 bend suffices. | Single L bend: exit top of Lint, bend right, enter bottom of Test. OR exit right of Lint, bend up, enter bottom of Test. | MAJOR |
| R5 | sec -> stage | Security Scan at row2 col3, Staging at row1 col4 (diagonal up right). Edge exits from Security Scan and routes up and right to Staging. The route appears to use 2 direction changes. No obstacle between them on the L path. | Single L bend for diagonal. | MAJOR |
| R5 | rollback -> stage | Rollback at row3 col5, Staging at row1 col4 (two rows up, one column left). The edge exits the left side of Rollback and routes up and left with multiple bends to reach Staging. Given that Production (row1 col5) and Monitoring (row2 col5) are in the same column as Rollback but above it, the path must route around them. However the route appears to take a longer path than necessary with extra bends. The leftward route goes far left before turning up. 2 bends should suffice for routing around Monitoring and Production. | Clean path with 2 bends routing around Monitoring: exit top of Rollback, go up past Monitoring on the left, bend left, enter bottom or right of Staging. | MAJOR |

- R1: PASS
- R2: PASS
- R5: FAIL (3 edges with extra bends on diagonal routes)
- R11: PASS
- R12: PASS
- R13: PASS

No change from iteration 03.

---

### 13. 29-mesh-topology

**Result**: PASS

Clean 3x3 mesh rendering. All horizontal edges (a -> b, b -> c, d -> e, e -> f, g -> h, h -> i) are straight horizontal with 0 bends. All vertical edges (a -> d, d -> g, b -> e, e -> h, c -> f, f -> i) are straight vertical with 0 bends. Perpendicular anchors correct throughout. Side selection is appropriate (right exit for horizontal targets, bottom exit for vertical targets). No edges pass through nodes. No edges too close to unconnected nodes. All arrowheads at correct ends.

- R1: PASS
- R2: PASS
- R5: PASS
- R11: PASS
- R12: PASS
- R13: PASS

---

### 14. 30-styled-nodes

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R12 | API Layer bottom side | API Layer (row2 col2) has 3 edges exiting its bottom side: api -> db1 (Users DB, row3 col1), api -> db2 (Orders DB, row3 col2), and api -> db3 (Analytics DB, row3 col3). The three anchor points on API Layer's bottom side appear somewhat crowded. With 3 edges, positions 2, 3, and 4 should be used. The rendering shows the anchors close together near center rather than well distributed. | Three edges on bottom should use positions 2, 3, and 4 with clear visual spacing across the bottom side. | MAJOR |

- R1: PASS
- R2: PASS
- R5: PASS (user -> auth login: exits left of User, routes left down to Auth Service, 1 bend correct for diagonal; user -> api requests: exits bottom of User, straight down to API Layer, 0 bends; api -> cache read: exits right, straight horizontal to Cache, 0 bends; api -> db2: exits bottom, straight vertical to Orders DB, 0 bends; auth -> db1 verify: exits bottom of Auth Service, straight vertical to Users DB, 0 bends; api -> db1: 2 bends, justified by Auth Service at row2 col1 as obstacle; api -> db3: 2 bends, justified by Cache at row2 col3 as obstacle; db3 -> report generate: 2 bends, justified by Orders DB at row3 col2 as obstacle)
- R11: PASS
- R12: FAIL (API Layer bottom side anchors crowded for 3 outgoing edges)
- R13: PASS

No change from iteration 03.

---

### 15. 31-harness-overview

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R2 | governance -> agency (dashed) | Governance at row1 col3, Agency at row2 col3 (same column, directly below). The dashed edge exits the bottom of Governance and enters the top of Agency. Looking at the rendered PNG, this dashed edge shows a visible curve or S shape rather than a straight vertical line. The edge appears to arc slightly as it descends, particularly near the top of Agency where it approaches at a slight angle rather than perfectly perpendicular. Per the D2 rendering notes, slight visual irregularities on dashed edges should be tolerated if the route is mathematically straight (same X coordinate). However, this edge shows a clearly visible lateral deviation, not just a dashed rendering artifact. The edge visibly curves to the right before reaching Agency, creating a non perpendicular approach. | Edge must be a clean straight vertical line with perpendicular contact at both top and bottom anchors. A same column connection should have 0 bends and be perfectly vertical. | CRITICAL |

- R1: PASS
- R2: FAIL (governance -> agency dashed edge shows a visible curve rather than straight vertical)
- R5: PASS (governance -> knowledge: straight horizontal left, 0 bends; interface -> agency: straight horizontal right, 0 bends; agency -> model: straight horizontal right, 0 bends; user -> interface ask: straight horizontal right, 0 bends; interface -> user respond: straight horizontal left, 0 bends; agency -> knowledge: exits top of Agency, routes up and left to Knowledge with 2 bends, justified by Governance at row1 col3 sitting directly above Agency as an obstacle; governance -> agency: should be 0 bends straight vertical)
- R11: PASS
- R12: PASS (Agency top side has governance -> agency and agency -> knowledge at distinct enough positions with adequate spacing)
- R13: PASS

No change from iteration 03.

---

## Issue Summary by Severity

| Severity | Iter 03 | Iter 04 | Delta |
|----------|---------|---------|-------|
| CRITICAL | 1 | 1 | 0 |
| MAJOR | 16 | 17 | +1 |
| MINOR | 0 | 0 | 0 |
| **Total** | **17** | **18** | **+1** |

---

## Comparison with Iteration 03

### Improvements

No rendering improvements detected. All diagrams produce the same visual output as iteration 03.

### Regressions

No regressions detected. All diagrams that passed in iteration 03 continue to pass.

### New Findings

1. **R6 on 25-cross-connections**: The crossing edges a -> d and b -> c share nearly the same horizontal corridor. Their horizontal segments are so close together that the two paths are hard to distinguish visually. This was not explicitly called out in iteration 03 but is present in the rendering. This adds 1 MAJOR finding to the count.

### Persistent Issues

1. **R5 diagonal routing with 2 bends instead of 1 (no obstacles)**: Remains the most common issue. Diagrams 07-diamond-pattern (path-a -> merge, path-b -> merge), 08-grid-3x3 (a -> e, c -> e, e -> g, e -> i), 22-pipeline (enrich -> errors), and 28-ci-cd-pipeline (lint -> test, sec -> stage, rollback -> stage) still show diagonal connections with 2 direction changes where 1 L bend would suffice and no obstacle justifies the extra bend.

2. **R12 anchor distribution**: Diagrams 08-grid-3x3 (E top side), 12-bidirectional (Cache top side), 21-self-loop (Queue top side), 25-cross-connections (all four nodes), and 30-styled-nodes (API Layer bottom side) show crowded anchor points when multiple edges connect to the same side of a node.

3. **R2 governance -> agency dashed edge**: The curved/non perpendicular approach on the dashed edge in 31-harness-overview persists as the only CRITICAL finding. The edge is between same column nodes and should be a clean straight vertical line.

4. **R13 clearance**: In 09-diagonal-connections, the long diagonal routes pass too close to corner nodes (Bottom Left and Bottom Right).

5. **R6 crossing edge distinction**: In 25-cross-connections, the crossing edges share a nearly identical horizontal corridor making them hard to trace individually.

---

## Most Common Issues

1. **R5 (Minimum Bends)**: 8 occurrences across 4 diagrams where diagonal connections use 2 bends with no obstacle justification.

2. **R12 (Anchor Distribution)**: 7 occurrences across 5 diagrams where multiple edges converge on the same node side with insufficient spacing between anchor points.

3. **R13 (Edge Clearance)**: 2 occurrences in 09-diagonal-connections and 12-bidirectional where edges pass too close to non connected nodes.

4. **R2 (Perpendicular Anchors)**: 1 CRITICAL occurrence on the governance -> agency dashed edge in 31-harness-overview.

5. **R6 (Edge Crossings)**: 1 occurrence in 25-cross-connections where crossing edges share nearly the same horizontal corridor.

---

## Recommendations

1. **Priority 1 (CRITICAL)**: Fix the R2 finding on governance -> agency dashed edge in 31-harness-overview. The edge between same column nodes must be a clean straight vertical line, not a curve.

2. **Priority 2**: Fix diagonal routing to prefer single L bends where no obstacle exists between source and destination. The 2 bend pattern (vertical, horizontal, vertical) on clear diagonal paths is the most common remaining defect. Affected diagrams: 07-diamond-pattern, 08-grid-3x3, 22-pipeline, 28-ci-cd-pipeline.

3. **Priority 3**: Improve anchor distribution when multiple edges enter/exit the same node side. Enforce the position spread rules (2 edges: positions 2 and 4; 3 edges: positions 2, 3, 4). Affected diagrams: 08-grid-3x3, 12-bidirectional, 21-self-loop, 25-cross-connections, 30-styled-nodes.

4. **Priority 4**: Improve edge clearance from non connected nodes on long diagonal routes. Affected diagram: 09-diagonal-connections.

5. **Priority 5**: Improve crossing edge visual separation in 25-cross-connections so both paths are individually traceable.
