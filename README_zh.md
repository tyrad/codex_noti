# go-codex-noti

为 [OpenAI Codex](https://github.com/openai/codex) 提供 macOS 通知触发器,支持 iTerm2 集成。

**用途**: 在 Codex 任务完成后向 macOS 电脑发送通知。仅支持 macOS 环境。如果在 iTerm2 中运行,点击通知可以返回之前的 iTerm tab(iTerm2 最小化时可能不支持)。

## 安装

```bash
./build.sh
```

会安装到 `~/.codex/codex-noti` 并复制 iTerm2 脚本到 `~/.codex/scripts/`。

## 配置 Codex

在 `~/.codex/config.toml` 中添加:

```toml
notify = ["/Users/you/.codex/codex-noti"]
```

## 功能特性

- ✅ Codex 完成任务时发送 macOS 通知
- ✅ 点击通知 → 自动跳转到 iTerm2 对应窗格(需安装 `terminal-notifier`)
- ✅ terminal-notifier 不可用时自动降级到 AppleScript
- ✅ 详细模式: `NOTIFY_VERBOSE=1` 显示完整 JSON 载荷

## 运行要求

- macOS
- Go 1.23+
- 可选: `brew install terminal-notifier` (用于 iTerm2 窗格激活)
- 必须在 iTerm2 中运行 Codex(其他终端无效)

## 手动测试

```bash
./codex-noti '{"type":"agent-turn-complete","thread-id":"test-123","turn-id":"1","cwd":"/tmp","input-messages":["测试问题"],"last-assistant-message":"已完成"}'
```

## 工作原理

1. Codex 调用 `codex-noti` 并传入 JSON 事件数据
2. 解析 JSON 中的用户问题 + AI 回复
3. 通过 terminal-notifier(附带 iTerm URL) 或 AppleScript 发送通知
4. 点击通知: AppleScript 根据 UUID 查找匹配的 iTerm 会话并激活
