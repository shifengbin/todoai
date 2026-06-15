## Context

当前应用是 Wails 桌面应用，后端 Go 管理 PTY/ConPTY shell session，前端 Vue/xterm.js 渲染终端并在 TODO 树中展示终端活动状态。现有状态来源有三类：

- 后端 `ShellSessionManager` 发出 shell lifecycle 状态，例如 `running` / `exited`。
- shell integration 通过应用私有 `777;tui-helper` payload 发出 `command-start` / `command-end`，用于维护 command label。
- 前端监听 xterm title change，并用标题文本启发式推导 `idle` / `busy` / `needs-input`。

这些信号目前分散处理，Claude/Codex 的标题格式规则也直接写在 `App.vue` 中。调研结果显示，Claude 和 Codex 都有更可靠的结构化状态来源：Claude 后台 session 可用 `claude agents --json`，Claude hooks 有 `Notification`、tool lifecycle、`Stop`、`SessionEnd`；Codex 可通过 `codex exec --json` 或 app-server JSON-RPC 得到 turn/item events，Codex hooks 可覆盖 turn/tool/subagent/stop 边界。结构化信号应优先于标题 fallback。

## Goals / Non-Goals

**Goals:**

- 定义统一 terminal agent status 模型，支持 Claude、Codex 和 unknown agent。
- 将 shell lifecycle、command-state、agent structured events、title fallback 合并到一个 reducer 中。
- 用明确优先级避免低置信度 title fallback 覆盖 hook、JSONL、app-server 或 Claude agents JSON 的结果。
- 保持现有 command label 行为和侧边栏显示体验。
- 让状态推导可通过纯函数单元测试验证。

**Non-Goals:**

- 不在本变更中把所有 Codex 终端改为 app-server 驱动。
- 不强制用户配置 Claude/Codex hooks；不可用时必须自动降级。
- 不解析终端屏幕内容或 OCR 内容来推断状态。
- 不持久化 agent runtime status；应用重启后仍以 restored terminal 的非运行状态为准。
- 不改变 launch profile 配置格式或默认启动命令。

## Decisions

### 1. 使用统一 AgentStatus reducer 作为唯一状态计算边界

前端新增一个状态 reducer 模块，接收标准化事件并输出 terminal runtime status：

```text
ShellStatusEvent
CommandStateEvent
AgentStructuredEvent
TitleActivityEvent
        │
        ▼
AgentStatusReducer
        │
        ▼
{ phase, source, confidence, reason, label, updatedAt }
```

`phase` 使用 `idle | busy | needs-input | done | failed | exited`。`source` 使用可追踪枚举，例如 `shell`, `command-state`, `claude-agents-json`, `claude-hook`, `codex-jsonl`, `codex-app-server`, `codex-hook`, `title-fallback`。`confidence` 使用 `authoritative | structured | heuristic`。

替代方案是继续在 `App.vue` 中串联多个 handler。这个方案短期改动少，但会继续让 title、command label、shell exit 互相覆盖，难以为 Claude/Codex 差异写清楚测试。

### 2. 结构化事件优先，title fallback 最低优先级

状态优先级从高到低：

1. shell/process 终止态：`exited` / startup failure。
2. authoritative agent state：Claude `agents --json`、Codex app-server/JSONL 的 turn/item completion/failure。
3. structured hook state：Claude/Codex hook events、permission/input notification、tool lifecycle。
4. shell command-state：维护 command label，不能单独证明 agent 内部 busy。
5. title fallback：只在当前没有更新鲜的高置信度 agent 状态时影响 `busy` / `needs-input`。

替代方案是给每个来源各自维护一份 UI 状态，然后在侧边栏合并。那会把优先级分散到多个组件，后续新增来源时容易出现不同视图不一致。

### 3. Claude 和 Codex 采集器先做可选适配层

Claude:

- 后台 sessions：提供可选轮询或手动刷新入口，运行 `claude agents --json --cwd <project path>`，按 session id/cwd 匹配 terminal 或 TODO project context。
- 前台嵌入式进程：通过 hook 脚本将 `UserPromptSubmit`、`PreToolUse`、`PostToolUse`、`Notification`、`Stop`、`SessionEnd` 等事件转换成应用私有结构化事件。

