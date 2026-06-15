## Why

当前终端活动状态由 shell 进程状态、命令标签、终端标题启发式和侧边栏汇总逻辑共同维护，Claude 和 Codex 的状态信号混在 `App.vue` 的 title classifier 中，容易把启动标题、路径、单帧动画或等待输入误判为忙碌状态。

Claude 和 Codex 都已经提供比标题文本更可靠的结构化状态来源。需要把状态系统重构为统一的 agent 状态模型，让结构化事件优先，终端标题只作为低置信度 fallback。

## What Changes

- 新增统一 agent activity/status 模型，表达 `idle`、`busy`、`needs-input`、`done`、`failed`、`exited`，并记录状态来源、置信度、原因和更新时间。
- 将 shell 生命周期、shell command-state、Claude/Codex 结构化事件、终端标题 fallback 统一进入一个 reducer，避免多个入口直接覆盖 UI 状态。
- 为 Claude 支持官方可用的准确状态来源：
  - 后台 Claude sessions 优先使用 `claude agents --json --cwd <path>` 的 `state/status/waitingFor`。
  - 前台嵌入式 Claude 进程通过 hooks 接收 `UserPromptSubmit`、tool lifecycle、`Notification`、`Stop`、`SessionEnd` 等状态边界。
- 为 Codex 支持官方可用的准确状态来源：
  - 非交互任务优先支持 `codex exec --json` JSONL 事件。
  - 深度集成预留 `codex app-server` JSON-RPC turn/item notifications。
  - 保留 Codex hooks 作为当前 PTY 启动方式下的结构化补强。
- 保留现有 command label 行为，但 launch profile command label 本身不得标记 busy。
- 降级终端标题分类的权重：仅在没有更高置信度结构化事件时用于提示活动状态。
- 不引入破坏性 API 变更；前端显示可继续消费 `activityState`，但其来源改为统一状态模型计算结果。

## Capabilities

### New Capabilities

- `agent-status`: 统一 Claude/Codex agent activity 状态、状态来源优先级、结构化事件映射和 UI 汇总规则。

### Modified Capabilities

- `embedded-shell-sessions`: 终端 command label、launch profile 和 shell lifecycle 不再直接决定 agent busy 状态，改为向统一 agent status reducer 提供输入信号。
- `embedded-terminal-emulation`: 终端 title activity classification 改为低置信度 fallback，不能覆盖更可靠的结构化 agent 状态。

## Impact

- 后端：`ShellSessionManager` 状态事件、command-state event shape、可选 Claude/Codex 状态采集器、Wails runtime event。
- 前端：`App.vue` 中 terminal runtime fields、title classifier、command-state handler、status reducer；`ProjectSidebar.vue` 中 terminal/TODO activity 汇总。
- 测试：新增状态 reducer 单元测试，覆盖 Claude/Codex 结构化事件优先级、title fallback、launch profile 不强制 busy、shell exit 清理状态。
- 文档/配置：需要定义 Claude/Codex hook 脚本或 JSONL/app-server 采集方式，但不要求用户必须配置所有来源；不可用来源应自动降级。
