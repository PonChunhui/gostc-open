package cache

import (
	"fmt"
	"server/global"
	"time"
)

const (
	user_ban_key   = "user:ban:%s"
	tunnel_ban_key = "tunnel:ban:%s"
)

// SetUserBan 缓存用户封禁状态(1正常 2封禁)，封禁/解封时立即写入，60秒后过期回源。
func SetUserBan(userCode string, status int) {
	global.Cache.SetStruct(fmt.Sprintf(user_ban_key, userCode), status, time.Second*60)
}

// GetUserBan 返回缓存的用户封禁状态；found=false 表示未命中(需回源加载)。
func GetUserBan(userCode string) (status int, found bool) {
	err := global.Cache.GetStruct(fmt.Sprintf(user_ban_key, userCode), &status)
	return status, err == nil
}

func DelUserBan(userCode string) {
	global.Cache.Del(fmt.Sprintf(user_ban_key, userCode))
}

// SetTunnelBan 缓存隧道封禁状态(1正常 2封禁)，封禁/解封时立即写入，60秒后过期回源。
func SetTunnelBan(tunnelCode string, status int) {
	global.Cache.SetStruct(fmt.Sprintf(tunnel_ban_key, tunnelCode), status, time.Second*60)
}

// GetTunnelBan 返回缓存的隧道封禁状态；found=false 表示未命中(需回源加载)。
func GetTunnelBan(tunnelCode string) (status int, found bool) {
	err := global.Cache.GetStruct(fmt.Sprintf(tunnel_ban_key, tunnelCode), &status)
	return status, err == nil
}

func DelTunnelBan(tunnelCode string) {
	global.Cache.Del(fmt.Sprintf(tunnel_ban_key, tunnelCode))
}