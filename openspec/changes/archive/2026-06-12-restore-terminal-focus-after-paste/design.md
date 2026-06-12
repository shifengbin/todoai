## Context

当前终端右键菜单在 `App.vue` 中打开，菜单 Paste 动作调用 `TerminalSessionManager.paste(terminalId)` 读取系统剪贴板并向对应 shell 发送输入，随后关闭菜单。由于点击菜单按钮会把浏览器焦点移出 xterm，菜单关闭后没有显式调用 xterm 的焦点恢复能力，用户需要再次点击终端才能继续输入。

相关约束：
- 前端使用 Vue 和 xterm.js，终端实例由 `TerminalSessionManager` 持有。
- 剪贴板读写走 Wails runtime API，后端 PTY 输入接口不需要变化。
- 修复应限定在前端交互层，避免改变 shell session 生命周期和输入路由。

## Goals / Non-Goals

**Goals:**
- 右键菜单 Paste 完成后，活动终端重新获得焦点，用户可以立即继续键盘输入。
- 保持现有 Paste 输入路由、空剪贴板忽略、菜单关闭和错误处理行为。
- 用单元测试或组件测试覆盖焦点恢复。

**Non-Goals:**
- 不改变快捷键 `Ctrl+Shift+V` 的语义。
- 不改变系统剪贴板 API 或 Go 后端 shell API。
- 不调整终端菜单的视觉样式、菜单项结构或其他右键菜单行为。

## Decisions

1. 在 `TerminalSessionManager` 中封装终端聚焦能力。

   `TerminalSessionManager` 已经是前端持有 xterm session 的边界，新增 `focus(terminalId)` 可以把对 `terminal.focus?.()` 的调用集中在同一处。相比在 `App.vue` 中直接访问 xterm 实例，这能保持 xterm 实现细节封装在 terminal manager 内。

2. 在右键菜单 Paste 流程关闭菜单后恢复焦点。

   菜单按钮点击期间焦点在菜单按钮上，先关闭菜单再等待 Vue 更新，然后调用 `terminalManager.focus(terminalId)`，可以避免焦点被即将卸载的菜单元素抢占。该流程应对剪贴板有内容和无内容都执行焦点恢复，因为用户触发 Paste 后的预期仍然是回到终端继续操作。

3. 不在 `paste()` 内部自动聚焦所有调用来源。

   `paste()` 当前表达的是“读取剪贴板并发送输入”，而焦点恢复是右键菜单造成的 UI 副作用。将焦点恢复放在菜单动作层，能避免无意改变快捷键粘贴或未来非菜单粘贴入口的焦点行为。

## Risks / Trade-offs

- 聚焦时机过早导致焦点仍停留在菜单按钮 → 在关闭菜单并等待 DOM 更新后再聚焦。
- 目标终端 session 已被释放或不存在 → `TerminalSessionManager.focus()` 应安全忽略缺失 session。
- 测试环境 fake xterm 没有 `focus` 方法 → 在测试 fake terminal 中添加可观察的 `focus` mock，验证调用次数和目标终端。

## Migration Plan

无需数据迁移。该变更仅影响前端运行时交互，回滚时移除焦点恢复调用和相关测试即可。

## Open Questions

None
