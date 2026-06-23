## ADDED Requirements

### Requirement: Start Shell In Workspace Directory

系统 SHALL 支持为 workspace 全局终端启动 embedded shell。全局终端 shell 进程 SHALL 使用当前 workspace 根目录作为工作目录，并 SHALL 与 TODO project 终端使用相同的 shell 设置、输入、输出、resize 和 clipboard 能力。

#### Scenario: Workspace global shell starts in workspace root

- **WHEN** 当前 workspace 路径为 `/home/user/work/customer-a`
- **AND** 用户创建全局终端
- **THEN** 系统启动 embedded shell
- **AND** shell 工作目录为 `/home/user/work/customer-a`
- **AND** shell 使用当前配置的终端 shell

#### Scenario: Workspace global shell uses launch fallback

- **WHEN** 当前 workspace 路径为 `/home/user/work/customer-a`
- **AND** 保存的终端 shell 不可用
- **AND** 自动检测选择 `/bin/sh` 作为 fallback
- **AND** 用户创建全局终端
- **THEN** 系统使用 `/bin/sh` 启动全局终端 shell
- **AND** shell 工作目录为 `/home/user/work/customer-a`

### Requirement: Isolate Workspace Global Terminal Sessions

系统 SHALL 将 workspace 全局终端与 TODO project 终端隔离。全局终端 SHALL 不属于任何 TODO project context，TODO 删除、TODO project 移除和项目候选删除 SHALL NOT 删除全局终端。关闭 workspace SHALL 关闭运行中的全局终端进程。

#### Scenario: Todo deletion preserves global terminals

- **WHEN** 全局终端 A 正在运行
- **AND** TODO `修复登录问题` 下的终端 B 正在运行
- **AND** 用户删除 TODO `修复登录问题`
- **THEN** 终端 B 被关闭并移除
- **AND** 全局终端 A 继续运行并保持显示

#### Scenario: Project candidate deletion preserves global terminals

- **WHEN** 全局终端 A 正在运行
- **AND** 用户删除全局项目候选 `frontend-app`
- **THEN** 全局终端 A 继续运行并保持显示
- **AND** 全局终端 A 的工作目录仍为当前 workspace 根目录

#### Scenario: Closing workspace closes global terminals

- **WHEN** 全局终端 A 正在运行
- **AND** 用户关闭当前 workspace
- **THEN** 系统关闭全局终端 A 的 shell 进程
- **AND** 运行时终端状态被清空
