## Why

当前用户在左侧 TODO 工作树中点击终端条目后，右侧 xterm 会切换到对应终端，但键盘焦点仍可能停留在侧边栏按钮上。用户需要再点击一次右侧终端区域才能继续输入，影响多终端切换时的连续操作。

## What Changes

- 点击左侧 TODO 树里的终端条目并成功切换到该终端后，系统 SHALL 自动将输入焦点转移到右侧对应的嵌入式终端。
- 自动聚焦只发生在用户明确选择终端的交互路径上。
- 保持现有终端选择、激活、历史回放、自动重启和右键菜单行为不变。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `embedded-terminal-emulation`: 增加用户从 TODO 终端树选择终端后，右侧嵌入式终端获得输入焦点的交互要求。

## Impact

- 影响前端终端选择流程：`frontend/src/App.vue`。
- 复用现有 xterm 聚焦能力：`frontend/src/terminalManager.js`。
- 更新前端单元测试：`frontend/src/App.test.js`。
- 不改变 Go 后端 API、终端会话模型、持久化格式或运行时依赖。
