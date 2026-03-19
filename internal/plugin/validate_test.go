package plugin

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/lib/geo"
)

type finding struct {
	rule     string
	severity string // CRITICAL, MAJOR, MINOR, INFO
	edge     string
	detail   string
}

func validateAllRoutes(g *d2graph.Graph) []finding {
	var findings []finding

	for _, edge := range g.Edges {
		if edge.Src == nil || edge.Dst == nil || len(edge.Route) < 2 {
			continue
		}
		if edge.Src == edge.Dst {
			continue // skip self loops for now
		}

		srcBox := edge.Src.Box
		dstBox := edge.Dst.Box
		edgeName := fmt.Sprintf("%s->%s", edge.Src.AbsID(), edge.Dst.AbsID())

		// R1: No edge through node (CRITICAL)
		for i := 1; i < len(edge.Route); i++ {
			for _, obj := range g.Objects {
				if obj.Box == nil || obj.IsContainer() || obj.Box == srcBox || obj.Box == dstBox {
					continue
				}
				if segmentIntersectsBox(edge.Route[i-1], edge.Route[i], obj.Box) {
					findings = append(findings, finding{"R1", "CRITICAL", edgeName,
						fmt.Sprintf("segment %d passes through %s", i, obj.AbsID())})
				}
			}
		}

		// R2: Perpendicular anchor contact (CRITICAL)
		p0, p1 := edge.Route[0], edge.Route[1]
		if math.Abs(p0.X-p1.X) > 1 && math.Abs(p0.Y-p1.Y) > 1 {
			findings = append(findings, finding{"R2", "CRITICAL", edgeName, "first segment is diagonal"})
		}
		pN := edge.Route[len(edge.Route)-1]
		pN1 := edge.Route[len(edge.Route)-2]
		if math.Abs(pN.X-pN1.X) > 1 && math.Abs(pN.Y-pN1.Y) > 1 {
			findings = append(findings, finding{"R2", "CRITICAL", edgeName, "last segment is diagonal"})
		}

		// R5: Bend analysis (sentiment based)
		realBends := countBends(edge.Route)
		srcCX := srcBox.TopLeft.X + srcBox.Width/2
		srcCY := srcBox.TopLeft.Y + srcBox.Height/2
		dstCX := dstBox.TopLeft.X + dstBox.Width/2
		dstCY := dstBox.TopLeft.Y + dstBox.Height/2
		sameCol := math.Abs(srcCX-dstCX) < 5
		sameRow := math.Abs(srcCY-dstCY) < 5

		idealBends := 0
		if !sameCol && !sameRow {
			idealBends = 1
			// Check for obstacles on L path
			if hasObstacleBetween(srcBox, dstBox, g.Objects) {
				idealBends = 2
			}
		}

		extraBends := realBends - idealBends
		if extraBends > 0 {
			severity := "INFO"
			if extraBends >= 2 {
				severity = "MAJOR"
			}
			findings = append(findings, finding{"R5", severity, edgeName,
				fmt.Sprintf("%d bends (ideal %d, %d unnecessary)", realBends, idealBends, extraBends)})
		}

		// R13: Clearance from non connected nodes (MAJOR)
		for i := 1; i < len(edge.Route); i++ {
			for _, obj := range g.Objects {
				if obj.Box == nil || obj.IsContainer() || obj.Box == srcBox || obj.Box == dstBox {
					continue
				}
				clearance := segmentClearance(edge.Route[i-1], edge.Route[i], obj.Box)
				if clearance >= 0 && clearance < 8 {
					findings = append(findings, finding{"R13", "MAJOR", edgeName,
						fmt.Sprintf("segment %d has only %.0fpx clearance from %s", i, clearance, obj.AbsID())})
				}
			}
		}
	}

	// R12: Anchor distribution
	type sideKey struct {
		nodeID string
		side   string
	}
	anchorPoints := make(map[sideKey][]float64)

	for _, edge := range g.Edges {
		if edge.Src == nil || edge.Dst == nil || len(edge.Route) < 2 || edge.Src == edge.Dst {
			continue
		}
		// Source exit point
		p0 := edge.Route[0]
		srcBox := edge.Src.Box
		srcSide := detectSide(p0, srcBox)
		key := sideKey{edge.Src.AbsID(), srcSide}
		anchorPoints[key] = append(anchorPoints[key], anchorOffset(p0, srcBox, srcSide))

		// Destination entry point
		pN := edge.Route[len(edge.Route)-1]
		dstBox := edge.Dst.Box
		dstSide := detectSide(pN, dstBox)
		key = sideKey{edge.Dst.AbsID(), dstSide}
		anchorPoints[key] = append(anchorPoints[key], anchorOffset(pN, dstBox, dstSide))
	}

	for key, offsets := range anchorPoints {
		if len(offsets) < 2 {
			continue
		}
		// Check minimum gap between anchor points
		for i := 0; i < len(offsets); i++ {
			for j := i + 1; j < len(offsets); j++ {
				gap := math.Abs(offsets[i] - offsets[j])
				if gap < 10 {
					findings = append(findings, finding{"R12", "MAJOR", key.nodeID,
						fmt.Sprintf("%s side has %d anchors with %.0fpx gap (need 10+)", key.side, len(offsets), gap)})
					goto nextKey
				}
			}
		}
	nextKey:
	}

	return findings
}

