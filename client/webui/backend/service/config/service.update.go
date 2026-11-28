package service

import (
	"gostc-sub/internal/common"
	"gostc-sub/internal/updater"
	"gostc-sub/webui/backend/global"
)

// UpdateSettingsReq 保存自更新设置。
type UpdateSettingsReq struct {
	Mode          string `json:"mode"`
	Channel       string `json:"channel"`
	IntervalHours int    `json:"intervalHours"`
	PanelURL      string `json:"panelUrl"`
	AutoRestart   bool   `json:"autoRestart"`
	PublicKey     string `json:"publicKey"`
}

// UpdateCheckReq 检查/升级时可覆盖面板地址与灰度标识。
type UpdateCheckReq struct {
	PanelURL string `json:"panelUrl"`
	UID      string `json:"uid"`
}

// UpdateQueryResp 设置 + 当前版本 + 待升级信息。
type UpdateQueryResp struct {
	Settings updater.Settings `json:"settings"`
	Version  string           `json:"version"`
	Pending  *updater.Info    `json:"pending"`
}

func ensureInit(panelURL, uid string) {
	updater.Init(updater.Options{
		DataDir:  global.DataDir,
		PanelURL: panelURL,
		UID:      uid,
		Current:  common.VERSION,
		Log:      func(msg string) { common.Logger.AddLog("update", msg) },
	})
}

func (*service) UpdateQuery() UpdateQueryResp {
	ensureInit("", "")
	return UpdateQueryResp{
		Settings: updater.GetSettings(),
		Version:  common.VERSION,
		Pending:  updater.Pending(),
	}
}

func (*service) UpdateSave(req UpdateSettingsReq) error {
	ensureInit(req.PanelURL, "")
	return updater.SaveSettings(updater.Settings{
		Mode:          req.Mode,
		Channel:       req.Channel,
		IntervalHours: req.IntervalHours,
		PanelURL:      req.PanelURL,
		AutoRestart:   req.AutoRestart,
		PublicKey:     req.PublicKey,
	})
}

func (*service) UpdateCheck(req UpdateCheckReq) (*updater.Info, error) {
	ensureInit(req.PanelURL, req.UID)
	return updater.CheckNow()
}

func (*service) UpdateApply(req UpdateCheckReq) error {
	ensureInit(req.PanelURL, req.UID)
	return updater.ApplyNow()
}
