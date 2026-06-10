## Context

项目树由 `ProjectSidebar.vue` 渲染，项目是顶层节点，终端是项目下的子节点。当前样式中，终端列表的竖向分支线由 `.terminal-list::before` 绘制，终端行的横向连接线由 `.terminal-row::before` 绘制。激活终端已经通过 `.terminal-row.active::before` 使用高亮色，但父级竖向分支线仍使用默认灰色。

这个变更只调整项目树视觉状态，不改变项目、终端、会话生命周期或 Wails API。

## Goals / Non-Goals

**Goals:**

- 让包含激活终端的项目分支竖线与激活终端横线使用一致高亮色。
- 保持没有激活终端的项目分支继续使用默认树形引导线颜色。
- 保持现有项目选择、终端选择、展开折叠和删除交互不变。
- 用组件测试覆盖可验证的激活分支状态。

**Non-Goals:**

- 不调整终端行、项目行的选择模型。
- 不新增后端状态或持久化字段。
- 不重新设计整个项目树样式或颜色系统。

## Decisions

### 在项目节点上显式标记 active terminal 分支

给 `ProjectSidebar.vue` 增加一个轻量判断：当某个项目的任一终端 ID 等于 `activeTerminalId` 时，在该项目的 `.project-node` 上加状态类，例如 `has-active-terminal`。样式再通过 `.project-node.has-active-terminal .terminal-list::before` 将竖线颜色设为激活终端横线的同一颜色。

这样比直接使用 CSS `:has(.terminal-row.active)` 更稳妥：组件测试可以直接断言状态类，且不依赖运行环境对 `:has()` 的支持或测试 DOM 实现细节。

### 高亮整条可见终端分支竖线

当前 DOM 使用一条 `.terminal-list::before` 绘制整个终端列表竖线，并没有按每个终端行分段。为了保持改动小且符合现有结构，当项目下有 active terminal 时，高亮该项目下可见终端列表的整条竖向分支线。

替代方案是只高亮从父节点到 active terminal 横线之间的竖线段，但这需要拆分竖线绘制或增加覆盖层，复杂度高于当前需求收益，也更容易产生不同终端数量下的定位问题。

### 复用现有 active branch 颜色

竖线使用与 `.terminal-row.active::before` 相同的颜色值，避免新增视觉 token 或引入新色值。后续如果项目建立 CSS 变量或设计 token，可以再把该颜色抽成共享变量。

## Risks / Trade-offs

- 整条分支竖线高亮可能让同一项目下的其他非激活终端看起来也属于 active path -> 通过只在包含 active terminal 的项目分支上生效，并保留单个 active 终端行的边框和横线，降低误读。
- 使用硬编码颜色会延续现有样式的重复色值 -> 本次保持小范围修改，后续统一颜色变量时再收敛。
- 只通过视觉颜色表达状态可能对低视觉敏感用户帮助有限 -> 现有 active 终端行仍保留 `active` 类和边框样式，本次只增强已有视觉路径，不作为唯一状态来源。
