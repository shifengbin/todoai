## 1. Status Model And Reducer

- [x] 1.1 Add a frontend agent status module with phase, source, confidence, reason, label, and updatedAt fields.
- [x] 1.2 Implement source priority rules so shell exit and structured Claude/Codex signals override title fallback.
- [x] 1.3 Move existing title activity classification into the reducer as `title-fallback` input without changing current fallback behavior.
- [x] 1.4 Add reducer tests for idle startup, shell exit cleanup, structured busy, structured needs-input, title fallback, and stale lower-priority events.

## 2. Existing Terminal Signal Integration

- [x] 2.1 Route Wails `terminal-status` events through the agent status reducer while preserving terminal shell state.
- [x] 2.2 Route `terminal-command-state` events through the reducer while keeping `currentCommand` as the display label source.
- [x] 2.3 Ensure launch profile command labels do not mark terminals busy until a recognized activity signal arrives.
- [x] 2.4 Update `applyState` terminal merge logic to preserve runtime agent status only for running terminals and clear it for restored/exited terminals.

## 3. Claude And Codex Structured Sources

- [x] 3.1 Add normalized event mapping for Claude hook events, including prompt submit, tool lifecycle, notification, stop, and session end.
- [x] 3.2 Add normalized event mapping for Claude `agents --json` state, status, and waitingFor values.
- [x] 3.3 Add normalized event mapping for Codex `exec --json` JSONL events, including thread, turn, item, and error events.
- [x] 3.4 Add normalized event mapping stubs for Codex app-server turn/item notifications without requiring app-server launch profiles in this change.
- [x] 3.5 Add capability detection and graceful fallback when Claude/Codex structured sources are unavailable or disabled.

## 4. UI Activity Summary

- [x] 4.1 Update terminal runtime descriptors so `activityState` is derived from unified agent status phase.
- [x] 4.2 Update TODO activity aggregation so `needs-input` takes precedence over `busy`, and exited/done/failed terminals do not remain busy.
- [x] 4.3 Preserve existing terminal row labels, activity icons, accessible labels, and stable row dimensions.
- [x] 4.4 Add or update frontend component tests for terminal rows, collapsed TODO activity, launch profile idle behavior, and structured status priority.

## 5. Backend And Event Plumbing

- [x] 5.1 Add a narrow Wails event shape for normalized agent status events if a structured source is collected outside the frontend.
- [x] 5.2 Keep terminal history persistence free of transient agent status fields.
- [x] 5.3 Add backend tests for any new event emission or parser paths introduced for structured Claude/Codex status.
- [x] 5.4 Verify existing command-state payload filtering tests still pass and invalid payloads do not change agent status.

## 6. Verification And Review

- [x] 6.1 Run Go tests covering shell sessions, command-state filtering, and terminal history.
- [x] 6.2 Run frontend automated tests for reducer, App, terminal manager, and ProjectSidebar behavior.
- [x] 6.3 Run OpenSpec status for `refactor-agent-status-system` and resolve incomplete artifacts.
- [x] 6.4 Run an automatic code review pass for the completed implementation and address actionable findings.
- [x] 6.5 Run `wails build -tags webkit2_41` to generate the executable.
