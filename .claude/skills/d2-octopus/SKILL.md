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
# 1. Declare all classes used in this diagram
classes: {
  row-1-col-1: {style.opacity: 1}
  row-1-col-2: {style.opacity: 1}
  row-2-col-1: {style.opacity: 1}
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

1. **Every class must be declared** in the `classes` block with `{style.opacity: 1}` (or `0` for config). D2 ignores undeclared classes.

2. **Grid coordinates are 1 indexed**. `row-1-col-1` is the top left cell.

3. **No two nodes can share the same grid position**. Each cell holds one node.

4. **D2 only passes one class per node** to the plugin. Use separate nodes for separate classes.

5. **Nodes without grid classes** are auto placed in the next available cell.

## Grid Size Configuration

Control cell dimensions with the `octopus-{width}x{height}` class. Assign it to a hidden config node. Gap between cells is derived automatically (20% of width).

```d2
classes: {
  octopus-300x150: {style.opacity: 0}
  row-1-col-1: {style.opacity: 1}
  row-1-col-2: {style.opacity: 1}
}

# Hidden config node (invisible, not positioned)
octopus-config: " " {
  class: octopus-300x150
  style.opacity: 0
}

a: API {class: row-1-col-1}
b: DB {class: row-1-col-2}
```

Default cell size is 200x120. Only add the config node when you need different spacing.

## Edge Anchor Control

By default, Octopus automatically selects which side of a shape an edge exits/enters and distributes multiple edges across 5 anchor points per side.

For precise control, use `src-anchor-*` and `dst-anchor-*` classes:

```d2
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
# WRONG
a -> b: { class: src-anchor-bottom-2 dst-anchor-top-4 }

# CORRECT
a -> b: {
  class: src-anchor-bottom-2
  class: dst-anchor-top-4
}
```

### Anchor Point Reference

Each side has 5 positions at 1/6, 2/6, 3/6, 4/6, 5/6 of the edge length:

**Top/Bottom** (left to right): `top-1`, `top-2`, `top-3`, `top-4`, `top-5`
**Left/Right** (top to bottom): `left-1`, `left-2`, `left-3`, `left-4`, `left-5`
**Corners**: `top-left`, `top-right`, `bottom-left`, `bottom-right`
**Centers**: `top-center`, `bottom-center`, `left-center`, `right-center`

### When to Use Explicit Anchors

- Multiple edges exit the same side: pick different numbered positions (e.g., `bottom-2` and `bottom-4`)
- Lines overlap visually: spread them with explicit positions
- The automatic distribution does not look right

## Rendering

Always render with `--layout=octopus`:

```bash
d2 --layout=octopus diagram.d2 output.svg
d2 --layout=octopus diagram.d2 output.png
```

## Design Guidelines

1. **Plan the grid first**. Sketch rows and columns before writing D2 code.

2. **Use rows for layers**. Row 1 = presentation, row 2 = services, row 3 = data, etc.

3. **Use columns for grouping**. Components in the same vertical stack share a column.

4. **Leave gaps for routing**. Skip row/column numbers between dense areas for cleaner edge paths.

5. **Use distinct stroke colors** for multiple edges:
```d2
a -> b: { style.stroke: "#E74C3C" }
c -> d: { style.stroke: "#3498DB" }
```

6. **Use explicit anchors** when edges going the same direction from a node need separation.

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

web -> auth: {
  class: src-anchor-bottom-2
  class: dst-anchor-top-3
}
web -> biz: {
  class: src-anchor-bottom-4
  class: dst-anchor-top-2
}
mobile -> biz
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