Codex:

- `codex exec --json`：当 launch profile 明确使用非交互 exec 模式时，解析 JSONL 事件并映射为 terminal agent status。
- app-server：设计事件映射和 reducer 输入，但不要求第一阶段替换当前交互 TUI launch profile。
- hooks：在继续使用 PTY 交互 Codex 时，将 hooks 作为结构化补强；没有 Notification 等待输入事件时继续依赖 title fallback 或 shell state。

替代方案是只支持 hooks。hooks 对当前 PTY 集成最轻，但 Codex `exec --json` 和 app-server 已经提供更强的 machine-readable event stream，不应该被降级为 title 解析。

### 4. 保持 command label 和 activity phase 分离

`currentCommand` 继续用于终端行主标签，例如 `codex`、`claude --dangerously-skip-permissions`、`npm test`。`phase` 独立驱动 activity indicator 和 TODO 汇总。

launch profile command label 只说明应用提交了命令，不得直接设置 `busy`。只有 command-state、结构化 agent event 或 title fallback 才能更新 phase。

替代方案是把 `currentCommand` 非空视为 busy。这个规则对普通 long-running command 直观，但对 Claude/Codex 启动后空闲等待输入的场景会持续误报 busy。

### 5. 后端事件保持窄接口，前端先承担归一化

第一阶段尽量不扩大 Go `ProjectTerminal` 持久化模型。后端继续发送 shell status、terminal output、command-state。新增 Claude/Codex 状态采集器如果落在后端，应通过单独 Wails event 发送标准化 agent event，而不是把 activity state 写入 terminal history。

替代方案是后端保存完整 agent state。它能让多窗口或重启恢复更强，但本应用目前 terminal runtime state 本身不恢复进程，持久化 agent state 容易制造“看似仍在运行”的误导。

## Risks / Trade-offs

- [Risk] Claude/Codex 官方事件在用户本机版本不可用。 -> 采集器必须能力检测，失败时回退到现有 shell/title 行为，并标记 source 为 fallback。
- [Risk] hook 配置需要信任或被用户禁用。 -> hooks 是增强来源，不作为唯一状态来源；缺失 hooks 不阻塞 terminal 使用。
- [Risk] `claude agents --json` 只覆盖后台 session。 -> 前台嵌入式 Claude 仍走 hook/title fallback，不把 agents JSON 误用于普通 foreground 进程。
- [Risk] Codex app-server 集成范围过大。 -> 本变更只定义映射和 reducer 接口，第一阶段可先支持 `codex exec --json`/hooks。
- [Risk] 结构化状态长时间不更新导致 stale busy。 -> reducer 应记录 `updatedAt` 和 source；可对 heuristic/title 状态设置较短过期策略，对 structured running 状态等待明确 stop/failure 或 shell exit。
- [Risk] 多来源同时到达造成闪烁。 -> reducer 按优先级和事件时间处理，低优先级事件不能覆盖更新鲜的高优先级状态。

## Migration Plan

1. 新增 reducer 和事件映射测试，不改变 UI。
2. 将现有 shell status、command-state、title classifier 接入 reducer，保持当前可见行为。
3. 增加 Claude/Codex 结构化事件输入通道和可选采集器。
4. 将侧边栏 activity summary 改为读取 reducer 输出。
5. 保留旧 title helper 一段时间作为 fallback 测试对象；确认覆盖后再内聚到 reducer 模块。

回滚方式：保留旧 `activityState` 字段和 title classifier 兼容路径。如果结构化事件导致异常，可禁用采集器，UI 回到 shell/title fallback。

## Open Questions

- Claude foreground hook 事件应通过 stdout 私有 payload、临时文件，还是本地 HTTP/Wails bridge 送入应用？
- Claude `agents --json` 的 session 与当前 terminal 的关联是否需要用户显式选择 background session id？
- Codex 交互 TUI 是否值得在后续阶段替换为 app-server 驱动，还是继续作为 launch profile 的普通进程？
