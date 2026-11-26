//go:build darwin

package main

import _ "embed"

// macOS 托盘使用 PNG（NSImage 对 ico 解码不可靠）
//
//go:embed static/logo.png
var trayIcon []byte
