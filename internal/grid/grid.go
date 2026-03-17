package grid

import (
	"fmt"
	"regexp"
	"strconv"
)

var gridClassPattern = regexp.MustCompile(`^row-(\d+)-col-(\d+)$`)

// Position represents a grid coordinate.
type Position struct {
	Row int
	Col int
}

// Alignment controls how nodes are positioned within their grid cell.
type Alignment string

const (
	AlignCenter       Alignment = "center"
	AlignTopLeft      Alignment = "top-left"
	AlignTopCenter    Alignment = "top-center"
	AlignTopRight     Alignment = "top-right"
	AlignLeftCenter   Alignment = "left-center"
	AlignRightCenter  Alignment = "right-center"
	AlignBottomLeft   Alignment = "bottom-left"
	AlignBottomCenter Alignment = "bottom-center"
	AlignBottomRight  Alignment = "bottom-right"
)

// Options configures the grid to pixel transformation.
type Options struct {
	CellWidth  int
	CellHeight int
	Gap        int
	Padding    int
	Align      Alignment
	Anchor     Anchor
}

// DefaultOptions returns sensible default grid options.
func DefaultOptions() Options {
	return Options{
		CellWidth:  200,
		CellHeight: 120,
		Gap:        40,
		Padding:    60,
		Align:      AlignCenter,
		Anchor:     AnchorCenter,
	}
}

// PixelPosition computes the top left pixel coordinates for a grid position.
func (o Options) PixelPosition(pos Position) (x float64, y float64) {
	x = float64((pos.Col-1)*(o.CellWidth+o.Gap) + o.Padding)
	y = float64((pos.Row-1)*(o.CellHeight+o.Gap) + o.Padding)
	return x, y
}

// GapCenterY returns the Y coordinate at the center of the horizontal gap
// between row r and row r+1. This is where horizontal edge segments should run.
func (o Options) GapCenterY(row int) float64 {
	// Bottom of row = padding + row * (cellHeight + gap) - gap
	// But simpler: top of row+1 minus half the gap
	topOfNextRow := float64(row*(o.CellHeight+o.Gap) + o.Padding)
	return topOfNextRow - float64(o.Gap)/2
}

// GapCenterX returns the X coordinate at the center of the vertical gap
// between col c and col c+1. This is where vertical edge segments should run.
func (o Options) GapCenterX(col int) float64 {
	leftOfNextCol := float64(col*(o.CellWidth+o.Gap) + o.Padding)
	return leftOfNextCol - float64(o.Gap)/2
}

// Anchor controls which point of the shape is used as the reference point.
// Each edge of the shape has 5 named anchor points evenly spaced.
// Naming: {edge}-{position} where position is 1 through 5 from left to right (top/bottom)
// or top to bottom (left/right).
// Shortcuts: {edge}-center = {edge}-3, {corner} = corner point.
type Anchor string

const (
	// Center
	AnchorCenter Anchor = "center"

	// Corners
	AnchorTopLeft     Anchor = "top-left"
	AnchorTopRight    Anchor = "top-right"
	AnchorBottomLeft  Anchor = "bottom-left"
	AnchorBottomRight Anchor = "bottom-right"

	// Top edge: 5 points left to right
	AnchorTop1      Anchor = "top-1"
	AnchorTop2      Anchor = "top-2"
	AnchorTopCenter Anchor = "top-center" // same as top-3
	AnchorTop3      Anchor = "top-3"
	AnchorTop4      Anchor = "top-4"
	AnchorTop5      Anchor = "top-5"

	// Bottom edge: 5 points left to right
	AnchorBottom1      Anchor = "bottom-1"
	AnchorBottom2      Anchor = "bottom-2"
	AnchorBottomCenter Anchor = "bottom-center" // same as bottom-3
	AnchorBottom3      Anchor = "bottom-3"
	AnchorBottom4      Anchor = "bottom-4"
	AnchorBottom5      Anchor = "bottom-5"

	// Left edge: 5 points top to bottom
	AnchorLeft1      Anchor = "left-1"
	AnchorLeft2      Anchor = "left-2"
	AnchorLeftCenter Anchor = "left-center" // same as left-3
	AnchorLeft3      Anchor = "left-3"
	AnchorLeft4      Anchor = "left-4"
	AnchorLeft5      Anchor = "left-5"

	// Right edge: 5 points top to bottom
	AnchorRight1      Anchor = "right-1"
	AnchorRight2      Anchor = "right-2"
	AnchorRightCenter Anchor = "right-center" // same as right-3
	AnchorRight3      Anchor = "right-3"
	AnchorRight4      Anchor = "right-4"
	AnchorRight5      Anchor = "right-5"
)

// NodeStyle holds per node alignment and anchor overrides.
type NodeStyle struct {
	Align  Alignment // where in the cell to target
	Anchor Anchor    // which point of the shape to place there
}

// ParseClasses extracts a grid Position from a list of D2 classes.
func ParseClasses(classes []string) (Position, bool) {
	p, _, found := ParseClassesFull(classes)
	return p, found
}

