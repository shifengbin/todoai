## MODIFIED Requirements

### Requirement: Import Projects From Parent Directory

系统 SHALL 允许用户选择一个父目录，并将该父目录下第一层 Git 仓库子目录批量导入为项目。系统 SHALL 跳过普通文件、不可访问目录、非 Git 子目录和已存在的项目路径。批量导入入口 SHALL 在用户 hover 时提示仅导入一级子目录中的 Git 仓库。

#### Scenario: User imports Git child directories from a parent directory

- **WHEN** 用户选择父目录 `/home/user/work`
- **AND** 该目录包含 Git 仓库子目录 `/home/user/work/frontend-app`
- **AND** 该目录包含 Git 仓库子目录 `/home/user/work/api-service`
- **THEN** 项目库包含项目 `frontend-app`，路径为 `/home/user/work/frontend-app`
- **AND** 项目库包含项目 `api-service`，路径为 `/home/user/work/api-service`

#### Scenario: Import skips non-Git child directories

- **WHEN** 用户选择父目录 `/home/user/work`
- **AND** 该目录包含 Git 仓库子目录 `/home/user/work/frontend-app`
- **AND** 该目录包含非 Git 子目录 `/home/user/work/docs`
- **THEN** 项目库包含项目 `frontend-app`，路径为 `/home/user/work/frontend-app`
- **AND** 项目库不包含项目 `docs`
- **AND** 导入结果显示 `/home/user/work/docs` 被跳过

#### Scenario: Import skips duplicate project paths

- **WHEN** 项目库已包含路径 `/home/user/work/frontend-app`
- **AND** 用户从父目录 `/home/user/work` 批量导入
- **THEN** 系统不会创建第二个 `/home/user/work/frontend-app` 项目
- **AND** 导入结果显示该路径被跳过

#### Scenario: Import ignores non-directory children

- **WHEN** 用户选择父目录 `/home/user/work`
- **AND** 该目录包含文件 `/home/user/work/readme.md`
- **THEN** 系统不会为 `readme.md` 创建项目

#### Scenario: Bulk import button explains Git-only behavior

- **WHEN** 用户将鼠标放到批量导入按钮上
- **THEN** 系统提示 `仅导入一级子目录中的 Git 仓库`

#### Scenario: User cancels parent directory import

- **WHEN** 用户打开父目录导入选择器并取消
- **THEN** 项目库保持不变
