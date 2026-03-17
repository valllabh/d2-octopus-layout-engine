# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Octopus is a D2 external layout engine plugin written in Go. It provides grid based node positioning for D2 diagrams. Users assign grid coordinates via D2 classes (e.g., `class: row-2-col-3`) and the plugin computes pixel positions. Edge routing is handled by D2 itself, not by Octopus.

The binary is `d2plugin-octopus`. Users select it with `d2 --layout=octopus diagram.d2`.

## Build and Run

```bash
make build          # compile to bin/d2plugin-octopus
make test           # run all tests
make test-single RUN=TestName PKG=./internal/grid  # run a single test
make lint           # go vet
make install        # copy binary to $GOPATH/bin
make clean          # remove bin/
```

To test with D2: `export PATH="$PWD/bin:$PATH" && d2 --layout=octopus examples/ai-orchestrator.d2 output.svg`

## Architecture

```
cmd/d2plugin-octopus/main.go  -- entry point, calls d2plugin.Serve()
internal/plugin/plugin.go     -- Plugin interface (Info, Flags, HydrateOpts, Layout, PostProcess)
internal/grid/grid.go         -- grid coordinate parsing and pixel math
```

**Layout pipeline** (all in `plugin.Layout()`):
1. Parse grid positions from node classes (`row-N-col-M` pattern)
2. Validate no position conflicts
3. Auto place unpositioned nodes in next available cell
4. Compute pixel coordinates from grid positions using configurable cell size, gap, padding
5. Expand containers to fit children
6. Route edges with orthogonal paths between positioned nodes

**Key D2 integration points:**
- Implements `d2plugin.Plugin` interface from `oss.terrastruct.com/d2/d2plugin`
- Uses `d2graph.Graph`, `d2graph.Object`, `d2graph.Edge` for graph manipulation
- Uses `lib/geo.Box` and `lib/geo.Point` for positioning
- Plugin communicates with D2 via JSON over stdin/stdout (handled by `d2plugin.Serve()`)

## BMAD Framework

The `_bmad/` directory contains the BMAD workflow framework for project planning. Planning artifacts (PRD, architecture) are in `_bmad-output/planning-artifacts/`. Do not modify files inside `_bmad/core/`.

## Conventions

- Use Makefile targets for all build, test, and run operations
- Do not use hyphens or dashes in prose sentences
- Keep project documents well organized
- Grid coordinates are 1 indexed: `[1,1]` is top left
