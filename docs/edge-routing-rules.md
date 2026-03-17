# Edge Routing Quality Rules

These rules define what constitutes a good diagram from the Octopus layout engine. Use these as pass/fail criteria when validating edge routing quality.

Rules are organized into categories and prioritized as CRITICAL, MAJOR, or MINOR. Critical rules represent correctness failures that make a diagram misleading or unusable. Major rules affect readability significantly. Minor rules affect aesthetic polish.

---

## Category 1: Edge Path Correctness

### R1: No Edge Through Node (CRITICAL)

An edge line must NEVER pass through any shape box other than its source and destination. If a straight or L shaped path crosses a box, the edge must reroute around it.

**Source**: Universal graph drawing constraint. ELK, yFiles, GoJS (AvoidsNodes routing), and Graphviz all treat node overlap avoidance as mandatory. Purchase et al. (1997) confirmed that node/edge overlap is the single most damaging aesthetic violation for comprehension.

**Pass**: All edge segments run through empty space or gaps between nodes.
**Fail**: Any edge segment visually crosses through a node box.

### R2: Perpendicular Anchor Contact (CRITICAL)

Every edge must leave its source anchor and arrive at its destination anchor perpendicular to the shape edge. This is a fundamental rule of orthogonal edge routing.

- Top/bottom anchors: the connecting segment must be vertical
- Left/right anchors: the connecting segment must be horizontal
- Corner anchors: either vertical or horizontal is acceptable

**Source**: ELK ORTHOGONAL routing mode enforces this. GoJS fromSpot/toSpot system produces perpendicular exits by default. yFiles orthogonal edge router guarantees perpendicular attachment. This is the defining property of orthogonal routing as described in the Handbook of Graph Drawing (Tamassia, 2013).

**Pass**: First segment from source is perpendicular to the source side. Last segment into destination is perpendicular to the destination side.
**Fail**: Edge approaches anchor at a diagonal or non perpendicular angle.

### R3: No Edge to Edge Crossing Through Node (CRITICAL)

When two edges cross, the crossing point must not occur inside a node bounding box. Crossings are tolerable in open space but never inside a shape.

**Source**: Derived from R1. If a crossing happens inside a node, at least one edge violates the no pass through rule.

**Pass**: All edge crossing points (if any) occur in open space between nodes.
**Fail**: An edge crossing point is visually inside a node boundary.

### R4: Correct Direction Indicators (CRITICAL)

Arrowheads must appear at the correct end of each edge. The arrowhead points toward the destination node. No arrowhead should point toward the source unless the connection is explicitly bidirectional.

**Source**: Fundamental correctness. All diagramming tools enforce this. Graphviz `dir` attribute controls forward/back/both/none.

**Pass**: Arrowhead is at the destination end and points into the destination node.
**Fail**: Arrowhead is missing, at the wrong end, or pointing in the wrong direction.

---

## Category 2: Edge Path Quality

### R5: Minimum Bends (MAJOR)

Routes should use the minimum number of bends necessary to reach the destination without violating other rules.

- Same row, adjacent columns: 0 bends (straight horizontal)
- Same column, adjacent rows: 0 bends (straight vertical)
- Diagonal adjacent (no obstacles): 1 bend (L shape)
- Same row/column with obstacle between: 2 bends (U shape)
- Diagonal with obstacle: 2 bends (U shape or Z shape)

**Source**: Bend minimization is a primary objective in orthogonal routing algorithms. ELK layered algorithm minimizes bends. Tamassia (1987) proved optimal bend minimization for orthogonal layouts. Purchase (1997) user studies confirmed bends degrade comprehension, second only to edge crossings.

**Pass**: Route uses the minimum bends for the geometry.
**Fail**: Route has unnecessary extra bends that could be eliminated.

### R6: Minimize Edge Crossings (MAJOR)

The total number of edge crossings in the diagram should be minimized. When crossings are unavoidable, they should occur in open space away from nodes.

**Source**: Edge crossing minimization is the most studied graph drawing aesthetic. Garey and Johnson (1983) proved crossing minimization is NP hard, but heuristics are effective. ELK provides greedy switch and sweepline crossing minimization. Purchase et al. (1997) found crossings have the single largest negative impact on diagram comprehension.

