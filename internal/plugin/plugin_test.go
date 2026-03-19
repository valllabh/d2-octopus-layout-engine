package plugin

import (
	"context"
	"testing"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/lib/geo"
)

func TestInfo(t *testing.T) {
	p := New()
	info, err := p.Info(context.Background())
	if err != nil {
		t.Fatalf("Info() error: %v", err)
	}
	if info.Name != "octopus" {
		t.Errorf("Name = %q, want %q", info.Name, "octopus")
	}
	if info.Type != "binary" {
		t.Errorf("Type = %q, want %q", info.Type, "binary")
	}
	if len(info.Features) != 2 {
		t.Errorf("Features count = %d, want 2", len(info.Features))
	}
}

func TestFlags(t *testing.T) {
	p := New()
	flags, err := p.Flags(context.Background())
	if err != nil {
		t.Fatalf("Flags() error: %v", err)
	}
	if len(flags) != 6 {
		t.Errorf("Flags count = %d, want 6", len(flags))
	}
	names := make(map[string]bool)
	for _, f := range flags {
		names[f.Name] = true
	}
	for _, expected := range []string{"octopus-cell-width", "octopus-cell-height", "octopus-gap", "octopus-padding", "octopus-align", "octopus-anchor"} {
		if !names[expected] {
			t.Errorf("missing flag %q", expected)
		}
	}
}

func TestHydrateOpts(t *testing.T) {
	p := New()
	err := p.HydrateOpts([]byte(`{"octopus-cell-width": 300, "octopus-gap": 20}`))
	if err != nil {
		t.Fatalf("HydrateOpts() error: %v", err)
	}
	if p.opts.CellWidth != 300 {
		t.Errorf("CellWidth = %d, want 300", p.opts.CellWidth)
	}
	if p.opts.Gap != 20 {
		t.Errorf("Gap = %d, want 20", p.opts.Gap)
	}
	// Unchanged defaults
	if p.opts.CellHeight != 120 {
		t.Errorf("CellHeight = %d, want 120 (default)", p.opts.CellHeight)
	}
}

func TestHydrateOptsEmpty(t *testing.T) {
	p := New()
	if err := p.HydrateOpts(nil); err != nil {
		t.Fatalf("HydrateOpts(nil) error: %v", err)
	}
	if err := p.HydrateOpts([]byte{}); err != nil {
		t.Fatalf("HydrateOpts(empty) error: %v", err)
	}
}

func TestLayoutBasic(t *testing.T) {
	p := New()
	g := &d2graph.Graph{
		Root: &d2graph.Object{
			ID: "root",
		},
	}

	node1 := &d2graph.Object{
		ID:    "server",
		Graph: g,
		Box:   geo.NewBox(geo.NewPoint(0, 0), 100, 60),
	}
	node1.Attributes.Classes = []string{"row-1-col-1"}

	node2 := &d2graph.Object{
		ID:    "db",
		Graph: g,
		Box:   geo.NewBox(geo.NewPoint(0, 0), 100, 60),
	}
	node2.Attributes.Classes = []string{"row-2-col-3"}

	g.Objects = []*d2graph.Object{node1, node2}
	g.Root.ChildrenArray = g.Objects

	err := p.Layout(context.Background(), g)
	if err != nil {
		t.Fatalf("Layout() error: %v", err)
	}

	// node1 at row 1, col 1: cell top left (60,60), node 100x60 centered in 200x120 cell
	// x = 60 + (200-100)/2 = 110, y = 60 + (120-60)/2 = 90
	if node1.Box.TopLeft.X != 110 {
		t.Errorf("node1 x = %v, want 110", node1.Box.TopLeft.X)
	}
	if node1.Box.TopLeft.Y != 90 {
		t.Errorf("node1 y = %v, want 90", node1.Box.TopLeft.Y)
	}

	// node2 at row 2, col 3: cell top left (540,220), centered
	// x = 540 + 50 = 590, y = 220 + 30 = 250
	if node2.Box.TopLeft.X != 590 {
		t.Errorf("node2 x = %v, want 590", node2.Box.TopLeft.X)
	}
	if node2.Box.TopLeft.Y != 250 {
		t.Errorf("node2 y = %v, want 250", node2.Box.TopLeft.Y)
	}
}

