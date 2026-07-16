# zsh ZDOTDIR 递归修复实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 修复 TodoAI zsh 包装脚本在用户配置加载期间暴露临时 `ZDOTDIR` 所导致的配置或 hook 递归，同时保证 `exec zsh` 后命令状态检测继续有效。

**Architecture:** 继续使用临时 `ZDOTDIR` 让 zsh 自动加载 TodoAI 包装 `.zshrc`。Go 启动层负责解开嵌套 TodoAI 环境中的真实用户配置目录；包装脚本通过非导出的进程内标记防止重复 source，并只在加载用户配置的作用域内切换到原始 `ZDOTDIR`，之后恢复临时目录以支持下一次 `exec zsh`。

**Tech Stack:** Go、zsh 5.x、Go `testing`、`os/exec`、Wails v2。

---

设计依据：`docs/plans/2026-07-16-zsh-zdotdir-recursion-design.md`。

### Task 1: 解开嵌套 TodoAI 的原始 ZDOTDIR

**Files:**
- Modify: `shell.go:1230-1252`
- Test: `shell_test.go:687-712`

**Step 1: 写入嵌套环境失败测试**

在 `shell_test.go` 增加 `path/filepath` import，并新增测试：

```go
func TestZshIntegratedLaunchPreservesOriginalZDOTDIRAcrossNestedTodoAI(t *testing.T) {
	home := t.TempDir()
	outerWrapper := t.TempDir()
	launch, err := IntegratedShellLaunch("/bin/zsh", []string{
		"HOME=" + home,
		"ZDOTDIR=" + outerWrapper,
		"TUI_HELPER_ORIGINAL_ZDOTDIR=" + home,
	})
	if err != nil {
		t.Fatalf("IntegratedShellLaunch() error = %v", err)
	}
	t.Cleanup(launch.Cleanup)

	if got := envValueFromList(launch.Env, "TUI_HELPER_ORIGINAL_ZDOTDIR"); got != home {
		t.Fatalf("TUI_HELPER_ORIGINAL_ZDOTDIR = %q, want %q", got, home)
	}
	if got := envValueFromList(launch.Env, "ZDOTDIR"); got == "" || got == outerWrapper {
		t.Fatalf("ZDOTDIR = %q, want a new wrapper directory", got)
	}
	if _, err := os.Stat(filepath.Join(envValueFromList(launch.Env, "ZDOTDIR"), ".zshrc")); err != nil {
		t.Fatalf("wrapper .zshrc is unavailable: %v", err)
	}
}
```

**Step 2: 运行测试并确认失败**

Run:

```bash
go test ./... -run '^TestZshIntegratedLaunchPreservesOriginalZDOTDIRAcrossNestedTodoAI$' -count=1
```

Expected: FAIL，当前实现把 `outerWrapper` 写入 `TUI_HELPER_ORIGINAL_ZDOTDIR`。

**Step 3: 最小实现原始目录解包**

修改 `zshIntegratedLaunch` 的目录解析顺序：

```go
originalZDOTDIR := envValueFromList(launch.Env, "TUI_HELPER_ORIGINAL_ZDOTDIR")
if originalZDOTDIR == "" {
	originalZDOTDIR = envValueFromList(launch.Env, "ZDOTDIR")
}
if originalZDOTDIR == "" {
	originalZDOTDIR = envValueFromList(launch.Env, "HOME")
}
```

其余临时目录创建、环境覆盖和 cleanup 行为保持不变。

**Step 4: 运行测试并确认通过**

Run:

```bash
go test ./... -run '^TestZshIntegratedLaunchPreservesOriginalZDOTDIRAcrossNestedTodoAI$' -count=1
```

Expected: PASS。

**Step 5: 提交**

```bash
git add shell.go shell_test.go
git commit -m "fix: preserve original zsh config directory"
```

### Task 2: 用户配置加载期间切换 ZDOTDIR 并防止重复 source

**Files:**
- Modify: `shell.go:1298-1320`
- Test: `shell_test.go`

**Step 1: 增加真实 zsh 测试辅助函数**

在 `shell_test.go` import 中增加 `os/exec` 和 `path/filepath`，添加以下辅助函数：

```go
func requireZsh(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh is unavailable")
	}
	return path
}

func runIntegratedZsh(t *testing.T, shellPath string, baseEnv []string, command string) (ShellLaunch, string) {
	t.Helper()
	launch, err := IntegratedShellLaunch(shellPath, baseEnv)
	if err != nil {
		t.Fatalf("IntegratedShellLaunch() error = %v", err)
	}
	t.Cleanup(launch.Cleanup)
	cmd := exec.Command(launch.Path, append(append([]string{}, launch.Args...), "-c", command)...)
	cmd.Env = launch.Env
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("integrated zsh failed: %v\n%s", err, output)
	}
	return launch, string(output)
}
```