**Pass**: No edge crossings exist that could be eliminated by rerouting without adding bends or violating other rules.
**Fail**: An edge crossing exists that a different route would avoid at equal or lower bend cost.

### R7: Crossing Angle Quality (MINOR)

When two edges must cross, they should cross at angles as close to 90 degrees as possible. Shallow angle crossings are harder to trace visually.

**Source**: Huang et al. (2008) demonstrated that crossing angles near 90 degrees cause significantly less comprehension degradation than acute angle crossings. Large crossing angles help the eye distinguish the two paths.

**Pass**: Edge crossings occur at angles of 45 degrees or greater.
**Fail**: Edge crossings occur at angles less than 45 degrees (near parallel overlap).

### R8: Clean U Shape Reroutes (MAJOR)

When an edge must detour around obstacles, the reroute should be a clean U shape through the nearest gap, not a series of micro bends hugging the obstacle.

**Source**: ELK and yFiles orthogonal routers produce clean detour shapes with consistent offsets. Micro bends create visual noise and degrade the "uniform edge density" aesthetic described in graph drawing literature.

**Pass**: Rerouted edge has a smooth U or Z shape with consistent gap clearance.
**Fail**: Rerouted edge has multiple small bends or runs too close to obstacle edges.

### R9: Edge Straightness Preference (MINOR)

When source and destination are aligned (same row or same column) with no obstacles between them, the edge should be a single straight segment. No unnecessary bends or curves on a clear path.

**Source**: Straightness is a core graph drawing aesthetic. Eades (1984) force directed methods optimize for straight edges. All orthogonal routers prefer straight lines when possible.

**Pass**: Aligned nodes with clear path between them are connected by a straight line.
**Fail**: A bend or curve exists on an edge that could be a straight line.

### R10: Consistent Edge Segment Alignment (MINOR)

Horizontal segments of different edges that run at similar Y positions should snap to the same Y coordinate. Vertical segments at similar X positions should snap to the same X coordinate. This creates visual channels that are easier to follow.

**Source**: ELK compaction strategies align parallel segments. yFiles channel routing produces shared routing corridors. This reduces visual complexity by creating implicit structure.

**Pass**: Parallel segments within a small tolerance share the same coordinate.
**Fail**: Nearly parallel segments run at slightly different coordinates creating visual noise.

---

## Category 3: Edge to Node Relationship

### R11: Smart Side Selection (MAJOR)

The exit side of source and entry side of destination should be chosen based on relative position to produce the cleanest route.

- Target is directly below: exit bottom, enter top
- Target is directly above: exit top, enter bottom
- Target is directly right: exit right, enter left
- Target is directly left: exit left, enter right
- Target is diagonally down right: exit right + enter top (L shape) OR exit bottom + enter left (L shape), whichever avoids obstacles
- Target is in same row but with obstacles between: exit bottom/top (go around), not through

**Source**: GoJS fromSpot/toSpot system. ELK port side constraints with FIXED_SIDE option. yFiles port candidate sets. All major layout engines select exit/entry sides based on relative node positions.

**Pass**: Side selection produces a natural looking route toward the destination.
**Fail**: Edge exits the wrong side creating unnecessary detours.

### R12: Anchor Distribution (MAJOR)

When multiple edges connect to the same side of a node, they must use different anchor points spread evenly across that side.

- 1 edge: center anchor (position 3)
- 2 edges: positions 2 and 4
- 3 edges: positions 2, 3, and 4
- 4 edges: positions 1, 2, 4, and 5
- 5 edges: all positions 1 through 5

**Source**: GoJS distributes connections along side spots (LeftSide, RightSide, etc.). ELK port spacing option controls distance between ports. yFiles distributes ports evenly across node sides. Even distribution prevents visual congestion at anchor points.

**Pass**: Multiple edges on the same side use distinct, evenly distributed anchor points.
**Fail**: Multiple edges stack on the same anchor point or are unevenly distributed.

### R13: Minimum Edge Clearance from Nodes (MAJOR)

Edge segments that run alongside a node (but do not connect to it) must maintain a minimum clearance gap. Edges should never graze or touch a node they are not connected to.

