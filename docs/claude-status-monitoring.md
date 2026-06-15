# Claude Code 状态外部监控方案

> 目标：从 Claude Code 进程外部，准确判断每一个正在运行的 Claude 实例当前的状态（空闲 / 思考中 / 工具执行 / 等待输入），并能区分多开的多个实例。

---

## 1. 总体思路

Claude Code 提供了**生命周期 Hook**机制：在会话发生关键事件（用户输入、工具调用、回答结束、等待确认等）时，会执行你在 `settings.json` 里配置好的 shell 命令，并通过 **stdin 传一个 JSON 负载**（包含 `session_id`、`cwd`、`transcript_path`、`tool_name` 等字段）。

我们利用这一点：

1. 在每个关键事件触发时，把当前状态写入按 `session_id` 命名的状态文件；
2. 外部监控程序读取或 `inotifywait` 监听这些文件，即可准确知道每个实例的状态。

这是延迟最低、最可靠的方式 —— 不依赖解析日志，也不依赖进程检测的猜测。

---

## 2. 可用的 Hook 事件

| 事件名 | 触发时机 | 对应状态 |
|---|---|---|
| `UserPromptSubmit` | 用户刚提交输入，Claude 开始处理 | `busy` |
| `PreToolUse` | 即将调用某个工具 | `tool:<name>` |
| `PostToolUse` | 工具调用结束 | `tool_done` |
| `Notification` | Claude 在等待用户确认权限或输入 | `waiting` |
| `Stop` | 一轮对话回合结束（变回空闲） | `idle` |
| `SubagentStop` | 子 Agent 结束 | （可选） |
| `SessionStart` | 会话开始 | `idle` |
| `SessionEnd` | 会话结束 | （删除状态文件） |

---

## 3. Hook 传入 stdin 的 JSON 字段

每次事件触发，Claude 都会向 hook 命令的 stdin 写入一个 JSON 对象，常见字段：

```json
{
  "session_id": "abc123...",          // 会话唯一 ID（区分多实例的关键）
  "transcript_path": "/home/.../xxx.jsonl",  // 该会话完整对话日志
  "cwd": "/home/shifengbin/project-a", // 该实例的工作目录
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",                 // 仅 PreToolUse/PostToolUse
  "tool_input": { ... },               // 仅 PreToolUse/PostToolUse
  "tool_response": { ... }             // 仅 PostToolUse
}
```

`session_id` 是区分多个并行 Claude 实例的**唯一可靠依据**。

---

## 4. 实现：状态采集脚本

把脚本放在 `~/.claude/bin/status-hook.sh`：

```bash
#!/usr/bin/env bash
# Claude Code 状态采集 hook
# 通过 stdin 接收 JSON，将状态写入 /tmp/claude/<session_id>.status

set -euo pipefail
mkdir -p /tmp/claude

payload=$(cat)

sid=$(jq -r '.session_id // "unknown"'        <<<"$payload")
event=$(jq -r '.hook_event_name // ""'         <<<"$payload")
cwd=$(jq -r '.cwd // ""'                       <<<"$payload")
tool=$(jq -r '.tool_name // ""'                <<<"$payload")
transcript=$(jq -r '.transcript_path // ""'    <<<"$payload")

# 把事件映射成人类可读的状态
case "$event" in
  UserPromptSubmit) status="busy" ;;
  PreToolUse)       status="tool:$tool" ;;
  PostToolUse)      status="tool_done" ;;
  Notification)     status="waiting" ;;
  Stop)             status="idle" ;;
  SessionStart)     status="idle" ;;
  SessionEnd)
    rm -f "/tmp/claude/$sid.status"
    exit 0
    ;;
  *)                status="$event" ;;
esac

# 写单个实例的状态文件
jq -n \
  --arg session    "$sid" \
  --arg status     "$status" \
  --arg event      "$event" \
  --arg cwd        "$cwd" \
  --arg tool       "$tool" \
  --arg transcript "$transcript" \
  --argjson ts     "$(date +%s)" \
  '{session:$session, status:$status, event:$event, cwd:$cwd, tool:$tool, transcript:$transcript, ts:$ts}' \
  > "/tmp/claude/$sid.status"

# 汇总所有活跃实例
jq -s '.' /tmp/claude/*.status 2>/dev/null > /tmp/claude/all.json || true
```

赋可执行权限：

```bash
chmod +x ~/.claude/bin/status-hook.sh
```

---

## 5. 配置：`~/.claude/settings.json`

把所有相关事件都指向同一个脚本：

```json
{
  "hooks": {
    "SessionStart":     [{"hooks":[{"type":"command","command":"~/.claude/bin/status-hook.sh"}]}],
    "UserPromptSubmit": [{"hooks":[{"type":"command","command":"~/.claude/bin/status-hook.sh"}]}],
    "PreToolUse":       [{"hooks":[{"type":"command","command":"~/.claude/bin/status-hook.sh"}]}],
    "PostToolUse":      [{"hooks":[{"type":"command","command":"~/.claude/bin/status-hook.sh"}]}],
    "Notification":     [{"hooks":[{"type":"command","command":"~/.claude/bin/status-hook.sh"}]}],
    "Stop":             [{"hooks":[{"type":"command","command":"~/.claude/bin/status-hook.sh"}]}],
    "SessionEnd":       [{"hooks":[{"type":"command","command":"~/.claude/bin/status-hook.sh"}]}]
  }
}
```

> 如果项目级别只想要这套行为，也可以放在项目目录的 `.claude/settings.json` 中。

