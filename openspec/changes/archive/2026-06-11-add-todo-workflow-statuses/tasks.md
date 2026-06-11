## 1. 后端状态模型

- [x] 1.1 将 TODO 状态常量扩展为 `not-started`、`in-progress`、`completed`，并保留旧 `active`、`archived` 读取兼容逻辑
- [x] 1.2 增加统一的开放 TODO 判断辅助函数，使 `not-started` 和 `in-progress` 都可被选择、编辑、关联项目和创建终端
- [x] 1.3 修改创建 TODO 逻辑：默认状态为 `not-started`，并且不改变当前 TODO、TODO 项目、项目或终端上下文
- [x] 1.4 增加手动切换 TODO 工作流状态的后端方法，仅允许 `not-started` 与 `in-progress` 互相切换
- [x] 1.5 修改完成逻辑：开放 TODO 完成后保存为 `completed`，保留完成时间和项目快照
- [x] 1.6 修改删除逻辑：开放 TODO 删除后从可见 TODO 集合移除，并移除关联项目
- [x] 1.7 增加旧数据规范化：`active` 映射为 `not-started`，`archived/completed` 映射为 `completed`，`archived/deleted` 不进入可见 TODO 列表

## 2. Wails API 与绑定

- [x] 2.1 在 App 层暴露手动切换 TODO 状态的方法，并确保切换状态后返回包含 shell 状态的 `ProjectState`
- [x] 2.2 更新完成和删除 API 的调用路径，继续关闭对应 TODO 下的终端
- [x] 2.3 重新生成或更新 Wails 前端绑定文件，使前端可以调用新的状态切换方法

## 3. 前端 TODO 工作区

- [x] 3.1 将 TODO 顶部视图从 `Active` / `Archived` 改为 `未执行`、`执行中`、`已完成`
- [x] 3.2 按 `todo.status` 分别计算未执行、执行中、已完成 TODO 列表，并让已完成列表只包含 `completed`
- [x] 3.3 为未执行和执行中视图复用现有优先级/创建时间排序控件和排序逻辑
- [x] 3.4 在 TODO 行增加手动状态切换入口：未执行可切为执行中，执行中可切回未执行
- [x] 3.5 调整完成和删除后的 UI 展示：完成进入已完成，删除不显示在任何 TODO 状态视图中
- [x] 3.6 实现新建 TODO 默认收起，并避免创建后的自动选中逻辑把新 TODO 立即展开
- [x] 3.7 调整批量展开/收起逻辑，使其作用于当前未执行或执行中视图中的 TODO
- [x] 3.8 更新空状态文案、按钮 `data-testid`、aria 标签和标题，使其匹配三栏状态视图

## 4. 自动化测试

- [x] 4.1 增加后端测试覆盖新建默认 `not-started`、手动状态切换、完成进入 `completed`、删除不进入完成列表
- [x] 4.2 增加后端测试覆盖旧 `active`、`archived/completed`、`archived/deleted` 数据兼容
- [x] 4.3 更新 App 层测试，覆盖状态切换 API、完成关闭终端、删除关闭终端和返回状态
- [x] 4.4 更新客户端组件测试，覆盖三栏切换、未执行/执行中排序、已完成只显示 completed、删除隐藏
- [x] 4.5 更新客户端组件测试，覆盖新建 TODO 默认收起和当前状态视图批量展开/收起
- [x] 4.6 更新客户端集成测试，覆盖创建、手动切换、完成、删除的用户路径

## 5. 验证与 Review

- [x] 5.1 运行 Go 测试，确认后端状态模型和兼容逻辑通过
- [x] 5.2 运行前端自动化测试，确认客户端三栏状态视图和交互通过
- [x] 5.3 运行 OpenSpec 校验，确认 proposal、design、specs、tasks 可被解析
- [x] 5.4 执行自动 review，检查状态迁移、删除语义、UI 可访问性和测试覆盖是否符合变更要求
