package plugin

import (
	"container/heap"
	"math"
	"sort"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/lib/geo"

	"github.com/valllabh/d2-octopus-layout-engine/internal/grid"
)

// direction encodes how we arrived at a routing grid node.
type direction int

const (
	dirNone  direction = 0
	dirUp    direction = 1
	dirDown  direction = 2
	dirLeft  direction = 3
	dirRight direction = 4
)

// routingGrid is a sparse grid of X/Y coordinates through gap channels.
// Nodes at intersections of these coordinates form the search space for A*.
type routingGrid struct {
	xs       []float64           // sorted unique X coordinates
	ys       []float64           // sorted unique Y coordinates
	blocked  map[[2]int]bool     // (xi, yi) -> blocked by obstacle
	obstacles []*geo.Box         // all obstacle boxes
	routed   []routedSegment     // segments from previously routed edges
}

// routedSegment records one segment of an already routed edge for crossing detection.
type routedSegment struct {
	x1, y1, x2, y2 float64
}

const (
	obstacleMargin  = 8.0   // margin around obstacles for blocked check
	bendPenalty     = 500.0  // cost penalty for changing direction
	crossingPenalty = 300.0  // cost penalty per crossing with existing edges
	proximityDist   = 20.0  // distance threshold for proximity penalty
	proximityPen    = 50.0   // cost penalty for being near an obstacle
)

// buildRoutingGrid creates a sparse routing grid from gap channels and anchor points.
func buildRoutingGrid(obstacles []*geo.Box, opts grid.Options) *routingGrid {
	// Find max row and col from obstacle positions
	maxRow, maxCol := 0, 0
	for _, obs := range obstacles {
		// Estimate grid position from pixel coordinates
		col := int(math.Round((obs.TopLeft.X-float64(opts.Padding))/float64(opts.CellWidth+opts.Gap))) + 1
		row := int(math.Round((obs.TopLeft.Y-float64(opts.Padding))/float64(opts.CellHeight+opts.Gap))) + 1
		if col > maxCol {
			maxCol = col
		}
		if row > maxRow {
			maxRow = row
		}
	}

	xSet := make(map[float64]bool)
	ySet := make(map[float64]bool)

	// Add gap center coordinates with sub-lanes for parallel edge separation.
	// Each gap channel gets 3 lanes: center-laneOffset, center, center+laneOffset.
	const laneOffset = 12.0
	for c := 0; c <= maxCol; c++ {
		cx := opts.GapCenterX(c)
		xSet[cx-laneOffset] = true
		xSet[cx] = true
		xSet[cx+laneOffset] = true
	}
	for r := 0; r <= maxRow; r++ {
		cy := opts.GapCenterY(r)
		ySet[cy-laneOffset] = true
		ySet[cy] = true
		ySet[cy+laneOffset] = true
	}

	// Add boundary margins (outside the grid)
	marginX := float64(opts.Padding) / 2
	marginY := float64(opts.Padding) / 2
	if marginX < 25 {
		marginX = 25
	}
	if marginY < 25 {
		marginY = 25
	}

	// Find bounding box of all obstacles
	minOX, minOY := math.MaxFloat64, math.MaxFloat64
	maxOX, maxOY := -math.MaxFloat64, -math.MaxFloat64
	for _, obs := range obstacles {
		if obs.TopLeft.X < minOX {
			minOX = obs.TopLeft.X
		}
		if obs.TopLeft.Y < minOY {
			minOY = obs.TopLeft.Y
		}
		r := obs.TopLeft.X + obs.Width
		b := obs.TopLeft.Y + obs.Height
		if r > maxOX {
			maxOX = r
		}
		if b > maxOY {
			maxOY = b
		}
	}

	// Add boundary lines outside all obstacles
	xSet[minOX-marginX] = true
	xSet[maxOX+marginX] = true
	ySet[minOY-marginY] = true
	ySet[maxOY+marginY] = true

	// Convert to sorted slices
	xs := make([]float64, 0, len(xSet))
	for x := range xSet {
		xs = append(xs, x)
	}
	sort.Float64s(xs)

	ys := make([]float64, 0, len(ySet))
	for y := range ySet {
		ys = append(ys, y)
	}
	sort.Float64s(ys)

	// Mark blocked nodes (inside any obstacle with margin)
	blocked := make(map[[2]int]bool)
	for xi, x := range xs {
		for yi, y := range ys {
			for _, obs := range obstacles {
				if pointInsideBoxWithMargin(x, y, obs, obstacleMargin) {
					blocked[[2]int{xi, yi}] = true
					break
				}
			}
		}
	}

	return &routingGrid{
		xs:        xs,
		ys:        ys,
		blocked:   blocked,
		obstacles: obstacles,
	}
}

