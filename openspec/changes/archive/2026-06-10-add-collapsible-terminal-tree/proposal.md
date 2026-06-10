## Why

The left sidebar now supports multiple terminals per project, but the current visual treatment reads like a flat list and becomes noisy as projects accumulate terminals. Users need to collapse each project's terminal children and quickly understand which terminal rows belong to which project.

## What Changes

- Add per-project expand/collapse controls for terminal child rows in the project sidebar.
- Automatically expand the relevant project when a project, terminal, or newly created terminal becomes active so the current terminal remains discoverable.
- Improve the sidebar tree styling with clearer parent/child structure, including indentation, branch guides, and distinct project and terminal row affordances.
- Preserve existing project selection, terminal selection, and add-terminal actions.
- No backend API, persistence, or shell-routing changes.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `project-workspace`: Project terminal tree rows become collapsible and visually communicate parent/child relationships.

## Impact

- Vue frontend: `ProjectSidebar` state, events, markup, and tests.
- CSS: left sidebar tree layout, row spacing, branch guides, active/hover states, and collapsed/expanded affordances.
- No Go backend or Wails binding changes expected.
