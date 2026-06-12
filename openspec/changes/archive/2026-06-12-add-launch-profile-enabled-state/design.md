## Context

Launch profiles 目前作为 terminal settings 的一部分持久化，Go 模型只包含 `name` 和 `command`，前端设置页直接编辑该列表，左侧启动菜单展示内置 `Terminal` 加全部自定义 profiles。

本次变更需要跨 Go settings persistence、Wails 绑定、Vue 设置页和 ProjectSidebar 启动菜单。主要约束是兼容旧设置文件：旧 JSON 中没有启用字段，升级后这些 profile 必须继续显示。

## Goals / Non-Goals

**Goals:**

- 为每个自定义 launch profile 持久化启用状态。
- 旧配置中缺少启用状态的 profile 默认视为启用。
- 设置页允许切换启用状态，并保留禁用 profile 的名称、命令、排序和删除能力。
- 启动菜单隐藏禁用的自定义 profiles，但始终保留内置 `Terminal`。
- 用后端和前端测试覆盖持久化、迁移、保存和菜单过滤。

**Non-Goals:**

- 不允许禁用内置 `Terminal`。
- 不改变终端 shell 选择、检测或 fallback 行为。
- 不改变 launch profile 命令执行方式。
- 不把“禁用”解释为删除或清空 profile。

## Decisions

1. 对外模型使用 `enabled` 字段，而不是持久化 `disabled`。

   `enabled` 与设置页“是否启用”的语义一致，Wails TypeScript 模型和前端代码也更直观。备选方案是使用 `disabled` 来利用 Go bool 零值兼容旧配置，但这会让 API 与 UI 语义相反，后续维护成本更高。

2. 加载旧配置时显式迁移缺省启用。

   Go 的 `bool` 零值无法区分 JSON 缺字段和用户保存的 `false`。实现时应在读取 settings JSON 时使用辅助持久化结构或自定义解码，让 profile 的 `enabled` 字段以 `*bool` 形式参与加载；缺字段时转换为 `true`，明确保存的 `false` 保持禁用。保存时写回普通 `enabled: true/false`。

3. 禁用 profile 只影响启动菜单可见性。

   禁用 profile 仍在 settings panel 中显示，可以编辑 name/command、调整顺序、删除或重新启用。这样用户可以临时隐藏入口，而不会丢失命令配置。启动菜单使用过滤后的 enabled profiles，内置 `Terminal` 不参与过滤。

4. 禁用 profile 仍执行现有验证。

   即使 profile 当前禁用，保存时也要求 name 和 command 非空、名称不重复且不与 `Terminal` 冲突。这样重新启用不会暴露无效配置，也避免禁用项占用重复名称后造成菜单行为不确定。

## Risks / Trade-offs

- [Risk] 旧配置被误判为禁用 → 使用辅助解码结构区分缺字段与显式 `false`，并添加旧 JSON 缺字段的回归测试。
- [Risk] 前后端默认值不一致 → 后端 `defaultTerminalLaunchProfiles()` 和前端 fallback defaults 都显式包含 `enabled: true`，保存前归一化同一语义。
- [Risk] 菜单过滤改变索引相关测试 → 测试应断言可见文本和触发结果，不依赖禁用项仍占据菜单 index。
- [Risk] Wails generated files 漏更新 → 实现任务中包含重新生成或同步 `frontend/wailsjs/go/models.ts` 与 API 类型。

## Migration Plan

- 首次加载旧 settings 文件时，缺少 `enabled` 的 launch profiles 在内存 state 中归一化为启用。
- 后续保存 settings 时写出包含 `enabled` 的 profile 列表。
- 回滚到旧版本时，旧版本会忽略多余的 `enabled` 字段，name/command 仍可读取。

## Open Questions

无。用户已确认禁用的 launch profile 应从启动菜单隐藏。
