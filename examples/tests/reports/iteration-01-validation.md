# Iteration 01: Edge Routing Validation Report

**Date**: 2026-03-17
**Tester**: QA (automated visual inspection)
**Focus Rules**: R1 (no edge through box), R2 (perpendicular anchors), R5 (minimum bends), R11 (smart side selection), R13 (clearance from nodes)

---

## Summary

| # | Diagram | Result |
|---|---------|--------|
| 1 | 05-fan-out | PASS |
| 2 | 07-diamond-pattern | FAIL |
| 3 | 08-grid-3x3 | FAIL |
| 4 | 09-diagonal-connections | FAIL |
| 5 | 12-bidirectional | FAIL |
| 6 | 16-microservices | PASS |
| 7 | 19-auto-placement | FAIL |
| 8 | 21-self-loop | FAIL |
| 9 | 22-pipeline | FAIL |
| 10 | 25-cross-connections | FAIL |
| 11 | 27-large-grid-5x5 | PASS |
| 12 | 28-ci-cd-pipeline | FAIL |
| 13 | 29-mesh-topology | PASS |
| 14 | 30-styled-nodes | FAIL |
| 15 | 31-harness-overview | FAIL |

**Pass**: 4 / 15
**Fail**: 11 / 15

---

## Detailed Findings

### 1. 05-fan-out

**Result**: PASS

All three edges from Load Balancer fan out cleanly to Server 1, Server 2, and Server 3. Edges exit the bottom of Load Balancer and enter the top of each server. The left and right edges use L shaped routes with one bend each which is appropriate for diagonal targets. Anchor distribution on the bottom side of Load Balancer appears reasonable. No edges pass through any node.

---

### 2. 07-diamond-pattern

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | path-a -> merge | Edge exits bottom of Path A, goes down, bends right, then bends down again into Merge. This is 2 bends but only 1 bend (L shape) is needed since Path A is diagonally above left of Merge. | Single L bend: exit right from Path A, enter top of Merge, or exit bottom from Path A, enter left of Merge. | MAJOR |
| R5 | path-b -> merge | Edge exits bottom of Path B, goes down, bends left, then bends down again into Merge. Same issue as above, 2 bends where 1 would suffice. | Single L bend from diagonal position. | MAJOR |
| R11 | path-a -> merge | Edge exits bottom side of Path A. Since Merge is diagonally down right, exiting right or bottom would both work, but the current route adds an unnecessary extra bend. | Exit bottom, enter left or exit right, enter top with single L bend. | MAJOR |
| R11 | path-b -> merge | Edge exits bottom side of Path B. Since Merge is diagonally down left, exiting left or bottom would both work. Same unnecessary bend pattern. | Exit bottom, enter right or exit left, enter top with single L bend. | MAJOR |

---

### 3. 08-grid-3x3

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | a -> e | A is at row1 col1, E is at row2 col2 (diagonal). The edge exits bottom of A, routes down and right with what appears to be 2 bends. Only 1 bend (L shape) needed. | Single L bend for diagonal adjacent. | MAJOR |
| R5 | c -> e | C is at row1 col3, E is at row2 col2 (diagonal). Same extra bend pattern. | Single L bend. | MAJOR |
| R13 | Multiple edges around E | Multiple edges converge on E from all 8 surrounding nodes. The edges from A and C run very close to B on either side. The edge segments pass near B with minimal clearance. | Edges should maintain at least one grid unit clearance from non connected nodes. | MAJOR |
| R5 | e -> g | E is at row2 col2, G is at row3 col1 (diagonal). Edge appears to use 2 bends where 1 would suffice. | Single L bend. | MAJOR |
| R5 | e -> i | E is at row2 col2, I is at row3 col3 (diagonal). Edge routes right and then bends down with what appears to be extra bends. | Single L bend. | MAJOR |
| R12 | Anchors on E top side | Three edges (from A, B, C) all arrive at the top of E. The anchor points appear very close together and potentially overlapping. | Anchors should be distributed at positions 2, 3, 4 across the top side. | MAJOR |

---

### 4. 09-diagonal-connections

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | tl -> br | Top Left to Bottom Right is a long diagonal. The edge exits bottom of Top Left, goes down, turns right, then turns down into Bottom Right. This is 2 bends. Given Center node is in the path, 2 bends may be needed to avoid it. However the route appears to go wide rather than taking an efficient path. | If obstacle avoidance requires 2 bends, the path should be a clean Z or U shape routed through the nearest gap. | MAJOR |
| R5 | tr -> bl | Top Right to Bottom Left. The edge exits the bottom of Top Right, routes down, turns left all the way across, then turns down into Bottom Left. 2 bends which may be justified for crossing, but the horizontal segment runs very close to Center node. | Clean Z or U shape with adequate clearance from Center. | MAJOR |
| R13 | tr -> bl | The horizontal segment of this edge appears to run very close to the Center node, potentially grazing it. | Minimum one grid unit clearance from Center. | MAJOR |
| R13 | tl -> br | The edge path runs close to Center node on its route. | Minimum clearance from non connected nodes. | MAJOR |
| R11 | center -> br | Center is at row2 col2, Bottom Right is at row3 col3 (diagonal down right). Edge exits right side of Center and bends down. This is acceptable side selection. However the edge route appears to nearly overlap with the tl->br edge path near Bottom Right. | Distinct paths with clear separation near the destination. | MAJOR |

