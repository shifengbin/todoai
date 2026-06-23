## 1. Zsh Command-State Integration

- [x] 1.1 Add failing Go coverage showing the zsh integration script gates `command-end` behind a prior `command-start`.
- [x] 1.2 Update `zshIntegrationScript()` to track whether `preexec` has started a command and only emit `command-end` from `precmd` when that flag is set.
- [x] 1.3 Verify zsh still emits `command-start` for executed commands and one `command-end` after the shell returns to the prompt.

## 2. Frontend Label Race Handling

- [x] 2.1 Add failing frontend coverage for a launch profile terminal whose optimistic `currentCommand` receives an unpaired `command-end` before any matching `command-start`.
- [x] 2.2 Preserve the launch profile label across the unpaired idle `command-end` while keeping normal command-start/command-end clearing behavior unchanged.
- [x] 2.3 Add or update frontend coverage proving a real `command-start` followed by `command-end` still returns the label to the shell name.

## 3. Verification And Review

- [x] 3.1 Run Go tests covering shell integration and command-state filtering.
- [x] 3.2 Run frontend automated tests covering App/agent status/terminal label behavior.
- [x] 3.3 Run an automated code review pass for the changed files and address any correctness or maintainability findings.

## 4. Build

- [x] 4.1 Run `wails build -tags webkit2_41` to generate the executable file.
