## ADDED Requirements

### Requirement: Manage Todo Action Confirmation Popovers

系统 SHALL 在 TODO item 的完成和删除操作上使用按钮旁确认气泡。系统 SHALL 在同一时间最多显示一个侧边栏浮层；打开 TODO 操作确认气泡 SHALL 关闭终端启动菜单和 TODO 项目移除确认气泡，打开其它侧边栏浮层 SHALL 关闭 TODO 操作确认气泡。确认气泡 SHALL 支持取消、点击外部关闭和确认成功后关闭。

#### Scenario: Complete confirmation popover opens beside complete action

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户点击该 TODO item 上的完成按钮
- **THEN** 系统在完成按钮旁显示完成确认气泡
- **AND** 系统不立即完成 TODO `修复登录问题`

#### Scenario: Delete confirmation popover opens beside delete action

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户点击该 TODO item 上的删除按钮
- **THEN** 系统在删除按钮旁显示删除确认气泡
- **AND** 系统不立即删除 TODO `修复登录问题`

#### Scenario: Opening another sidebar popover closes todo action popover

- **WHEN** TODO `修复登录问题` 的完成确认气泡已显示
- **AND** 用户打开 TODO 下项目 `frontend-app` 的终端启动菜单
- **THEN** 系统关闭 TODO `修复登录问题` 的完成确认气泡
- **AND** 系统显示 `frontend-app` 的终端启动菜单

#### Scenario: Outside click closes todo action popover without action

- **WHEN** TODO `修复登录问题` 的删除确认气泡已显示
- **AND** 用户点击确认气泡外部
- **THEN** 系统关闭删除确认气泡
- **AND** TODO `修复登录问题` 保持在活动任务列表中

### Requirement: Display Todo Project Row State

系统 SHALL 在 TODO 工作树中用整行背景表达 TODO 下项目行的悬停和选中状态。TODO 项目行背景 SHALL 覆盖该项目 header 的项目信息区域、创建终端按钮区域和删除按钮区域。TODO 项目行上的创建终端按钮和删除按钮 SHALL 在整行背景上保持可读，并 SHALL 保留各自的 hover 和 focus 反馈。

#### Scenario: Active todo project row background covers action buttons

- **WHEN** TODO `修复登录问题` 下的项目 `frontend-app` 是当前选中的 TODO 项目上下文
- **THEN** `frontend-app` 项目行的选中背景覆盖项目名称和路径区域
- **AND** 选中背景覆盖创建终端按钮区域
- **AND** 选中背景覆盖删除按钮区域

#### Scenario: Hovered todo project row background covers action buttons

- **WHEN** 用户将鼠标悬停在 TODO `修复登录问题` 下的项目 `frontend-app` 行上
- **THEN** `frontend-app` 项目行的悬停背景覆盖项目名称和路径区域
- **AND** 悬停背景覆盖创建终端按钮区域
- **AND** 悬停背景覆盖删除按钮区域

## MODIFIED Requirements

### Requirement: Complete Todo

系统 SHALL 允许用户通过 TODO item 完成按钮旁的确认气泡完成活动 TODO。完成 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 归档为已完成状态。用户取消确认 SHALL 不改变 TODO 或其终端状态。

#### Scenario: User completes a todo

- **WHEN** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的完成按钮
- **AND** 系统在完成按钮旁显示完成确认气泡
- **AND** 用户在确认气泡中确认完成
- **THEN** 系统关闭该 TODO 下所有运行中 shell 进程
- **AND** 系统从运行时状态移除该 TODO 下所有终端
- **AND** TODO `修复登录问题` 不再显示在活动任务列表中
- **AND** TODO `修复登录问题` 显示在归档列表中且归档原因为 `completed`

#### Scenario: User cancels todo completion

- **WHEN** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的完成按钮
- **AND** 系统在完成按钮旁显示完成确认气泡
- **AND** 用户取消确认
- **THEN** TODO `修复登录问题` 保持在活动任务列表中
- **AND** 该 TODO 下的终端保持运行时状态不变

### Requirement: Delete Todo

系统 SHALL 允许用户通过 TODO item 删除按钮旁的确认气泡删除活动 TODO。删除 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 归档为删除状态。用户取消确认 SHALL 不改变 TODO 或其终端状态。

#### Scenario: User deletes a todo

- **WHEN** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的删除按钮
- **AND** 系统在删除按钮旁显示删除确认气泡
- **AND** 用户在确认气泡中确认删除
- **THEN** 系统关闭该 TODO 下所有运行中 shell 进程
- **AND** 系统从运行时状态移除该 TODO 下所有终端
- **AND** TODO `修复登录问题` 不再显示在活动任务列表中
- **AND** TODO `修复登录问题` 显示在归档列表中且归档原因为 `deleted`

#### Scenario: User cancels todo deletion

- **WHEN** 用户点击 TODO `修复登录问题` 的删除按钮
- **AND** 系统在删除按钮旁显示删除确认气泡
- **AND** 用户取消确认
- **THEN** TODO `修复登录问题` 保持在活动任务列表中
- **AND** 该 TODO 下的终端保持运行时状态不变
