## ADDED Requirements

### Requirement: Inspect And Copy Complete Lifecycle Script Failure Output

系统 SHALL 在开始 TODO 触发的初始化脚本或完成 TODO 触发的完成脚本执行失败时，于当前应用运行期间保留该次执行未经截断的 stdout/stderr 完整合并输出。系统 SHALL 继续在普通失败状态中展示截断后的错误摘要，并 SHALL 仅在用户请求查看或复制时按需读取完整输出。失败状态 SHALL 在 Retry 操作旁提供复制错误日志按钮，复制内容 SHALL 与该次执行的完整原始合并输出完全一致且 SHALL NOT 追加阶段、脚本名称或其他格式化文本；当进程没有产生输出时，系统 SHALL 回退使用执行错误信息。系统 SHALL 在用户悬浮失败状态时以保留换行、尺寸受限且可滚动的悬浮详情展示同一份完整输出。系统 SHALL 在该阶段重新执行、执行成功、TODO 被删除、TODO 成功完成归档或 workspace 切换时清除已经失效的完整输出，且 SHALL NOT 跨应用重启持久化该输出。

#### Scenario: Starting todo exposes complete initialization failure output

- **WHEN** 用户点击开始 TODO 并触发初始化脚本
- **AND** 初始化脚本执行失败且产生超过 4096 字节的多行合并输出
- **THEN** 失败状态继续展示截断后的单行错误摘要
- **AND** 用户悬浮失败状态后可以在可滚动详情中查看未经截断的完整多行输出
- **AND** 详情包含该次输出的开头、结尾和原始换行

#### Scenario: Completing todo allows exact failure output to be copied

- **WHEN** 用户点击完成 TODO 并触发完成脚本
- **AND** 完成脚本执行失败并产生原始合并输出
- **THEN** 系统在 Retry 操作旁展示复制错误日志按钮
- **WHEN** 用户点击复制错误日志按钮
- **THEN** 系统将该次完成脚本的完整原始合并输出写入剪贴板
- **AND** 复制内容不包含系统追加的阶段、脚本名称或其他说明文本

#### Scenario: Failure without process output copies error message

- **WHEN** 初始化脚本或完成脚本在产生进程输出前执行失败
- **AND** 当前失败没有 stdout/stderr 合并输出
- **WHEN** 用户查看或复制该失败的错误日志
- **THEN** 系统展示或复制该次执行错误信息

#### Scenario: Retrying replaces obsolete failure output

- **WHEN** 生命周期脚本失败状态具有完整错误输出
- **AND** 用户点击 Retry 重新执行同一阶段脚本
- **THEN** 系统立即使上一次失败的完整输出失效
- **AND** 运行中状态不提供上一次失败日志的查看或复制入口
- **WHEN** 本次重新执行再次失败
- **THEN** 系统只提供本次失败的完整原始输出

#### Scenario: Runtime cleanup removes complete failure output

- **WHEN** 生命周期脚本失败状态具有完整错误输出
- **AND** 对应脚本随后执行成功、TODO 被删除、TODO 成功完成归档或用户切换 workspace
- **THEN** 系统清除该失败状态对应的完整输出
- **AND** 后续请求不能读取已经失效的日志

#### Scenario: Complete failure output is not persisted

- **WHEN** 生命周期脚本执行失败并保留了完整原始输出
- **AND** 用户关闭并重新启动应用
- **THEN** 系统不从持久化数据恢复该次完整输出
