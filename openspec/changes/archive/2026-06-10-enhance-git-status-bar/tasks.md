## 1. Backend Git Initialization

- [x] 1.1 Add a testable Git initialization runner that executes `git -C <project path> init`.
- [x] 1.2 Add `InitializeProjectGitRepository(projectID)` on the Wails app, including project lookup and unavailable-path handling.
- [x] 1.3 Add Go tests for successful initialization, missing project errors, unavailable project paths, and Git command failures.

## 2. Wails Bindings

- [x] 2.1 Regenerate Wails frontend bindings so `InitializeProjectGitRepository` is available in `frontend/wailsjs/go/main/App.js` and `App.d.ts`.
- [x] 2.2 Confirm generated binding changes are limited to the new app method.

## 3. Frontend Status Model

- [x] 3.1 Replace the single Git status text computed value with structured status chips and action state.
- [x] 3.2 Include branch and changed-count chips for repositories, and show staged, unstaged, untracked, ahead, and behind chips only when their values are non-zero.
- [x] 3.3 Preserve stable empty/error states for no project, unavailable path, loading status, and Git status query failures.

## 4. Frontend Git Initialization Action

- [x] 4.1 Show `Initialize Git Repository` only when the active project is available and `isRepo` is false.
- [x] 4.2 Call `InitializeProjectGitRepository(activeProjectId)` directly when the action is clicked.
- [x] 4.3 Disable the action and show initialization progress while the request is pending.
- [x] 4.4 Refresh the active project Git status after successful initialization.
- [x] 4.5 Display a non-blocking error and keep the non-repository state visible when initialization fails.

## 5. Status Bar Styling

- [x] 5.1 Add compact rounded chip styles with distinct colors for branch, changed, staged, unstaged, untracked, sync, neutral, warning, and error states.
- [x] 5.2 Add status-bar action button styling that fits the fixed-height footer.
- [x] 5.3 Ensure long branch names and narrow widths do not break the fixed status-bar layout.

## 6. Frontend Tests

- [x] 6.1 Update existing Git status tests to assert branch and changed-count chip content.
- [x] 6.2 Add tests for staged, unstaged, untracked, ahead, and behind chip visibility.
- [x] 6.3 Add tests for non-Git repository initialization button visibility and direct click behavior.
- [x] 6.4 Add tests for initialization loading state, successful refresh, and failure error handling.

## 7. Verification

- [x] 7.1 Run `go test ./...`.
- [x] 7.2 Run `npm run test` from `frontend`.
- [x] 7.3 Run `npm run build` from `frontend`.
