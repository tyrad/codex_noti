package go_codex_noti

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Event struct {
	Type                 string   `json:"type"`
	ThreadID             string   `json:"thread-id"`
	TurnID               string   `json:"turn-id"`
	Cwd                  string   `json:"cwd"`
	InputMessages        []string `json:"input-messages"`
	LastAssistantMessage string   `json:"last-assistant-message"`
}

func main() {
	jsonArg := flag.String("json", "", "JSON payload (if empty, read from stdin)")
	titleFlag := flag.String("title", "", "override notification title (optional)")
	flag.Parse()

	var r io.Reader
	switch {
	case *jsonArg != "":
		r = strings.NewReader(*jsonArg)
	default:
		// read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			exitErr("no input: provide --json or pipe JSON via stdin")
		}
		r = bufio.NewReader(os.Stdin)
	}

	var ev Event
	if err := json.NewDecoder(r).Decode(&ev); err != nil {
		exitErr("invalid JSON: %v", err)
	}

	// Compose notification fields
	title := *titleFlag
	if title == "" {
		if ev.Type != "" {
			title = ev.Type
		} else {
			title = "Agent Notification"
		}
	}
	subtitle := composeSubtitle(ev)
	body := composeBody(ev)

	if err := notifyMacOS(title, subtitle, body); err != nil {
		exitErr("notify failed: %v", err)
	}

	// 输出一点简短信息便于调用方确认
	fmt.Println("ok:", time.Now().Format(time.RFC3339))
}

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

func composeBody(ev Event) string {
	if ev.LastAssistantMessage != "" {
		return ev.LastAssistantMessage
	}
	if len(ev.InputMessages) > 0 {
		return strings.Join(ev.InputMessages, " ")
	}
	parts := []string{}
	if ev.Cwd != "" {
		parts = append(parts, "cwd: "+ev.Cwd)
	}
	if len(parts) == 0 {
		return "Task notification"
	}
	return strings.Join(parts, " • ")
}

func notifyMacOS(title, subtitle, body string) error {
	// AppleScript: display notification "body" with title "title" subtitle "subtitle"
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
	// 供 AppleScript 字符串使用的简单转义：先转反斜杠，再转双引号
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func exitErr(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", a...)
	os.Exit(1)
}