// withAnchorCoordinates creates a copy of the routing grid with anchor X/Y values added.
// Returns the new grid and the indices for the anchor points.
func (rg *routingGrid) withAnchorCoordinates(srcPt, dstPt *geo.Point, srcBox, dstBox *geo.Box) (*routingGrid, int, int, int, int) {
	// Insert source and destination coordinates
	xSet := make(map[float64]bool)
	ySet := make(map[float64]bool)
	for _, x := range rg.xs {
		xSet[x] = true
	}
	for _, y := range rg.ys {
		ySet[y] = true
	}
	xSet[srcPt.X] = true
	xSet[dstPt.X] = true
	ySet[srcPt.Y] = true
	ySet[dstPt.Y] = true

	xs := make([]float64, 0, len(xSet))
	for x := range xSet {
		xs = append(xs, x)
	}
	sort.Float64s(xs)

	ys := make([]float64, 0, len(ySet))
	for y := range ySet {
		ys = append(ys, y)
	}
	sort.Float64s(ys)

	// Build blocked map, excluding src and dst boxes from blocking
	blocked := make(map[[2]int]bool)
	for xi, x := range xs {
		for yi, y := range ys {
			for _, obs := range rg.obstacles {
				if obs == srcBox || obs == dstBox {
					continue
				}
				if pointInsideBoxWithMargin(x, y, obs, obstacleMargin) {
					blocked[[2]int{xi, yi}] = true
					break
				}
			}
		}
	}

	newRG := &routingGrid{
		xs:        xs,
		ys:        ys,
		blocked:   blocked,
		obstacles: rg.obstacles,
		routed:    rg.routed,
	}

	srcXI := findIndex(xs, srcPt.X)
	srcYI := findIndex(ys, srcPt.Y)
	dstXI := findIndex(xs, dstPt.X)
	dstYI := findIndex(ys, dstPt.Y)

	return newRG, srcXI, srcYI, dstXI, dstYI
}

func findIndex(sorted []float64, val float64) int {
	for i, v := range sorted {
		if math.Abs(v-val) < 0.01 {
			return i
		}
	}
	return 0
}

func pointInsideBoxWithMargin(x, y float64, box *geo.Box, margin float64) bool {
	return x > box.TopLeft.X-margin &&
		x < box.TopLeft.X+box.Width+margin &&
		y > box.TopLeft.Y-margin &&
		y < box.TopLeft.Y+box.Height+margin
}

// segmentCrossesObstacle checks if a horizontal or vertical segment between two grid nodes
// passes through any obstacle.
func (rg *routingGrid) segmentCrossesObstacle(x1, y1, x2, y2 float64) bool {
	p1 := geo.NewPoint(x1, y1)
	p2 := geo.NewPoint(x2, y2)
	for _, obs := range rg.obstacles {
		if segmentIntersectsBox(p1, p2, obs) {
			return true
		}
	}
	return false
}

// astarState is the search state: grid position plus arrival direction.
type astarState struct {
	xi, yi int
	dir    direction
}

// astarNode is a node in the A* priority queue.
type astarNode struct {
	state    astarState
	gCost    float64     // cost from start
	fCost    float64     // gCost + heuristic
	parent   *astarNode
	index    int         // heap index
}

// priorityQueue implements heap.Interface for A* open set.
type priorityQueue []*astarNode

func (pq priorityQueue) Len() int            { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool  { return pq[i].fCost < pq[j].fCost }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *priorityQueue) Push(x interface{}) {
	n := x.(*astarNode)
	n.index = len(*pq)
	*pq = append(*pq, n)
}
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}

