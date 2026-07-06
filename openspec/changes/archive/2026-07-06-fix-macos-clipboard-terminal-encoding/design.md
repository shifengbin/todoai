## Context

当前应用是 Go + Wails + Vue 桌面应用，终端渲染由 xterm.js 承担，系统剪贴板主要通过 Wails runtime 的 `ClipboardGetText`/`ClipboardSetText` 访问。已有代码中自定义应用菜单只包含“文件”菜单，会覆盖 Wails 在 macOS 上默认补齐的 App/Edit/Window 菜单；终端 PTY 输出读取后直接按 `string(buffer[:n])` 进入过滤器、事件和历史；xterm 字体栈只包含少量等宽字体；嵌入式 shell 环境只强制 `TERM`/`COLORTERM`。

用户反馈集中在 macOS：创建 TODO 名称/描述无法粘贴，右键复制 TODO 标题和描述后乱码，终端只出现部分字符异常且看起来像缺字。该问题跨越原生菜单、剪贴板、shell 环境、PTY 输出字节处理和前端字体 fallback，不能只用单点修复。

## Goals / Non-Goals

**Goals:**

- 在 macOS 上恢复标准编辑菜单角色，使 TODO 创建和编辑输入框可使用系统复制、粘贴、全选等行为。
- 确保 TODO 标题/描述和终端选中文本通过系统剪贴板复制、粘贴时保持 Unicode 文本不乱码。
- 确保嵌入式 shell 进程具备 UTF-8 locale，避免 GUI 启动环境缺少 `LANG`/`LC_CTYPE` 时影响命令输出和剪贴板工具。
- 确保 PTY 输出跨 read chunk 的 UTF-8 多字节字符不会被提前转换成损坏字符串。
- 提升 xterm 字体 fallback，覆盖中文、emoji、Powerline/Nerd Font 图标和常见 macOS 等宽字体。
- 通过单元测试覆盖菜单、剪贴板、UTF-8 分片、locale 和字体栈，并执行 Wails 打包验证。

**Non-Goals:**

- 不引入新的终端渲染库或替换 xterm.js。
- 不修改 Wails 依赖源码。
- 不提供用户可配置字体界面；本次只调整默认字体 fallback。
- 不处理非 UTF-8 legacy 编码终端程序的主动转码。
- 不改变 TODO、终端、工作区的数据模型。

## Decisions

### 1. 在应用菜单中保留自定义文件菜单并补齐 macOS 标准编辑角色

在 `applicationMenu()` 中根据平台构建菜单。macOS 下包含 Wails 的 `AppMenu`、`EditMenu` 和 `WindowMenu` 角色，同时保留现有“文件”菜单动作。其他平台保持现有菜单行为。

备选方案是完全依赖浏览器默认 context menu 或自行监听 `Cmd+V`。前者无法覆盖应用菜单快捷键缺失，后者容易绕过 WebView/系统输入法的默认行为。使用 Wails 原生 Edit 角色更接近 macOS 桌面应用预期。

### 2. 在应用层确保 UTF-8 进程环境，并在 shell 环境中补齐 UTF-8 locale

启动桌面应用时，对缺失或非 UTF-8 的 `LANG`/`LC_CTYPE` 进行保守补齐，优先不覆盖用户已有 UTF-8 设置。`EmbeddedTerminalEnv` 同样保证 shell 进程环境包含 UTF-8 locale。这样 Wails runtime 调用的 `pbcopy`/`pbpaste`、后端启动的 shell、后台命令都能继承合理的字符环境。

备选方案是在每个 `pbcopy`/`pbpaste` 调用点包一层自定义剪贴板 API。由于项目目前直接使用 Wails runtime 剪贴板，且 runtime 内部在当前进程环境下启动命令，统一修正进程和 shell 环境更小、更一致。

### 3. 在 PTY 输出进入字符串处理前做 UTF-8 分片保护

为 `ShellSession` 增加一个小型 UTF-8 chunk decoder。每次 PTY read 后先和上次 pending bytes 合并，只向 `commandStateOutputFilter` 输出完整 UTF-8 字符组成的字符串；read 末尾如果是不完整多字节序列，保留到下一次 read。遇到真正非法字节时才按替换字符处理，避免无限 pending。

备选方案是把 `TerminalOutputEvent.Data` 改为 base64 字节并由前端解码。这会扩大 API 和前端复杂度，也会影响历史记录格式。后端流式 UTF-8 decoder 可以保持现有事件和历史模型不变。

### 4. 强化 xterm 字体 fallback，而不是强制单一字体

默认字体栈保留现有等宽字体，同时增加 macOS 常见等宽、中文、emoji 和终端符号字体 fallback，例如 Nerd Font/Symbols、`PingFang SC`、`Apple Color Emoji`、`Menlo` 等。这样已安装 Nerd Font 的用户可显示 Powerline/图标，未安装时仍能尽量用系统字体显示中文和 emoji。

备选方案是打包字体或强制使用某个 Nerd Font。打包会增加体积和授权复杂度，强制字体也不能保证所有平台存在。字体栈 fallback 是最小可行改动。

### 5. 剪贴板写入必须等待并处理错误

TODO 标题/描述复制从 fire-and-forget 改为 `await ClipboardSetText(text)`，失败时通过现有错误提示通道反馈。终端剪贴板路径继续复用 `TerminalSessionManager` 的错误处理，并补充 Unicode 测试。

备选方案是只依赖 runtime 返回值但不等待。这样无法发现 macOS `pbcopy` 失败，也不利于用户判断复制是否成功。

## Risks / Trade-offs

- **某些终端图标仍缺字** → 字体 fallback 只能使用用户系统已安装字体；Nerd Font 私有区图标仍依赖本机安装对应字体。通过常见 Nerd Font 名称提高命中率，但不打包字体。
- **locale 补齐影响少数依赖非 UTF-8 locale 的程序** → 仅在缺失或非 UTF-8 时补齐，并保留已有 UTF-8 环境，降低破坏性。
- **UTF-8 decoder 处理非法字节需要取舍** → 本设计只保证合法 UTF-8 不因 chunk 分割损坏；真正非法字节仍会替换显示，不尝试 legacy 编码猜测。
- **macOS 菜单测试无法在 Linux CI 完全验证原生行为** → Go 单元测试覆盖菜单结构，前端测试覆盖输入和剪贴板调用；最终仍需 Wails 打包和 Mac 手工验收。
- **修改菜单构建可能影响现有文件菜单测试** → 保留现有文件菜单标签和动作，并新增平台角色断言，避免回归。

## Migration Plan

无需数据迁移。实现后重新打包应用即可。若发布后出现平台菜单异常，可回滚菜单构建改动；若终端输出出现异常，可回滚 UTF-8 decoder 并保留字体/locale 修复作为独立改动。

## Open Questions

无阻塞问题。实现阶段需要在 macOS 上手工确认 `Cmd+V`、右键复制中文 TODO、`printf '中文 ✓ 🔧   \n'` 的显示效果。