**Source**: Graphviz `esep` attribute (default +3 points). ELK `edgeNodeSpacing` and `edgeNodeBetweenLayersSpacing` options. GoJS AvoidsNodes routing respects node bounds plus margin. yFiles minimum distance parameters. Standard practice is 8 to 16 pixels of clearance.

**Pass**: All edge segments maintain at least one grid unit of clearance from non connected nodes.
**Fail**: An edge segment runs closer than the minimum clearance to a node it does not connect to.

### R14: End Segment Minimum Length (MINOR)

The first segment leaving a source node and the last segment entering a destination node must have a minimum length. Extremely short end segments make it hard to see which side an edge connects to.

**Source**: GoJS `fromEndSegmentLength` and `toEndSegmentLength` properties enforce this. Standard practice is at least 10 to 20 pixels of end segment length before the first bend.

**Pass**: End segments are at least 10 pixels long (or one quarter of the minimum cell gap).
**Fail**: An end segment is so short that the connection side is ambiguous.

---

## Category 4: Label Quality

### R15: Label on Longest Segment (MAJOR)

Edge labels must be placed on the longest straight segment of the route, not at bend points, not at the very start or end of the edge.

**Source**: GoJS segmentIndex/segmentFraction system places labels along specific segments. ELK edgeLabelPlacement with CENTER strategy. Graphviz label placement at edge midpoint. Placing labels on the longest segment maximizes readability and minimizes overlap risk.

**Pass**: Label sits clearly on the longest segment with padding on both sides.
**Fail**: Label sits at a bend point, at an anchor point, or on a short segment.

### R16: No Label Overlap with Nodes (CRITICAL)

Edge labels must not overlap with any node bounding box. A label that covers part of a node makes both the label and the node text unreadable.

**Source**: Universal constraint. ELK label node spacing option. Graphviz xlabel uses an overlap avoidance algorithm. GoJS avoids label overlap through segment offset positioning.

**Pass**: No edge label overlaps any node bounding box.
**Fail**: Any edge label visually overlaps a node.

### R17: No Label Overlap with Other Labels (MAJOR)

Edge labels must not overlap with other edge labels or node labels. When labels compete for space, they must be offset or repositioned.

**Source**: Graphviz xlabel overlap avoidance. ELK labelLabelSpacing option. This is a standard requirement in all commercial diagramming tools.

**Pass**: All labels are fully readable with no overlap between any two labels.
**Fail**: Two or more labels overlap making at least one partially unreadable.

### R18: No Label Overlap with Edges (MINOR)

Edge labels should not overlap with other edges (edges they are not labeling). A label crossed by an unrelated edge is harder to read.

**Source**: ELK edge label side selection places labels on the less congested side. GoJS segmentOffset moves labels away from crossing edges.

**Pass**: No edge label is crossed by an unrelated edge.
**Fail**: An unrelated edge visually crosses through a label.

### R19: Label Orientation Readability (MINOR)

Labels on horizontal segments should be horizontal. Labels on vertical segments should be either horizontal (offset to the side) or rotated to read bottom to top. Labels should never be upside down.

**Source**: GoJS segmentOrientation with Upright option ensures text never renders upside down. Standard typography convention requires text to read left to right or bottom to top, never top to bottom or right to left.

**Pass**: All labels read in a natural direction (left to right, or bottom to top for vertical).
**Fail**: Any label is upside down or reads top to bottom.

---

## Category 5: Multi Edge Handling

### R20: No Edge Overlap (MAJOR)

Two distinct edges should not share the same pixel path. Even if two edges connect the same pair of nodes in the same direction, they must be visually distinguishable.

**Source**: Graphviz `concentrate` attribute merges shared paths (opt in only). By default, all tools render distinct paths. yFiles parallel edge router offsets parallel edges symmetrically.

**Pass**: Every edge has its own distinct visual path.
**Fail**: Two edges overlap making them indistinguishable.

### R21: Parallel Edge Separation (MAJOR)

When two or more edges run between the same pair of nodes (or along the same corridor), they must be offset from each other by a consistent amount. The offset should be symmetric around the center line.

**Source**: yFiles parallel edge router. GoJS Link.curve with curviness offset. ELK handles parallel edges via port distribution. Standard offset is 8 to 16 pixels between parallel edges.

