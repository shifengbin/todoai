## Why

Mac 上存在剪贴板和终端字符显示异常：新建 TODO 名称/描述无法稳定粘贴，TODO 右键复制标题和描述后出现乱码，嵌入式终端部分字符显示异常。当前问题影响中文文本、符号字体、终端输出和剪贴板交互，需要同时覆盖字体缺失、UTF-8 分片、UTF-8 locale 和 macOS 标准编辑菜单路径。

## What Changes

- 强化 Mac 桌面输入框的系统编辑能力，确保创建 TODO 表单中的名称和描述字段可通过系统剪贴板粘贴文本。
- 强化 TODO 右键菜单的“复制标题和描述”行为，确保中文标题和描述原样写入系统剪贴板，并在剪贴板写入失败时反馈错误。
- 强化嵌入式终端输出处理，避免 PTY 输出 read chunk 切开 UTF-8 多字节字符后产生替换字符或乱码。
- 强化嵌入式 shell 的 UTF-8 环境，避免 Mac GUI 启动环境缺少 `LANG`/`LC_CTYPE` 时影响终端命令、`pbcopy`/`pbpaste` 或程序输出。
- 强化 xterm 字体栈，增加 macOS 中文、emoji、Powerline/Nerd Font 图标和常见等宽字体 fallback，降低“部分字符像乱码”的缺字风险。
- 不引入破坏性变更；既有 TODO、终端、工作区和 shell API 行为保持兼容。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: 强化创建 TODO 表单的系统剪贴板粘贴能力，以及 TODO 标题/描述复制到系统剪贴板时的 Unicode 保真和错误反馈。
- `embedded-shell-sessions`: 强化终端剪贴板和 shell 会话字符环境，确保终端输入/输出和剪贴板交互使用 UTF-8 语义。
- `embedded-terminal-emulation`: 强化 xterm 渲染能力，确保 UTF-8 输出分片安全，并提供覆盖中文、emoji 和终端符号的字体 fallback。

## Impact

- 前端：`frontend/src/App.vue`、`frontend/src/terminalManager.js`、`frontend/src/xtermFactory.js`、相关 Vitest 测试。
- 后端：`app.go` 菜单构建，`shell.go` PTY 输出读取、环境变量处理，相关 Go 测试。
- 平台：macOS Wails 原生菜单、Wails runtime 剪贴板桥接、系统 `pbcopy`/`pbpaste` 调用环境。
- 规格：`todo-workspace`、`embedded-shell-sessions`、`embedded-terminal-emulation` 的增量需求。
- 验证：Go 单元测试、前端单元测试、`wails build -tags webkit2_41` 打包验证。
