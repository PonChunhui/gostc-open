// Package ban 封禁校验：为登录、JWT中间件、客户端注册、FRP连接钩子提供带缓存的封禁状态读取。
// 缓存60秒过期回源查库；管理端封禁/解封时立即写缓存，保证即时生效。
package ban

import (
	"server/model"
	"server/repository"
	"server/repository/cache"
)

// UserBanned 判断用户是否被封禁；未命中缓存时回源查库并回填。
func UserBanned(userCode string) bool {
	if userCode == "" {
		return false
	}
	if status, found := cache.GetUserBan(userCode); found {
		return status == model.STATUS_BANNED
	}
	db, _, _ := repository.Get("")
	status := model.STATUS_NORMAL
	user, _ := db.SystemUser.Where(db.SystemUser.Code.Eq(userCode)).First()
	if user != nil {
		status = user.Status
	}
	cache.SetUserBan(userCode, status)
	return status == model.STATUS_BANNED
}

// TunnelBanned 判断隧道(五类任一)是否被封禁；未命中缓存时按编号回源查库并回填。
func TunnelBanned(tunnelCode string) bool {
	if tunnelCode == "" {
		return false
	}
	if status, found := cache.GetTunnelBan(tunnelCode); found {
		return status == model.STATUS_BANNED
	}
	db, _, _ := repository.Get("")
	status := model.STATUS_NORMAL
	if tunnel, _ := db.GostClientTunnel.Where(db.GostClientTunnel.Code.Eq(tunnelCode)).First(); tunnel != nil {
		status = tunnel.Status
	} else if forward, _ := db.GostClientForward.Where(db.GostClientForward.Code.Eq(tunnelCode)).First(); forward != nil {
		status = forward.Status
	} else if host, _ := db.GostClientHost.Where(db.GostClientHost.Code.Eq(tunnelCode)).First(); host != nil {
		status = host.Status
	} else if proxy, _ := db.GostClientProxy.Where(db.GostClientProxy.Code.Eq(tunnelCode)).First(); proxy != nil {
		status = proxy.Status
	} else if p2p, _ := db.GostClientP2P.Where(db.GostClientP2P.Code.Eq(tunnelCode)).First(); p2p != nil {
		status = p2p.Status
	}
	cache.SetTunnelBan(tunnelCode, status)
	return status == model.STATUS_BANNED
}