**Pass**: Parallel edges are evenly spaced and symmetric around the center path.
**Fail**: Parallel edges are unevenly spaced or asymmetric.

### R22: Bidirectional Edge Distinction (MAJOR)

When two nodes have edges in both directions (A to B and B to A), the two edges must follow different paths so both arrowheads are visible. They should not be rendered as a single line with arrowheads at both ends unless the connection is semantically bidirectional.

**Source**: Graphviz `dir=both` for semantic bidirectional edges. yFiles and GoJS offset reverse edges to separate paths. Distinct directed connections require distinct visual paths.

**Pass**: Both edges are visible with distinct paths and arrowheads pointing in correct directions.
**Fail**: Reverse edges overlap, share a path, or have ambiguous direction.

### R23: Self Loop Rendering (MINOR)

When a node has a connection to itself, the self loop must be rendered as a visible loop that exits and returns to the same node. The loop should have a consistent shape (typically a rounded rectangle or curve) and should not overlap with other edges.

**Source**: ELK selfLoopDistribution and selfLoopOrdering options. Graphviz renders self loops as small circular arcs. Standard self loop rendering exits from one anchor, arcs outward, and returns to an adjacent anchor on the same or different side.

**Pass**: Self loop is visible, does not overlap other edges, exits and returns to distinct anchor points.
**Fail**: Self loop is invisible, overlaps other edges, or exits and returns at the same point.

---

## Category 6: Node Layout Correctness

### R24: Shape Centering (MAJOR)

Nodes should be centered within their grid cell (default alignment). The center of the shape should align with the center of the cell.

**Source**: Octopus grid layout model. All grid based layout systems center elements within cells by default.

**Pass**: Shape is visually centered in its cell.
**Fail**: Shape is offset to one corner of its cell.

### R25: No Node Overlap (CRITICAL)

No two node bounding boxes should overlap. If two shapes overlap due to custom sizes exceeding cell dimensions, the engine should report an error rather than render overlapping shapes.

**Source**: ELK overlap removal algorithms. Graphviz `overlap` attribute with removal strategies. GoJS avoidable node bounds. Overlapping nodes produce unreadable diagrams.

**Pass**: No two node bounding boxes overlap, OR overlapping shapes produce a clear error message.
**Fail**: Overlapping shapes render without warning.

### R26: Uniform Cell Sizing (MINOR)

Grid cells in the same row should share the same height. Grid cells in the same column should share the same width. This ensures visual consistency of the grid structure.

**Source**: Standard grid layout behavior. HTML table cell sizing rules. ELK layered algorithm equalizes layer heights.

**Pass**: All cells in a row have the same height. All cells in a column have the same width.
**Fail**: Cells in the same row/column have inconsistent dimensions.

---

## Category 7: Aesthetic Quality

### R27: Visual Symmetry (MINOR)

When the graph structure is symmetric, the layout should reflect that symmetry. If nodes A and B are both children of C, and both have the same number of connections, their edge routes should be mirror images of each other.

**Source**: Symmetry is one of the classical graph drawing aesthetics (Eades, 1984). Purchase (1997) found symmetry has a positive but smaller effect than crossing or bend reduction. ELK and yFiles support symmetric layouts through balanced placement.

**Pass**: Structurally symmetric subgraphs have visually symmetric edge routes.
**Fail**: Structurally symmetric subgraphs have asymmetric routes when symmetric routes are possible.

### R28: Uniform Edge Spacing (MINOR)

Edges in the same routing corridor should be evenly spaced. The gap between adjacent parallel edge segments should be consistent throughout the diagram.

**Source**: ELK edgeEdgeBetweenLayerSpacing option. yFiles channel routing. Consistent spacing reduces visual noise and makes edges easier to trace.

**Pass**: Parallel edges in the same corridor maintain consistent spacing.
**Fail**: Spacing between parallel edges varies noticeably within the same corridor.

### R29: Minimal Total Edge Length (MINOR)

The total combined length of all edges should be reasonably short. Edges should not take unnecessarily long detours when shorter routes exist that satisfy all other rules.

**Source**: Total edge length is a standard graph drawing optimization metric. Shorter edges are easier to trace. ELK and force directed algorithms minimize total edge length as a secondary objective.

