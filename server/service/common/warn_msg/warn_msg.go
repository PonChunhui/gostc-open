package warn_msg

import (
	"server/model"
	"server/service/common/tunnel_quota"
	"time"
)

func banMsg(status int, reason string) string {
	if status != model.STATUS_BANNED {
		return ""
	}
	if reason != "" {
		return "已被管理员封禁：" + reason
	}
	return "已被管理员封禁"
}

func GetHostWarnMsg(host model.GostClientHost) string {
	if msg := banMsg(host.Status, host.BanReason); msg != "" {
		return msg
	}
	if host.Enable != 1 {
		return "已停用"
	}
	if host.ChargingType == model.GOST_CONFIG_CHARGING_CUCLE_DAY && host.ExpAt < time.Now().Unix() {
		return "已到期"
	}
	if tunnel_quota.Exhausted(host.Code, host.LimitTotal) {
		return "流量包已用完"
	}
	return ""
}
func GetForwardWarnMsg(forward model.GostClientForward) string {
	if msg := banMsg(forward.Status, forward.BanReason); msg != "" {
		return msg
	}
	if forward.Enable != 1 {
		return "已停用"
	}
	if forward.ChargingType == model.GOST_CONFIG_CHARGING_CUCLE_DAY && forward.ExpAt < time.Now().Unix() {
		return "已到期"
	}
	if tunnel_quota.Exhausted(forward.Code, forward.LimitTotal) {
		return "流量包已用完"
	}
	return ""
}

func GetTunnelWarnMsg(tunnel model.GostClientTunnel) string {
	if msg := banMsg(tunnel.Status, tunnel.BanReason); msg != "" {
		return msg
	}
	if tunnel.Enable != 1 {
		return "已停用"
	}
	if tunnel.ChargingType == model.GOST_CONFIG_CHARGING_CUCLE_DAY && tunnel.ExpAt < time.Now().Unix() {
		return "已到期"
	}
	if tunnel_quota.Exhausted(tunnel.Code, tunnel.LimitTotal) {
		return "流量包已用完"
	}
	return ""
}

func GetProxyWarnMsg(proxy model.GostClientProxy) string {
	if msg := banMsg(proxy.Status, proxy.BanReason); msg != "" {
		return msg
	}
	if proxy.Enable != 1 {
		return "已停用"
	}
	if proxy.ChargingType == model.GOST_CONFIG_CHARGING_CUCLE_DAY && proxy.ExpAt < time.Now().Unix() {
		return "已到期"
	}
	if tunnel_quota.Exhausted(proxy.Code, proxy.LimitTotal) {
		return "流量包已用完"
	}
	return ""
}

func GetP2PWarnMsg(p2p model.GostClientP2P) string {
	if msg := banMsg(p2p.Status, p2p.BanReason); msg != "" {
		return msg
	}
	if p2p.Enable != 1 {
		return "已停用"
	}
	if p2p.ChargingType == model.GOST_CONFIG_CHARGING_CUCLE_DAY && p2p.ExpAt < time.Now().Unix() {
		return "已到期"
	}
	if tunnel_quota.Exhausted(p2p.Code, p2p.LimitTotal) {
		return "流量包已用完"
	}
	return ""
}

func GetCfgWarnMsg(p2p model.FrpClientCfg) string {
	if p2p.Enable != 1 {
		return "已停用"
	}
	return ""
}