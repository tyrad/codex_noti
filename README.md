# agent-notify

一个极简的 macOS 通知触发器，服务于 [OpenAI Codex](https://github.com/openai/codex) 的 `notify` 配置能力（参见 [`docs/config.md#notify`](https://github.com/openai/codex/blob/main/docs/config.md#notify)）。它从 Codex 提供的 JSON 载荷读取一条任务/对话上下文消息，并通过 AppleScript 触发系统通知。

## 功能
- 从位置参数或标准输入读取通知事件（Codex notify 的 JSON）。
- 事件包含类型、线程/轮次 ID、工作目录、用户输入与助手回复等信息。
- 自动组装通知标题、副标题与正文，并调用 macOS 通知。

## 运行
```bash
# 构建（示例路径可换成你习惯的）
env -u GOROOT go build -o ~/bin/codex-noti .

# 运行：接受 Codex 的 JSON 作为第一个位置参数
~/bin/codex-noti '{
  "type": "agent-turn-complete",
  "thread-id": "b5f6c1c2-1111-2222-3333-444455556666",
  "turn-id": "12345",
  "cwd": "/Users/alice/projects/example",
  "input-messages": ["Rename foo to bar and update the callsites."],
  "last-assistant-message": "Rename complete and verified cargo build succeeds."
}'
```

> 提示：如果本机环境设置了指向旧版本的 `GOROOT`，构建时使用 `env -u GOROOT` 可避免版本不匹配报错。

## Codex 配置示例
在 `~/.codex/config.toml` 中配置：
```toml
notify = ["/Users/you/bin/codex-noti"]
```
建议使用绝对路径，避免 Codex 的运行环境找不到可执行文件。

## 提示开关（输出完整入参）
设置环境变量 `NOTIFY_VERBOSE=1`，通知正文会直接显示 Codex 传入的完整 JSON，并打印 `ok:` 确认行（便于调试）：
```bash
NOTIFY_VERBOSE=1 ~/bin/codex-noti '<JSON>'
```

也可以在 `main.go` 内将 `debugVerbose` 设为 `true`，无需环境变量即可始终显示完整入参并输出确认行。默认关闭时，不会打印 `ok:` 行，输出更干净。

## iTerm2 点击通知自动返回 Pane（可选）

本功能需安装 `terminal-notifier`，并提供一个 AppleScript 来激活 iTerm2 并跳转到指定的 window/tab/session。默认会尝试使用“可执行文件同目录的 scripts/bring_iterm.scpt”；使用 `build.sh` 安装时会将脚本复制到 `~/.codex/scripts/bring_iterm.scpt`。

### 准备 AppleScript（示例）

已附带示例脚本：`scripts/bring_iterm.scpt`（可复制到 `~/scripts/bring_iterm.scpt` 或直接使用绝对路径）。
```applescript
-- 用法：osascript bring_iterm.scpt <windowId> <tabId> [<sessionId>]
on run argv
    if (count of argv) < 2 then
        error "需要至少 2 个参数：windowId tabId"
    end if

    set winId to item 1 of argv
    set tabId to item 2 of argv
    set sessId to ""
    if (count of argv) ≥ 3 then
        set sessId to item 3 of argv
    end if

    tell application "iTerm2"
        activate
        repeat with w in windows
            if (id of w as string) is equal to winId then
                tell w
                    repeat with t in tabs
                        if (id of t as string) is equal to tabId then
                            set current tab to t
                            tell t
                                if sessId is not "" then
                                    repeat with s in sessions
                                        try
                                            if (id of s as string) is equal to sessId then
                                                select s
                                                exit repeat
                                            end if
                                        end try
                                    end repeat
                                end if
                            end tell
                            exit repeat
                        end if
                    end repeat
                end tell
                exit repeat
            end if
        end repeat
    end tell
end run
```

### 运行前设置环境变量
确保当前 shell 是在 iTerm2 打开的，这样才有 `ITERM_*` 变量。如果你将二进制安装到 `~/.codex/codex-noti`，并保留默认的 `~/.codex/scripts/bring_iterm.scpt`，可省略此变量；否则可以显式指定：
```bash
export ITERM_ACTIVATE_SCRIPT="$HOME/scripts/bring_iterm.scpt"
```

### 工作原理
- 程序会检测到存在 `terminal-notifier` 和 `ITERM_ACTIVATE_SCRIPT`，且环境中有 `ITERM_WINDOW_ID`/`ITERM_TAB_ID`（可选 `ITERM_SESSION_ID`）。满足时发送通知附带 `-execute "osascript <script> <win> <tab> <sess>"`。
- 用户点击通知后，AppleScript 会激活 iTerm2 并切到对应 window/tab/session。
- 若条件不满足或执行失败，自动回退到 AppleScript 原生通知（仍会弹出通知，但不跳回 Pane）。