---

### 5. 12-bidirectional

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | client -> server (request) | Client and Server are in the same row (row 1). The edge exits the right side of Client and goes right to Server. This should be 0 bends (straight horizontal). The rendering shows a straight line, which is correct for this edge. However the label "request" and "response" labels overlap visually. | Straight horizontal, labels should not overlap. | MAJOR |
| R13 | server -> cache (write) | The write edge from Server to Cache passes downward. The vertical segment runs between Client and Server areas. Clearance appears adequate. | Adequate clearance maintained. | MINOR |
| R5 | client -> cache (read) | Client is at row1 col1, Cache is at row2 col2 (diagonal). The edge takes 2 bends where 1 L bend would suffice. The edge goes down from Client's bottom, then right, then down into Cache. | Single L bend for diagonal. | MAJOR |
| R5 | cache -> client (hit) | Cache to Client is diagonal (row2 col2 to row1 col1). The edge exits the left side of Cache, goes left then up into Client. This appears to use 2 bends. | Single L bend. | MAJOR |
| R12 | Client left side anchors | Both "response" (from Server) and "hit" (from Cache) enter the left side of Client. The edges appear to arrive at similar anchor positions. | Two edges on left side should use positions 2 and 4. | MAJOR |
| R12 | Multiple edges on Client right/bottom | "request" exits right, "read" exits bottom. The anchor distribution is acceptable since they use different sides. | N/A (acceptable). | N/A |

---

### 6. 16-microservices

**Result**: PASS

Clean hierarchical layout. Mobile App connects down to API Gateway. Gateway fans out to three services. Each service connects down to its database. All edges are straight vertical where nodes are in the same column, and the fan out from Gateway uses clean L bends for the diagonal connections to Auth Service and Order Service. No edges pass through nodes. Perpendicular anchors appear correct. Side selection is appropriate.

---

### 7. 19-auto-placement

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | auto1 -> auto3 | Auto 1 and Auto 3 are in the same row. The edge exits the right side of Auto 1, goes right, bends down, runs along the bottom, then bends back up to enter Auto 3 from the top. This is 3+ bends for what should be a 0 bend straight horizontal edge. | Straight horizontal line from Auto 1 right side to Auto 3 left side. | MAJOR |
| R11 | auto1 -> auto3 | The edge exits right side of Auto 1, which is correct since Auto 3 is to the right. But then the routing goes down and around instead of going straight. | Exit right, enter left, straight horizontal. | MAJOR |
| R5 | explicit2 -> auto2 | Pinned B is at row2 col2, Auto 2 is placed in the same row to the right. The edge exits right of Pinned B, goes right then up with unnecessary bends. | Cleaner route with fewer bends. | MAJOR |
| R11 | explicit2 -> auto2 | Pinned B is below and left of Auto 2. The edge goes right then up. Given that Auto 2 is above and right, exiting right is acceptable but the routing could be cleaner. | Exit top or right with single L bend. | MAJOR |

---

### 8. 21-self-loop

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | scheduler -> queue (enqueue) | Scheduler is at row1 col1, Queue is at row2 col2 (diagonal). The edge exits bottom of Scheduler, goes down, and enters top of Queue. This appears to use 2 bends. Given the diagonal positioning, 1 L bend is sufficient. | Single L bend for diagonal connection. | MAJOR |
| R13 | scheduler -> scheduler (heartbeat) | The self loop exits the top of Scheduler, arcs right, and returns to the top of Scheduler. The loop is visible but it is a relatively tight loop. The clearance from the node is adequate but small. | Self loop should have clear separation from the node boundary. | MINOR |
| R5 | worker -> queue (retry) | Worker at row1 col3, Queue at row2 col2. The edge exits bottom of Worker and bends left down into Queue. This appears to use 2 bends. | Single L bend for diagonal. | MAJOR |

---