func countBends(route []*geo.Point) int {
	bends := 0
	for i := 1; i < len(route)-1; i++ {
		prev := route[i-1]
		curr := route[i]
		next := route[i+1]
		prevHoriz := math.Abs(prev.Y-curr.Y) < 1
		nextHoriz := math.Abs(curr.Y-next.Y) < 1
		if prevHoriz != nextHoriz {
			bends++
		}
	}
	return bends
}

func hasObstacleBetween(srcBox, dstBox *geo.Box, objects []*d2graph.Object) bool {
	srcCX := srcBox.TopLeft.X + srcBox.Width/2
	srcCY := srcBox.TopLeft.Y + srcBox.Height/2
	dstCX := dstBox.TopLeft.X + dstBox.Width/2
	dstCY := dstBox.TopLeft.Y + dstBox.Height/2
	for _, obj := range objects {
		if obj.Box == nil || obj.IsContainer() || obj.Box == srcBox || obj.Box == dstBox {
			continue
		}
		ox := obj.Box.TopLeft.X + obj.Box.Width/2
		oy := obj.Box.TopLeft.Y + obj.Box.Height/2
		betweenX := ox > math.Min(srcCX, dstCX)-10 && ox < math.Max(srcCX, dstCX)+10
		betweenY := oy > math.Min(srcCY, dstCY)-10 && oy < math.Max(srcCY, dstCY)+10
		if betweenX && betweenY {
			return true
		}
	}
	return false
}

// segmentClearance returns how close a segment comes to a box.
// Returns -1 if the segment does not come near the box at all.
func segmentClearance(p1, p2 *geo.Point, box *geo.Box) float64 {
	bx1 := box.TopLeft.X
	by1 := box.TopLeft.Y
	bx2 := bx1 + box.Width
	by2 := by1 + box.Height

	// Horizontal segment
	if math.Abs(p1.Y-p2.Y) < 1 {
		y := p1.Y
		minX := math.Min(p1.X, p2.X)
		maxX := math.Max(p1.X, p2.X)
		// Check if segment overlaps box X range
		if maxX < bx1 || minX > bx2 {
			return -1 // no X overlap
		}
		// Vertical distance from segment to box
		if y < by1 {
			return by1 - y
		} else if y > by2 {
			return y - by2
		}
		return 0 // inside box Y range (should be caught by R1)
	}

	// Vertical segment
	if math.Abs(p1.X-p2.X) < 1 {
		x := p1.X
		minY := math.Min(p1.Y, p2.Y)
		maxY := math.Max(p1.Y, p2.Y)
		if maxY < by1 || minY > by2 {
			return -1
		}
		if x < bx1 {
			return bx1 - x
		} else if x > bx2 {
			return x - bx2
		}
		return 0
	}

	return -1
}

