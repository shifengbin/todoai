# global-project-candidates Specification

## Purpose
TBD - created by archiving change separate-global-project-candidates. Update Purpose after archive.
## Requirements
### Requirement: Manage Global Project Candidates

系统 SHALL 在应用级配置中维护跨所有 workspace 共享的全局项目候选库。全局项目候选 SHALL 保存项目 ID、名称、绝对路径、路径可用性、创建时间和最近选择时间。全局项目候选 SHALL 只作为创建 TODO 或给 TODO 添加工程时的候选项，不作为已加入 TODO 工程的事实数据源。

#### Scenario: Global candidates are shared across workspaces

- **WHEN** 用户在 workspace `/work/customer-a` 中导入候选项目 `/repo/frontend-app`
- **AND** 用户打开 workspace `/work/customer-b`
- **THEN** 创建 TODO 或添加工程弹窗中仍可选择候选项目 `frontend-app`
- **AND** 候选项目路径为 `/repo/frontend-app`

#### Scenario: No workspace still has global candidates available for management

- **WHEN** 当前没有打开 workspace
- **THEN** 系统仍可加载全局项目候选库
- **AND** 系统 SHALL NOT 允许创建 TODO 或向 TODO 添加工程

### Requirement: Import Global Project Candidates

系统 SHALL 允许用户从单个工程目录或父目录导入一个或多个全局项目候选。导入目标 SHALL 是 Git 仓库。单个工程目录导入 SHALL 只判断用户选择的目录本身是否为 Git 仓库；若所选目录不是 Git 仓库，系统 SHALL 询问用户是否初始化 Git 仓库后导入。父目录导入 SHALL 遍历父目录下的直接子目录，并只保存直接子目录中已经是 Git 仓库的目录。导入 SHALL 按规范化绝对路径去重；已存在的路径 SHALL 不创建重复候选。

#### Scenario: User imports a single Git global candidate project

- **WHEN** 用户从创建 TODO、编辑 TODO 或添加工程弹窗选择工程目录 `/repo/frontend-app`
- **AND** `/repo/frontend-app` 是 Git 仓库
- **THEN** 全局候选项目库包含 `frontend-app`
- **AND** 候选项目路径为 `/repo/frontend-app`

#### Scenario: User initializes a single non-Git candidate before import

- **WHEN** 用户从创建 TODO、编辑 TODO 或添加工程弹窗选择工程目录 `/repo/frontend-app`
- **AND** `/repo/frontend-app` 不是 Git 仓库
- **AND** 用户确认初始化 Git 仓库后导入
- **THEN** 系统在 `/repo/frontend-app` 执行 Git 初始化
- **AND** 全局候选项目库包含 `frontend-app`
- **AND** 候选项目路径为 `/repo/frontend-app`

#### Scenario: User declines initialization for a single non-Git candidate

- **WHEN** 用户从创建 TODO、编辑 TODO 或添加工程弹窗选择工程目录 `/repo/frontend-app`
- **AND** `/repo/frontend-app` 不是 Git 仓库
- **AND** 用户拒绝初始化 Git 仓库
- **THEN** 全局候选项目库不包含 `/repo/frontend-app`
- **AND** 系统显示 toast `只能导入 Git 仓库`
- **AND** 该 toast 在 2 秒后自动消失

#### Scenario: User imports Git global candidates from parent directory

- **WHEN** 用户从添加工程弹窗选择父目录 `/repo`
- **AND** `/repo` 下包含 Git 仓库目录 `/repo/frontend-app`
- **AND** `/repo` 下包含 Git 仓库目录 `/repo/api-service`
- **THEN** 全局候选项目库包含 `frontend-app`
- **AND** 全局候选项目库包含 `api-service`

#### Scenario: Parent directory import skips non-Git child directories

- **WHEN** 用户从添加工程弹窗选择父目录 `/repo`
- **AND** `/repo` 下包含 Git 仓库目录 `/repo/frontend-app`
- **AND** `/repo` 下包含非 Git 目录 `/repo/docs`
- **THEN** 全局候选项目库包含 `frontend-app`
- **AND** 全局候选项目库不包含 `docs`
- **AND** 导入摘要显示 `/repo/docs` 被跳过或未新增

#### Scenario: Duplicate imported paths are skipped

- **WHEN** 全局候选项目库已包含路径 `/repo/frontend-app`
- **AND** 用户再次从父目录 `/repo` 导入
- **THEN** 全局候选项目库中只有一个路径为 `/repo/frontend-app` 的候选项目
- **AND** 导入摘要显示该路径被跳过或未新增

#### Scenario: Importing candidates does not refresh active git status immediately

- **WHEN** 用户导入一个或多个全局候选项目
- **THEN** 系统更新全局候选项目列表和导入摘要
- **AND** 系统不立即查询任何导入候选项目的 Git 状态

