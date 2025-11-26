#!/bin/bash
set -e

# 编译 Go 二进制文件
go build -o codex-noti main.go

# 确保 ~/.codex 目录存在
mkdir -p ~/.codex/codex-noti/scripts

# 复制到目标路径
cp codex-noti ~/.codex/codex-noti/codex-noti
cp scripts/bring_iterm.scpt ~/.codex/codex-noti/scripts/bring_iterm.scpt

# 添加执行权限
chmod +x ~/.codex/codex-noti/codex-noti
chmod +x ~/.codex/codex-noti/scripts/bring_iterm.scpt

echo "✓ Built and installed to ~/.codex/codex-noti/codex-noti"
echo "✓ iTerm2 AppleScript installed to ~/.codex/codex-noti/scripts/bring_iterm.scpt"
