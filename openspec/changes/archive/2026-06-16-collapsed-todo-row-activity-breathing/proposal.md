## Why

折叠 TODO 后，隐藏终端的活动状态目前复用终端行的小图标提示，尤其是 `busy` 的转圈图标在 TODO 父级行上不够明显。TODO item 作为聚合摘要层级，应使用更容易扫到的整行状态反馈，并与终端行的精细状态图标区分开。

## What Changes

- 折叠 TODO 下存在隐藏终端活动时，TODO 行使用整行呼吸效果表达聚合活动状态。
- `busy` 和 `needs-input` 使用不同强度或色彩的整行呼吸反馈，且 `needs-input` 继续高于 `busy`。
- 展开 TODO 时，活动状态继续由具体终端行展示，TODO 父级行不重复显示聚合活动反馈。
- 移除折叠 TODO 行上复用终端活动图标的呈现；终端行自身的转圈和警告图标保持不变。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: 调整收起 TODO 分支时隐藏终端聚合活动状态的视觉呈现要求。

## Impact

- 影响前端 TODO 工作区侧边栏中折叠 TODO 行的活动状态渲染和样式。
- 影响 `ProjectSidebar` 相关组件测试中对折叠 TODO 活动提示的断言。
- 不改变后端数据结构、终端活动状态模型、OpenSpec `agent-status` 的聚合优先级或终端行自身状态提示。
