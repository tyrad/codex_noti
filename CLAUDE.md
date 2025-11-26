# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**go-codex-noti** is a minimal macOS notification trigger for [OpenAI Codex](https://github.com/openai/codex)'s `notify` capability. It reads JSON payloads from Codex (task/conversation context) and triggers system notifications via AppleScript, with optional iTerm2 pane auto-focus on click.

## Core Architecture

### Single-File Design (main.go:1-229)
- **JSON Parsing**: Reads Codex event JSON from `os.Args[1]` or stdin (main.go:29-56)
- **Notification Strategy**: Terminal-notifier (preferred) with AppleScript fallback (main.go:108-137)
- **iTerm2 Integration**: Uses `iterm2://session?id=` URL scheme for pane activation (main.go:139-181)

### Event Structure
```go
type Event struct {
    Type                 string   `json:"type"`                    // e.g., "agent-turn-complete"
    ThreadID             string   `json:"thread-id"`               // Conversation UUID
    TurnID               string   `json:"turn-id"`                 // Turn number
    Cwd                  string   `json:"cwd"`                     // Working directory
    InputMessages        []string `json:"input-messages"`          // User questions
    LastAssistantMessage string   `json:"last-assistant-message"`  // AI reply
}
```

### Notification Flow
1. **Parse JSON** → validate Event struct
2. **Compose Content**:
   - Title: "AI处理完成通知" (hardcoded)
   - Subtitle: `turn {id} • thread {last-8-chars}` (main.go:78-92)
   - Body: User question + AI reply (main.go:94-105)
3. **Send via terminal-notifier** (if installed):
   - Adds `-execute` with iTerm URL if `ITERM_SESSION_ID` exists (main.go:139-181)
   - Falls back to `-activate com.googlecode.iterm2` if no script
4. **Fallback to AppleScript** if terminal-notifier fails/missing (main.go:192-204)

### iTerm2 Pane Activation Logic
- **Primary Method** (main.go:180): Calls AppleScript `bring_iterm.scpt` with session ID
  - Session ID sourced from `$ITERM_SESSION_ID` env var
  - Fallback: Query via AppleScript `tell application "iTerm2" to unique id of current session...` (main.go:156)
- **AppleScript Strategy** (scripts/bring_iterm.scpt:45-49):
  - Traverses all windows/tabs/sessions to find matching UUID
  - Uses `set index of theWindow to 1` + `select theTab` (bypasses AppleEvent permission issues)
  - **Why not URL scheme**: `iterm2://session?id=` only activates iTerm2, cannot jump to specific tab

## Common Commands

### Build & Install
```bash
# Standard build (respects local Go installation)
go build -o codex-noti main.go

# Automated install to ~/.codex/ (recommended)
./build.sh

# Manual install with GOROOT workaround
env -u GOROOT go build -o ~/bin/codex-noti .
```

### Testing Notifications
```bash
# Basic test
./codex-noti '{"type":"agent-turn-complete","thread-id":"b5f6c1c2-1111-2222-3333-444455556666","turn-id":"12345","cwd":"/tmp","input-messages":["Test question"],"last-assistant-message":"Test reply"}'

# Verbose mode (shows full JSON in notification body)
NOTIFY_VERBOSE=1 ./codex-noti '<JSON>'
```

### iTerm2 Debug Scripts
```bash
# Test iTerm URL schemes (tries id=/session=/sid= variants)
./debug.sh

# Test AppleScript-based focus switching (loops through windows/tabs/sessions)
./debug2.sh
```

### Codex Configuration
Add to `~/.codex/config.toml`:
```toml
notify = ["/Users/you/.codex/codex-noti"]
```

## Development Notes

### Verbose Mode Toggle
- **Environment Variable**: `NOTIFY_VERBOSE=1` (runtime)
- **Code Constant**: `debugVerbose = true` in main.go:16 (compile-time)
- Effect: Displays raw JSON in notification body + prints `ok: <timestamp>` to stdout

### Error Handling Philosophy
- **Graceful Degradation**: terminal-notifier errors → silent fallback to AppleScript (main.go:128-135)
- **No Partial Failures**: AppleScript errors → exit with `exitErr()` (main.go:200-202)

### String Escaping
- **AppleScript**: Escape `\` then `"` for `display notification` command (main.go:206-214)
- **Shell Execute**: Escape `\`, `"`, `` ` ``, `$` for terminal-notifier `-execute` arg (main.go:170-177)

### Session ID Resolution Priority
1. `$ITERM_SESSION_ID` env var (format: `w0t0p0:UUID`)
2. AppleScript query: `tell application "iTerm2" to unique id of current session...` (returns UUID only)
3. If both fail → no iTerm activation
4. **Critical**: Must run Codex in iTerm2, not other terminals (e.g., GoLand embedded terminal won't have `$ITERM_SESSION_ID`)

## Architecture Decisions

### Why Single main.go?
- Tool is **stateless** (one-shot notification)
- No complex dependencies (only stdlib)
- Deployment is simpler (single binary)

### Why terminal-notifier Over Pure AppleScript?
- **Clickable Actions**: AppleScript's `display notification` cannot execute commands on click
- **Better UX**: terminal-notifier supports `-execute` for iTerm pane jumping
- **Fallback Safety**: AppleScript ensures notifications work even without Homebrew

### Why AppleScript Over URL Scheme?
- **URL Scheme Limitation**: `iterm2://session?id={UUID}` only activates iTerm2, does not jump to specific tab/pane
- **AppleScript Precision**: Traverses window/tab hierarchy to find exact session, then focuses that tab
- **Permission Workaround**: Uses `set index` + `select` instead of `tell session to select` (avoids AppleEvent -10000 error)
- **Trade-off**: Requires iTerm2 Automation permission in System Settings (one-time prompt)

## Potential Issues

### terminal-notifier Not Found
- Symptom: Falls back to AppleScript (no pane auto-focus)
- Fix: `brew install terminal-notifier`

### ITERM_SESSION_ID Empty
- Symptom: Notification works but no pane jump
- Cause: Not running in iTerm2 or env var not propagated
- Fix: Ensure Codex runs in iTerm2 session (check `echo $ITERM_SESSION_ID`)

### AppleScript Permission Issues
- Symptom: Notification works but no tab jump, or `-10000` error in logs
- Fix: System Settings → Privacy & Security → Automation → Allow Terminal (or your shell's parent app) to control iTerm2
- Note: First-time use will auto-prompt for permission
