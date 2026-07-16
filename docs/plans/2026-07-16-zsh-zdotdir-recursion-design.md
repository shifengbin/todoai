# 修复 zsh 临时 ZDOTDIR 递归设计

## 背景

TodoAI 为 zsh 终端创建临时目录，将 `ZDOTDIR` 指向该目录，并在临时 `.zshrc` 中先加载用户原有配置，再注册 `preexec` 和 `precmd` hook。这样可以在不修改用户 `~/.zshrc` 的情况下发送命令开始和结束状态。

当前包装脚本在加载用户配置时仍保留临时 `ZDOTDIR`。Oh My Zsh、Powerlevel10k 等启动组件会把该目录视为真实配置目录；如果配置加载期间或之后再次进入交互 zsh，可能重新进入 TodoAI 包装脚本，引发 hook 或配置递归。嵌套启动 TodoAI 时，现有 Go 代码还可能把上一层 TodoAI 临时目录误认为用户原始 `ZDOTDIR`。

## 目标

- 用户 `.zshrc` 加载期间看到真实的原始 `ZDOTDIR`。
- 同一 zsh 进程重复加载包装脚本时不重复加载用户配置或注册 hook。
- 在内嵌终端执行 `exec zsh` 后，TodoAI 命令状态检测继续有效。
- 从另一个 TodoAI 终端启动 TodoAI 时，不继承上一层临时目录作为用户配置目录。
- 用户配置报错时恢复包装目录并继续尝试注册 TodoAI hook。
- 终端退出或删除后继续清理本次创建的临时目录。

## 非目标

- 不修改用户 `~/.zshrc`、Oh My Zsh、Powerlevel10k 或 vfox 配置。
- 不改变 Bash、PowerShell 或其他 shell 的集成方式。
- 不在本次修复中清理由历史版本或测试遗留的临时目录。
- 不处理终端历史 JSON 并发写入损坏问题。

## 方案比较

### 方案一：作用域切换 ZDOTDIR 并增加进程内防重入

继续使用临时 `.zshrc`，加载用户配置前暂时把 `ZDOTDIR` 切回原始目录，加载结束后恢复临时目录。使用非导出的 zsh 全局参数标记当前进程已经完成集成，避免同一进程重复 source。新 zsh 进程不会继承该标记，因此 `exec zsh` 仍会重新安装集成。

该方案不修改用户配置，改动集中，兼容现有清理和终端生命周期，作为采用方案。

### 方案二：向用户配置安装永久 TodoAI 插件

把 hook 脚本写入稳定路径，并修改用户 `.zshrc` 或 Oh My Zsh 插件列表。该方案天然支持 `exec zsh`，但会永久修改用户配置，并引入升级、卸载和多版本兼容问题，因此不采用。

### 方案三：首次加载后永久恢复原始 ZDOTDIR

启动后不再保留临时 `ZDOTDIR`。实现简单，但用户执行 `exec zsh` 后会直接加载原始配置并失去 TodoAI hook，不满足已确认的兼容性要求，因此不采用。

## 设计

### 原始目录解析

`zshIntegratedLaunch` 按以下顺序计算用户原始配置目录：

1. 已存在的 `TUI_HELPER_ORIGINAL_ZDOTDIR`。
2. 启动环境中的 `ZDOTDIR`。
3. 启动环境中的 `HOME`。

优先复用 `TUI_HELPER_ORIGINAL_ZDOTDIR` 可以解开嵌套 TodoAI 环境，避免把上一层 `/tmp/todoai-zsh-*` 目录继续向下传递。

### 包装脚本执行流

临时 `.zshrc` 使用一个非导出的进程内标记，例如 `__tui_helper_zsh_integrated`。标记不存在时执行以下流程：

1. 立即设置标记，阻止用户配置在同一进程内再次 source 包装文件。
2. 保存当前临时 `ZDOTDIR`。
3. 将 `ZDOTDIR` 临时设置为 `TUI_HELPER_ORIGINAL_ZDOTDIR`。
4. source 原始目录下的 `.zshrc`。
5. 通过 zsh 的 `always` 清理块恢复临时 `ZDOTDIR`。
6. 定义 TodoAI 状态输出函数并注册 `preexec`、`precmd` hook。

该标记不得导出。`source` 同一包装文件时标记仍存在，因此跳过重复初始化；`exec zsh` 创建的新进程不会继承标记，但会继承恢复后的临时 `ZDOTDIR`，因此新进程会再次加载包装脚本并正确安装 hook。

### 错误处理

原始 `.zshrc` 不存在时跳过用户配置加载，仍安装 TodoAI hook。用户配置返回非零或输出错误时不吞掉原始诊断；`always` 块负责恢复临时 `ZDOTDIR`，随后继续注册 hook。用户配置显式执行 `exit` 时遵循用户意图，终端退出并由现有会话清理流程删除临时目录。

## 测试

在 `shell_test.go` 增加针对 zsh 包装的测试：

- 无显式 `ZDOTDIR` 时使用 `HOME` 作为原始目录。
- 已有 `TUI_HELPER_ORIGINAL_ZDOTDIR` 时优先复用它，而不是继承旧临时 `ZDOTDIR`。
- 用户 `.zshrc` 执行期间观察到原始 `ZDOTDIR`，执行完成后包装环境恢复临时目录。
- 同一进程重复 source 包装文件时，用户配置只加载一次，hook 不重复。
- 真实 zsh 子进程执行 `exec zsh` 后仍包含 TodoAI 的 `preexec` 和 `precmd` hook。
- 用户配置返回非零时仍恢复临时目录并安装 hook。
- 调用 cleanup 后临时目录被删除。

验证范围包括相关 Go 单测、`go test ./...`，以及最终 Wails 构建，确认前端绑定和打包未受影响。
