## Context

当前 TODO 工作树已经支持在 TODO 收起时聚合隐藏终端的活动状态，聚合状态来自 `busy`、`needs-input`、`idle`，并且 `needs-input` 高于 `busy`。现有前端呈现把折叠 TODO 的聚合状态放在标题行右侧的小图标上，并复用终端行的转圈和警告图标。

这个呈现方式在父级 TODO 摘要层级不够醒目，也容易让用户把 TODO 聚合状态和具体终端状态混为一谈。TODO item 本身已经有整条 header 背景承载优先级，因此新的活动状态反馈应叠加在 TODO header 整行上，而不是只在标题区域添加一个局部图标。

## Goals / Non-Goals

**Goals:**

- 折叠 TODO 存在隐藏终端活动时，用整条 TODO header 的呼吸效果表达聚合状态。
- 区分 `busy` 和 `needs-input` 两种聚合活动状态，并保持 `needs-input` 的最高优先级。
- 保留展开 TODO 时由具体终端行展示活动状态的行为。
- 保留终端行自身的转圈和警告图标。

**Non-Goals:**

- 不改变终端活动状态模型、聚合优先级或后端数据结构。
- 不改变 TODO 工作流状态，例如 `not-started`、`in-progress`、`completed`。
- 不新增外部动画或图标依赖。
- 不重新设计 TODO 优先级颜色体系。

## Decisions

1. 折叠 TODO 的活动反馈挂到 TODO header 整行。

   TODO header 覆盖展开控件、标题摘要和行外操作按钮，是用户识别一个 TODO item 的完整横向区域。将呼吸效果挂在 header 上可以避免只提示标题区域造成的视觉割裂，也符合当前优先级背景覆盖整行 header 的设计。

   备选方案是只在标题区域 `.todo-row` 上呼吸，改动更小，但不符合“整行”反馈；另一种方案是在左侧增加呼吸状态条，比较克制但仍不够明显。

2. 折叠 TODO 不再渲染终端活动小图标。

   终端行的小图标适合表达具体终端的精细状态；TODO 行是父级聚合摘要，使用整行背景更容易被扫到。移除 TODO 行上的 `LoaderCircle` / `CircleAlert` 复用后，终端行仍保留现有图标。

3. 呼吸效果只表达有意义的活动状态。

   `busy` 使用较轻的 accent 色背景或边框呼吸；`needs-input` 使用更醒目的 warning 色背景或边框呼吸。`idle` 或没有隐藏终端时不显示呼吸动画，避免静止状态也持续吸引注意。

4. 语义状态继续保留在 DOM 上。

   折叠 TODO 仍应暴露聚合活动状态，例如通过 `data-activity-state`、title 或 aria label，让测试和辅助技术能够判断状态。视觉变化不应变成只能通过 CSS 动画推断的状态。

## Risks / Trade-offs

- [风险] 呼吸背景可能与 TODO 优先级背景冲突。→ 使用透明度较低的叠加色、边框或伪元素动画，并验证高/中/低优先级下的可读性。
- [风险] 动画过强会干扰列表扫描。→ `busy` 使用较弱节奏，`needs-input` 更明显但仍保持文本可读。
- [风险] 测试仍依赖旧的 `todo-activity-*` 图标。→ 更新组件测试，断言行级状态和旧图标不存在。
- [风险] reduced motion 用户可能不适合持续动画。→ 样式应通过 `prefers-reduced-motion` 降低或取消呼吸动画，同时保留静态状态色。