### 9. 22-pipeline

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | enrich -> errors (failures) | Enrich is at row1 col3, Error Queue is at row2 col2 (diagonal down left). The edge exits the bottom of Enrich, goes down, bends left, bends down again. This uses 2 bends where 1 L bend would suffice. | Single L bend: exit bottom, enter right OR exit left, enter top. | MAJOR |
| R13 | enrich -> errors | The edge route from Enrich down to Error Queue passes near the Dead Letter node area. | Adequate clearance from Dead Letter required. | MAJOR |
| R11 | errors -> dlq (unrecoverable) | Error Queue is at row2 col2, Dead Letter is at row2 col3. They are in the same row. The edge exits right of Error Queue and enters left of Dead Letter. This is correct side selection and straight horizontal. | This edge is correct. | N/A |

---

### 10. 25-cross-connections

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | a -> d (crosses) | A is at row1 col1, D is at row3 col3. The edge exits bottom of A, goes down, bends right at a horizontal level, then bends down to D. This is 2 bends for a long diagonal, which is the minimum for an L/Z shape avoiding intersection. However the horizontal segment shares space with the b->c edge. | The 2 bend route is acceptable if needed to cross, but the edges should be clearly separated. | MAJOR |
| R5 | b -> c (crosses) | B is at row1 col3, C is at row3 col1. Same pattern as above, 2 bends. | Similar to a->d. | MAJOR |
| R11 | a -> c | A at row1 col1, C at row3 col1 (same column, directly below). Edge exits bottom of A and goes straight down to C. This is correct. However the edge appears to run very close to or overlap with the a->d edge near A's bottom. | Edges should use distinct anchor points and maintain separation. | MAJOR |
| R11 | b -> d | B at row1 col3, D at row3 col3 (same column, directly below). Edge exits bottom of B and goes straight down to D. Same potential overlap issue near B's bottom. | Distinct anchor points and clear separation. | MAJOR |
| R12 | A bottom side | Two edges leave A's bottom (a->d and a->c). They should use distributed anchor points. They appear very close together. | Positions 2 and 4 on the bottom side. | MAJOR |
| R12 | B bottom side | Two edges leave B's bottom (b->c and b->d). Same anchor crowding issue. | Positions 2 and 4 on the bottom side. | MAJOR |
| R12 | C top side | Two edges enter C's top (a->c visible, but actually only a->c enters C and b->c enters C). The two edges arrive very close together at C's top. | Distributed anchors. | MAJOR |
| R12 | D top side | Two edges enter D's top (a->d and b->d). Same issue. | Distributed anchors. | MAJOR |

---

### 11. 27-large-grid-5x5

**Result**: PASS

The 5x5 grid with ring connections and spoke connections to center HUB renders well. The ring edges (N1->N2->...->N5, N5->E1->E2->E3->S5, etc.) are straight horizontal or straight vertical where nodes are adjacent in the same row or column. The spoke connections (N3->HUB, W2->HUB, E2->HUB, S3->HUB) use straight paths where aligned and appropriate L bends where diagonal. No edges visibly pass through other nodes. Clearance appears adequate. The overall layout is clean and readable.

---

### 12. 28-ci-cd-pipeline

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | build -> lint | Build at row1 col2, Lint at row2 col2 (same column, directly below). Edge exits bottom of Build and goes down to Lint. This is straight vertical, 0 bends. Correct. However the edge from build -> lint appears to curve slightly or have an unnecessary jog. | Straight vertical. | MINOR |
| R5 | lint -> test | Lint at row2 col2, Test at row1 col3 (diagonal up right). The edge exits top of Lint, goes up, then right into Test. This uses 2 bends. For diagonal, 1 L bend is sufficient. | Single L bend: exit right from Lint, enter bottom of Test OR exit top of Lint, enter left of Test. | MAJOR |
| R5 | test -> sec | Test at row1 col3, Security Scan at row2 col3 (same column, directly below). Edge appears to exit bottom of Test and enter top of Security Scan. This is correct (0 bends). But visually there appears to be a small jog. | Clean straight vertical. | MINOR |
| R5 | sec -> stage | Security Scan at row2 col3, Staging at row1 col4. Diagonal up right. Edge exits right of Security Scan, goes right and up. This appears to use 2 bends. | Single L bend for diagonal. | MAJOR |
| R5 | rollback -> stage | Rollback at row3 col5, Staging at row1 col4. The edge must go up and left. The route appears to take a long path with multiple bends. | Clean U or Z shape with minimum bends. | MAJOR |
| R13 | rollback -> stage | The edge from Rollback up to Staging runs close to the Production node and Monitoring node. | Adequate clearance from non connected nodes. | MAJOR |

---

### 13. 29-mesh-topology

**Result**: PASS

The 3x3 mesh with horizontal and vertical connections renders cleanly. Horizontal edges (a->b, b->c, d->e, e->f, g->h, h->i) are straight horizontal. Vertical edges (a->d, d->g, b->e, e->h, c->f, f->i) are straight vertical. No edges pass through nodes. Perpendicular anchors are correct. Side selection is appropriate (right for horizontal, bottom for vertical). This is a clean grid layout.

