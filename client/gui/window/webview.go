package window

import (
	"net"

	"gostc-sub/gui/global"
)

// mainURL 返回内嵌管理界面地址（本地 Web 服务）。
func mainURL() string {
	_, port, err := net.SplitHostPort(global.Config.Address)
	if err != nil || port == "" {
		port = "18080"
	}
	return "http://127.0.0.1:" + port
}
