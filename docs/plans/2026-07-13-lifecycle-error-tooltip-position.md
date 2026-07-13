# 生命周期脚本错误悬浮窗定位修复 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 使用悬浮窗真实渲染尺寸，使生命周期脚本错误详情紧贴失败状态行并正确处理窗口边界。

**Architecture:** `ProjectSidebar` 在悬浮开始时保存触发元素，完整日志加载并挂载后测量触发元素与悬浮窗的真实矩形。一个纯定位函数按“下方优先、空间不足翻到上方、最后限制在视口安全边距内”的顺序计算固定坐标。

**Tech Stack:** Vue 3 Composition API、Vue Test Utils、Vitest、jsdom。

---

### Task 1: 记录 OpenSpec 回归任务

**Files:**
- Modify: `openspec/changes/add-lifecycle-script-error-log-copy/design.md`
- Modify: `openspec/changes/add-lifecycle-script-error-log-copy/tasks.md`

**Step 1: 补充定位设计**

在失败日志悬浮层决策中记录：必须在日志节点挂载后使用真实尺寸计算位置，测量前隐藏，下方优先且保留 12px 间距。

**Step 2: 增加待完成任务**

新增回归测试、最小定位修复和验证三个未完成任务，完成每一步后立即勾选对应项。

**Step 3: 检查文档差异**

Run: `git diff --check -- openspec/changes/add-lifecycle-script-error-log-copy`

Expected: 无输出，退出码为 0。

### Task 2: 添加真实高度定位失败测试

**Files:**
- Test: `frontend/src/App.test.js`

**Step 1: 写入失败测试**

在生命周期完整错误日志悬浮测试附近新增测试。通过 `getBoundingClientRect` mock 返回：

```js
const triggerRect = { left: 200, top: 440, right: 760, bottom: 470, width: 560, height: 30 }
const tooltipRect = { left: 0, top: 0, right: 640, bottom: 60, width: 640, height: 60 }
```

将视口高度设为 500px，触发 600ms 悬浮并等待日志加载，然后断言：

```js
expect(tooltip.style.top).toBe('368px')
expect(tooltip.style.left).toBe('200px')
```

其中 `368 = 440 - 60 - 12`，证明使用实际高度而不是最大高度。

**Step 2: 运行测试确认 RED**

Run: `npm test -- --run frontend/src/App.test.js -t "positions a short lifecycle error tooltip directly above its trigger"`

Expected: FAIL，当前实现返回按 300px/320px 预估的过高坐标。

### Task 3: 使用真实尺寸计算悬浮位置

**Files:**
- Modify: `frontend/src/components/ProjectSidebar.vue`
- Test: `frontend/src/App.test.js`

**Step 1: 保存锚点并隐藏未测量浮层**

将位置初始值改为 `null`，保存当前触发元素。未测量时返回：

```js
return { visibility: 'hidden' }
```

隐藏悬浮层时同时清除触发元素和位置。

**Step 2: 实现真实矩形定位**

定位函数接收 `triggerRect` 和 `tooltipRect`，按真实尺寸计算：

```js
const belowTop = triggerRect.bottom + lifecycleErrorTooltipOffset
const aboveTop = triggerRect.top - tooltipRect.height - lifecycleErrorTooltipOffset
const rawTop = belowTop + tooltipRect.height <= viewportHeight - lifecycleErrorTooltipOffset
  ? belowTop
  : aboveTop
```

横向和纵向最终值都使用视口 12px 安全边距限制。

**Step 3: 日志挂载后测量**

监听当前可见错误键对应的完整输出。输出存在时等待 `nextTick()`，从 tooltip layer 取得当前 `.todo-lifecycle-error-tooltip`，再次确认错误键仍有效，然后读取两个矩形并写入最终位置。

**Step 4: 运行聚焦测试确认 GREEN**

Run: `npm test -- --run frontend/src/App.test.js -t "positions a short lifecycle error tooltip directly above its trigger"`

Expected: PASS。

**Step 5: 增加下方优先断言**

使用下方空间充足的触发矩形，断言 `top` 等于 `triggerRect.bottom + 12`，防止修复只覆盖向上翻转。

**Step 6: 复跑生命周期悬浮测试**

Run: `npm test -- --run frontend/src/App.test.js -t "lifecycle"`

Expected: 相关测试全部通过。

### Task 4: 完整验证与交付检查

**Files:**
- Modify: `openspec/changes/add-lifecycle-script-error-log-copy/tasks.md`

**Step 1: 运行前端全量测试**

Run: `npm test`

Expected: 退出码为 0，无失败测试。

**Step 2: 运行前端生产构建**

Run: `npm run build`

Expected: Vite 构建成功。

**Step 3: 运行差异检查**

Run: `git diff --check`

Expected: 无输出，退出码为 0。

**Step 4: 完成 OpenSpec 任务**

勾选本轮三个任务，并运行：

Run: `openspec instructions apply --change add-lifecycle-script-error-log-copy --json`

Expected: `state` 为 `all_done`，`remaining` 为 0。

**Step 5: 审查检查点**

检查最终差异，不执行 Git 提交；仅在用户明确授权后提交。

