//go:build !windows && !darwin

package main

import _ "embed"

//go:embed static/logo.ico
var trayIcon []byte
