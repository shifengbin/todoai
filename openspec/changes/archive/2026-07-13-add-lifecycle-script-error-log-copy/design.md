## Context

TODO 从 `not-started` 进入 `in-progress` 时会异步执行初始化脚本，用户完成 `in-progress` TODO 时会先异步执行完成脚本。两条链路最终都由 `TodoLifecycleScriptExecutor` 执行。命令运行器通过 `CombinedOutput` 已经获得 stdout/stderr 的完整合并输出，但执行器写入 `TodoLifecycleScriptStatus` 前只保留最后 4096 字节的 `outputTail`，该状态又会随 `ProjectState` 和运行时事件频繁传给前端。

前端 `ProjectSidebar` 当前在失败状态中展示单行省略的 `outputTail` 和 Retry 按钮；`App.vue` 已经统一封装 Wails API 调用、错误展示和 `ClipboardSetText`。侧栏还具有延迟 600ms、挂载到 `body` 的 TODO 描述悬浮层，可以复用其定位和可见性管理模式。

完整错误输出只服务于当前应用运行期间仍然有效的生命周期脚本失败状态，不需要跨应用重启或 workspace 重新打开后恢复。

## Goals / Non-Goals

**Goals:**

- 对开始 TODO 触发的初始化脚本和完成 TODO 触发的完成脚本，在执行失败后保留该次完整 stdout/stderr 合并输出。
- 保持现有失败摘要和状态事件轻量，只在用户悬浮查看或点击复制时传输完整输出。
- 在 Retry 旁提供明确的复制入口，并以保留原始换行、可滚动的悬浮详情展示长日志。
- 使完整输出与当前失败状态具有一致的创建、替换和清理生命周期。
- 沿用现有 Wails 剪贴板能力和全局错误反馈方式。

**Non-Goals:**

- 不持久化生命周期脚本日志，也不在应用重启后恢复失败日志。
- 不为成功执行保留或展示完整输出。
- 不提供实时日志流、日志搜索、日志下载或历史执行记录。
- 不分别展示 stdout 和 stderr；继续使用 `CombinedOutput` 返回的原始合并顺序。
- 不扩展到终端命令、worktree 准备或其他非生命周期脚本错误。

## Decisions

### 1. 完整失败输出与状态摘要分离存储

`TodoLifecycleScriptExecutor` 增加受现有互斥锁保护的完整失败输出映射，键继续使用 `todoId + phase`。脚本失败时，执行器将未经截断的 `TodoLifecycleScriptRunResult.Output` 写入该映射，同时保持 `TodoLifecycleScriptStatus.OutputTail` 的现有截断逻辑。

每次执行同时分配单调递增的 `runId`，每次 workspace scope 重置时推进 `scopeEpoch`。这两个轻量身份随状态传给前端，但完整输出仍只保留在后端。执行器在 queued 写回和异步结果提交前校验当前 `runId`，从而阻止清理前启动的旧 goroutine 复活状态或覆盖新执行。

选择分离存储而不是向 `TodoLifecycleScriptStatus` 增加完整输出字段，是为了避免完整日志进入每次 `ProjectState` 返回和状态事件。与写入临时文件相比，内存存储符合“不跨重启保留”的范围，且无需引入文件权限、路径管理和过期文件清理。

### 2. 通过 Wails API 按需读取当前失败日志

应用层增加按 `todoId` 和 `phase` 获取当前失败输出的方法。方法只在对应状态仍为 `failed` 时返回数据；存在完整进程输出时原样返回该字符串，进程未产生输出时返回状态中的执行错误 `message`。不存在当前失败状态时返回可识别错误，避免前端复制已经失效的旧日志。

前端在悬浮延迟到期或点击复制后调用该 API。`App.vue` 负责 Wails 调用、以 `scopeEpoch + todoId + phase + runId` 为身份的短期缓存和错误处理，并将已加载日志传给 `ProjectSidebar`；侧栏继续通过事件请求数据，不直接承担后端 API 与剪贴板依赖。复制操作使用 `ClipboardSetText` 写入 API 返回的字符串，不追加阶段、脚本名称或其他格式化文本，并将剪贴板返回 `false` 与 Promise rejection 都视为失败。

### 3. 完整日志与失败状态同步清理

