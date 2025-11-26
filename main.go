package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// 将 debugVerbose 设为 true 可强制输出完整入参（无需环境变量）
const debugVerbose = false

// Event 对应传入的通知 JSON 载荷
type Event struct {
	Type                 string   `json:"type"`
	ThreadID             string   `json:"thread-id"`
	TurnID               string   `json:"turn-id"`
	Cwd                  string   `json:"cwd"`
	InputMessages        []string `json:"input-messages"`
	LastAssistantMessage string   `json:"last-assistant-message"`
}

// 入口：接受 Codex 传入的 JSON（位置参数或 stdin），触发通知
func main() {
	var r io.Reader
	switch {
	case len(os.Args) >= 2:
		// Codex notify: JSON 作为第一个位置参数传入
		r = strings.NewReader(os.Args[1])
	default:
		// 回退：允许从 stdin 读取 JSON
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			exitErr("no input: pass Codex JSON as first argument")
		}
		r = bufio.NewReader(os.Stdin)
	}

	// 读入原始 JSON，既用于解析也可在“提示开关”开启时直接展示
	raw, err := io.ReadAll(r)
	if err != nil {
		exitErr("read input: %v", err)
	}
	if len(raw) == 0 {
		exitErr("no input: empty payload")
	}

	var ev Event
	if err := json.Unmarshal(raw, &ev); err != nil {
		exitErr("invalid JSON: %v", err)
	}

	// 组装通知文案
	title := "Codex Task Completed"
	subtitle := composeSubtitle(ev)
	body := composeBody(ev)

	// 开关开启时，正文直接展示原始入参 JSON
	if verboseEnabled() {
		body = strings.TrimSpace(string(raw))
	}

	if err := sendNotification(title, subtitle, body); err != nil {
		exitErr("notify failed: %v", err)
	}

	// 调试时输出确认信息
	if verboseEnabled() {
		fmt.Println("ok:", time.Now().Format(time.RFC3339))
	}
}

// composeSubtitle 拼出 turn/thread 信息
func composeSubtitle(ev Event) string {
	tid := ev.TurnID
	if tid == "" {
		tid = "-"
	}
	th := ev.ThreadID
	if len(th) > 8 {
		th = th[len(th)-8:]
	}
	if th == "" {
		th = "-"
	}
	return fmt.Sprintf("turn %s • thread %s", tid, th)
}

// composeBody 优先展示"问题 + AI 回复"，缺失时回退到工作目录提示
func composeBody(ev Event) string {
	question := strings.TrimSpace(strings.Join(ev.InputMessages, " "))
	if question == "" {
		question = "(none)"
	}
	answer := strings.TrimSpace(ev.LastAssistantMessage)
	if answer == "" {
		answer = "(no reply)"
	}
	return fmt.Sprintf("Question: %s\nReply: %s", question, answer)
}

// sendNotification 优先使用 terminal-notifier，缺失或失败时回退 AppleScript
func sendNotification(title, subtitle, body string) error {
	if path, err := exec.LookPath("terminal-notifier"); err == nil {
		execArg := buildExecuteArg()
		activate := detectActivateApp()

		args := []string{"-title", title, "-message", body}
		if subtitle != "" {
			args = append(args, "-subtitle", subtitle)
		}
		if execArg != "" {
			args = append(args, "-execute", execArg)
		} else if activate != "" {
			// 没有脚本时至少激活 iTerm2
			args = append(args, "-activate", activate)
		}

		if verboseEnabled() {
			fmt.Fprintf(os.Stderr, "[notify] terminal-notifier %v\n", args)
		}
		cmd := exec.Command(path, args...)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			if verboseEnabled() {
				fmt.Fprintf(os.Stderr, "[notify] terminal-notifier error: %v, out: %s\n", err, strings.TrimSpace(string(out)))
			}
		}
	}
	return notifyMacOS(title, subtitle, body)
}

// buildExecuteArg 如果设置了 ITERM_ACTIVATE_SCRIPT 且存在 ITERM_* 环境变量，则构造点击后激活 iTerm 的命令
func buildExecuteArg() string {
	// 必须在 iTerm2 环境中运行才启用点击跳转
	sess := os.Getenv("ITERM_SESSION_ID")
	if sess == "" {
		return ""
	}

	script := os.Getenv("ITERM_ACTIVATE_SCRIPT")
	// 默认查找可执行同目录下的 scripts/bring_iterm.scpt
	if script == "" {
		if self, err := os.Executable(); err == nil {
			if base := filepath.Dir(self); base != "" {
				defaultScript := filepath.Join(base, "scripts", "bring_iterm.scpt")
				if _, err := os.Stat(defaultScript); err == nil {
					script = defaultScript
				}
			}
		}
	}
	if script == "" {
		return ""
	}
	if _, err := os.Stat(script); err != nil {
		return ""
	}

	q := func(s string) string {
		// 为双引号包裹的 shell 参数做转义，适配 -execute "osascript ..."
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		s = strings.ReplaceAll(s, "`", "\\`")
		s = strings.ReplaceAll(s, "$", "\\$")
		return s
	}

	// 调用 AppleScript 精确跳转到 session（比 URL scheme 可靠）
	return fmt.Sprintf(`osascript "%s" "%s"`, q(script), q(sess))
}

// detectActivateApp 返回 -activate 的 bundle id（仅在 iTerm 环境尝试）
func detectActivateApp() string {
	if os.Getenv("ITERM_SESSION_ID") != "" {
		return "com.googlecode.iterm2"
	}
	return ""
}

// notifyMacOS 利用 AppleScript 调用系统通知
func notifyMacOS(title, subtitle, body string) error {
	// AppleScript：display notification "body" with title "title" subtitle "subtitle"
	script := fmt.Sprintf(
		`display notification "%s" with title "%s" subtitle "%s"`,
		(escapeAS(body)), escapeAS(title), escapeAS(subtitle))

	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func escapeAS(s string) string {
	// 保留换行，便于多行展示，同时规整回车符
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	// 供 AppleScript 字符串使用的简单转义：先转反斜杠，再转双引号
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// verboseEnabled 判断是否开启“提示开关”，将入参全文输出到通知正文
func verboseEnabled() bool {
	if debugVerbose {
		return true
	}
	v := strings.ToLower(os.Getenv("NOTIFY_VERBOSE"))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func exitErr(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "error: "+format+"\n", a...)
	os.Exit(1)
}
