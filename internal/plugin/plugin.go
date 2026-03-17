package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2plugin"
	"oss.terrastruct.com/d2/lib/geo"

	"github.com/valllabh/octopus-layout-engine/internal/grid"
)

// OctopusPlugin implements the d2plugin.Plugin interface for grid based layout.
type OctopusPlugin struct {
	opts grid.Options
}

// New creates a new OctopusPlugin with default options.
func New() *OctopusPlugin {
	return &OctopusPlugin{
		opts: grid.DefaultOptions(),
	}
}

func (p *OctopusPlugin) Info(_ context.Context) (*d2plugin.PluginInfo, error) {
	return &d2plugin.PluginInfo{
		Name:      "octopus",
		ShortHelp: "Grid based layout engine with explicit node positioning",
		LongHelp: `Octopus places nodes at user defined grid coordinates.
Add a class like "row-2-col-3" to any node to place it at row 2, column 3.
Nodes without grid classes are auto placed in the next available cell.
Edge routing is handled by D2.`,
		Type: "binary",
		Features: []d2plugin.PluginFeature{
			d2plugin.TOP_LEFT,
			d2plugin.CONTAINER_DIMENSIONS,
		},
	}, nil
}

func (p *OctopusPlugin) Flags(_ context.Context) ([]d2plugin.PluginSpecificFlag, error) {
	return []d2plugin.PluginSpecificFlag{
		{
			Name:    "octopus-cell-width",
			Type:    "int",
			Default: interface{}(200),
			Usage:   "Width of each grid cell in pixels",
			Tag:     "",
		},
		{
			Name:    "octopus-cell-height",
			Type:    "int",
			Default: interface{}(120),
			Usage:   "Height of each grid cell in pixels",
			Tag:     "",
		},
		{
			Name:    "octopus-gap",
			Type:    "int",
			Default: interface{}(40),
			Usage:   "Gap between grid cells in pixels",
			Tag:     "",
		},
		{
			Name:    "octopus-padding",
			Type:    "int",
			Default: interface{}(60),
			Usage:   "Padding around the entire grid in pixels",
			Tag:     "",
		},
		{
			Name:    "octopus-align",
			Type:    "string",
			Default: interface{}("center"),
			Usage:   "Cell alignment point: center, top-left, top-center, top-right, left-center, right-center, bottom-left, bottom-center, bottom-right",
			Tag:     "",
		},
		{
			Name:    "octopus-anchor",
			Type:    "string",
			Default: interface{}("center"),
			Usage:   "Shape anchor point: center, top-left, top-center, top-right, left-center, right-center, bottom-left, bottom-center, bottom-right",
			Tag:     "",
		},
	}, nil
}

