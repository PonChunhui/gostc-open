//go:build windows

package main

import _ "embed"

//go:embed static/logo.ico
var trayIcon []byte