func detectSide(pt *geo.Point, box *geo.Box) string {
	cx := box.TopLeft.X + box.Width/2
	cy := box.TopLeft.Y + box.Height/2
	dx := math.Abs(pt.X - cx)
	dy := math.Abs(pt.Y - cy)
	halfW := box.Width / 2
	halfH := box.Height / 2

	if math.Abs(pt.Y-box.TopLeft.Y) < 2 {
		return "top"
	}
	if math.Abs(pt.Y-(box.TopLeft.Y+box.Height)) < 2 {
		return "bottom"
	}
	if math.Abs(pt.X-box.TopLeft.X) < 2 {
		return "left"
	}
	if math.Abs(pt.X-(box.TopLeft.X+box.Width)) < 2 {
		return "right"
	}

	_ = dx
	_ = dy
	_ = halfW
	_ = halfH
	return "unknown"
}

func anchorOffset(pt *geo.Point, box *geo.Box, side string) float64 {
	switch side {
	case "top", "bottom":
		return pt.X - box.TopLeft.X
	case "left", "right":
		return pt.Y - box.TopLeft.Y
	}
	return 0
}

func layoutAndValidate(t *testing.T, name string, nodes map[string][2]int, edges [][2]string) {
	t.Helper()
	p := New()
	g := &d2graph.Graph{Root: &d2graph.Object{ID: "root"}}

	objs := make(map[string]*d2graph.Object)
	for id, pos := range nodes {
		o := &d2graph.Object{ID: id, Graph: g, Box: geo.NewBox(geo.NewPoint(0, 0), 100, 50)}
		o.Attributes.Classes = []string{fmt.Sprintf("row-%d-col-%d", pos[0], pos[1])}
		objs[id] = o
		g.Objects = append(g.Objects, o)
	}
	for _, e := range edges {
		g.Edges = append(g.Edges, &d2graph.Edge{Src: objs[e[0]], Dst: objs[e[1]]})
	}

	err := p.Layout(context.Background(), g)
	if err != nil {
		t.Fatalf("%s: Layout error: %v", name, err)
	}

	findings := validateAllRoutes(g)

	criticals := 0
	majors := 0
	infos := 0
	var report strings.Builder
	for _, f := range findings {
		report.WriteString(fmt.Sprintf("  [%s] %s %s: %s\n", f.severity, f.rule, f.edge, f.detail))
		switch f.severity {
		case "CRITICAL":
			criticals++
		case "MAJOR":
			majors++
		case "INFO":
			infos++
		}
	}

	if criticals > 0 {
		t.Errorf("%s: %d CRITICAL issues\n%s", name, criticals, report.String())
	}
	if majors > 0 {
		t.Errorf("%s: %d MAJOR issues\n%s", name, majors, report.String())
	}
	if infos > 0 {
		t.Logf("%s: %d edges, %d info notes (acceptable)\n%s", name, len(g.Edges), infos, report.String())
	}
	if criticals == 0 && majors == 0 && infos == 0 {
		t.Logf("%s: CLEAN (%d edges, 0 issues)", name, len(g.Edges))
	}
}

func TestValidate05FanOut(t *testing.T) {
	layoutAndValidate(t, "05-fan-out",
		map[string][2]int{"lb": {1, 2}, "s1": {2, 1}, "s2": {2, 2}, "s3": {2, 3}},
		[][2]string{{"lb", "s1"}, {"lb", "s2"}, {"lb", "s3"}})
}

func TestValidate07Diamond(t *testing.T) {
	layoutAndValidate(t, "07-diamond",
		map[string][2]int{"begin": {1, 2}, "pa": {2, 1}, "pb": {2, 3}, "merge": {3, 2}},
		[][2]string{{"begin", "pa"}, {"begin", "pb"}, {"pa", "merge"}, {"pb", "merge"}})
}

func TestValidate08Grid3x3(t *testing.T) {
	layoutAndValidate(t, "08-grid-3x3",
		map[string][2]int{
			"a": {1, 1}, "b": {1, 2}, "c": {1, 3},
			"d": {2, 1}, "e": {2, 2}, "f": {2, 3},
			"g": {3, 1}, "h": {3, 2}, "i": {3, 3},
		},
		[][2]string{{"a", "e"}, {"b", "e"}, {"c", "e"}, {"e", "g"}, {"e", "h"}, {"e", "i"}, {"d", "e"}, {"f", "e"}})
}

