# go-codex-noti

macOS notification trigger for [OpenAI Codex](https://github.com/openai/codex) with iTerm2 integration.

**Purpose**: Send macOS notifications when Codex tasks complete. Only supports macOS. When running in iTerm2, clicking the notification jumps back to the original iTerm tab (may not work if iTerm2 is minimized).

## Install

```bash
./build.sh
```

Installs to `~/.codex/codex-noti` and copies iTerm2 script to `~/.codex/scripts/`.

## Configure Codex

Add to `~/.codex/config.toml`:

```toml
notify = ["/Users/you/.codex/codex-noti"]
```

## Features

- ✅ Sends macOS notifications when Codex completes tasks
- ✅ Click notification → jumps to iTerm2 pane (requires `terminal-notifier`)
- ✅ Fallback to AppleScript if terminal-notifier unavailable
- ✅ Verbose mode: `NOTIFY_VERBOSE=1` shows full JSON payload

## Requirements

- macOS
- Go 1.23+
- Optional: `brew install terminal-notifier` (for iTerm2 pane activation)
- Must run Codex in iTerm2 (not other terminals)

## Manual Test

```bash
./codex-noti '{"type":"agent-turn-complete","thread-id":"test-123","turn-id":"1","cwd":"/tmp","input-messages":["Test"],"last-assistant-message":"Done"}'
```

## How It Works

1. Codex calls `codex-noti` with JSON event data
2. Parses user question + AI reply from JSON
3. Sends notification via terminal-notifier (with iTerm URL) or AppleScript
4. On click: AppleScript finds matching iTerm session by UUID and activates it