### Requirement: Clear Global Project Candidates

系统 SHALL 允许用户清空全局项目候选库。清空全局候选 SHALL 只删除候选项记录，SHALL NOT 删除磁盘目录，SHALL NOT 删除任何 workspace 中已经加入 TODO 的工程副本，SHALL NOT 关闭已有 TODO 工程终端。

#### Scenario: User clears global candidates

- **WHEN** 全局候选项目库包含 `frontend-app`
- **AND** TODO `修复登录问题` 下已保存工程副本 `frontend-app`
- **AND** 用户确认清空全局候选项目
- **THEN** 全局候选项目库为空
- **AND** TODO `修复登录问题` 下仍显示工程 `frontend-app`
- **AND** 该 TODO 工程的路径仍为原添加时保存的路径

#### Scenario: Clearing candidates does not delete directories

- **WHEN** 全局候选项目库包含路径 `/repo/frontend-app`
- **AND** 用户确认清空全局候选项目
- **THEN** 系统从候选库移除 `/repo/frontend-app`
- **AND** 系统不删除磁盘目录 `/repo/frontend-app`

### Requirement: Clear Single Global Project Candidate

系统 SHALL 允许用户从候选项目列表中清除单个全局项目候选。清除单个候选 SHALL 使用应用内自定义确认弹窗，SHALL NOT 使用系统原生确认框。清除单个候选 SHALL 只删除该候选项记录，SHALL NOT 删除磁盘目录，SHALL NOT 删除任何 workspace 中已经加入 TODO 的工程副本，SHALL NOT 关闭已有 TODO 工程终端。若被清除候选已经在当前未提交的创建 TODO、编辑 TODO 或添加工程弹窗中被选中，系统 SHALL 同步移除该临时选择。

#### Scenario: User clears one global candidate

- **WHEN** 全局候选项目库包含 `frontend-app` 和 `api-service`
- **AND** 用户在候选项目列表中请求清除候选项目 `frontend-app`
- **AND** 系统显示应用内清除确认弹窗
- **AND** 用户在该弹窗中确认清除
- **THEN** 全局候选项目库不再包含 `frontend-app`
- **AND** 全局候选项目库仍包含 `api-service`

#### Scenario: User cancels clearing one global candidate

- **WHEN** 全局候选项目库包含 `frontend-app`
- **AND** 用户在候选项目列表中请求清除候选项目 `frontend-app`
- **AND** 系统显示应用内清除确认弹窗
- **AND** 用户在该弹窗中取消确认
- **THEN** 全局候选项目库仍包含 `frontend-app`

#### Scenario: Clearing selected candidate removes pending selection

- **WHEN** 用户在创建 TODO、编辑 TODO 或添加工程弹窗中已选择候选项目 `frontend-app`
- **AND** 用户确认清除候选项目 `frontend-app`
- **THEN** 当前弹窗的待提交项目选择不再包含 `frontend-app`
- **AND** 后续提交不会引用已清除的候选项目 ID

#### Scenario: Clearing one candidate preserves TODO project copy and terminals

- **WHEN** 全局候选项目库包含路径 `/repo/frontend-app`
- **AND** TODO `修复登录问题` 下已保存工程副本 `frontend-app`
- **AND** 该 TODO 工程已有运行中的终端
- **AND** 用户确认清除候选项目 `frontend-app`
- **THEN** 系统从候选库移除 `/repo/frontend-app`
- **AND** 系统不删除磁盘目录 `/repo/frontend-app`
- **AND** TODO `修复登录问题` 下仍显示工程 `frontend-app`
- **AND** 该 TODO 工程终端保持运行

### Requirement: Migrate Workspace Projects Into Global Candidates

系统 SHALL 在打开旧 workspace 时将该 workspace 中已有项目库记录按路径合并到全局项目候选库。迁移 SHALL 按规范化绝对路径去重，并 SHALL 保留已存在全局候选的 ID。

#### Scenario: Legacy workspace projects are merged into global candidates

- **WHEN** 旧 workspace `/work/customer-a` 的持久化数据包含项目 `frontend-app` 路径 `/repo/frontend-app`
- **AND** 全局候选项目库尚不包含 `/repo/frontend-app`
- **AND** 用户打开 workspace `/work/customer-a`
- **THEN** 全局候选项目库包含 `frontend-app`
- **AND** 候选项目路径为 `/repo/frontend-app`

#### Scenario: Legacy migration deduplicates by path

- **WHEN** 全局候选项目库已包含路径 `/repo/frontend-app`
- **AND** 旧 workspace `/work/customer-a` 的持久化数据也包含路径 `/repo/frontend-app`
- **AND** 用户打开 workspace `/work/customer-a`
- **THEN** 全局候选项目库中只有一个路径为 `/repo/frontend-app` 的候选项目
- **AND** 已存在候选项目的 ID 保持不变
