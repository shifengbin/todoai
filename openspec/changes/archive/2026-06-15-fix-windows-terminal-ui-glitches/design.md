## Context

当前应用由 Go 后端通过 PTY/ConPTY 启动 shell，前端用 xterm.js 渲染终端输出。PowerShell/pwsh 集成会通过 OSC 777 发出应用私有 `command-start` / `command-end` 事件，后端 `commandStateOutputFilter` 会在输出渲染和历史持久化前剥离这些 payload。前端还会监听 xterm 标题变化，并把标题中的 `working`、`thinking`、spinner 或 `!` 推断为终端活动状态。

Windows WebView + ConPTY 环境下出现三个相关问题：

- TODO 删除菜单或删除确认气泡仍可能显示在全局右上角，而不是对应 TODO 附近。
- 仅在 Windows 上通过创建终端下拉菜单启动非纯 `Terminal` 的 launch profile 时，用户可能看到一长串类似 base64 的未知控制文本；例如设置中配置 `calc` 后从下拉菜单启动也会出现。纯 `Terminal` 启动不触发该现象。该文本尚未确认是 base64、OSC 52、应用私有协议，还是其他 ConPTY 文本降级形态。
- Claude 打开后空闲状态被误判为 `busy`，导致 TODO 终端树长期显示运行中；实际忙碌状态表现为一个点在左/中/右位置之间循环切换。

这些问题分别落在侧栏浮层定位、后端输出过滤、前端 xterm 标题分类三个边界，但用户感知上都属于 Windows 终端工作流不稳定。

## Goals / Non-Goals

**Goals:**

- Windows 下 TODO 删除菜单和确认气泡 SHALL 贴近触发它们的 TODO 操作上下文，不得脱离到应用全局角落。
- 应用私有命令状态 payload SHALL 不被渲染或持久化为普通终端文本；Windows launch profile 中未知的类似 base64 文本 SHALL NOT 仅因形态相似而被特殊隐藏。
- Claude 启动后的初始/空闲标题变化 SHALL 不让终端长期显示 `busy`。
- 明确的工作中、spinner 和需要输入信号 SHALL 继续反映到终端树状态。
- 用单元测试覆盖可自动化的定位结构、payload 过滤和标题分类；用 Windows 手测覆盖 WebView 实际坐标、launch profile 启动路径和 Claude 活动状态行为。

**Non-Goals:**

- 不重写全局浮层系统，不引入 UI 组件库。
- 不改变 TODO 删除、项目删除、终端删除的后端语义。
- 不替换 ConPTY 后端，不改变 launch profile 数据格式。
- 不尝试完整解析所有第三方 OSC 协议；只处理已知会泄漏且可安全识别的形态。

## Decisions

### 1. 先明确“菜单定位”和“确认气泡定位”两个锚点

TODO 右键菜单使用 viewport 坐标和 `position: fixed`，确认气泡使用 TODO 操作区内的相对定位容器。实现时应分别验证：

- 右键菜单和三点按钮菜单是否使用正确的触发元素坐标。
- 删除确认气泡是否位于对应 TODO 的 action/control 容器内。
- 菜单和气泡是否会被侧栏滚动容器、transform 或 overflow 影响。

备选方案是把所有侧栏浮层改成 Teleport 到全局浮层层。这样可以绕过 overflow/stacking context，但变更面会扩大到项目删除、批量删除、添加终端菜单等既有浮层。本次优先修正 TODO 删除路径，并保留后续统一浮层系统的空间。

### 2. 只过滤应用私有 command-state，不对未知 base64-like 文本做特殊处理

后端仍是第一道过滤层：所有应用私有 OSC 777 payload 继续在 `commandStateOutputFilter` 中剥离，并产生命令状态事件。Windows 文本降级形态必须只在能确认是应用私有协议时消费，避免吞掉普通命令输出。

对于 Windows 下通过 launch profile 自动提交任意命令时出现的类 base64 未知文本，本次取消启发式过滤。即使文本看起来像 base64、OSC 降级片段或其它控制文本，只要它不匹配当前支持的应用私有 command-state 协议，就作为普通输出渲染和持久化。