// anchorExitDirection returns the required exit direction for a given anchor.
// Top anchors exit upward, bottom exit downward, left exit leftward, right exit rightward.
func anchorExitDirection(anchor grid.Anchor) direction {
	if isVerticalAnchor(anchor) {
		if isBottomAnchor(anchor) {
			return dirDown
		}
		return dirUp
	}
	if isRightAnchor(anchor) {
		return dirRight
	}
	return dirLeft
}

// anchorApproachDirection returns the direction of travel when arriving at a destination anchor.
// To reach a left anchor, you travel rightward. To reach a top anchor, you travel downward.
func anchorApproachDirection(anchor grid.Anchor) direction {
	if isVerticalAnchor(anchor) {
		if isBottomAnchor(anchor) {
			return dirUp // arrive at bottom by traveling up
		}
		return dirDown // arrive at top by traveling down
	}
	if isRightAnchor(anchor) {
		return dirLeft // arrive at right by traveling left
	}
	return dirRight // arrive at left by traveling right
}

// astarRouteEdge routes a single edge using A* search on the routing grid.
func astarRouteEdge(edge *d2graph.Edge, srcBox, dstBox *geo.Box, anchors grid.EdgeAnchors, rg *routingGrid) {
	// Resolve anchor points
	sx, sy := grid.AnchorPoint(anchors.SrcAnchor, srcBox.TopLeft.X, srcBox.TopLeft.Y, srcBox.Width, srcBox.Height)
	srcPt := geo.NewPoint(sx, sy)

	dx, dy := grid.AnchorPoint(anchors.DstAnchor, dstBox.TopLeft.X, dstBox.TopLeft.Y, dstBox.Width, dstBox.Height)
	dstPt := geo.NewPoint(dx, dy)

	srcVertical := isVerticalAnchor(anchors.SrcAnchor)
	dstVertical := isVerticalAnchor(anchors.DstAnchor)

	srcCX := srcBox.TopLeft.X + srcBox.Width/2
	srcCY := srcBox.TopLeft.Y + srcBox.Height/2
	dstCX := dstBox.TopLeft.X + dstBox.Width/2
	dstCY := dstBox.TopLeft.Y + dstBox.Height/2
	sameCol := math.Abs(srcCX-dstCX) < 5
	sameRow := math.Abs(srcCY-dstCY) < 5

	// Same row with both horizontal anchors: use straight horizontal if no obstacle between them
	if sameRow && !srcVertical && !dstVertical {
		avgY := (srcPt.Y + dstPt.Y) / 2
		straightPt1 := geo.NewPoint(srcPt.X, avgY)
		straightPt2 := geo.NewPoint(dstPt.X, avgY)
		if !segmentCrossesOtherObstacles(straightPt1.X, straightPt1.Y, straightPt2.X, straightPt2.Y, rg.obstacles, srcBox, dstBox) {
			midX := (srcPt.X + dstPt.X) / 2
			edge.Route = []*geo.Point{
				straightPt1,
				geo.NewPoint(midX, avgY),
				straightPt2,
			}
			edge.IsCurve = false
			registerRouteSegments(edge.Route, rg)
			return
		}
	}

	// Same column with both vertical anchors: use straight vertical if no obstacle between them
	if sameCol && srcVertical && dstVertical {
		avgX := (srcPt.X + dstPt.X) / 2
		straightPt1 := geo.NewPoint(avgX, srcPt.Y)
		straightPt2 := geo.NewPoint(avgX, dstPt.Y)
		if !segmentCrossesOtherObstacles(straightPt1.X, straightPt1.Y, straightPt2.X, straightPt2.Y, rg.obstacles, srcBox, dstBox) {
			midY := (srcPt.Y + dstPt.Y) / 2
			edge.Route = []*geo.Point{
				straightPt1,
				geo.NewPoint(avgX, midY),
				straightPt2,
			}
			edge.IsCurve = false
			registerRouteSegments(edge.Route, rg)
			return
		}
	}

	// Create a per edge copy of the grid with anchor coordinates
	edgeGrid, srcXI, srcYI, dstXI, dstYI := rg.withAnchorCoordinates(srcPt, dstPt, srcBox, dstBox)

	// Determine required directions
	srcExitDir := anchorExitDirection(anchors.SrcAnchor)
	dstApproachDir := anchorApproachDirection(anchors.DstAnchor)

	// Run A*
	path := runAStar(edgeGrid, srcXI, srcYI, dstXI, dstYI, srcExitDir, dstApproachDir, srcBox, dstBox)

	if path == nil {
		// Fallback: direct L shape outside the grid boundary
		path = fallbackRoute(srcPt, dstPt, anchors.SrcAnchor, anchors.DstAnchor, edgeGrid)
	}

	// Convert grid indices to pixel points
	route := make([]*geo.Point, len(path))
	for i, idx := range path {
		route[i] = geo.NewPoint(edgeGrid.xs[idx[0]], edgeGrid.ys[idx[1]])
	}

	// Simplify: remove collinear intermediate waypoints
	route = simplifyRoute(route)

	// Register segments for crossing detection
	registerRouteSegments(route, rg)

	edge.Route = route
	edge.IsCurve = false
}