---

## 6. 外部读取与监听

### 6.1 一次性查看

```bash
# 看有几个活跃实例
ls /tmp/claude/*.status

# 看每个实例当前状态
jq . /tmp/claude/all.json
```

输出示例：

```json
[
  {"session":"a1b2","status":"tool:Bash","cwd":"/home/me/proj-a","tool":"Bash","ts":1718450010},
  {"session":"c3d4","status":"waiting","cwd":"/home/me/proj-b","tool":"","ts":1718450005},
  {"session":"e5f6","status":"idle","cwd":"/home/me/proj-c","tool":"","ts":1718449998}
]
```

### 6.2 实时监听变化

```bash
# 装一次 inotify-tools
sudo apt install inotify-tools

# 实时打印状态变化
inotifywait -m -e close_write,delete /tmp/claude --format '%w%f %e' |
  while read -r path event; do
    [[ "$path" == *.status ]] || continue
    if [[ "$event" == "DELETE" ]]; then
      echo "[exit] $(basename "$path" .status)"
    else
      jq -c '{session, status, cwd, tool}' "$path"
    fi
  done
```

### 6.3 用作终端 statusline / 通知

举两个常见用法：

- **桌面通知**：在监听循环里看到 `status == "waiting"`，发 `notify-send "Claude 需要确认"`。
- **tmux 状态栏**：让 `status-right` 显示 `jq -r '.[].status' /tmp/claude/all.json` 的结果。

---

## 7. 把会话关联到终端 / 窗口

`session_id` 唯一但不直观。如果想知道"这个 Claude 是我哪个终端里启动的"，在 hook 脚本里额外记录环境变量即可：

```bash
# 在 status-hook.sh 里追加这些字段
tty=$(tty 2>/dev/null || echo "")
ppid="${PPID:-}"
tmux_pane="${TMUX_PANE:-}"
window="${WINDOWID:-}"
```

然后写进 JSON：

```bash
jq -n \
  --arg session "$sid" \
  --arg status  "$status" \
  ...
  --arg tty       "$tty" \
  --arg tmux_pane "$tmux_pane" \
  --arg window    "$window" \
  '{session:$session, status:$status, ..., tty:$tty, tmux_pane:$tmux_pane, window:$window}'
```

外部就能根据 `tmux_pane` 反查到具体面板。

---

## 8. 其他可选方案对比

| 方案 | 准确性 | 延迟 | 实现成本 | 备注 |
|---|---|---|---|---|
| **Hooks + 状态文件**（推荐） | ★★★★★ | 毫秒级 | 低 | 本文方案 |
| 解析 `~/.claude/projects/*/[sid].jsonl` | ★★★★ | 取决于 tail | 中 | 适合需要完整对话历史的场景 |
| `pgrep -f claude` | ★ | 秒级 | 极低 | 只能知道有没有进程在跑，区分不出状态 |
| 终端输出截屏 / OCR | ★★ | 高 | 高 | 不推荐 |

---

## 9. 常见坑

1. **stdin 只能读一次**：`cat` 之后存到变量里再用 `jq` 反复处理，别在两条 hook 命令里都 `jq` stdin。
2. **`Stop` 不代表 Claude 完全退出**，只代表"一回合"结束 —— 进入空闲，等待下一次 `UserPromptSubmit`。真正退出对应 `SessionEnd`。
3. **状态文件残留**：进程异常退出可能不触发 `SessionEnd`，文件会留下。可以在监控端用 `mtime` 判断"超过 N 小时无更新即视为僵尸"做清理。
4. **多个 settings.json 会合并**：用户级 `~/.claude/settings.json` 和项目级 `.claude/settings.json` 的 hooks 是**叠加**的，不会互相覆盖 —— 注意别让同一个脚本跑两次。
5. **权限提示也是 Notification**：`waiting` 状态既可能是等输入，也可能是等工具权限确认，必要时可读 `transcript_path` 末尾几行 JSON 进一步区分。

---

## 10. 一键安装脚本（可选）

```bash
#!/usr/bin/env bash
set -e
mkdir -p ~/.claude/bin /tmp/claude

# 1. 写状态采集脚本
cat > ~/.claude/bin/status-hook.sh <<'EOF'
# ... 第 4 节的脚本内容
EOF
chmod +x ~/.claude/bin/status-hook.sh

# 2. 合并 hooks 到 settings.json
python3 - <<'PY'
import json, os, pathlib
p = pathlib.Path.home() / ".claude/settings.json"
data = json.loads(p.read_text()) if p.exists() else {}
hooks = data.setdefault("hooks", {})
cmd = {"type": "command", "command": "~/.claude/bin/status-hook.sh"}
for ev in ["SessionStart","UserPromptSubmit","PreToolUse","PostToolUse",
           "Notification","Stop","SessionEnd"]:
    hooks.setdefault(ev, [{"hooks": []}])[0]["hooks"].append(cmd)
p.write_text(json.dumps(data, indent=2))
PY

echo "Done. 重启 Claude Code 后 /tmp/claude/ 即会出现状态文件。"
```

---

## 11. 小结

- **核心机制**：Claude Code 的生命周期 hooks + 通过 stdin 传入的 JSON 负载。
- **区分多实例**：以 `session_id` 为主键写状态文件，配合 `cwd` / `tmux_pane` 等辅助字段。
- **外部消费**：直接读 `/tmp/claude/*.status`，或 `inotifywait` 实时推送，对接通知、状态栏、监控面板都很简单。
