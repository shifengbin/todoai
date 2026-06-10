## Context

`ProjectSidebar` currently renders each project as a top-level row and renders terminal rows beneath it. The interaction model is already terminal-scoped, but the sidebar lacks per-project collapse state and the CSS relies mostly on indentation, so terminal rows do not clearly read as children of their project. This change is limited to the Vue sidebar and shared stylesheet.

## Goals / Non-Goals

**Goals:**

- Let users expand and collapse terminal children for each project independently.
- Reveal a project's terminal branch when that project or one of its terminals becomes active.
- Make the tree relationship visually clear through row structure, indentation, and branch guides.
- Preserve existing project selection, terminal selection, and add-terminal behavior.
- Keep the implementation local to frontend component state and CSS.

**Non-Goals:**

- Persist collapse state across application restarts.
- Add terminal close, rename, drag-and-drop, or reordering behavior.
- Change backend shell session models, Wails bindings, or terminal routing.
- Redesign the whole application shell outside the sidebar tree.

## Decisions

### Use local sidebar collapse state

`ProjectSidebar` should own a set of collapsed project IDs. This keeps the behavior close to the UI that renders it and avoids adding persistence or app-wide state for a preference that is currently only presentational.

Alternative considered: store collapsed IDs in `App.vue`. That would make the state easier to preserve later, but it adds parent props/events for no current need. Local state is smaller and fits this change.

### Reveal active branches automatically

When `activeProjectId` changes, the sidebar should remove that project from the collapsed set. When a terminal becomes active, its owning project should also be expanded. Creating a new terminal already makes it active, so the same rule keeps the new terminal visible.

Alternative considered: never auto-expand and let the user manage every branch manually. That gives stricter control, but it can hide the selected terminal immediately after a project or terminal selection, making the current state harder to understand.

### Separate toggle, project, and add actions

The expand/collapse control should be its own icon button in the project row. Clicking it toggles the branch without selecting the project. Clicking the project content still selects the project. The add-terminal action remains available for available projects and continues to stop propagation.

Alternative considered: make the whole project row toggle expansion. That is simpler visually, but it conflicts with the existing project selection action.

### Use CSS branch guides instead of extra data

The tree relationship should be expressed with CSS classes and pseudo-elements: a vertical guide under each expanded project and short branch connectors for terminal rows. Terminal labels continue to derive from the existing terminal descriptor fields.

Alternative considered: render textual tree characters such as `└─`. That would be easy, but it tends to look cramped in proportional UI fonts and gives less control over active and hover states.

## Risks / Trade-offs

- Hidden terminal rows may make background work less visible -> auto-expand active branches and keep the project row visible even when children are collapsed.
- More buttons in the project row can feel crowded -> use compact icon buttons with stable dimensions and clear hover/focus states.
- Visual tree guides can break under long labels or narrow sidebars -> keep label overflow behavior and test truncation with long project paths and command labels.
- Unavailable projects may have no terminal children -> keep their project row visible and avoid showing add-terminal actions for unavailable projects.

## Migration Plan

No data migration is required. The change can ship as a frontend-only update. Rollback is a revert of `ProjectSidebar` and sidebar CSS changes.

## Open Questions

- None.
