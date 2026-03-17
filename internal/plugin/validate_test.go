package plugin

import (
	"context"
	"fmt"
	"math"
	"testing"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/lib/geo"
)

// validateRoute checks a single edge route against quality rules.
// Returns a list of violations.
func validateRoute(edge *d2graph.Edge, allObjects []*d2graph.Object) []string {
	var issues []string
	if edge.Src == nil || edge.Dst == nil || len(edge.Route) < 2 {
		return issues
	}

	srcBox := edge.Src.Box
	dstBox := edge.Dst.Box
	srcID := edge.Src.AbsID()
	dstID := edge.Dst.AbsID()

	// R1: No edge through node
	for i := 1; i < len(edge.Route); i++ {
		for _, obj := range allObjects {
			if obj.Box == nil || obj.IsContainer() {
				continue
			}
			if obj.Box == srcBox || obj.Box == dstBox {
				continue
			}
			if segmentIntersectsBox(edge.Route[i-1], edge.Route[i], obj.Box) {
				issues = append(issues, fmt.Sprintf("R1 CRITICAL: %s->%s segment %d passes through %s", srcID, dstID, i, obj.AbsID()))
			}
		}
	}

	// R2: Perpendicular anchor contact
	// First segment must be perpendicular to source side
	if len(edge.Route) >= 2 {
		p0, p1 := edge.Route[0], edge.Route[1]
		if math.Abs(p0.X-p1.X) > 1 && math.Abs(p0.Y-p1.Y) > 1 {
			issues = append(issues, fmt.Sprintf("R2 CRITICAL: %s->%s first segment is diagonal (not perpendicular)", srcID, dstID))
		}
	}
	// Last segment must be perpendicular to destination side
	if len(edge.Route) >= 2 {
		pN := edge.Route[len(edge.Route)-1]
		pN1 := edge.Route[len(edge.Route)-2]
		if math.Abs(pN.X-pN1.X) > 1 && math.Abs(pN.Y-pN1.Y) > 1 {
			issues = append(issues, fmt.Sprintf("R2 CRITICAL: %s->%s last segment is diagonal (not perpendicular)", srcID, dstID))
		}
	}

	// R5: Minimum bends
	bends := len(edge.Route) - 2 // N points = N-2 direction changes (bends) for orthogonal
	// Actually count real direction changes
	realBends := 0
	for i := 1; i < len(edge.Route)-1; i++ {
		prev := edge.Route[i-1]
		curr := edge.Route[i]
		next := edge.Route[i+1]
		prevHoriz := math.Abs(prev.Y-curr.Y) < 1
		nextHoriz := math.Abs(curr.Y-next.Y) < 1
		if prevHoriz != nextHoriz {
			realBends++
		}
	}
	_ = bends

	// Determine minimum expected bends
	srcCX := srcBox.TopLeft.X + srcBox.Width/2
	srcCY := srcBox.TopLeft.Y + srcBox.Height/2
	dstCX := dstBox.TopLeft.X + dstBox.Width/2
	dstCY := dstBox.TopLeft.Y + dstBox.Height/2
	sameCol := math.Abs(srcCX-dstCX) < 5
	sameRow := math.Abs(srcCY-dstCY) < 5

	var minBends int
	if sameCol || sameRow {
		minBends = 0
	} else {
		// Diagonal: check for obstacles on L path
		hasObstacle := false
		for _, obj := range allObjects {
			if obj.Box == nil || obj.IsContainer() || obj.Box == srcBox || obj.Box == dstBox {
				continue
			}
			// Check if obj is between src and dst on the grid
			ox := obj.Box.TopLeft.X + obj.Box.Width/2
			oy := obj.Box.TopLeft.Y + obj.Box.Height/2
			betweenX := (ox > math.Min(srcCX, dstCX)-10) && (ox < math.Max(srcCX, dstCX)+10)
			betweenY := (oy > math.Min(srcCY, dstCY)-10) && (oy < math.Max(srcCY, dstCY)+10)
			if betweenX && betweenY {
				hasObstacle = true
				break
			}
		}
		if hasObstacle {
			minBends = 2
		} else {
			minBends = 1
		}
	}

	if realBends > minBends+1 { // Allow 1 extra bend for rerouting
		issues = append(issues, fmt.Sprintf("R5 MAJOR: %s->%s has %d bends, expected at most %d", srcID, dstID, realBends, minBends+1))
	}

	return issues
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

	allIssues := 0
	for _, edge := range g.Edges {
		issues := validateRoute(edge, g.Objects)
		for _, issue := range issues {
			t.Errorf("%s: %s", name, issue)
			allIssues++
		}
	}
	if allIssues == 0 {
		t.Logf("%s: PASS (all %d edges clean)", name, len(g.Edges))
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
	nodes := map[string][2]int{"sched": {1, 1}, "worker": {1, 3}, "queue": {2, 2}}
	edges := [][2]string{{"sched", "queue"}, {"queue", "worker"}, {"worker", "queue"}, {"sched", "sched"}}
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
		t.Fatal(err)
	}
	// Skip self-loop edge for route validation (different rules)
	allIssues := 0
	for _, edge := range g.Edges {
		if edge.Src == edge.Dst {
			continue
		}
		issues := validateRoute(edge, g.Objects)
		for _, issue := range issues {
			t.Errorf("21-self-loop: %s", issue)
			allIssues++
		}
	}
	if allIssues == 0 {
		t.Logf("21-self-loop: PASS")
	}
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