// registerRouteSegments adds the segments of a route to the routing grid for crossing detection.
func registerRouteSegments(route []*geo.Point, rg *routingGrid) {
	for i := 1; i < len(route); i++ {
		rg.routed = append(rg.routed, routedSegment{
			x1: route[i-1].X, y1: route[i-1].Y,
			x2: route[i].X, y2: route[i].Y,
		})
	}
}

// overlapSegInfo identifies a segment within an edge's route.
type overlapSegInfo struct {
	edgeIdx int
	segIdx  int
}

// spreadOverlappingSegments detects parallel segments from different edges that
// share the same gap channel coordinate and spreads them evenly across the
// available gap width. Horizontal segments use the vertical gap (between rows)
// and vertical segments use the horizontal gap (between columns).
// Only interior segments (not first or last) are adjusted to preserve anchors.
func spreadOverlappingSegments(edges []*d2graph.Edge, opts grid.Options) {
	// Collect all obstacle boxes for computing actual gaps between shapes
	var obstacles []*geo.Box
	for _, edge := range edges {
		if edge.Src != nil && edge.Src.Box != nil {
			obstacles = append(obstacles, edge.Src.Box)
		}
		if edge.Dst != nil && edge.Dst.Box != nil {
			obstacles = append(obstacles, edge.Dst.Box)
		}
	}

	type segInfo = overlapSegInfo

	type hGroup struct {
		y    float64
		segs []segInfo
	}
	type vGroup struct {
		x    float64
		segs []segInfo
	}

	var hGroups []hGroup
	var vGroups []vGroup

	// Use a wider tolerance to group segments from sub-lanes into the same channel
	const groupTolerance = 15.0

	for ei, edge := range edges {
		if edge.Route == nil || len(edge.Route) < 4 {
			continue
		}
		for si := 1; si < len(edge.Route)-2; si++ {
			p1 := edge.Route[si]
			p2 := edge.Route[si+1]
			if math.Abs(p1.Y-p2.Y) < 1 {
				y := (p1.Y + p2.Y) / 2
				added := false
				for gi := range hGroups {
					if math.Abs(hGroups[gi].y-y) < groupTolerance {
						hGroups[gi].segs = append(hGroups[gi].segs, segInfo{ei, si})
						added = true
						break
					}
				}
				if !added {
					hGroups = append(hGroups, hGroup{y: y, segs: []segInfo{{ei, si}}})
				}
			} else if math.Abs(p1.X-p2.X) < 1 {
				x := (p1.X + p2.X) / 2
				added := false
				for gi := range vGroups {
					if math.Abs(vGroups[gi].x-x) < groupTolerance {
						vGroups[gi].segs = append(vGroups[gi].segs, segInfo{ei, si})
						added = true
						break
					}
				}
				if !added {
					vGroups = append(vGroups, vGroup{x: x, segs: []segInfo{{ei, si}}})
				}
			}
		}
	}

	// Spread horizontal segments evenly across the actual vertical gap.
	// Find the nearest shape above and below the channel to determine real gap.
	for _, g := range hGroups {
		if len(g.segs) < 2 {
			continue
		}
		gapTop, gapBottom := findVerticalGapBounds(g.y, obstacles)
		gapHeight := gapBottom - gapTop
		if gapHeight < 4 {
			gapHeight = float64(opts.Gap)
		}
		gapCenter := (gapTop + gapBottom) / 2
		n := len(g.segs)
		spacing := gapHeight / float64(n+1)
		for i, si := range g.segs {
			newY := gapTop + spacing*float64(i+1)
			_ = gapCenter // use gapTop based positioning
			p1 := edges[si.edgeIdx].Route[si.segIdx]
			p2 := edges[si.edgeIdx].Route[si.segIdx+1]
			p1.Y = newY
			p2.Y = newY
		}
	}

	// Spread vertical segments evenly across the actual horizontal gap.
	for _, g := range vGroups {
		if len(g.segs) < 2 {
			continue
		}
		gapLeft, gapRight := findHorizontalGapBounds(g.x, obstacles)
		gapWidth := gapRight - gapLeft
		if gapWidth < 4 {
			gapWidth = float64(opts.Gap)
		}
		n := len(g.segs)
		spacing := gapWidth / float64(n+1)
		for i, si := range g.segs {
			newX := gapLeft + spacing*float64(i+1)
			p1 := edges[si.edgeIdx].Route[si.segIdx]
			p2 := edges[si.edgeIdx].Route[si.segIdx+1]
			p1.X = newX
			p2.X = newX
		}
	}
}


