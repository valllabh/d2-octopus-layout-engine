---
name: d2-octopus
description: Create D2 diagrams using the Octopus grid layout engine. Generates .d2 files with grid position classes, edge anchor control, and renders to SVG/PNG. Triggers on "octopus diagram", "grid diagram", "d2 octopus", "create octopus diagram", "layout diagram".
argument-hint: [description of the diagram to create]
allowed-tools:
  - Bash
  - Read
  - Write
  - Glob
  - Grep
---

# D2 Octopus Layout Engine Skill

You help users create D2 diagrams using the Octopus external layout engine, which provides explicit grid based node positioning with A* edge routing.

## Prerequisites

The user must have `d2` and `d2plugin-octopus` installed. If rendering fails, suggest:

```bash
go install github.com/valllabh/d2-octopus-layout-engine/cmd/d2plugin-octopus@latest
```

## How Octopus Works

Octopus places nodes on a grid using `row-N-col-M` classes (1 indexed). Row 1 is the top, column 1 is the left. The engine computes pixel positions and routes edges automatically with orthogonal paths that avoid obstacles.

Unlike D2's built in engines (dagre, elk, tala), Octopus gives the user full control over node placement. Every node goes exactly where specified.

## D2 File Structure

Every Octopus D2 file follows this pattern:

```d2
# 1. Declare all grid position classes used in this diagram
classes: {
  row-1-col-1: {style.opacity: 1}
  row-1-col-2: {style.opacity: 1}
  row-2-col-1: {style.opacity: 1}
  # ... declare every row-N-col-M combination used
}

# 2. Define nodes with grid positions
node-a: Label A {class: row-1-col-1}
node-b: Label B {class: row-1-col-2}
node-c: Label C {class: row-2-col-1}

# 3. Define edges
node-a -> node-b
node-a -> node-c
```

### Critical Rules

1. **Every grid class must be declared** in the `classes` block with `{style.opacity: 1}`. This is a D2 requirement. Without the declaration, D2 ignores the class.

2. **Grid coordinates are 1 indexed**. `row-1-col-1` is the top left cell.

3. **No two nodes can share the same grid position**. Each cell holds one node.

4. **Nodes without grid classes** are auto placed in the next available cell.

## Edge Anchor Control

By default, Octopus automatically selects which side of a shape an edge exits/enters and distributes multiple edges across 5 anchor points per side.

For precise control, use `src-anchor-*` and `dst-anchor-*` classes:

```d2
# Declare anchor classes
classes: {
  src-anchor-bottom-2: {style.opacity: 1}
  dst-anchor-top-4: {style.opacity: 1}
}

a -> b: {
  class: src-anchor-bottom-2
  class: dst-anchor-top-4
}
```

**IMPORTANT**: Use separate `class:` lines for edge anchors. D2 does NOT split space separated class values.

```d2
# WRONG: D2 treats this as one class name
a -> b: { class: src-anchor-bottom-2 dst-anchor-top-4 }

# CORRECT: separate class lines
a -> b: {
  class: src-anchor-bottom-2
  class: dst-anchor-top-4
}
```

### Anchor Point Reference

Each side has 5 named positions distributed at 1/6, 2/6, 3/6, 4/6, 5/6 of the edge length:

**Top/Bottom edge** (left to right): `top-1`, `top-2`, `top-3`, `top-4`, `top-5`
**Left/Right edge** (top to bottom): `left-1`, `left-2`, `left-3`, `left-4`, `left-5`
**Corners**: `top-left`, `top-right`, `bottom-left`, `bottom-right`
**Edge centers**: `top-center` (= top-3), `bottom-center`, `left-center`, `right-center`
**Shape center**: `center`

### When to Use Explicit Anchors

Use explicit anchors when:
- Multiple edges exit the same side and you want specific ordering
- You want to prevent overlapping lines by spreading edges across different anchor points
- The automatic distribution does not produce the desired visual result

For edges going through the same gap channel between shapes, use different anchor positions (e.g., `bottom-2` and `bottom-4`) to visually separate the lines.

## Node Alignment and Anchor

Override how a node sits within its grid cell:

```d2
classes: {
  align-top-left: {style.opacity: 1}
  anchor-bottom-center: {style.opacity: 1}
}

node: {
  class: row-1-col-1
  class: align-top-left
  class: anchor-bottom-center
}
```