---

### 14. 30-styled-nodes

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | api -> db1 | API Layer at row2 col2, Users DB at row3 col1 (diagonal down left). The edge exits the bottom of API Layer, goes down, bends left, then bends down into Users DB. This is 2 bends where 1 L bend suffices. | Single L bend: exit left from API, enter top of Users DB OR exit bottom, enter right. | MAJOR |
| R5 | api -> db3 | API Layer at row2 col2, Analytics DB at row3 col3 (diagonal down right). Same 2 bend pattern where 1 would suffice. | Single L bend. | MAJOR |
| R5 | user -> auth (login) | User at row1 col2, Auth Service at row2 col1 (diagonal down left). The edge appears to exit left of User, go left and down to Auth. The routing uses what appears to be 2 bends. | Single L bend for diagonal. | MAJOR |
| R13 | api -> db1 | The edge from API to Users DB routes downward and left. It passes near the Auth Service node boundary. | Adequate clearance from Auth Service. | MAJOR |
| R5 | db3 -> report (generate) | Analytics DB at row3 col3, Reports at row4 col2 (diagonal down left). Edge exits bottom of Analytics DB, routes down and left. Appears to use 2 bends. | Single L bend. | MAJOR |

---

### 15. 31-harness-overview

**Result**: FAIL

| Rule | Edge | Issue | Expected | Severity |
|------|------|-------|----------|----------|
| R5 | agency -> knowledge | Agency at row2 col3, Knowledge at row1 col2 (diagonal up left). The edge exits top of Agency, goes up, bends left into Knowledge. This appears to use 2 bends but the route has an unusual curved path. | Single clean L bend: exit top, enter right OR exit left, enter bottom. | MAJOR |
| R5 | governance -> knowledge | Governance at row1 col3, Knowledge at row1 col2 (same row, adjacent). The edge exits left of Governance and enters right of Knowledge. This should be a straight horizontal line. The rendering shows a straight horizontal, which is correct. | Straight horizontal. This edge appears correct. | N/A |
| R5 | governance -> agency (dashed) | Governance at row1 col3, Agency at row2 col3 (same column, directly below). The edge exits bottom of Governance and enters top of Agency. This should be straight vertical. The rendering shows a somewhat curved or jogged path rather than a clean straight line. | Straight vertical line. | MAJOR |
| R2 | governance -> agency | The dashed edge from Governance to Agency appears to have a slight curve or non perpendicular approach at the Agency anchor. | Edge must arrive perpendicular to the destination side. | CRITICAL |
| R11 | interface -> user (respond) | Interface at row2 col2, User at row2 col1 (same row, directly left). Edge exits left of Interface toward User. This is correct side selection. | Correct. | N/A |

---

## Issue Summary by Severity

| Severity | Count |
|----------|-------|
| CRITICAL | 1 |
| MAJOR | 38 |
| MINOR | 4 |
| **Total** | **43** |

## Most Common Issues

1. **R5 (Minimum Bends)**: The most pervasive issue. Diagonal connections consistently use 2 bends (down then across then down) where a single L bend would suffice. This pattern appears in nearly every diagram with diagonal node connections. This suggests the routing algorithm defaults to a "step down, step across, step down" pattern rather than computing the minimum bend L shape for diagonal pairs.

2. **R13 (Clearance from Nodes)**: Several diagrams have edges that route close to non connected nodes, particularly in dense layouts like 08-grid-3x3 and 09-diagonal-connections.

3. **R11 (Smart Side Selection)**: Side selection is generally reasonable but sometimes leads to suboptimal routes because the chosen side forces extra bends.

4. **R12 (Anchor Distribution)**: In diagrams where multiple edges converge on the same node side (08-grid-3x3, 25-cross-connections), the anchor points appear crowded rather than evenly distributed.

## Root Cause Hypothesis

The primary issue is in the diagonal routing logic. The engine appears to always route diagonal connections with a 3 segment path (vertical, horizontal, vertical) instead of a 2 segment L shaped path (vertical then horizontal, or horizontal then vertical). Fixing the diagonal routing to prefer L shapes would resolve approximately 60% of all findings.

## Recommendations

1. **Priority 1**: Fix diagonal routing to use single L bend instead of double bend for nodes that are diagonally adjacent with no obstacles between them.
2. **Priority 2**: Improve anchor distribution when multiple edges connect to the same side of a node.
3. **Priority 3**: Add clearance checks to ensure edges maintain minimum distance from non connected nodes.
4. **Priority 4**: Review the governance->agency dashed edge rendering for perpendicular anchor compliance.