// findVerticalGapBounds finds the actual vertical gap around a Y coordinate
// by looking for the nearest shape bottom above and shape top below.
func findVerticalGapBounds(y float64, obstacles []*geo.Box) (top, bottom float64) {
	top = -math.MaxFloat64
	bottom = math.MaxFloat64
	for _, obs := range obstacles {
		obsBottom := obs.TopLeft.Y + obs.Height
		obsTop := obs.TopLeft.Y
		// Nearest bottom edge above y
		if obsBottom <= y+1 && obsBottom > top {
			top = obsBottom
		}
		// Nearest top edge below y
		if obsTop >= y-1 && obsTop < bottom {
			bottom = obsTop
		}
	}
	// Fallback if no bounds found
	if top == -math.MaxFloat64 {
		top = y - 20
	}
	if bottom == math.MaxFloat64 {
		bottom = y + 20
	}
	return top, bottom
}

// findHorizontalGapBounds finds the actual horizontal gap around an X coordinate
// by looking for the nearest shape right edge to the left and shape left edge to the right.
func findHorizontalGapBounds(x float64, obstacles []*geo.Box) (left, right float64) {
	left = -math.MaxFloat64
	right = math.MaxFloat64
	for _, obs := range obstacles {
		obsRight := obs.TopLeft.X + obs.Width
		obsLeft := obs.TopLeft.X
		// Nearest right edge to the left of x
		if obsRight <= x+1 && obsRight > left {
			left = obsRight
		}
		// Nearest left edge to the right of x
		if obsLeft >= x-1 && obsLeft < right {
			right = obsLeft
		}
	}
	if left == -math.MaxFloat64 {
		left = x - 20
	}
	if right == math.MaxFloat64 {
		right = x + 20
	}
	return left, right
}

