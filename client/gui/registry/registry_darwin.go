//go:build darwin

package registry

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// macOS 通过用户级 LaunchAgent 实现开机（登录）自启，
// 与 Windows 版写 HKLM Run 注册表项对应。
const label = "com.gostc.gui"

func plistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", label+".plist"), nil
}

func Registered() bool {
	p, err := plistPath()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return false
	}
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	return strings.Contains(string(data), exe)
}

func Register() error {
	p, err := plistPath()
	if err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	var escaped bytes.Buffer
	if err := xml.EscapeText(&escaped, []byte(exe)); err != nil {
		return err
	}
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>
`, label, escaped.String())
	return os.WriteFile(p, []byte(plist), 0644)
}

func UnRegister() {
	p, err := plistPath()
	if err != nil {
		return
	}
	_ = os.Remove(p)
}