func TestLayoutAutoPlace(t *testing.T) {
	p := New()
	g := &d2graph.Graph{
		Root: &d2graph.Object{ID: "root"},
	}

	// node1 explicitly placed
	node1 := &d2graph.Object{
		ID:    "a",
		Graph: g,
		Box:   geo.NewBox(geo.NewPoint(0, 0), 100, 60),
	}
	node1.Attributes.Classes = []string{"row-1-col-1"}

	// node2 has no grid class, should be auto placed
	node2 := &d2graph.Object{
		ID:    "b",
		Graph: g,
		Box:   geo.NewBox(geo.NewPoint(0, 0), 100, 60),
	}

	g.Objects = []*d2graph.Object{node1, node2}

	err := p.Layout(context.Background(), g)
	if err != nil {
		t.Fatalf("Layout() error: %v", err)
	}

	// node2 should be auto placed at row 1, col 2 (next available)
	// cell x = 300, centered: 300 + (200-100)/2 = 350
	expectedX := float64(350)
	if node2.Box.TopLeft.X != expectedX {
		t.Errorf("auto placed node x = %v, want %v", node2.Box.TopLeft.X, expectedX)
	}
}

func TestLayoutConflict(t *testing.T) {
	p := New()
	g := &d2graph.Graph{
		Root: &d2graph.Object{ID: "root"},
	}

	node1 := &d2graph.Object{
		ID:    "a",
		Graph: g,
		Box:   geo.NewBox(geo.NewPoint(0, 0), 100, 60),
	}
	node1.Attributes.Classes = []string{"row-1-col-1"}

	node2 := &d2graph.Object{
		ID:    "b",
		Graph: g,
		Box:   geo.NewBox(geo.NewPoint(0, 0), 100, 60),
	}
	node2.Attributes.Classes = []string{"row-1-col-1"}

	g.Objects = []*d2graph.Object{node1, node2}

	err := p.Layout(context.Background(), g)
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
}

func TestPostProcess(t *testing.T) {
	p := New()
	input := []byte("<svg>test</svg>")
	output, err := p.PostProcess(context.Background(), input)
	if err != nil {
		t.Fatalf("PostProcess() error: %v", err)
	}
	if string(output) != string(input) {
		t.Errorf("PostProcess() modified SVG, expected passthrough")
	}
}

func TestLayoutContainerWithNilChildBoxes(t *testing.T) {
	// Regression: layoutContainer panicked when children had nil Box or nil TopLeft
	p := New()
	g := &d2graph.Graph{Root: &d2graph.Object{ID: "root"}}

	parent := &d2graph.Object{ID: "parent", Graph: g}
	child1 := &d2graph.Object{ID: "child1", Graph: g, Box: nil} // nil Box
	child2 := &d2graph.Object{ID: "child2", Graph: g, Box: geo.NewBox(geo.NewPoint(100, 100), 80, 40)}
	child2.Attributes.Classes = []string{"row-1-col-1"}

	parent.ChildrenArray = append(parent.ChildrenArray, child1, child2)
	g.Objects = append(g.Objects, parent, child1, child2)

	err := p.Layout(context.Background(), g)
	if err != nil {
		t.Fatalf("Layout() with nil child boxes should not crash: %v", err)
	}
	if parent.Box == nil {
		t.Error("parent container should have a Box after layout")
	}
}

func TestLayoutContainerAllNilChildren(t *testing.T) {
	// Container where ALL children have nil boxes should get default size
	p := New()
	g := &d2graph.Graph{Root: &d2graph.Object{ID: "root"}}

	parent := &d2graph.Object{ID: "parent", Graph: g}
	child := &d2graph.Object{ID: "child", Graph: g, Box: nil}
	parent.ChildrenArray = append(parent.ChildrenArray, child)
	g.Objects = append(g.Objects, parent, child)

	err := p.Layout(context.Background(), g)
	if err != nil {
		t.Fatalf("Layout() error: %v", err)
	}
	if parent.Box == nil {
		t.Error("container with all nil children should get a default Box")
	}
}