// ParseClassesFull extracts grid Position, alignment, and anchor from D2 classes.
func ParseClassesFull(classes []string) (Position, NodeStyle, bool) {
	var pos Position
	var style NodeStyle
	found := false

	for _, class := range classes {
		// Check for grid position
		matches := gridClassPattern.FindStringSubmatch(class)
		if matches != nil {
			row, err1 := strconv.Atoi(matches[1])
			col, err2 := strconv.Atoi(matches[2])
			if err1 == nil && err2 == nil && row >= 1 && col >= 1 {
				pos = Position{Row: row, Col: col}
				found = true
			}
			continue
		}
		// Check for alignment (where in cell)
		switch class {
		case "align-center":
			style.Align = AlignCenter
		case "align-top-left":
			style.Align = AlignTopLeft
		case "align-top-center":
			style.Align = AlignTopCenter
		case "align-top-right":
			style.Align = AlignTopRight
		case "align-left-center":
			style.Align = AlignLeftCenter
		case "align-right-center":
			style.Align = AlignRightCenter
		case "align-bottom-left":
			style.Align = AlignBottomLeft
		case "align-bottom-center":
			style.Align = AlignBottomCenter
		case "align-bottom-right":
			style.Align = AlignBottomRight
		}
		// Check for anchor (which point of shape)
		if len(class) > 7 && class[:7] == "anchor-" {
			style.Anchor = Anchor(class[7:])
		}
	}
	return pos, style, found
}

// EdgeAnchors holds per edge source and destination anchor overrides.
type EdgeAnchors struct {
	SrcAnchor Anchor // where the edge exits the source shape
	DstAnchor Anchor // where the edge enters the destination shape
}

// ParseEdgeClasses extracts source and destination anchors from edge D2 classes.
// Classes: src-anchor-{name}, dst-anchor-{name}
// where {name} is any valid anchor name (center, top-left, top-1, right-4, etc.)
func ParseEdgeClasses(classes []string) EdgeAnchors {
	var anchors EdgeAnchors
	for _, class := range classes {
		if len(class) > 11 && class[:11] == "src-anchor-" {
			anchors.SrcAnchor = Anchor(class[11:])
		} else if len(class) > 11 && class[:11] == "dst-anchor-" {
			anchors.DstAnchor = Anchor(class[11:])
		}
	}
	return anchors
}

// AnchorPoint computes the pixel coordinates of an anchor point on a box.
func AnchorPoint(a Anchor, boxX, boxY, boxW, boxH float64) (float64, float64) {
	px, py := resolveAnchor(a, boxX, boxY, boxW, boxH)
	return px, py
}

func resolveAnchor(a Anchor, x, y, w, h float64) (float64, float64) {
	// Edge fraction helper: position 1..5 maps to 1/6..5/6 of edge length
	edgeFrac := func(pos int, length float64) float64 {
		return length * float64(pos) / 6.0
	}

	switch a {
	// Center
	case AnchorCenter:
		return x + w/2, y + h/2

	// Corners
	case AnchorTopLeft:
		return x, y
	case AnchorTopRight:
		return x + w, y
	case AnchorBottomLeft:
		return x, y + h
	case AnchorBottomRight:
		return x + w, y + h

	// Top edge (5 points, left to right)
	case AnchorTop1:
		return x + edgeFrac(1, w), y
	case AnchorTop2:
		return x + edgeFrac(2, w), y
	case AnchorTopCenter, AnchorTop3:
		return x + w/2, y
	case AnchorTop4:
		return x + edgeFrac(4, w), y
	case AnchorTop5:
		return x + edgeFrac(5, w), y

	// Bottom edge (5 points, left to right)
	case AnchorBottom1:
		return x + edgeFrac(1, w), y + h
	case AnchorBottom2:
		return x + edgeFrac(2, w), y + h
	case AnchorBottomCenter, AnchorBottom3:
		return x + w/2, y + h
	case AnchorBottom4:
		return x + edgeFrac(4, w), y + h
	case AnchorBottom5:
		return x + edgeFrac(5, w), y + h

	// Left edge (5 points, top to bottom)
	case AnchorLeft1:
		return x, y + edgeFrac(1, h)
	case AnchorLeft2:
		return x, y + edgeFrac(2, h)
	case AnchorLeftCenter, AnchorLeft3:
		return x, y + h/2
	case AnchorLeft4:
		return x, y + edgeFrac(4, h)
	case AnchorLeft5:
		return x, y + edgeFrac(5, h)

	// Right edge (5 points, top to bottom)
	case AnchorRight1:
		return x + w, y + edgeFrac(1, h)
	case AnchorRight2:
		return x + w, y + edgeFrac(2, h)
	case AnchorRightCenter, AnchorRight3:
		return x + w, y + h/2
	case AnchorRight4:
		return x + w, y + edgeFrac(4, h)
	case AnchorRight5:
		return x + w, y + edgeFrac(5, h)

	default:
		return x + w/2, y + h/2
	}
}

// ValidatePositions checks for duplicate grid positions and returns an error if found.
func ValidatePositions(positions map[string]Position) error {
	seen := make(map[Position]string)
	for id, pos := range positions {
		if existing, ok := seen[pos]; ok {
			return fmt.Errorf("grid conflict: nodes %q and %q both at row %d, col %d", existing, id, pos.Row, pos.Col)
		}
		seen[pos] = id
	}
	return nil
}
