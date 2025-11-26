-- iTerm2 激活并定位到指定 session（使用原生 AppleScript 选择）
-- 用法：osascript bring_iterm.scpt <sessionId>
--
-- 设计说明：
--   1. sessionId 可以是完整格式 "w0t0p0:UUID" 或纯 UUID
--   2. 遍历所有窗口/tab/session 查找匹配的 unique id
--   3. 找到后通过 "set index + select tab" 切换焦点（避免 AppleEvent 权限错误）
--   4. 不使用 URL scheme（iterm2://session?id= 只能激活应用，不能跳转 tab）
--
-- 权限说明：
--   首次运行会提示授予 "Terminal 控制 iTerm2" 权限
--   也可手动设置：系统设置 → 隐私与安全性 → 自动化

on run argv
	if (count of argv) < 1 then
		error "需要 1 个参数：sessionId"
	end if
	
	-- 提取 session ID（支持 "w0t0p0:UUID" 和纯 "UUID" 两种格式）
	set sessId to item 1 of argv
	set sessTail to sessId
	set oldDelims to AppleScript's text item delimiters
	try
		set AppleScript's text item delimiters to {":"}
		set sessTail to last text item of sessId -- 取冒号后的 UUID 部分
	end try
	set AppleScript's text item delimiters to oldDelims
	
	tell application "iTerm2"
		activate -- 先激活 iTerm2 应用
		
		-- 三层遍历：窗口 → tab → session
		repeat with w in windows
			repeat with t in tabs of w
				repeat with s in sessions of t
					try
						-- 获取当前 session 的 unique id（UUID 格式）
						set uid to unique id of s as string
						
						-- 匹配逻辑：UUID 完全匹配，或 sessId 以 UUID 结尾（兼容 w0t0p0:UUID 格式）
						if uid is equal to sessTail or sessId ends with uid then
							-- 关键：通过设置窗口索引和选择 tab 来切换焦点
							-- 为什么不用 "tell session to select"？
							--   - 会触发 AppleEvent -10000 权限错误
							--   - macOS 沙盒限制跨进程控制 session 级对象
							-- 为什么不用 URL scheme？
							--   - iterm2://session?id= 只能激活应用，不能跳转到具体 tab
							set index of w to 1 -- 将窗口移到最前
							select t -- 选择匹配的 tab
							return -- 找到后立即退出
						end if
					end try
				end repeat
			end repeat
		end repeat
	end tell
end run
