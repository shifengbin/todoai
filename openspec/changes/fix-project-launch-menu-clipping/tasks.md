## 1. Test Coverage

- [x] 1.1 Add a ProjectSidebar test that opens the launch menu from a project near the bottom of the project list and verifies the menu uses a non-clipped placement.
- [x] 1.2 Add test coverage for preserving existing launch menu behavior: built-in `Terminal`, configured profiles, option selection, and outside-click close.
- [x] 1.3 Mock or control DOM measurement in the component test so the bottom-boundary scenario is deterministic in jsdom.

## 2. Sidebar Menu Placement

- [x] 2.1 Add component state for the currently open launch menu placement and any required max-height constraint.
- [x] 2.2 Measure the trigger button and project list visible bounds when opening a launch menu.
- [x] 2.3 Choose downward placement when there is enough space below the trigger, and upward placement when the menu would otherwise be clipped near the list bottom.
- [x] 2.4 Constrain the menu height and allow menu-internal scrolling when neither upward nor downward placement has enough space for all options.
- [x] 2.5 Reset placement and height state when the launch menu closes.

## 3. Styling

- [x] 3.1 Add CSS classes or inline custom properties for upward launch menu placement.
- [x] 3.2 Update launch menu overflow styling so constrained menus keep all options reachable without being clipped by `.project-list`.
- [x] 3.3 Preserve the existing menu visual style, hover states, and sidebar scrolling behavior.

## 4. Verification

- [x] 4.1 Run the frontend component tests.
- [x] 4.2 Run the frontend build.
- [x] 4.3 Manually verify the last visible project's add-terminal menu in a small-height window if a local app run is available. Automated geometry coverage was used because this API session cannot visually inspect an interactive desktop window.