// runAStar performs A* search on the routing grid.
func runAStar(rg *routingGrid, srcXI, srcYI, dstXI, dstYI int, srcExitDir, dstApproachDir direction, srcBox, dstBox *geo.Box) [][2]int {
	maxXI := len(rg.xs) - 1
	maxYI := len(rg.ys) - 1

	// Heuristic: Manhattan distance in pixels
	heuristic := func(xi, yi int) float64 {
		return math.Abs(rg.xs[xi]-rg.xs[dstXI]) + math.Abs(rg.ys[yi]-rg.ys[dstYI])
	}

	startState := astarState{xi: srcXI, yi: srcYI, dir: dirNone}
	startNode := &astarNode{
		state: startState,
		gCost: 0,
		fCost: heuristic(srcXI, srcYI),
	}

	open := &priorityQueue{startNode}
	heap.Init(open)

	// Best g cost seen for each state
	best := make(map[astarState]float64)
	best[startState] = 0

	// Neighbor deltas: dx, dy, direction
	neighbors := [4]struct {
		dxi, dyi int
		dir      direction
	}{
		{0, -1, dirUp},
		{0, 1, dirDown},
		{-1, 0, dirLeft},
		{1, 0, dirRight},
	}

	for open.Len() > 0 {
		current := heap.Pop(open).(*astarNode)
		xi, yi := current.state.xi, current.state.yi

		// Goal check: reached destination with correct approach direction
		if xi == dstXI && yi == dstYI {
			if current.state.dir == dstApproachDir || current.state.dir == dirNone {
				return reconstructPath(current)
			}
		}

		for _, nb := range neighbors {
			nxi := xi + nb.dxi
			nyi := yi + nb.dyi

			// Bounds check
			if nxi < 0 || nxi > maxXI || nyi < 0 || nyi > maxYI {
				continue
			}

			// Enforce source exit direction on first move
			if current.state.dir == dirNone && nb.dir != srcExitDir {
				continue
			}

			// Skip if destination node is blocked (unless it is the goal)
			if rg.blocked[[2]int{nxi, nyi}] && !(nxi == dstXI && nyi == dstYI) {
				continue
			}

			// Check segment does not cross an obstacle
			x1, y1 := rg.xs[xi], rg.ys[yi]
			x2, y2 := rg.xs[nxi], rg.ys[nyi]

			// Allow segments that start or end at src/dst box (they are on the anchor)
			if !isEndpointBox(x1, y1, x2, y2, srcBox) && !isEndpointBox(x1, y1, x2, y2, dstBox) {
				if rg.segmentCrossesObstacle(x1, y1, x2, y2) {
					continue
				}
			} else {
				// Even for endpoint segments, check against OTHER obstacles
				if segmentCrossesOtherObstacles(x1, y1, x2, y2, rg.obstacles, srcBox, dstBox) {
					continue
				}
			}

			// Compute cost
			segDist := math.Abs(x2-x1) + math.Abs(y2-y1)
			moveCost := segDist

			// Bend penalty
			if current.state.dir != dirNone && nb.dir != current.state.dir {
				moveCost += bendPenalty
			}

			// Crossing penalty
			crossings := countCrossings(x1, y1, x2, y2, rg.routed)
			moveCost += float64(crossings) * crossingPenalty

			// Parallel overlap penalty: penalize using a lane that already has a
			// routed segment running in the same direction with overlapping range
			moveCost += parallelOverlapPenalty(x1, y1, x2, y2, rg.routed)

			// Proximity penalty
			moveCost += proximityPenalty(x1, y1, x2, y2, rg.obstacles, srcBox, dstBox)

			newG := current.gCost + moveCost
			nState := astarState{xi: nxi, yi: nyi, dir: nb.dir}

			if prev, ok := best[nState]; ok && newG >= prev {
				continue
			}
			best[nState] = newG

			node := &astarNode{
				state:  nState,
				gCost:  newG,
				fCost:  newG + heuristic(nxi, nyi),
				parent: current,
			}
			heap.Push(open, node)
		}
	}

	return nil // no path found
}

func isEndpointBox(x1, y1, x2, y2 float64, box *geo.Box) bool {
	if box == nil {
		return false
	}
	return pointInsideBoxWithMargin(x1, y1, box, 1) || pointInsideBoxWithMargin(x2, y2, box, 1)
}

func segmentCrossesOtherObstacles(x1, y1, x2, y2 float64, obstacles []*geo.Box, srcBox, dstBox *geo.Box) bool {
	p1 := geo.NewPoint(x1, y1)
	p2 := geo.NewPoint(x2, y2)
	for _, obs := range obstacles {
		if obs == srcBox || obs == dstBox {
			continue
		}
		if segmentIntersectsBox(p1, p2, obs) {
			return true
		}
	}
	return false
}

