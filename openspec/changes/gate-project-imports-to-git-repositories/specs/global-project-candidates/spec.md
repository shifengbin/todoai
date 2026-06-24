## MODIFIED Requirements

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
