package window

import (
	"gostc-sub/gui/global"
	"gostc-sub/webui/backend/bootstrap"
)

// applyAddress 保存监听地址并重启本地 Web 服务，返回异常消息（成功返回"无异常消息"）。
// 供 Windows 的 gioui 设置窗口与 macOS 的原生设置对话框共用。
func applyAddress(address string) string {
	global.Config.Address = address
	_ = global.SaveConfig(global.BasePath + "/config.yaml")
	if bootstrap.IsRunningServer() {
		bootstrap.StopServer()
	}
	var errMsg string
	if err := bootstrap.StartServer(global.Config.Address); err != nil {
		errMsg = err.Error()
	} else {
		errMsg = "无异常消息"
	}
	RefreshMain()
	return errMsg
}