// countCrossings counts how many previously routed segments this segment crosses.
func countCrossings(x1, y1, x2, y2 float64, routed []routedSegment) int {
	count := 0
	for _, seg := range routed {
		if segmentsIntersect(x1, y1, x2, y2, seg.x1, seg.y1, seg.x2, seg.y2) {
			count++
		}
	}
	return count
}

// segmentsIntersect checks if two orthogonal segments cross each other.
// One must be horizontal and the other vertical to cross.
func segmentsIntersect(ax1, ay1, ax2, ay2, bx1, by1, bx2, by2 float64) bool {
	aHoriz := math.Abs(ay1-ay2) < 1
	bHoriz := math.Abs(by1-by2) < 1

	// Both same orientation cannot cross (they are parallel)
	if aHoriz == bHoriz {
		return false
	}

	// Make a the horizontal one, b the vertical one
	if !aHoriz {
		ax1, ay1, ax2, ay2, bx1, by1, bx2, by2 = bx1, by1, bx2, by2, ax1, ay1, ax2, ay2
	}

	// a is horizontal (same Y), b is vertical (same X)
	hY := ay1
	vX := bx1
	hMinX := math.Min(ax1, ax2)
	hMaxX := math.Max(ax1, ax2)
	vMinY := math.Min(by1, by2)
	vMaxY := math.Max(by1, by2)

	return vX > hMinX && vX < hMaxX && hY > vMinY && hY < vMaxY
}

// parallelOverlapPenalty penalizes a segment that runs parallel to and overlaps
// with an already routed segment. This pushes the A* to use a different sub-lane.
func parallelOverlapPenalty(x1, y1, x2, y2 float64, routed []routedSegment) float64 {
	const parallelPen = 400.0
	const nearThreshold = 2.0 // within 2px = same lane
	penalty := 0.0

	aHoriz := math.Abs(y1-y2) < 1
	aVert := math.Abs(x1-x2) < 1

	for _, seg := range routed {
		bHoriz := math.Abs(seg.y1-seg.y2) < 1
		bVert := math.Abs(seg.x1-seg.x2) < 1

		if aHoriz && bHoriz && math.Abs(y1-seg.y1) < nearThreshold {
			// Both horizontal at the same Y. Check X range overlap.
			aMin := math.Min(x1, x2)
			aMax := math.Max(x1, x2)
			bMin := math.Min(seg.x1, seg.x2)
			bMax := math.Max(seg.x1, seg.x2)
			if aMax > bMin && bMax > aMin {
				penalty += parallelPen
			}
		}
		if aVert && bVert && math.Abs(x1-seg.x1) < nearThreshold {
			// Both vertical at the same X. Check Y range overlap.
			aMin := math.Min(y1, y2)
			aMax := math.Max(y1, y2)
			bMin := math.Min(seg.y1, seg.y2)
			bMax := math.Max(seg.y1, seg.y2)
			if aMax > bMin && bMax > aMin {
				penalty += parallelPen
			}
		}
	}
	return penalty
}

// proximityPenalty adds cost when a segment runs close to an obstacle.
func proximityPenalty(x1, y1, x2, y2 float64, obstacles []*geo.Box, srcBox, dstBox *geo.Box) float64 {
	penalty := 0.0
	for _, obs := range obstacles {
		if obs == srcBox || obs == dstBox {
			continue
		}
		dist := segmentToBoxDistance(x1, y1, x2, y2, obs)
		if dist < proximityDist && dist > 0 {
			penalty += proximityPen * (1 - dist/proximityDist)
		}
	}
	return penalty
}

