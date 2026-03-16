# Octopus Layout Engine

## Problem

Every diagramming tool forces you to choose: either you manually drag boxes in a GUI (Figma, draw.io), or you write code and surrender control to an auto-layout algorithm (D2, Mermaid, Graphviz).

Auto-layout engines (dagre, elk, dot) decide where your boxes go. You can hint with "direction: right" or "rank", but you cannot say "put this box at row 2, column 3." When the algorithm disagrees with your intent, you fight it. You add invisible nodes, phantom edges, and hacks. The result is fragile and unpredictable.

Manual GUI tools give you full control but produce opaque binary files that cannot be versioned, diffed, generated programmatically, or maintained by AI.

There is no tool where you write a simple text definition, control exactly where things go, and get a clean SVG or PNG out.

## Solution

Octopus is a text-to-diagram layout engine where **you define the grid, you place the components, and the engine renders it.** The engine draws connections between components but never moves them.

You write a YAML or JSON file. You get an SVG.

## Core Principles

1. **You own the layout.** The engine never repositions your components. You say where things go, they stay there.
2. **Text in, SVG out.** Everything is a plain text file. Version it, diff it, generate it with scripts or AI.
3. **Connections are smart, placement is manual.** You place boxes. The engine routes arrows between them cleanly (no overlaps, smooth curves or right-angle paths).
4. **Theming, not styling.** You pick a theme (sketch, clean, blueprint). Individual component styling is minimal. Consistency by default.
5. **Composable.** A diagram can include another diagram. A legend is a diagram. A component can be a group containing sub-components.

## Input Format

```yaml
diagram:
  width: 800
  height: 600
  theme: sketch

components:
  user:
    label: User
    shape: person
    position: [1, 3]     # grid row, column
    color: gray

  orchestrator:
    label: Orchestrator
    position: [2, 3]
    color: blue

  context:
    label: Context Manager
    position: [3, 2]
    color: green

  router:
    label: Model Router
    position: [3, 3]
    color: green

  tools:
    label: Tool Executor
    position: [3, 4]
    color: green

  llm:
    label: LLM API
    position: [4, 3]
    color: lightblue

  response:
    label: Response
    position: [5, 3]
    color: gray

connections:
  - from: user
    to: orchestrator
  - from: orchestrator
    to: context
  - from: orchestrator
    to: router
  - from: orchestrator
    to: tools
  - from: router
    to: llm
  - from: orchestrator
    to: response

legend:
  position: bottom
  items:
    - color: blue
      label: Core Engine
    - color: green
      label: Services
    - color: lightblue
      label: External API
    - color: gray
      label: I/O
```

## Output

- SVG (default)
- PNG (via headless browser or sharp)
- Configurable scale (1x, 2x, 3x)
- Configurable padding

## Key Features

### Grid-based positioning
Components are placed on a logical grid (row, column). The engine calculates pixel positions from grid coordinates. Grid cell sizes adapt to content. You can also use absolute pixel positions if needed.

### Smart edge routing
The only "auto" behavior. Given fixed component positions, the engine routes connection lines to avoid overlapping components. Supports straight, orthogonal (right-angle), and curved routing styles.

### Themes
Built-in visual themes that control fonts, colors, stroke styles, and shapes:
- **clean** - crisp lines, solid fills, professional
- **sketch** - hand-drawn look, rough edges, casual
- **blueprint** - white on blue, technical drawing style
- **minimal** - thin lines, no fills, just outlines

### Color palettes
Named color sets (not hex codes in every file):
- `blue`, `green`, `yellow`, `red`, `gray`, `lightblue`
- Each theme maps these names to specific hex values

### Component shapes
- `box` (default) - rectangle with rounded corners
- `person` - stick figure / user icon
- `cylinder` - database
- `diamond` - decision
- `circle` - event / trigger
- `group` - container holding sub-components

### Groups and nesting
A component can contain other components:
```yaml
context:
  label: Context Manager
  shape: group
  position: [3, 2]
  children:
    - system-prompt
    - history
    - memory
```

### Composability
Include one diagram inside another:
```yaml
includes:
  - file: legend.yaml
    position: bottom
```

## Tech Stack

- **Language:** TypeScript
- **SVG generation:** Direct SVG string building (no DOM dependency)
- **PNG export:** Sharp or Playwright for rasterization
- **CLI:** Node.js CLI tool
- **Input:** YAML (primary), JSON (also supported)

## CLI Usage

```bash
# Basic
octopus render diagram.yaml -o diagram.svg

# PNG at 3x scale
octopus render diagram.yaml -o diagram.png --scale 3

# With theme override
octopus render diagram.yaml -o diagram.svg --theme sketch

# Watch mode
octopus render diagram.yaml -o diagram.svg --watch
```

## Non-Goals

- No interactive/web editor (this is a CLI tool)
- No auto-layout (the whole point is you control layout)
- No drag and drop
- No animation
- No real-time collaboration

## Success Criteria

- Can recreate all 13 AI Harness diagrams from YAML in under 5 minutes each
- Output quality matches or exceeds D2 sketch theme
- Zero layout surprises (what you write is what you get)
- Files are readable by humans and writable by AI
