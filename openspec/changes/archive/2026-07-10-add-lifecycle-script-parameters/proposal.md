## Why

当前 TODO 生命周期脚本只能保存固定脚本文本，同一个脚本模板在不同 TODO 上需要不同输入时，用户必须复制并修改脚本模板或在脚本中手动编辑值。为生命周期脚本增加参数能力，可以让脚本模板复用固定流程，并在创建 TODO 时填写本次任务专属参数。

## What Changes

- 脚本管理支持在生命周期脚本模板中定义明文字符串参数，包括参数名、显示名称、描述、默认值和是否必填。
- 创建 TODO 选择生命周期脚本后，表单根据该脚本的参数定义展示参数输入项，并允许用户设置本次 TODO 的参数值。
- 创建 TODO 时保存脚本内容、参数定义和参数值的快照，后续全局脚本模板修改不影响已创建 TODO。
- 执行初始化脚本或完成脚本前，系统将脚本文本中的参数占位符渲染为当前 shell 可安全执行的字符串字面量，用于隔离 bash/zsh/fish、PowerShell 和 cmd 的语法差异。
- 参数值以明文保存和展示，不支持 secret 参数类型。
- 不引入破坏性变更；未定义参数的现有生命周期脚本继续按当前行为执行。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `todo-lifecycle-scripts`: 增加生命周期脚本参数定义、创建 TODO 时参数取值、快照保存，以及脚本执行前的参数占位符渲染要求。

## Impact

- 后端 Go 数据模型：`TodoLifecycleScriptTemplate`、`TodoLifecycleScriptSnapshot`、`CreateTodoRequest`/项目状态持久化、生命周期脚本执行请求。
- 后端脚本执行：参数校验、占位符渲染、shell 字符串字面量转义、初始化/完成脚本重试时使用同一快照参数。
- 前端 Vue：脚本管理表单、创建 TODO 表单、脚本快照组装、Wails 生成模型。
- 测试：设置持久化、TODO 创建快照、生命周期脚本执行渲染、跨 shell 转义规则、前端参数输入交互。