**Pass**: No edge takes a detour longer than 2x the shortest compliant path.
**Fail**: An edge route is more than 2x the length of the shortest path that satisfies all higher priority rules.

### R30: Consistent Routing Style (MINOR)

All edges in the diagram should use the same routing style (all orthogonal, all polyline, or all spline). Mixing styles within a single diagram creates visual inconsistency.

**Source**: ELK edgeRouting option applies per parent (all children get the same style). Graphviz splines attribute is graph wide. Consistent style is a basic design principle.

**Pass**: All edges use the same routing style throughout the diagram.
**Fail**: Some edges are orthogonal while others are diagonal or curved within the same diagram.

---

## Category 8: Spacing and Clearance

### R31: Minimum Gap Between Parallel Edge Segments (MINOR)

When two edge segments run parallel to each other (same direction, different edges), there must be a minimum visible gap between them so a viewer can distinguish them as separate lines.

**Source**: ELK edgeEdgeBetweenLayerSpacing. Graphviz esep. Standard minimum is 4 to 8 pixels depending on stroke width.

**Pass**: Parallel segments from different edges have at least 4 pixels of gap.
**Fail**: Parallel segments are so close they appear as a single thick line.

### R32: Bend Point Minimum Spacing (MINOR)

Two consecutive bends in the same edge should be separated by a minimum distance. Bends that are too close together create a visual zigzag that is hard to follow.

**Source**: yFiles minimum segment length parameters. GoJS corner property smooths tight bends. Short segments between bends degrade readability.

**Pass**: Every straight segment between two bends is at least 10 pixels long.
**Fail**: Two bends are so close that the segment between them is nearly invisible.

### R33: Edge to Diagram Boundary Clearance (MINOR)

Edges should not run along or touch the diagram boundary. A minimum margin should exist between any edge segment and the diagram edge.

**Source**: Graphviz pad attribute. ELK padding options. Standard practice is at least the same padding used for nodes.

**Pass**: All edge segments maintain at least the diagram padding distance from the boundary.
**Fail**: An edge segment touches or extends beyond the diagram boundary.

---

## Rule Priority Summary

When rules conflict, resolve in this order:

1. CRITICAL rules (R1, R2, R3, R4, R16, R25) are never violated
2. MAJOR rules (R5, R6, R8, R11, R12, R13, R15, R17, R20, R21, R22, R24) are violated only when necessary to satisfy a CRITICAL rule
3. MINOR rules (R7, R9, R10, R14, R18, R19, R23, R26, R27, R28, R29, R30, R31, R32, R33) are violated only when necessary to satisfy a CRITICAL or MAJOR rule

When two rules of the same priority conflict, prefer the rule with the lower number.

---

## Research Sources

These rules are synthesized from the following sources:

- **Purchase, Cohen, James (1997)**: "Which Aesthetic Has the Greatest Effect on Human Understanding?" foundational user study ranking aesthetic criteria by impact on comprehension. Edge crossings ranked first, followed by bends, then symmetry.
- **Tamassia (1987)**: Optimal bend minimization algorithm for orthogonal graph drawing. Proved polynomial time solution for fixed embedding.
- **Huang, Hong, Eades (2008)**: Crossing angle research demonstrating large angle crossings are significantly less harmful than acute angle crossings.
- **Eades (1984)**: Force directed graph drawing introducing spring model aesthetics including edge straightness, uniform edge length, and symmetry.
- **Handbook of Graph Drawing and Visualization (Tamassia, 2013)**: Comprehensive reference covering orthogonal drawing, edge routing, and aesthetic criteria.
- **Eclipse Layout Kernel (ELK)**: Open source layout framework with documented options for edge routing, spacing, crossing minimization, label placement, port constraints, and compaction.
- **yFiles**: Commercial graph visualization library with parallel edge routing, orthogonal edge routing, channel routing, and port distribution algorithms.
- **GoJS**: Diagramming library with AvoidsNodes routing, Spot based connection points, segmentIndex label placement, and end segment length parameters.
- **JointJS**: Diagramming library with Manhattan routing, orthogonal routing, metro routing, and obstacle avoidance.
- **Graphviz**: Open source graph visualization with splines, esep/sep spacing, xlabel overlap avoidance, concentrate edge merging, and dir arrowhead control.
