## ADDED Requirements

### Requirement: Sort Active Todos

系统 SHALL 在 TODO tab 的活动 TODO 列表中提供排序切换控件，并 SHALL 支持按任务优先级排序和按创建时间排序。系统 SHALL 默认选择优先级排序。优先级排序 SHALL 为 `高`、`中`、`低`，相同优先级的活动 TODO SHALL 按 `createdAt` 创建时间正序展示，先创建的 TODO 排在前面。时间排序 SHALL 按 `createdAt` 创建时间正序展示，先创建的 TODO 排在前面。归档 TODO 列表 SHALL 不受活动 TODO 排序规则影响。

#### Scenario: Active todo sort control defaults to priority

- **WHEN** 用户打开 TODO tab 的活动 TODO 列表
- **THEN** 系统显示活动 TODO 排序切换控件
- **AND** 排序切换控件默认选中优先级排序

#### Scenario: Active todos are ordered by priority

- **WHEN** 活动 TODO 列表包含优先级为 `低` 的 TODO `整理文档`
- **AND** 活动 TODO 列表包含优先级为 `高` 的 TODO `修复登录问题`
- **AND** 活动 TODO 列表包含优先级为 `中` 的 TODO `升级依赖`
- **THEN** TODO tab 的活动 TODO 列表依次显示 `修复登录问题`、`升级依赖`、`整理文档`

#### Scenario: Active todos with same priority are ordered by creation time

- **WHEN** 活动 TODO 列表包含优先级同为 `高` 的 TODO `修复登录问题` 和 `排查线上报警`
- **AND** TODO `修复登录问题` 的 `createdAt` 早于 TODO `排查线上报警`
- **THEN** TODO tab 的活动 TODO 列表中 `修复登录问题` 排在 `排查线上报警` 前面

#### Scenario: User switches active todos to creation time order

- **WHEN** 活动 TODO 列表包含创建时间更晚且优先级为 `高` 的 TODO `修复登录问题`
- **AND** 活动 TODO 列表包含创建时间更早且优先级为 `低` 的 TODO `整理文档`
- **AND** 用户选择时间排序
- **THEN** TODO tab 的活动 TODO 列表中 `整理文档` 排在 `修复登录问题` 前面

#### Scenario: Archived todo order is unaffected

- **WHEN** 用户查看归档 TODO 列表
- **THEN** 系统不按活动 TODO 的优先级排序规则重排归档 TODO 列表
