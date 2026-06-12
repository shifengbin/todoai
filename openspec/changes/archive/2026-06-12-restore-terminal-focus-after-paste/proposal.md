## Why

用户通过终端右键菜单执行 Paste 后，焦点会停留在菜单按钮或页面上，导致粘贴完成后必须再次点击终端才能继续输入。这个交互中断了终端的连续操作体验，尤其影响频繁复制粘贴命令的场景。

## What Changes

- 终端右键菜单的 Paste 操作完成后，系统 SHALL 将焦点恢复到触发粘贴的活动 xterm 终端。
- 右键菜单 Paste 仍 SHALL 保持现有行为：读取系统剪贴板、向对应终端 shell 发送非空文本、关闭菜单。
- 空剪贴板或剪贴板读取失败时，不 SHALL 发送终端输入；菜单关闭和错误处理行为保持现有模式。
- 增加前端测试覆盖右键菜单 Paste 后的焦点恢复行为。

## Capabilities

### New Capabilities

- None

### Modified Capabilities

- `embedded-shell-sessions`: 扩展终端剪贴板右键菜单 Paste 的交互要求，确保菜单粘贴后终端重新获得焦点。

## Impact

- 影响前端终端会话管理：`frontend/src/terminalManager.js` 需要提供或使用 xterm 的聚焦能力。
- 影响终端菜单交互：`frontend/src/App.vue` 的右键菜单 Paste 流程需要在菜单关闭后恢复终端焦点。
- 影响测试：更新 `frontend/src/terminalManager.test.js` 和/或 `frontend/src/App.test.js`，验证焦点恢复。
- 不影响 Go 后端 shell/session API，不引入新依赖，不改变剪贴板 API。
