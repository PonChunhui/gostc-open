// Package updater 客户端自更新：检查面板新版本，按模式（关闭/提示/自动）下载校验并原子替换重启。
package updater

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// 更新模式。
const (
	ModeOff    = "off"    // 不检查
	ModeNotify = "notify" // 发现新版本提示，用户确认后升级
	ModeAuto   = "auto"   // 发现新版本静默自动升级
)

// 更新通道。
const (
	ChannelStable = "stable"
	ChannelBeta   = "beta"
)

// Settings 自更新设置，三端（GUI/WebUI/CLI）共享，持久化于 data/update.json。
type Settings struct {
	Mode          string `json:"mode"`          // off|notify|auto
	Channel       string `json:"channel"`       // stable|beta
	IntervalHours int    `json:"intervalHours"` // 检查间隔（小时）
	PanelURL      string `json:"panelUrl"`      // 面板基址覆盖（空则用连接地址）
	AutoRestart   bool   `json:"autoRestart"`   // 升级后自动重启
	PublicKey     string `json:"publicKey"`     // ed25519 公钥(hex)，配置后强制校验签名
}

// DefaultSettings 返回默认设置（默认提示模式，安全）。
func DefaultSettings() Settings {
	return Settings{
		Mode:          ModeNotify,
		Channel:       ChannelStable,
		IntervalHours: 24,
		AutoRestart:   true,
	}
}

// Normalize 补齐空字段为默认值。
func (s Settings) Normalize() Settings {
	d := DefaultSettings()
	if s.Mode == "" {
		s.Mode = d.Mode
	}
	if s.Mode != ModeOff && s.Mode != ModeNotify && s.Mode != ModeAuto {
		s.Mode = d.Mode
	}
	if s.Channel == "" {
		s.Channel = d.Channel
	}
	if s.IntervalHours <= 0 {
		s.IntervalHours = d.IntervalHours
	}
	return s
}

// SettingsPath 返回设置文件路径。
func SettingsPath(dataDir string) string {
	return filepath.Join(dataDir, "update.json")
}

// LoadSettings 读取设置；不存在或损坏时返回默认值（不报错，便于首次运行）。
func LoadSettings(dataDir string) Settings {
	data, err := os.ReadFile(SettingsPath(dataDir))
	if err != nil {
		return DefaultSettings()
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings()
	}
	return s.Normalize()
}

// SaveSettingsFile 原子写入设置（temp+rename）。
func SaveSettingsFile(dataDir string, s Settings) error {
	s = s.Normalize()
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	path := SettingsPath(dataDir)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