// segmentToBoxDistance computes the minimum distance from an orthogonal segment to a box.
func segmentToBoxDistance(x1, y1, x2, y2 float64, box *geo.Box) float64 {
	bx1 := box.TopLeft.X
	by1 := box.TopLeft.Y
	bx2 := bx1 + box.Width
	by2 := by1 + box.Height

	// Horizontal segment
	if math.Abs(y1-y2) < 1 {
		y := y1
		minX := math.Min(x1, x2)
		maxX := math.Max(x1, x2)

		// Check if segment X range overlaps box X range
		if maxX > bx1 && minX < bx2 {
			// Vertical distance to box
			if y < by1 {
				return by1 - y
			}
			if y > by2 {
				return y - by2
			}
			return 0 // inside
		}
		return math.MaxFloat64
	}

	// Vertical segment
	if math.Abs(x1-x2) < 1 {
		x := x1
		minY := math.Min(y1, y2)
		maxY := math.Max(y1, y2)

		if maxY > by1 && minY < by2 {
			if x < bx1 {
				return bx1 - x
			}
			if x > bx2 {
				return x - bx2
			}
			return 0
		}
		return math.MaxFloat64
	}

	return math.MaxFloat64
}

// reconstructPath traces back from the goal node to the start.
func reconstructPath(node *astarNode) [][2]int {
	var path [][2]int
	for n := node; n != nil; n = n.parent {
		path = append(path, [2]int{n.state.xi, n.state.yi})
	}
	// Reverse
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// simplifyRoute removes collinear intermediate waypoints.
func simplifyRoute(route []*geo.Point) []*geo.Point {
	if len(route) <= 2 {
		return route
	}

	result := []*geo.Point{route[0]}
	for i := 1; i < len(route)-1; i++ {
		prev := result[len(result)-1]
		next := route[i+1]
		curr := route[i]

		// Keep point if direction changes
		sameX := math.Abs(prev.X-curr.X) < 0.5 && math.Abs(curr.X-next.X) < 0.5
		sameY := math.Abs(prev.Y-curr.Y) < 0.5 && math.Abs(curr.Y-next.Y) < 0.5

		if !sameX && !sameY {
			result = append(result, curr)
		} else if !sameX && !sameY {
			result = append(result, curr)
		} else if sameX || sameY {
			// Collinear, skip
			continue
		} else {
			result = append(result, curr)
		}
	}
	result = append(result, route[len(route)-1])

	// Ensure we have at least 3 points (D2 requirement)
	if len(result) == 2 {
		midX := (result[0].X + result[1].X) / 2
		midY := (result[0].Y + result[1].Y) / 2
		result = []*geo.Point{result[0], geo.NewPoint(midX, midY), result[1]}
	}

	return result
}

// fallbackRoute creates a direct L shape route outside the grid boundary when A* fails.
func fallbackRoute(srcPt, dstPt *geo.Point, srcAnchor, dstAnchor grid.Anchor, rg *routingGrid) [][2]int {
	// Find boundary coordinates (first and last in xs/ys)
	srcXI := findIndex(rg.xs, srcPt.X)
	srcYI := findIndex(rg.ys, srcPt.Y)
	dstXI := findIndex(rg.xs, dstPt.X)
	dstYI := findIndex(rg.ys, dstPt.Y)

	srcVertical := isVerticalAnchor(srcAnchor)
	dstVertical := isVerticalAnchor(dstAnchor)

	if srcVertical && !dstVertical {
		// L shape: vertical from src, horizontal to dst
		// Use dst Y level as the bend
		return [][2]int{
			{srcXI, srcYI},
			{srcXI, dstYI},
			{dstXI, dstYI},
		}
	}
	if !srcVertical && dstVertical {
		// L shape: horizontal from src, vertical to dst
		return [][2]int{
			{srcXI, srcYI},
			{dstXI, srcYI},
			{dstXI, dstYI},
		}
	}

	// Both vertical or both horizontal: use boundary for U shape
	if srcVertical {
		// Pick a boundary Y (top or bottom of grid)
		boundYI := 0
		if isBottomAnchor(srcAnchor) {
			boundYI = len(rg.ys) - 1
		}
		return [][2]int{
			{srcXI, srcYI},
			{srcXI, boundYI},
			{dstXI, boundYI},
			{dstXI, dstYI},
		}
	}
	// Both horizontal
	boundXI := 0
	if isRightAnchor(srcAnchor) {
		boundXI = len(rg.xs) - 1
	}
	return [][2]int{
		{srcXI, srcYI},
		{boundXI, srcYI},
		{boundXI, dstYI},
		{dstXI, dstYI},
	}
}
