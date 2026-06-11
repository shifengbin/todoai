## 1. Dependency And Platform Boundaries

- [x] 1.1 Evaluate and add the selected Go ConPTY dependency, keeping third-party types out of non-Windows business code.
- [x] 1.2 Split the current real PTY implementation into Unix and Windows platform files so Unix/Linux/macOS continue to use `creack/pty`.
- [x] 1.3 Keep `PtyProcess` and `ShellStarter` unchanged and verify `ShellSessionManager` does not require Windows-specific branching.

## 2. Windows ConPTY Adapter

- [x] 2.1 Implement a Windows `NewPtyProcess` adapter that starts the configured shell in the requested working directory with the requested environment and initial terminal size.
- [x] 2.2 Implement ConPTY-backed `Read`, `Write`, `Resize`, `Wait`, and idempotent `Close` behavior matching the existing `PtyProcess` contract.
- [x] 2.3 Map unsupported ConPTY capability errors to `ErrEmbeddedShellUnsupported` while preserving non-support startup errors as ordinary startup errors.
- [x] 2.4 Ensure terminal deletion and app shutdown close Windows ConPTY processes without leaving blocked waits or orphaned shell processes.

## 3. Tests

- [x] 3.1 Add backend tests for unsupported ConPTY error mapping and preservation of non-support startup errors.
- [x] 3.2 Add backend tests or platform-compiled test seams for Windows ConPTY resize, close, wait, and shell start request wiring.
- [x] 3.3 Run existing Go shell session tests to confirm Unix/fake-starter behavior remains unchanged.
- [x] 3.4 Run frontend automated tests to confirm the existing terminal unsupported state, terminal creation, input, resize, and deletion UI flows still pass.
- [x] 3.5 Run `GOOS=windows GOARCH=amd64 go build ./...` to verify Windows platform code compiles.
- [x] 3.6 Run `GOOS=windows GOARCH=amd64 go test -c -o /tmp/tui-helper-windows.test.exe .` to verify the Windows test package compiles.

## 4. Review And Manual Verification

- [x] 4.1 Run an automated code review pass focused on platform build tags, dependency leakage, lifecycle cleanup, and error classification.
- [ ] 4.2 On Windows 10 1809+ or Windows 11, manually verify creating a TODO project terminal starts PowerShell/Cmd/pwsh inside the app.
- [ ] 4.3 On Windows 10 1809+ or Windows 11, manually verify terminal input, command output, resize, terminal deletion, and app shutdown behavior.
- [ ] 4.4 On an unsupported or simulated unsupported Windows ConPTY environment, verify the terminal enters the existing `unsupported` state and does not auto-retry.

## 5. Final Verification

- [x] 5.1 Run the full relevant Go test suite.
- [x] 5.2 Run the full relevant frontend test suite.
- [x] 5.3 Run `openspec validate add-windows-conpty-terminal-backend --strict`.
- [x] 5.4 Run `wails build -tags webkit2_41`.