func (p *OctopusPlugin) HydrateOpts(opts []byte) error {
	if len(opts) == 0 {
		return nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(opts, &raw); err != nil {
		return fmt.Errorf("octopus: failed to parse options: %w", err)
	}
	if v, ok := raw["octopus-cell-width"]; ok {
		p.opts.CellWidth = toInt(v, p.opts.CellWidth)
	}
	if v, ok := raw["octopus-cell-height"]; ok {
		p.opts.CellHeight = toInt(v, p.opts.CellHeight)
	}
	if v, ok := raw["octopus-gap"]; ok {
		p.opts.Gap = toInt(v, p.opts.Gap)
	}
	if v, ok := raw["octopus-padding"]; ok {
		p.opts.Padding = toInt(v, p.opts.Padding)
	}
	if v, ok := raw["octopus-align"]; ok {
		if s, isStr := v.(string); isStr {
			p.opts.Align = grid.Alignment(s)
		}
	}
	if v, ok := raw["octopus-anchor"]; ok {
		if s, isStr := v.(string); isStr {
			p.opts.Anchor = grid.Anchor(s)
		}
	}
	return nil
}

func (p *OctopusPlugin) Layout(_ context.Context, g *d2graph.Graph) error {
	positions := make(map[string]grid.Position)
	nodeStyles := make(map[string]grid.NodeStyle)
	occupied := make(map[grid.Position]bool)

	// First pass: extract explicit grid positions, alignment, and anchor from classes.
	for _, obj := range g.Objects {
		if obj.IsContainer() {
			continue
		}
		pos, style, found := grid.ParseClassesFull(obj.Attributes.Classes)
		if !found {
			continue
		}
		positions[obj.AbsID()] = pos
		if style.Align != "" || style.Anchor != "" {
			nodeStyles[obj.AbsID()] = style
		}
		occupied[pos] = true
	}

	// Validate no conflicts among explicit positions.
	if err := grid.ValidatePositions(positions); err != nil {
		return err
	}

	// Second pass: auto place nodes without explicit positions.
	autoRow, autoCol := 1, 1
	for _, obj := range g.Objects {
		if obj.IsContainer() {
			continue
		}
		if _, hasPos := positions[obj.AbsID()]; hasPos {
			continue
		}
		for occupied[grid.Position{Row: autoRow, Col: autoCol}] {
			autoCol++
			if autoCol > 50 {
				autoCol = 1
				autoRow++
			}
		}
		pos := grid.Position{Row: autoRow, Col: autoCol}
		positions[obj.AbsID()] = pos
		occupied[pos] = true
	}

	// Third pass: set pixel positions on all non container objects.
	// Nodes are centered within their grid cell.
	cellW := float64(p.opts.CellWidth)
	cellH := float64(p.opts.CellHeight)

	for _, obj := range g.Objects {
		if obj.IsContainer() {
			continue
		}
		pos, ok := positions[obj.AbsID()]
		if !ok {
			continue
		}
		// cellX, cellY is the top left corner of the grid cell
		cellX, cellY := p.opts.PixelPosition(pos)

		// Use the node's own dimensions if available, otherwise use cell size
		width := cellW
		height := cellH
		if obj.Box != nil {
			if obj.Box.Width > 0 {
				width = obj.Box.Width
			}
			if obj.Box.Height > 0 {
				height = obj.Box.Height
			}
		}

		// Determine alignment and anchor: per node override or global default
		align := p.opts.Align
		anchor := p.opts.Anchor
		if ns, ok := nodeStyles[obj.AbsID()]; ok {
			if ns.Align != "" {
				align = ns.Align
			}
			if ns.Anchor != "" {
				anchor = ns.Anchor
			}
		}

		x, y := positionInCell(align, anchor, cellX, cellY, cellW, cellH, width, height)

		obj.Box = geo.NewBox(geo.NewPoint(x, y), width, height)

		lp := "INSIDE_MIDDLE_CENTER"
		obj.LabelPosition = &lp
	}

	// Fourth pass: expand containers to fit their children.
	for _, obj := range g.Objects {
		if !obj.IsContainer() {
			continue
		}
		p.layoutContainer(obj)
	}

	// Fifth pass: detect shape collisions (overlapping boxes).
	if err := detectCollisions(g.Objects); err != nil {
		return err
	}

	// Sixth pass: route edges with auto distributed anchors and obstacle avoidance.
	routeEdges(g.Edges, g.Objects)

	return nil
}

// positionInCell computes the node top left position within a cell.
// align = where in the cell to place the anchor point.
// anchor = which point of the shape is the anchor.
func positionInCell(align grid.Alignment, anchor grid.Anchor, cellX, cellY, cellW, cellH, nodeW, nodeH float64) (float64, float64) {
	// Step 1: compute the target point in the cell based on alignment
	targetX, targetY := resolvePoint(grid.Alignment(align), cellX, cellY, cellW, cellH)

	// Step 2: compute the offset from node top left to the anchor point
	anchorOffX, anchorOffY := resolvePoint(grid.Alignment(anchor), 0, 0, nodeW, nodeH)

	// Step 3: position node so that anchor point lands on target point
	return targetX - anchorOffX, targetY - anchorOffY
}

// resolvePoint computes a point within a rectangle based on a named position.
// Works for both cell alignment and shape anchor since both use the same 9 positions.
func resolvePoint(pos grid.Alignment, x, y, w, h float64) (float64, float64) {
	switch pos {
	case grid.AlignTopLeft:
		return x, y
	case grid.AlignTopCenter:
		return x + w/2, y
	case grid.AlignTopRight:
		return x + w, y
	case grid.AlignLeftCenter:
		return x, y + h/2
	case grid.AlignRightCenter:
		return x + w, y + h/2
	case grid.AlignBottomLeft:
		return x, y + h
	case grid.AlignBottomCenter:
		return x + w/2, y + h
	case grid.AlignBottomRight:
		return x + w, y + h
	default: // center
		return x + w/2, y + h/2
	}
}

func (p *OctopusPlugin) layoutContainer(container *d2graph.Object) {
	if len(container.ChildrenArray) == 0 {
		// Empty container: give it a default size at origin.
		if container.Box == nil {
			container.Box = geo.NewBox(geo.NewPoint(0, 0), float64(p.opts.CellWidth), float64(p.opts.CellHeight))
		}
		return
	}

	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := float64(0), float64(0)

	for _, child := range container.ChildrenArray {
		if child.Box == nil {
			continue
		}
		cx := child.Box.TopLeft.X
		cy := child.Box.TopLeft.Y
		cw := child.Box.Width
		ch := child.Box.Height

		if cx < minX {
			minX = cx
		}
		if cy < minY {
			minY = cy
		}
		if cx+cw > maxX {
			maxX = cx + cw
		}
		if cy+ch > maxY {
			maxY = cy + ch
		}
	}

	containerPad := float64(p.opts.Padding) / 2
	container.Box = geo.NewBox(
		geo.NewPoint(minX-containerPad, minY-containerPad),
		(maxX-minX)+2*containerPad,
		(maxY-minY)+2*containerPad,
	)

	lp := "OUTSIDE_TOP_CENTER"
	container.LabelPosition = &lp
}

// detectCollisions checks if any two positioned nodes physically overlap.
func detectCollisions(objects []*d2graph.Object) error {
	type rect struct {
		id                 string
		x, y, right, bottom float64
	}
	var rects []rect
	for _, obj := range objects {
		if obj.Box == nil || obj.IsContainer() {
			continue
		}
		rects = append(rects, rect{
			id:     obj.AbsID(),
			x:      obj.Box.TopLeft.X,
			y:      obj.Box.TopLeft.Y,
			right:  obj.Box.TopLeft.X + obj.Box.Width,
			bottom: obj.Box.TopLeft.Y + obj.Box.Height,
		})
	}
	for i := 0; i < len(rects); i++ {
		for j := i + 1; j < len(rects); j++ {
			a, b := rects[i], rects[j]
			if a.x < b.right && a.right > b.x && a.y < b.bottom && a.bottom > b.y {
				return fmt.Errorf("shape collision: %q and %q overlap", a.id, b.id)
			}
		}
	}
	return nil
}

// sideAnchors returns the 5 anchor names for a given side, ordered 1 through 5.
func sideAnchors(s side) [5]grid.Anchor {
	switch s {
	case sideTop:
		return [5]grid.Anchor{grid.AnchorTop1, grid.AnchorTop2, grid.AnchorTop3, grid.AnchorTop4, grid.AnchorTop5}
	case sideBottom:
		return [5]grid.Anchor{grid.AnchorBottom1, grid.AnchorBottom2, grid.AnchorBottom3, grid.AnchorBottom4, grid.AnchorBottom5}
	case sideLeft:
		return [5]grid.Anchor{grid.AnchorLeft1, grid.AnchorLeft2, grid.AnchorLeft3, grid.AnchorLeft4, grid.AnchorLeft5}
	case sideRight:
		return [5]grid.Anchor{grid.AnchorRight1, grid.AnchorRight2, grid.AnchorRight3, grid.AnchorRight4, grid.AnchorRight5}
	}
	return [5]grid.Anchor{grid.AnchorCenter, grid.AnchorCenter, grid.AnchorCenter, grid.AnchorCenter, grid.AnchorCenter}
}

// distributeAnchors picks N anchors from a side spread evenly across the 5 slots.
// For 1 edge: slot 3 (center). For 2: slots 2,4. For 3: slots 2,3,4. For 4: slots 1,2,4,5. For 5: all.
func distributeAnchors(count int) []int {
	switch count {
	case 1:
		return []int{2} // 0-indexed position 2 = slot 3 (center)
	case 2:
		return []int{0, 4} // slots 1, 5 (widest spread)
	case 3:
		return []int{0, 2, 4} // slots 1, 3, 5 (even spread)
	case 4:
		return []int{0, 1, 3, 4} // slots 1, 2, 4, 5
	default:
		return []int{0, 1, 2, 3, 4} // all 5
	}
}

// edgeEndpoint identifies one end of an edge at a specific node and side.
type edgeEndpoint struct {
	nodeID string
	side   side
}

// routeEdges routes all edges with automatic anchor distribution.
// Edges connecting to the same side of the same node are automatically spread
// across different anchor points. Manual anchor overrides take priority.
func routeEdges(edges []*d2graph.Edge, allObjects []*d2graph.Object) {
	// Phase 1: Categorize all edge endpoints by (nodeID, side) to count how many
	// edges share each side of each node.
	type edgeInfo struct {
		edge        *d2graph.Edge
		srcSide     side
		dstSide     side
		edgeAnchors grid.EdgeAnchors
		isSelfLoop  bool
	}

	// Collect obstacle boxes for routing
	var obstacles []*geo.Box
	for _, obj := range allObjects {
		if obj.Box != nil && !obj.IsContainer() {
			obstacles = append(obstacles, obj.Box)
		}
	}

	infos := make([]edgeInfo, 0, len(edges))
	sideCount := make(map[edgeEndpoint]int)
	sideIndex := make(map[edgeEndpoint]int)

	for _, edge := range edges {
		if edge.Src == nil || edge.Dst == nil {
			continue
		}
		if edge.Src.Box == nil || edge.Dst.Box == nil {
			continue
		}

		srcID := edge.Src.AbsID()
		dstID := edge.Dst.AbsID()
		anchors := grid.ParseEdgeClasses(edge.Attributes.Classes)

		if srcID == dstID {
			infos = append(infos, edgeInfo{edge: edge, isSelfLoop: true, edgeAnchors: anchors})
			continue
		}

		srcBox := edge.Src.Box
		dstBox := edge.Dst.Box

		ss, ds := chooseSides(srcBox, dstBox, obstacles)

		info := edgeInfo{
			edge:        edge,
			srcSide:     ss,
			dstSide:     ds,
			edgeAnchors: anchors,
		}
		infos = append(infos, info)

		// Only count for auto distribution if no manual anchor override
		if anchors.SrcAnchor == "" {
			sideCount[edgeEndpoint{srcID, ss}]++
		}
		if anchors.DstAnchor == "" {
			sideCount[edgeEndpoint{dstID, ds}]++
		}
	}

	// Phase 2: Pre compute anchor distributions for each (node, side) group.
	distributions := make(map[edgeEndpoint][]int)
	for ep, count := range sideCount {
		distributions[ep] = distributeAnchors(count)
	}

	// Phase 3: Route each edge using distributed anchors.
	for _, info := range infos {
		if info.isSelfLoop {
			routeSelfLoop(info.edge, info.edge.Src.Box)
			setEdgeLabel(info.edge)
			continue
		}

		edge := info.edge
		srcBox := edge.Src.Box
		dstBox := edge.Dst.Box
		srcID := edge.Src.AbsID()
		dstID := edge.Dst.AbsID()

		// Resolve source anchor
		var srcAnchor grid.Anchor
		if info.edgeAnchors.SrcAnchor != "" {
			srcAnchor = info.edgeAnchors.SrcAnchor
		} else {
			srcEP := edgeEndpoint{srcID, info.srcSide}
			idx := sideIndex[srcEP]
			sideIndex[srcEP] = idx + 1
			slots := distributions[srcEP]
			slotIdx := slots[idx%len(slots)]
			srcAnchor = sideAnchors(info.srcSide)[slotIdx]
		}

		// Resolve destination anchor
		var dstAnchor grid.Anchor
		if info.edgeAnchors.DstAnchor != "" {
			dstAnchor = info.edgeAnchors.DstAnchor
		} else {
			dstEP := edgeEndpoint{dstID, info.dstSide}
			idx := sideIndex[dstEP]
			sideIndex[dstEP] = idx + 1
			slots := distributions[dstEP]
			slotIdx := slots[idx%len(slots)]
			dstAnchor = sideAnchors(info.dstSide)[slotIdx]
		}

		routeEdgeWithAnchors(edge, srcBox, dstBox, grid.EdgeAnchors{
			SrcAnchor: srcAnchor,
			DstAnchor: dstAnchor,
		}, allObjects)
		setEdgeLabel(edge)
	}
}

type side int

const (
	sideTop    side = 0
	sideBottom side = 1
	sideLeft   side = 2
	sideRight  side = 3
)

// chooseSides picks the best exit/entry side pair by generating candidate routes
// and scoring them based on obstacle hits and bend count. Picks the cleanest route.
func chooseSides(srcBox, dstBox *geo.Box, obstacles []*geo.Box) (srcSide, dstSide side) {
	dx := (dstBox.TopLeft.X + dstBox.Width/2) - (srcBox.TopLeft.X + srcBox.Width/2)
	dy := (dstBox.TopLeft.Y + dstBox.Height/2) - (srcBox.TopLeft.Y + srcBox.Height/2)

	type candidate struct {
		src, dst side
		score    int // lower is better
	}

	// Generate all sensible L shape candidates based on direction
	var candidates []candidate

	// Horizontal sides based on dx direction
	hSrc, hDst := sideRight, sideLeft
	if dx < 0 {
		hSrc, hDst = sideLeft, sideRight
	}
	// Vertical sides based on dy direction
	vSrc, vDst := sideBottom, sideTop
	if dy < 0 {
		vSrc, vDst = sideTop, sideBottom
	}

	if math.Abs(dy) < 5 {
		// Same row: horizontal is primary, vertical as fallback
		candidates = append(candidates,
			candidate{hSrc, hDst, 0},
			candidate{vSrc, vSrc, 10}, // go around via below/above
		)
	} else if math.Abs(dx) < 5 {
		// Same column: vertical is primary
		candidates = append(candidates,
			candidate{vSrc, vDst, 0},
		)
	} else {
		// Diagonal: strongly prefer mixed axis L shapes (1 bend) over same axis U shapes (2 bends)
		// L1: horizontal exit, vertical entry
		candidates = append(candidates, candidate{hSrc, vDst, 0})
		// L2: vertical exit, horizontal entry
		candidates = append(candidates, candidate{vSrc, hDst, 0})
		// L3: horizontal both (2 bends, fallback only)
		candidates = append(candidates, candidate{hSrc, hDst, 50})
		// L4: vertical both (2 bends, fallback only)
		candidates = append(candidates, candidate{vSrc, vDst, 50})
	}

	// Score each candidate: check obstacles on the L path
	best := candidates[0]
	bestScore := math.MaxInt32

	for _, c := range candidates {
		srcPt := sideCenter(srcBox, c.src)
		dstPt := sideCenter(dstBox, c.dst)
		score := c.score

		obstacleCount := countObstaclesOnLPath(srcPt, dstPt, srcBox, dstBox, obstacles)
		score += obstacleCount * 100

		if score < bestScore {
			bestScore = score
			best = c
		}
	}

	return best.src, best.dst
}

func sideCenter(box *geo.Box, s side) *geo.Point {
	cx := box.TopLeft.X + box.Width/2
	cy := box.TopLeft.Y + box.Height/2
	switch s {
	case sideTop:
		return geo.NewPoint(cx, box.TopLeft.Y)
	case sideBottom:
		return geo.NewPoint(cx, box.TopLeft.Y+box.Height)
	case sideLeft:
		return geo.NewPoint(box.TopLeft.X, cy)
	case sideRight:
		return geo.NewPoint(box.TopLeft.X+box.Width, cy)
	}
	return geo.NewPoint(cx, cy)
}

// countObstaclesOnLPath counts how many obstacles an L shaped path would cross.
// For mixed axis sides (src horizontal, dst vertical or vice versa), there is only
// one natural L bend point. For same axis sides, checks the bend near destination.
func countObstaclesOnLPath(srcPt, dstPt *geo.Point, srcBox, dstBox *geo.Box, obstacles []*geo.Box) int {
	// Try both L orientations and return the better one
	paths := []*geo.Point{
		geo.NewPoint(dstPt.X, srcPt.Y), // bend option 1: horizontal first
		geo.NewPoint(srcPt.X, dstPt.Y), // bend option 2: vertical first
	}

	bestCount := math.MaxInt32
	for _, mid := range paths {
		count := 0
		for _, obs := range obstacles {
			if obs == srcBox || obs == dstBox {
				continue
			}
			if segmentIntersectsBox(srcPt, mid, obs) || segmentIntersectsBox(mid, dstPt, obs) {
				count++
			}
		}
		if count < bestCount {
			bestCount = count
		}
	}
	return bestCount
}


// routeEdgeWithAnchors creates a route using anchor points on source and destination.
// Uses gap based bending to avoid routing through intermediate boxes.
func routeEdgeWithAnchors(edge *d2graph.Edge, srcBox, dstBox *geo.Box, anchors grid.EdgeAnchors, allObjects []*d2graph.Object) {
	// Resolve source point
	sx, sy := grid.AnchorPoint(anchors.SrcAnchor, srcBox.TopLeft.X, srcBox.TopLeft.Y, srcBox.Width, srcBox.Height)
	srcPt := geo.NewPoint(sx, sy)

	// Resolve destination point
	dx, dy := grid.AnchorPoint(anchors.DstAnchor, dstBox.TopLeft.X, dstBox.TopLeft.Y, dstBox.Width, dstBox.Height)
	dstPt := geo.NewPoint(dx, dy)

	srcVertical := isVerticalAnchor(anchors.SrcAnchor)
	dstVertical := isVerticalAnchor(anchors.DstAnchor)

	// Check if boxes are in the same column or row (centers within half cell dimension).
	// This determines if the route should try to be straight despite anchor offsets.
	srcCX := srcBox.TopLeft.X + srcBox.Width/2
	srcCY := srcBox.TopLeft.Y + srcBox.Height/2
	dstCX := dstBox.TopLeft.X + dstBox.Width/2
	dstCY := dstBox.TopLeft.Y + dstBox.Height/2
	sameColumn := math.Abs(srcCX-dstCX) < 5
	sameRow := math.Abs(srcCY-dstCY) < 5

	if sameColumn && srcVertical && dstVertical {
		// Same column, both vertical anchors: force straight vertical using the average X
		avgX := (srcPt.X + dstPt.X) / 2
		midY := (srcPt.Y + dstPt.Y) / 2
		edge.Route = []*geo.Point{
			geo.NewPoint(avgX, srcPt.Y),
			geo.NewPoint(avgX, midY),
			geo.NewPoint(avgX, dstPt.Y),
		}
	} else if sameRow && !srcVertical && !dstVertical {
		// Same row, both horizontal anchors: force straight horizontal using the average Y
		avgY := (srcPt.Y + dstPt.Y) / 2
		midX := (srcPt.X + dstPt.X) / 2
		edge.Route = []*geo.Point{
			geo.NewPoint(srcPt.X, avgY),
			geo.NewPoint(midX, avgY),
			geo.NewPoint(dstPt.X, avgY),
		}
	} else if math.Abs(srcPt.X-dstPt.X) < 1 {
		// Vertically aligned anchors: straight line with midpoint
		midY := (srcPt.Y + dstPt.Y) / 2
		edge.Route = []*geo.Point{srcPt, geo.NewPoint(srcPt.X, midY), dstPt}
	} else if math.Abs(srcPt.Y-dstPt.Y) < 1 {
		// Horizontally aligned anchors: straight line with midpoint
		midX := (srcPt.X + dstPt.X) / 2
		edge.Route = []*geo.Point{srcPt, geo.NewPoint(midX, srcPt.Y), dstPt}
	} else if srcVertical && dstVertical {
		// Both exit/enter vertically (top/bottom): bend near the destination.
		// Routing near the destination avoids crossing intermediate boxes
		// because the horizontal segment is adjacent to the entry point.
		var bendY float64
		if srcPt.Y < dstPt.Y {
			// Going down: bend just above destination
			bendY = dstBox.TopLeft.Y - 20
		} else {
			// Going up: bend just below destination
			bendY = dstBox.TopLeft.Y + dstBox.Height + 20
		}
		edge.Route = []*geo.Point{
			srcPt,
			geo.NewPoint(srcPt.X, bendY),
			geo.NewPoint(dstPt.X, bendY),
			dstPt,
		}
	} else if !srcVertical && !dstVertical {
		// Both exit/enter horizontally (left/right): bend in gap between boxes.
		var bendX float64
		if srcPt.X < dstPt.X {
			// Going right: bend in gap between src right and dst left
			srcRight := srcBox.TopLeft.X + srcBox.Width
			dstLeft := dstBox.TopLeft.X
			bendX = (srcRight + dstLeft) / 2
		} else {
			// Going left: bend in gap between dst right and src left
			dstRight := dstBox.TopLeft.X + dstBox.Width
			srcLeft := srcBox.TopLeft.X
			bendX = (dstRight + srcLeft) / 2
		}
		edge.Route = []*geo.Point{
			srcPt,
			geo.NewPoint(bendX, srcPt.Y),
			geo.NewPoint(bendX, dstPt.Y),
			dstPt,
		}
	} else if srcVertical {
		// Source exits vertically, destination enters horizontally: L shape
		edge.Route = []*geo.Point{
			srcPt,
			geo.NewPoint(srcPt.X, dstPt.Y),
			dstPt,
		}
	} else {
		// Source exits horizontally, destination enters vertically: L shape
		edge.Route = []*geo.Point{
			srcPt,
			geo.NewPoint(dstPt.X, srcPt.Y),
			dstPt,
		}
	}
	edge.IsCurve = false

	// Post process: if route crosses any obstacle, reroute with a clean U shape
	edge.Route = rerouteAroundObstacles(edge.Route, srcBox, dstBox, allObjects)

	// Enforce perpendicular contact at both anchors.
	// After rerouting, the first/last segments may no longer be perpendicular.
	edge.Route = enforcePerpendicularAnchors(edge.Route, anchors.SrcAnchor, anchors.DstAnchor)
}

// enforcePerpendicularAnchors ensures the first and last segments of a route
// are perpendicular to the shape edge at the anchor point.
// If the first segment is not perpendicular, a short stub is inserted.
func enforcePerpendicularAnchors(route []*geo.Point, srcAnchor, dstAnchor grid.Anchor) []*geo.Point {
	if len(route) < 2 {
		return route
	}

	stubLen := 20.0

	// Check source: first segment must be perpendicular to source edge
	p0, p1 := route[0], route[1]
	srcVert := isVerticalAnchor(srcAnchor)
	firstIsVert := math.Abs(p0.X-p1.X) < 1
	firstIsHoriz := math.Abs(p0.Y-p1.Y) < 1

	if srcVert && !firstIsVert && !firstIsHoriz {
		// Anchor on top/bottom but first segment is diagonal. Insert vertical stub.
		stubY := p0.Y - stubLen
		if isBottomAnchor(srcAnchor) {
			stubY = p0.Y + stubLen
		}
		stub := geo.NewPoint(p0.X, stubY)
		route = append([]*geo.Point{p0, stub}, route[1:]...)
	} else if !srcVert && !firstIsHoriz && !firstIsVert {
		// Anchor on left/right but first segment is diagonal. Insert horizontal stub.
		stubX := p0.X - stubLen
		if isRightAnchor(srcAnchor) {
			stubX = p0.X + stubLen
		}
		stub := geo.NewPoint(stubX, p0.Y)
		route = append([]*geo.Point{p0, stub}, route[1:]...)
	} else if srcVert && firstIsHoriz {
		// Anchor on top/bottom but first segment goes horizontal. Insert vertical stub.
		stubY := p0.Y - stubLen
		if isBottomAnchor(srcAnchor) {
			stubY = p0.Y + stubLen
		}
		stub := geo.NewPoint(p0.X, stubY)
		route = append([]*geo.Point{p0, stub}, route[1:]...)
	} else if !srcVert && firstIsVert {
		// Anchor on left/right but first segment goes vertical. Insert horizontal stub.
		stubX := p0.X - stubLen
		if isRightAnchor(srcAnchor) {
			stubX = p0.X + stubLen
		}
		stub := geo.NewPoint(stubX, p0.Y)
		route = append([]*geo.Point{p0, stub}, route[1:]...)
	}

	// Check destination: last segment must be perpendicular to dest edge
	pN := route[len(route)-1]
	pN1 := route[len(route)-2]
	dstVert := isVerticalAnchor(dstAnchor)
	lastIsVert := math.Abs(pN.X-pN1.X) < 1
	lastIsHoriz := math.Abs(pN.Y-pN1.Y) < 1

	if dstVert && !lastIsVert && !lastIsHoriz {
		stubY := pN.Y + stubLen
		if isBottomAnchor(dstAnchor) {
			stubY = pN.Y - stubLen
		}
		stub := geo.NewPoint(pN.X, stubY)
		route = append(route[:len(route)-1], stub, pN)
	} else if !dstVert && !lastIsHoriz && !lastIsVert {
		stubX := pN.X + stubLen
		if isRightAnchor(dstAnchor) {
			stubX = pN.X - stubLen
		}
		stub := geo.NewPoint(stubX, pN.Y)
		route = append(route[:len(route)-1], stub, pN)
	} else if dstVert && lastIsHoriz {
		stubY := pN.Y + stubLen
		if isBottomAnchor(dstAnchor) {
			stubY = pN.Y - stubLen
		}
		stub := geo.NewPoint(pN.X, stubY)
		route = append(route[:len(route)-1], stub, pN)
	} else if !dstVert && lastIsVert {
		stubX := pN.X + stubLen
		if isRightAnchor(dstAnchor) {
			stubX = pN.X - stubLen
		}
		stub := geo.NewPoint(stubX, pN.Y)
		route = append(route[:len(route)-1], stub, pN)
	}

	return route
}

func isBottomAnchor(a grid.Anchor) bool {
	switch a {
	case grid.AnchorBottomLeft, grid.AnchorBottomCenter, grid.AnchorBottomRight,
		grid.AnchorBottom1, grid.AnchorBottom2, grid.AnchorBottom3, grid.AnchorBottom4, grid.AnchorBottom5:
		return true
	}
	return false
}

func isRightAnchor(a grid.Anchor) bool {
	switch a {
	case grid.AnchorTopRight, grid.AnchorBottomRight, grid.AnchorRightCenter,
		grid.AnchorRight1, grid.AnchorRight2, grid.AnchorRight3, grid.AnchorRight4, grid.AnchorRight5:
		return true
	}
	return false
}

// rerouteAroundObstacles checks if any route segment crosses an obstacle.
// If so, it replaces the entire route with a clean U shape detour that goes
// through the gap below or above all obstacles.
func rerouteAroundObstacles(route []*geo.Point, srcBox, dstBox *geo.Box, allObjects []*d2graph.Object) []*geo.Point {
	if len(route) < 2 {
		return route
	}

	var obstacles []*geo.Box
	for _, obj := range allObjects {
		if obj.Box == nil || obj.IsContainer() {
			continue
		}
		if obj.Box == srcBox || obj.Box == dstBox {
			continue
		}
		obstacles = append(obstacles, obj.Box)
	}

	if len(obstacles) == 0 {
		return route
	}

	// Check if any segment hits an obstacle
	hasHit := false
	for i := 1; i < len(route); i++ {
		for _, obs := range obstacles {
			if segmentIntersectsBox(route[i-1], route[i], obs) {
				hasHit = true
				break
			}
		}
		if hasHit {
			break
		}
	}

	if !hasHit {
		return route
	}

	// Route is blocked. Find only the obstacles that the route actually crosses.
	srcPt := route[0]
	dstPt := route[len(route)-1]

	var hitObstacles []*geo.Box
	for _, obs := range obstacles {
		for i := 1; i < len(route); i++ {
			if segmentIntersectsBox(route[i-1], route[i], obs) {
				hitObstacles = append(hitObstacles, obs)
				break
			}
		}
	}

	if len(hitObstacles) == 0 {
		return route
	}

	// Find bounding box of ONLY the hit obstacles
	minY, maxY := math.MaxFloat64, -math.MaxFloat64
	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	for _, obs := range hitObstacles {
		if obs.TopLeft.Y < minY {
			minY = obs.TopLeft.Y
		}
		if obs.TopLeft.X < minX {
			minX = obs.TopLeft.X
		}
		b := obs.TopLeft.Y + obs.Height
		r := obs.TopLeft.X + obs.Width
		if b > maxY {
			maxY = b
		}
		if r > maxX {
			maxX = r
		}
	}

	// Determine if the route is primarily horizontal or vertical
	isMainlyHorizontal := math.Abs(srcPt.X-dstPt.X) > math.Abs(srcPt.Y-dstPt.Y)

	if isMainlyHorizontal {
		// Detour above or below the hit obstacles
		detourAbove := minY - 25
		detourBelow := maxY + 25

		// Pick whichever is closer to the route midpoint
		midY := (srcPt.Y + dstPt.Y) / 2
		detourY := detourBelow
		if math.Abs(detourAbove-midY) < math.Abs(detourBelow-midY) {
			detourY = detourAbove
		}

		return []*geo.Point{
			srcPt,
			geo.NewPoint(srcPt.X, detourY),
			geo.NewPoint(dstPt.X, detourY),
			dstPt,
		}
	}

	// Mainly vertical: detour left or right of the hit obstacles
	detourLeft := minX - 25
	detourRight := maxX + 25

	midX := (srcPt.X + dstPt.X) / 2
	detourX := detourRight
	if math.Abs(detourLeft-midX) < math.Abs(detourRight-midX) {
		detourX = detourLeft
	}

	return []*geo.Point{
		srcPt,
		geo.NewPoint(detourX, srcPt.Y),
		geo.NewPoint(detourX, dstPt.Y),
		dstPt,
	}
}

// segmentIntersectsBox checks if a horizontal or vertical line segment passes through a box.
func segmentIntersectsBox(p1, p2 *geo.Point, box *geo.Box) bool {
	bx1 := box.TopLeft.X
	by1 := box.TopLeft.Y
	bx2 := bx1 + box.Width
	by2 := by1 + box.Height

	// Horizontal segment
	if math.Abs(p1.Y-p2.Y) < 1 {
		y := p1.Y
		if y < by1 || y > by2 {
			return false
		}
		minX := math.Min(p1.X, p2.X)
		maxX := math.Max(p1.X, p2.X)
		return maxX > bx1 && minX < bx2
	}

	// Vertical segment
	if math.Abs(p1.X-p2.X) < 1 {
		x := p1.X
		if x < bx1 || x > bx2 {
			return false
		}
		minY := math.Min(p1.Y, p2.Y)
		maxY := math.Max(p1.Y, p2.Y)
		return maxY > by1 && minY < by2
	}

	return false
}

// isVerticalAnchor returns true if the anchor is on the top or bottom edge of the shape.
func isVerticalAnchor(a grid.Anchor) bool {
	switch a {
	case grid.AnchorTopLeft, grid.AnchorTopCenter, grid.AnchorTopRight,
		grid.AnchorTop1, grid.AnchorTop2, grid.AnchorTop3, grid.AnchorTop4, grid.AnchorTop5,
		grid.AnchorBottomLeft, grid.AnchorBottomCenter, grid.AnchorBottomRight,
		grid.AnchorBottom1, grid.AnchorBottom2, grid.AnchorBottom3, grid.AnchorBottom4, grid.AnchorBottom5:
		return true
	}
	return false
}

// routeSelfLoop creates a loop route for an edge where src == dst.
func routeSelfLoop(edge *d2graph.Edge, box *geo.Box) {
	right := box.TopLeft.X + box.Width
	top := box.TopLeft.Y
	cy := top + box.Height/2
	loopSize := 40.0

	// Loop out the right side and back in from the top
	edge.Route = []*geo.Point{
		geo.NewPoint(right, cy-10),                     // exit right side, above center
		geo.NewPoint(right+loopSize, cy-10),            // go right
		geo.NewPoint(right+loopSize, top-loopSize),     // go up
		geo.NewPoint(right-box.Width/4, top-loopSize),  // go left
		geo.NewPoint(right-box.Width/4, top),           // enter from top
	}
	edge.IsCurve = false
}

func setEdgeLabel(edge *d2graph.Edge) {
	if edge.Attributes.Label.Value == "" {
		return
	}
	lp := "INSIDE_MIDDLE_CENTER"
	edge.LabelPosition = &lp

	// Place label at the midpoint of the longest segment in the route.
	// This avoids placing labels at bend points or near nodes.
	if len(edge.Route) < 2 {
		pct := 0.5
		edge.LabelPercentage = &pct
		return
	}

	// Calculate segment lengths and total length
	type segment struct {
		length     float64
		cumBefore  float64 // cumulative length before this segment starts
	}
	segments := make([]segment, 0, len(edge.Route)-1)
	totalLen := 0.0

	for i := 1; i < len(edge.Route); i++ {
		dx := edge.Route[i].X - edge.Route[i-1].X
		dy := edge.Route[i].Y - edge.Route[i-1].Y
		segLen := math.Sqrt(dx*dx + dy*dy)
		segments = append(segments, segment{length: segLen, cumBefore: totalLen})
		totalLen += segLen
	}

	if totalLen < 1 {
		pct := 0.5
		edge.LabelPercentage = &pct
		return
	}

	// Find the longest segment
	longestIdx := 0
	longestLen := 0.0
	for i, seg := range segments {
		if seg.length > longestLen {
			longestLen = seg.length
			longestIdx = i
		}
	}

	// Place label at midpoint of the longest segment
	midOfLongest := segments[longestIdx].cumBefore + segments[longestIdx].length/2
	pct := midOfLongest / totalLen
	edge.LabelPercentage = &pct
}

func (p *OctopusPlugin) PostProcess(_ context.Context, svg []byte) ([]byte, error) {
	return svg, nil
}

func toInt(v interface{}, fallback int) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case json.Number:
		n, err := val.Int64()
		if err != nil {
			return fallback
		}
		return int(n)
	default:
		return fallback
	}
}