func TestValidate09Diagonal(t *testing.T) {
	layoutAndValidate(t, "09-diagonal",
		map[string][2]int{"tl": {1, 1}, "tr": {1, 3}, "bl": {3, 1}, "br": {3, 3}, "center": {2, 2}},
		[][2]string{{"tl", "br"}, {"tr", "bl"}, {"tl", "center"}, {"center", "br"}})
}

func TestValidate12Bidirectional(t *testing.T) {
	layoutAndValidate(t, "12-bidirectional",
		map[string][2]int{"client": {1, 1}, "server": {1, 3}, "cache": {2, 2}},
		[][2]string{{"client", "server"}, {"server", "client"}, {"client", "cache"}, {"cache", "client"}, {"server", "cache"}})
}

func TestValidate16Microservices(t *testing.T) {
	layoutAndValidate(t, "16-microservices",
		map[string][2]int{
			"app": {1, 2}, "gw": {2, 2},
			"auth": {3, 1}, "users": {3, 2}, "orders": {3, 3},
			"adb": {4, 1}, "udb": {4, 2}, "odb": {4, 3},
		},
		[][2]string{{"app", "gw"}, {"gw", "auth"}, {"gw", "users"}, {"gw", "orders"}, {"auth", "adb"}, {"users", "udb"}, {"orders", "odb"}})
}

func TestValidate19AutoPlacement(t *testing.T) {
	layoutAndValidate(t, "19-auto-placement",
		map[string][2]int{"pin1": {1, 1}, "pin2": {2, 2}},
		[][2]string{{"pin1", "auto1"}, {"pin2", "auto2"}, {"auto1", "auto3"}})
}

func TestValidate21SelfLoop(t *testing.T) {
	layoutAndValidate(t, "21-self-loop",
		map[string][2]int{"sched": {1, 1}, "worker": {1, 3}, "queue": {2, 2}},
		[][2]string{{"sched", "queue"}, {"queue", "worker"}, {"worker", "queue"}})
}

func TestValidate22Pipeline(t *testing.T) {
	layoutAndValidate(t, "22-pipeline",
		map[string][2]int{"ingest": {1, 1}, "transform": {1, 2}, "enrich": {1, 3}, "load": {1, 4}, "errors": {2, 2}, "dlq": {2, 3}},
		[][2]string{{"ingest", "transform"}, {"transform", "enrich"}, {"enrich", "load"}, {"transform", "errors"}, {"enrich", "errors"}, {"errors", "dlq"}})
}

func TestValidate25Cross(t *testing.T) {
	layoutAndValidate(t, "25-cross",
		map[string][2]int{"a": {1, 1}, "b": {1, 3}, "c": {3, 1}, "d": {3, 3}},
		[][2]string{{"a", "d"}, {"b", "c"}, {"a", "c"}, {"b", "d"}})
}

func TestValidate27Large5x5(t *testing.T) {
	layoutAndValidate(t, "27-large-5x5",
		map[string][2]int{
			"n1": {1, 1}, "n2": {1, 2}, "n3": {1, 3}, "n4": {1, 4}, "n5": {1, 5},
			"w1": {2, 1}, "e1": {2, 5},
			"w2": {3, 1}, "hub": {3, 3}, "e2": {3, 5},
			"w3": {4, 1}, "e3": {4, 5},
			"s1": {5, 1}, "s2": {5, 2}, "s3": {5, 3}, "s4": {5, 4}, "s5": {5, 5},
		},
		[][2]string{
			{"n1", "n2"}, {"n2", "n3"}, {"n3", "n4"}, {"n4", "n5"},
			{"n5", "e1"}, {"e1", "e2"}, {"e2", "e3"}, {"e3", "s5"},
			{"s5", "s4"}, {"s4", "s3"}, {"s3", "s2"}, {"s2", "s1"},
			{"s1", "w3"}, {"w3", "w2"}, {"w2", "w1"}, {"w1", "n1"},
			{"n3", "hub"}, {"w2", "hub"}, {"e2", "hub"}, {"s3", "hub"},
		})
}