开始任一阶段执行时先清除该键的旧完整输出，失败时写入本次输出，成功时清除该键。现有 `Clear`、`ClearTodo` 行为同步清理完整输出，并增加原子 `ResetScope` 能力供 workspace 切换或关闭成功后调用。每个启动请求携带其来源 workspace scope；scope 已变化的迟到启动会被拒绝，旧 scope 的迟到事件通过 `scopeEpoch` 被应用层和前端丢弃，从而防止相同 TODO ID 在不同 workspace 间读取到旧数据。workspace 关闭失败时不重置 scope，保留当前状态和日志。

应用层使用同一互斥锁串行化 workspace transition 与生命周期状态回调的 epoch 校验、完成脚本成功后的 TODO 完成副作用，避免校验后切换 workspace 的 TOCTOU。执行器通过单锁 `Snapshot` 同时返回状态列表和 scope epoch，避免项目状态携带旧 statuses 与新 epoch 的撕裂组合。

前端在用户点击 Retry 时立即把本地状态切换为运行中，并在状态 run、scope 变化、状态被清除或切换 workspace 时同步丢弃对应缓存。异步读取返回后必须再次确认 `scopeEpoch` 和 `runId` 仍对应当前失败状态，避免 Retry 期间或新一轮失败后的迟到响应重新显示旧日志。

### 4. 失败状态行提供复制按钮和悬浮详情

只在 `failed` 状态的 Retry 按钮旁显示 Copy 图标按钮，并提供可访问名称和简短提示。复制按钮不改变 Retry 行为；剪贴板写入失败时沿用全局错误提示。

鼠标停留在失败状态行达到现有 600ms 延迟后请求并展示完整日志。悬浮层沿用现有挂载到 `body` 的定位模式，使用等宽字体和 `white-space: pre-wrap` 保留换行，设置稳定的最大宽度和最大高度并允许滚动，避免长行或大量输出撑开侧栏。鼠标离开、状态失效或 workspace 切换时关闭悬浮层。

悬浮层 SHALL 在完整日志节点挂载后读取自身真实渲染尺寸，再以失败状态行为锚点计算固定坐标。定位优先使用状态行下方 12px；下方空间不足时，使用真实高度紧贴状态行上方 12px，并将最终坐标限制在视口安全边距内。测量完成前隐藏悬浮层，避免使用最大高度预估导致短日志远离触发行或短暂显示在窗口左上角。

### 5. 保持现有失败摘要兼容性

侧栏失败行仍立即展示 `outputTail || message` 的单行摘要，完整日志 API 不参与普通状态渲染。现有状态 JSON、持久化数据和成功清理行为保持不变，因此无需数据迁移。

## Risks / Trade-offs

- [完整输出大小不受限制，失败脚本可能增加进程内存占用] -> 仅保留当前失败状态的输出，并在重试、成功、删除、完成或 workspace 切换时立即清理；不把完整输出复制到状态事件和项目状态中。
- [按需读取与 Retry 并发时可能返回已经过期的日志] -> 后端读取时校验当前状态，前端接收后再次按失败状态键校验并在状态变化时清理缓存。
- [超长日志渲染可能影响悬浮层性能] -> 仅在用户悬浮后创建日志节点，限制容器尺寸并使用滚动；普通侧栏继续只渲染摘要。
- [CombinedOutput 无法区分 stdout 和 stderr] -> 明确保持命令实际合并输出，不在本次变更中改变进程采集模型。
- [workspace 切换时现有运行时状态可能与新 workspace 的 TODO ID 冲突] -> workspace 切换和关闭时统一清除生命周期脚本状态与完整失败输出。

## Migration Plan

1. 扩展执行器的运行时失败输出存储、查询和清理接口，并补充 Go 测试。
2. 在应用层暴露按需读取 API，更新 Wails 生成绑定并补充应用 API 测试。
3. 增加前端缓存、复制操作、悬浮详情和对应交互测试。
4. 运行 Go 与前端测试，确认现有 Retry、成功归档和失败摘要行为不回归。

本变更不修改持久化文件，无需数据迁移。回滚时移除新增运行时映射、Wails API 和前端控件即可，现有截断摘要与 Retry 行为仍可独立工作。

## Open Questions

无。