```
PTY output
    │
    ├─ raw OSC 777 tui-helper         -> consume + emit command state
    ├─ textual 777;tui-helper         -> consume + emit command state
    └─ unknown/base64-like output      -> render + persist
```

备选方案是在前端 xterm.write 前统一正则删除所有类 base64 文本，或继续增加 Windows 专用 launch profile 文本形态。两个方案都有误删真实命令输出、日志和 token 的风险，本次不采用。

### 3. Claude 活动状态采用基线优先和连续动画识别的标题分类

当前规则只要标题包含 `working`、`thinking` 或 spinner 就立即标记 `busy`。对 Claude 这类 TUI，启动或空闲标题也可能包含这些片段，因此分类应先建立 idle baseline，再处理活动信号：

- 与 shell 名称、当前 command label、launch profile 名称、稳定程序标题匹配的标题视为 idle baseline。
- 没有 idle baseline 时，首个 Claude 或其他交互式程序的稳定标题不得直接让终端长期进入 `busy`。
- Claude 的点状忙碌信号应按连续 title 帧识别：只有观察到同一 terminal 的点在左/中/右位置之间切换，才将其视为 busy。
- 单个点帧、初始标题或没有后续动画变化的标题不得让 Claude 长期保持 busy。
- 短暂 busy 标题可以显示 busy，但当标题回到 idle baseline、命令结束、shell 退出或收到明确 idle/title reset 时必须恢复 idle。
- `!` 仍优先表示 `needs-input`。

备选方案是为 Claude 禁用所有标题活动推断。这样能避免误报，但会丢失真实工作中和等待输入提示，不采用。

### 4. 测试覆盖根因而不是只覆盖表象

jsdom 无法真实测量 Windows WebView 坐标，因此前端单元测试应锁住 DOM 归属、定位类名、坐标计算和状态转换；Go 单元测试应锁住过滤器输入输出和历史持久化。Windows 手测用于验证真实渲染：

- TODO 右键菜单和三点菜单触发删除，确认气泡出现在对应 TODO 附近。
- 在 Windows 上通过 launch profile 启动 `calc`、Claude、Codex 或其他命令后不显示受支持的应用私有 `777;tui-helper` command-state payload；未知的类 base64 文本不由过滤器特殊隐藏。
- Claude 空闲时终端树为 idle，真实工作中或等待输入时状态正确变化。

## Risks / Trade-offs

- [Windows 实际泄漏形态与现有假设不同] → 先用最小诊断或可复现样例捕获原始 title/output；本次不基于未知形态新增过滤。
- [过滤器误删普通输出] → 只消费当前支持的应用私有 command-state 协议；普通或未知类 base64 文本必须保留。
- [Claude 真正忙碌时被判断为 idle] → 保留明确 spinner、working/thinking、连续左/中/右点动画和 `!` 信号，但要求能回到 idle baseline。
- [浮层修复影响其他侧栏弹层] → 先限定 TODO 删除路径；如果需要共享 helper，只抽取坐标/placement 纯函数并用测试保护。
- [自动化无法覆盖 Windows WebView 坐标] → 在 tasks 中加入 Windows 手测清单，并把可验证的 DOM/CSS 约束纳入单元测试。

## Migration Plan

不需要数据迁移。变更只影响前端渲染、终端输出过滤和运行时活动状态。

发布后新建或重启的终端使用新的过滤和标题分类。已有终端历史不会被批量重写，但新的输出不会继续追加已识别的应用私有 command-state payload；未知的类 base64 文本不做特殊清理。

回滚时恢复前端定位/状态分类代码和后端过滤器改动即可；项目、TODO、settings、launch profile 和终端历史数据格式保持兼容。

## Open Questions

- Windows 下通过 launch profile 启动任意命令时实际泄漏的可见字符串前缀是 `777;tui-helper`、OSC 文本降级，还是其他控制文本形态？
- Claude 左/中/右点动画在 xterm title 中的精确字符串帧是什么，需要实现阶段用日志或手测样例确认。
- TODO 删除问题在 Windows 上指的是右键菜单错位、三点菜单错位，还是点击 `Delete TODO` 后的确认气泡错位？