func TestValidate28CICD(t *testing.T) {
	layoutAndValidate(t, "28-cicd",
		map[string][2]int{
			"commit": {1, 1}, "build": {1, 2}, "test": {1, 3}, "stage": {1, 4}, "prod": {1, 5},
			"lint": {2, 2}, "sec": {2, 3}, "monitor": {2, 5},
			"rollback": {3, 5},
		},
		[][2]string{
			{"commit", "build"}, {"build", "lint"}, {"build", "test"},
			{"lint", "test"}, {"test", "sec"}, {"sec", "stage"},
			{"stage", "prod"}, {"prod", "monitor"}, {"monitor", "rollback"},
			{"rollback", "stage"},
		})
}

func TestValidate29Mesh(t *testing.T) {
	layoutAndValidate(t, "29-mesh",
		map[string][2]int{
			"a": {1, 1}, "b": {1, 2}, "c": {1, 3},
			"d": {2, 1}, "e": {2, 2}, "f": {2, 3},
			"g": {3, 1}, "h": {3, 2}, "i": {3, 3},
		},
		[][2]string{
			{"a", "b"}, {"b", "c"}, {"d", "e"}, {"e", "f"}, {"g", "h"}, {"h", "i"},
			{"a", "d"}, {"d", "g"}, {"b", "e"}, {"e", "h"}, {"c", "f"}, {"f", "i"},
		})
}

func TestValidate30Styled(t *testing.T) {
	layoutAndValidate(t, "30-styled",
		map[string][2]int{
			"user": {1, 2}, "auth": {2, 1}, "api": {2, 2}, "cache": {2, 3},
			"db1": {3, 1}, "db2": {3, 2}, "db3": {3, 3},
			"report": {4, 2},
		},
		[][2]string{
			{"user", "auth"}, {"user", "api"}, {"api", "cache"},
			{"api", "db1"}, {"api", "db2"}, {"api", "db3"},
			{"db3", "report"}, {"auth", "db1"},
		})
}

func TestValidate31Harness(t *testing.T) {
	layoutAndValidate(t, "31-harness",
		map[string][2]int{
			"knowledge": {1, 2}, "governance": {1, 3},
			"user": {2, 1}, "interface": {2, 2}, "agency": {2, 3}, "model": {2, 4},
		},
		[][2]string{
			{"user", "interface"}, {"interface", "user"},
			{"interface", "agency"}, {"agency", "model"},
			{"governance", "knowledge"}, {"governance", "agency"},
			{"agency", "knowledge"},
		})
}

func TestValidate33Containers(t *testing.T) {
	p := New()
	g := &d2graph.Graph{Root: &d2graph.Object{ID: "root"}}

	frontend := &d2graph.Object{ID: "frontend", Graph: g}
	backend := &d2graph.Object{ID: "backend", Graph: g}

	web := &d2graph.Object{ID: "web", Graph: g, Box: geo.NewBox(geo.NewPoint(0, 0), 100, 50)}
	web.Attributes.Classes = []string{"row-1-col-1"}
	mobile := &d2graph.Object{ID: "mobile", Graph: g, Box: geo.NewBox(geo.NewPoint(0, 0), 100, 50)}
	mobile.Attributes.Classes = []string{"row-1-col-2"}
	api := &d2graph.Object{ID: "api", Graph: g, Box: geo.NewBox(geo.NewPoint(0, 0), 100, 50)}
	api.Attributes.Classes = []string{"row-2-col-1"}
	worker := &d2graph.Object{ID: "worker", Graph: g, Box: geo.NewBox(geo.NewPoint(0, 0), 100, 50)}
	worker.Attributes.Classes = []string{"row-2-col-2"}

	frontend.ChildrenArray = []*d2graph.Object{web, mobile}
	backend.ChildrenArray = []*d2graph.Object{api, worker}

	g.Objects = append(g.Objects, frontend, backend, web, mobile, api, worker)
	g.Edges = append(g.Edges,
		&d2graph.Edge{Src: web, Dst: api},
		&d2graph.Edge{Src: mobile, Dst: api},
		&d2graph.Edge{Src: api, Dst: worker},
	)

	err := p.Layout(context.Background(), g)
	if err != nil {
		t.Fatalf("Layout with containers should not error: %v", err)
	}
	if frontend.Box == nil {
		t.Error("frontend container should have a Box")
	}
	if backend.Box == nil {
		t.Error("backend container should have a Box")
	}
}
