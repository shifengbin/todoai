# application-appearance Specification

## Purpose
Defines persisted Light/Dark application appearance behavior for non-terminal UI surfaces while preserving embedded terminal colors.
## Requirements
### Requirement: Load Application Appearance Theme
系统 SHALL 在应用启动时加载已保存的外观主题，并 SHALL 将主题应用到终端以外的主界面。外观主题 SHALL 支持 `light` 和 `dark` 两个值。

#### Scenario: Missing saved theme uses light
- **WHEN** 应用加载设置
- **AND** 设置文件中没有已保存的外观主题
- **THEN** 系统使用 `light` 作为当前外观主题
- **AND** 主界面以浅色主题渲染

#### Scenario: Saved dark theme is restored
- **WHEN** 用户之前保存的外观主题是 `dark`
- **AND** 应用启动并加载设置
- **THEN** 系统将当前外观主题设为 `dark`
- **AND** 主界面以深色主题渲染

#### Scenario: Invalid saved theme falls back to light
- **WHEN** 设置文件中的外观主题不是 `light` 或 `dark`
- **AND** 应用加载设置
- **THEN** 系统使用 `light` 作为当前外观主题
- **AND** 应用启动不因无效主题值失败

### Requirement: Change Application Appearance Theme
系统 SHALL 允许用户从设置界面选择并保存应用外观主题。保存成功后，所选主题 SHALL 立即应用到当前应用窗口。

#### Scenario: User saves dark theme
- **WHEN** 设置界面显示外观主题选项
- **AND** 用户选择 `dark` 并保存设置
- **THEN** 系统持久化外观主题为 `dark`
- **AND** 当前应用窗口立即切换为深色主题

#### Scenario: User saves light theme
- **WHEN** 当前应用外观主题是 `dark`
- **AND** 用户在设置界面选择 `light` 并保存设置
- **THEN** 系统持久化外观主题为 `light`
- **AND** 当前应用窗口立即切换为浅色主题

#### Scenario: User cancels theme edit
- **WHEN** 当前应用外观主题是 `light`
- **AND** 用户在设置界面选择 `dark`
- **AND** 用户取消设置修改
- **THEN** 当前应用外观主题仍为 `light`
- **AND** 系统不持久化 `dark` 主题

### Requirement: Theme Covers Application Surfaces
系统 SHALL 在当前外观主题下统一渲染终端以外的应用主要表面，包括项目侧栏、工作区头部、状态栏和设置弹窗。

#### Scenario: Dark theme covers shell surfaces
- **WHEN** 当前外观主题是 `dark`
- **THEN** 项目侧栏使用深色表面颜色
- **AND** 工作区头部使用深色表面颜色
- **AND** 状态栏使用深色表面颜色
- **AND** 设置弹窗使用深色表面颜色

#### Scenario: Light theme covers shell surfaces
- **WHEN** 当前外观主题是 `light`
- **THEN** 项目侧栏使用浅色表面颜色
- **AND** 工作区头部使用浅色表面颜色
- **AND** 状态栏使用浅色表面颜色
- **AND** 设置弹窗使用浅色表面颜色

### Requirement: Preserve Embedded Terminal Colors
系统 SHALL 保持嵌入式终端内容区、终端占位状态、终端右键菜单和 xterm 配色不随应用外观主题变化。

#### Scenario: Existing terminals keep colors after theme save
- **WHEN** 当前应用存在已打开的嵌入式终端
- **AND** 用户保存不同的外观主题
- **THEN** 已打开终端的 xterm 配色保持不变
- **AND** 终端会话不重启

#### Scenario: New terminal keeps fixed terminal theme
- **WHEN** 当前外观主题是 `dark`
- **AND** 用户创建新的嵌入式终端
- **THEN** 新终端使用固定的嵌入式终端配色
