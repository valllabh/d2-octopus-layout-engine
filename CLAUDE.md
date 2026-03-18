# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

D2 Octopus is an external layout engine plugin for D2, written in Go. It provides grid based node positioning with A* edge routing. Users assign grid coordinates via D2 classes (e.g., `class: row-2-col-3`) and the plugin computes pixel positions and routes edges.

The binary is `d2plugin-octopus`. Users select it with `d2 --layout=octopus diagram.d2`.

## Build and Run

```bash
make build          # compile to bin/d2plugin-octopus
make test           # run all tests
make test-single RUN=TestName PKG=./internal/grid  # run a single test
make lint           # go vet
make install        # copy binary to $GOPATH/bin
make render         # render all test diagrams (SVG + PNG)
make clean          # remove bin/
```

## Project Structure

```
cmd/d2plugin-octopus/main.go  -- entry point, calls d2plugin.Serve()
internal/plugin/plugin.go     -- Plugin interface, layout pipeline, side selection
internal/plugin/astar.go      -- A* router, routing grid, overlap spreading
internal/grid/grid.go         -- grid coordinate parsing, anchors, pixel math
tests/
  input/                        -- 32 D2 test diagrams
  output/                       -- rendered SVGs (gitignored)
  png/                          -- rendered PNGs (gitignored)
docs/
  images/                       -- README images and comparisons
```

## Layout Pipeline (plugin.Layout())

1. Parse grid positions from node classes (`row-N-col-M` pattern)
2. Validate no position conflicts
3. Auto place unpositioned nodes in next available cell
4. Compute pixel coordinates using configurable cell size, gap, padding
5. Expand containers to fit children
6. Detect shape collisions
7. Route edges with A* pathfinder (obstacle avoidance, bend minimization)
8. Spread overlapping segments proportionally across gap channels

## Plugin Flags

| Flag | Default | Description |
|------|---------|-------------|
| `octopus-cell-width` | 200 | Cell width in pixels |
| `octopus-cell-height` | 120 | Cell height in pixels |
| `octopus-gap` | 40 | Gap between cells |
| `octopus-padding` | 60 | Padding around grid |
| `octopus-align` | center | Cell alignment |
| `octopus-anchor` | center | Shape anchor |

## Per Node/Edge Classes

- Grid position: `class: row-N-col-M`
- Node alignment: `class: align-top-left`
- Node anchor: `class: anchor-bottom-center`
- Edge anchors: use separate `class:` lines per anchor (D2 does not split space separated values)

Each shape edge has 5 named anchor points (1 through 5) plus corners and edge centers.

## Conventions

- Use Makefile targets for all build, test, and run operations
- Do not use hyphens or dashes in prose sentences
- Keep project documents well organized
- Grid coordinates are 1 indexed: `[1,1]` is top left
- After modifying edge routing, run all 32 tests and validate
- Keep `bin/d2plugin-octopus` and `/opt/homebrew/bin/d2plugin-octopus` in sync