不要使用开发机真实 `~/.zshrc`；每个测试都在 `t.TempDir()` 中生成最小用户配置。

**Step 2: 写入作用域和防重入失败测试**

```go
func TestZshIntegrationScopesZDOTDIRAndLoadsUserConfigOnce(t *testing.T) {
	shellPath := requireZsh(t)
	originalZDOTDIR := t.TempDir()
	loadLog := filepath.Join(t.TempDir(), "loads.log")
	userRC := `
print -r -- "$ZDOTDIR" >> "$TUI_HELPER_TEST_LOAD_LOG"
`
	if err := os.WriteFile(filepath.Join(originalZDOTDIR, ".zshrc"), []byte(userRC), 0o600); err != nil {
		t.Fatal(err)
	}

	launch, output := runIntegratedZsh(t, shellPath, []string{
		"HOME=" + originalZDOTDIR,
		"ZDOTDIR=" + originalZDOTDIR,
		"TUI_HELPER_TEST_LOAD_LOG=" + loadLog,
	}, `
source "$ZDOTDIR/.zshrc"
print -r -- "FINAL_ZDOTDIR=$ZDOTDIR"
print -r -- "PREEXEC=${preexec_functions[(I)__tui_helper_preexec]}"
print -r -- "PRECMD=${precmd_functions[(I)__tui_helper_precmd]}"
`)

	loads, err := os.ReadFile(loadLog)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Fields(string(loads)); len(got) != 1 || got[0] != originalZDOTDIR {
		t.Fatalf("user config ZDOTDIR loads = %#v, want [%q]", got, originalZDOTDIR)
	}
	wrapperZDOTDIR := envValueFromList(launch.Env, "ZDOTDIR")
	for _, want := range []string{
		"FINAL_ZDOTDIR=" + wrapperZDOTDIR,
		"PREEXEC=1",
		"PRECMD=1",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}
```

**Step 3: 运行测试并确认失败**

Run:

```bash
go test ./... -run '^TestZshIntegrationScopesZDOTDIRAndLoadsUserConfigOnce$' -count=1
```

Expected: FAIL；当前用户配置看到临时目录，且再次 source 包装文件会再次加载用户配置。

**Step 4: 最小实现包装脚本**

将 `zshIntegrationScript` 改为以下结构。标记必须是非导出的 zsh 参数，hook 定义和注册必须位于保护块内：

```go
func zshIntegrationScript() string {
	return `
if [[ -z ${__tui_helper_zsh_integrated-} ]]; then
  typeset -g __tui_helper_zsh_integrated=1
  typeset __tui_helper_wrapper_zdotdir="$ZDOTDIR"

  if [[ -n "$TUI_HELPER_ORIGINAL_ZDOTDIR" && -f "$TUI_HELPER_ORIGINAL_ZDOTDIR/.zshrc" ]]; then
    {
      export ZDOTDIR="$TUI_HELPER_ORIGINAL_ZDOTDIR"
      source "$TUI_HELPER_ORIGINAL_ZDOTDIR/.zshrc"
    } always {
      export ZDOTDIR="$__tui_helper_wrapper_zdotdir"
    }
  fi
  unset __tui_helper_wrapper_zdotdir

  autoload -Uz add-zsh-hook
  __tui_helper_emit_command_start() {
    printf '\033]777;todoai;command-start;%s\a' "$(printf '%s' "$1" | base64 | tr -d '\n')"
  }
  __tui_helper_emit_command_end() {
    printf '\033]777;todoai;command-end\a'
  }
  __tui_helper_preexec() {
    __tui_helper_emit_command_start "$1"
  }
  __tui_helper_precmd() {
    __tui_helper_emit_command_end
  }
  add-zsh-hook preexec __tui_helper_preexec
  add-zsh-hook precmd __tui_helper_precmd
fi
`
}
```

**Step 5: 运行测试并确认通过**

Run:

```bash
go test ./... -run '^TestZshIntegrationScopesZDOTDIRAndLoadsUserConfigOnce$' -count=1
```

Expected: PASS。

**Step 6: 运行已有 shell 集成测试**

Run:

```bash
go test ./... -run 'TestShellSessionManagerStartsSupportedShellsWithCommandLabelIntegration|TestZshIntegration' -count=1
```

Expected: PASS。

**Step 7: 提交**

```bash
git add shell.go shell_test.go
git commit -m "fix: isolate zsh wrapper configuration"
```

### Task 3: 验证 exec zsh、错误恢复和 cleanup

**Files:**
- Test: `shell_test.go`

**Step 1: 写入 exec zsh 失败测试**

