## 1. Branch Picker State

- [x] 1.1 Add shared branch picker state in `frontend/src/App.vue` for the open picker key, loaded branch options, loading/error state, and a fixed visible candidate limit.
- [x] 1.2 Add shared helpers to filter branch candidates by current input, cap rendered candidates, select a candidate, close the picker, and preserve manual input values.
- [x] 1.3 Ensure `ensureProjectBranchesLoaded(projectId)` records failed loads without blocking the input or repeatedly retrying on every render.

## 2. UI Implementation

- [x] 2.1 Replace the create TODO branch `datalist` markup with the application-rendered branch picker and candidate menu.
- [x] 2.2 Replace the TODO detail branch `datalist` markup with the same branch picker behavior.
- [x] 2.3 Replace the add-project-to-TODO branch `datalist` markup with the same branch picker behavior.
- [x] 2.4 Add concise loading, empty, truncated, and load-failed states for the candidate menu while keeping the branch input editable.
- [x] 2.5 Update branch picker styling so candidate menus fit inside existing dialogs, remain readable, and do not overlap adjacent controls incoherently.

## 3. Automated Tests

- [x] 3.1 Update existing `frontend/src/App.test.js` branch picker tests to assert application-rendered candidates instead of native `datalist` attributes.
- [x] 3.2 Add a create TODO test covering a large branch list and verifying only the capped number of filtered candidates is rendered.
- [x] 3.3 Add an add-project-to-TODO test covering branch candidate selection and preservation of the submitted base branch.
- [x] 3.4 Add a TODO detail test covering manual branch input after `ListProjectBranches` rejects.
- [x] 3.5 Run the frontend automated test suite for the affected app tests.

## 4. Quality And Packaging

- [x] 4.1 Run OpenSpec validation for `stabilize-todo-project-branch-picker`.
- [x] 4.2 Run an automated code review pass over the frontend branch picker changes and address any correctness, accessibility, or regression findings.
- [x] 4.3 Run the relevant Go tests if any backend branch-list behavior is changed.
- [x] 4.4 Build the Wails app with `wails build -tags webkit2_41` to generate the executable.
