package bootstrap

import (
	"server/global"
	paycommon "server/service/common/pay"
)

// InitPay 注册支付渠道（沙箱始终可用，真实渠道按配置启用）。
func InitPay() {
	paycommon.Init(global.Config.Pay)
	global.Logger.Info("init pay providers finish")
}