- **align**: where in the cell to place the anchor point (center, top-left, top-center, etc.)
- **anchor**: which point of the shape sits at the alignment target

## Plugin Flags

Users can customize grid sizing:

| Flag | Default | Description |
|------|---------|-------------|
| `octopus-cell-width` | 200 | Cell width in pixels |
| `octopus-cell-height` | 120 | Cell height in pixels |
| `octopus-gap` | 40 | Gap between cells |
| `octopus-padding` | 60 | Padding around the grid |

```bash
d2 --layout=octopus --octopus-gap=60 --octopus-cell-width=250 diagram.d2 output.svg
```

## Rendering

Always render with `--layout=octopus`:

```bash
# SVG output
d2 --layout=octopus diagram.d2 output.svg

# PNG output
d2 --layout=octopus diagram.d2 output.png
```

## Design Guidelines

When creating diagrams:

1. **Plan the grid first**. Sketch which row and column each node should occupy before writing D2 code.

2. **Use rows for layers**. Put related components on the same row (e.g., row 1 = presentation, row 2 = services, row 3 = data).

3. **Use columns for grouping**. Components in the same vertical stack should share a column.

4. **Leave gaps for routing**. If you have many cross connections, leave empty cells between dense areas so edges have room to route cleanly.

5. **Use distinct stroke colors** for edges when multiple connections exist. This makes the diagram easier to read.

```d2
a -> b: { style.stroke: "#E74C3C" }
c -> d: { style.stroke: "#3498DB" }
```

6. **Use explicit anchors** when two or more edges exit the same node toward the same direction. Pick different numbered positions (e.g., `bottom-2` and `bottom-4`) to keep lines separated.

## Example: Layered Architecture

```d2
classes: {
  row-1-col-1: {style.opacity: 1}
  row-1-col-2: {style.opacity: 1}
  row-1-col-3: {style.opacity: 1}
  row-2-col-1: {style.opacity: 1}
  row-2-col-2: {style.opacity: 1}
  row-2-col-3: {style.opacity: 1}
  row-3-col-1: {style.opacity: 1}
  row-3-col-2: {style.opacity: 1}
  row-4-col-1: {style.opacity: 1}
  row-4-col-2: {style.opacity: 1}
  src-anchor-bottom-2: {style.opacity: 1}
  src-anchor-bottom-4: {style.opacity: 1}
  dst-anchor-top-2: {style.opacity: 1}
  dst-anchor-top-3: {style.opacity: 1}
  dst-anchor-top-4: {style.opacity: 1}
  src-anchor-right-3: {style.opacity: 1}
  dst-anchor-left-3: {style.opacity: 1}
}

# Presentation Layer
web: Web UI {class: row-1-col-1}
mobile: Mobile {class: row-1-col-2}
cli: CLI {class: row-1-col-3}

# Service Layer
auth: Auth {class: row-2-col-1}
biz: Business Logic {class: row-2-col-2}
notif: Notifications {class: row-2-col-3}

# Data Layer
repo: Repository {class: row-3-col-1}
cache: Cache Layer {class: row-3-col-2}

# Storage Layer
pg: PostgreSQL {
  shape: cylinder
  class: row-4-col-1
}
redis: Redis {
  shape: cylinder
  class: row-4-col-2
}

# Edges with anchor control to prevent overlap
web -> auth: {
  class: src-anchor-bottom-2
  class: dst-anchor-top-3
}
web -> biz: {
  class: src-anchor-bottom-4
  class: dst-anchor-top-2
}
mobile -> biz: {
  class: src-anchor-bottom-2
  class: dst-anchor-top-3
}
cli -> biz: {
  class: src-anchor-bottom-4
  class: dst-anchor-top-4
}
biz -> repo
biz -> cache
biz -> notif: {
  class: src-anchor-right-3
  class: dst-anchor-left-3
}
repo -> pg
cache -> redis
```

## Workflow

1. User describes the diagram they want
2. Plan the grid layout (rows and columns)
3. Generate the D2 file with all required class declarations
4. Render with `d2 --layout=octopus`
5. Show the rendered output to the user
6. Iterate on anchor placement if edges overlap