```go
func TestZshIntegrationSurvivesExecZsh(t *testing.T) {
	shellPath := requireZsh(t)
	originalZDOTDIR := t.TempDir()
	if err := os.WriteFile(filepath.Join(originalZDOTDIR, ".zshrc"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	_, output := runIntegratedZsh(t, shellPath, []string{
		"HOME=" + originalZDOTDIR,
		"ZDOTDIR=" + originalZDOTDIR,
		"TUI_HELPER_TEST_ZSH=" + shellPath,
	}, `
exec "$TUI_HELPER_TEST_ZSH" -i -c '
  print -r -- "EXEC_PREEXEC=${preexec_functions[(I)__tui_helper_preexec]}"
  print -r -- "EXEC_PRECMD=${precmd_functions[(I)__tui_helper_precmd]}"
'
`)

	for _, want := range []string{"EXEC_PREEXEC=1", "EXEC_PRECMD=1"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}
```

**Step 2: 运行测试并确认通过**

Run:

```bash
go test ./... -run '^TestZshIntegrationSurvivesExecZsh$' -count=1
```

Expected: PASS。若失败，不得通过导出防重入标记规避；必须保证新进程重新执行包装脚本。

**Step 3: 写入错误恢复测试**

用户 `.zshrc` 返回非零，确认包装目录恢复且 hook 仍注册：

```go
func TestZshIntegrationRestoresWrapperAfterUserConfigError(t *testing.T) {
	shellPath := requireZsh(t)
	originalZDOTDIR := t.TempDir()
	if err := os.WriteFile(filepath.Join(originalZDOTDIR, ".zshrc"), []byte("return 23\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	launch, output := runIntegratedZsh(t, shellPath, []string{
		"HOME=" + originalZDOTDIR,
		"ZDOTDIR=" + originalZDOTDIR,
	}, `
print -r -- "FINAL_ZDOTDIR=$ZDOTDIR"
print -r -- "PREEXEC=${preexec_functions[(I)__tui_helper_preexec]}"
print -r -- "PRECMD=${precmd_functions[(I)__tui_helper_precmd]}"
`)

	for _, want := range []string{
		"FINAL_ZDOTDIR=" + envValueFromList(launch.Env, "ZDOTDIR"),
		"PREEXEC=1",
		"PRECMD=1",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}
```

**Step 4: 运行错误恢复测试并确认通过**

Run:

```bash
go test ./... -run '^TestZshIntegrationRestoresWrapperAfterUserConfigError$' -count=1
```

Expected: PASS。

**Step 5: 增加 cleanup 断言**

在 Task 1 的测试末尾显式验证临时目录删除：

```go
wrapperZDOTDIR := envValueFromList(launch.Env, "ZDOTDIR")
launch.Cleanup()
if _, err := os.Stat(wrapperZDOTDIR); !errors.Is(err, os.ErrNotExist) {
	t.Fatalf("wrapper directory still exists after cleanup: %v", err)
}
```

调用 cleanup 后不要再依赖 `t.Cleanup` 进行唯一验证；cleanup 本身允许重复调用或重复删除目录。

**Step 6: 运行 zsh 测试集合**

Run:

```bash
go test ./... -run 'TestZshIntegratedLaunch|TestZshIntegration|TestShellSessionManagerStartsSupportedShellsWithCommandLabelIntegration' -count=1
```

Expected: PASS。

**Step 7: 提交**

```bash
git add shell_test.go
git commit -m "test: cover zsh integration re-entry"
```

### Task 4: 全量验证和交付检查

**Files:**
- Verify: `shell.go`
- Verify: `shell_test.go`
- Verify: `docs/plans/2026-07-16-zsh-zdotdir-recursion-design.md`

**Step 1: 运行 Go 全量测试**

Run:

```bash
go test ./... -count=1
```

Expected: PASS，输出包含 `ok  todoai`。

**Step 2: 运行前端测试**

Run:

```bash
cd frontend && npm test
```

Expected: 所有 Vitest 测试通过。

**Step 3: 运行前端生产构建**

Run:

```bash
cd frontend && npm run build
```

Expected: Vite 构建成功；现有 chunk 体积警告不阻断交付。

**Step 4: 检查补丁格式**

Run:

```bash
git diff --check
```

Expected: 无输出，退出码为 0。

**Step 5: 最终运行 Wails 构建**

Run:

```bash
wails build -tags webkit2_41
```

Expected: 构建 `build/bin/todoai` 成功。

**Step 6: 提交必要的最终调整**

只有在验证步骤产生必要修复时执行：

```bash
git add shell.go shell_test.go
git commit -m "fix: harden zsh integration startup"
```

不要提交构建产物、临时 zsh 目录或与本修复无关的现有工作区改动。
