## 1. State Tracking

- [x] 1.1 Add a frontend runtime collection for terminal confirmation state keyed by `terminalId`.
- [x] 1.2 Update terminal event application logic to compare previous and next visible activity state.
- [x] 1.3 Mark non-active terminals as `needs-ack` only when they transition from `busy` to a non-busy state.
- [x] 1.4 Clear stale `needs-ack` state when the terminal becomes `busy` again.
- [x] 1.5 Clear `needs-ack` state when the user selects the corresponding terminal.

## 2. Sidebar Presentation

- [x] 2.1 Pass the derived confirmation state from `App.vue` to `ProjectSidebar.vue`.
- [x] 2.2 Extend terminal activity helpers to return and label `needs-ack`.
- [x] 2.3 Add the terminal row triangular warning marker for `needs-ack`.
- [x] 2.4 Extend collapsed TODO aggregation priority to `needs-input > needs-ack > busy > idle`.
- [x] 2.5 Ensure expanded TODO rows continue to show confirmation state only on terminal rows, not on the parent TODO row.

## 3. Styling

- [x] 3.1 Add terminal row color styling for `needs-ack` that differs from busy and needs-input.
- [x] 3.2 Add collapsed TODO `todo-activity-needs-ack` styling with urgent breathing animation.
- [x] 3.3 Add reduced-motion fallback for the confirmation breathing state.
- [x] 3.4 Verify CSS does not reuse terminal row icons on collapsed TODO rows.

## 4. Tests

- [x] 4.1 Add App tests for background `busy -> idle`, `busy -> done`, `busy -> failed`, and `busy -> exited` confirmation transitions.
- [x] 4.2 Add App tests proving active terminal transitions do not create `needs-ack`.
- [x] 4.3 Add App tests proving selecting a terminal clears `needs-ack`.
- [x] 4.4 Add App tests proving `needs-input -> idle` does not create `needs-ack`.
- [x] 4.5 Add ProjectSidebar tests for terminal row `needs-ack` icon, label, class, and data attribute.
- [x] 4.6 Add ProjectSidebar tests for collapsed TODO `needs-ack` aggregation and priority over `busy`.
- [x] 4.7 Add ProjectSidebar tests for `needs-input` priority over `needs-ack`.
- [x] 4.8 Add style tests for terminal confirmation color, urgent breathing keyframes, and reduced-motion fallback.
- [x] 4.9 Run the frontend automated test suite.

## 5. Review And Packaging

- [x] 5.1 Run an automated code review pass for the frontend state and sidebar changes.
- [x] 5.2 Run OpenSpec validation for `add-background-terminal-ack-state`.
- [x] 5.3 Run `wails build -tags webkit2_41` to generate the executable.
