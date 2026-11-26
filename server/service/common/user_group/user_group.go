// Package user_group 用户组限制：按组控制可用节点/套餐、隧道数量上限、每月流量上限。
// 与套餐流量包(tunnel_quota)、节点流量独立叠加，连接须全部通过。
package user_group

import (
	"errors"
	"server/model"
	"server/repository"
	"server/repository/cache"
	"server/repository/query"
)

// Load 直查数据库返回用户所属组；管理员、无组或组不存在返回 nil(=不限制)。
// 用于创建隧道、节点列表、管理端等低频路径，保证即时生效。
func Load(db *query.Query, userCode string) *model.SystemUserGroup {
	user, err := db.SystemUser.Where(db.SystemUser.Code.Eq(userCode)).First()
	if err != nil || user == nil {
		return nil
	}
	if user.Admin == model.SYSTEM_IS_ADMIN || user.GroupCode == "" {
		return nil
	}
	group, err := db.SystemUserGroup.Where(db.SystemUserGroup.Code.Eq(user.GroupCode)).First()
	if err != nil || group == nil {
		return nil
	}
	return group
}

// AllowNode 节点白名单：组为 nil 或白名单为空表示全部允许。
func AllowNode(group *model.SystemUserGroup, nodeCode string) bool {
	if group == nil {
		return true
	}
	codes := group.GetNodeCodes()
	if len(codes) == 0 {
		return true
	}
	for _, code := range codes {
		if code == nodeCode {
			return true
		}
	}
	return false
}

// AllowConfig 套餐白名单：组为 nil 或白名单为空表示全部允许。
func AllowConfig(group *model.SystemUserGroup, configCode string) bool {
	if group == nil {
		return true
	}
	codes := group.GetConfigCodes()
	if len(codes) == 0 {
		return true
	}
	for _, code := range codes {
		if code == configCode {
			return true
		}
	}
	return false
}

// TunnelCount 统计用户全部 5 类隧道数量合计。
func TunnelCount(db *query.Query, userCode string) (total int64) {
	tunnel, _ := db.GostClientTunnel.Where(db.GostClientTunnel.UserCode.Eq(userCode)).Count()
	forward, _ := db.GostClientForward.Where(db.GostClientForward.UserCode.Eq(userCode)).Count()
	host, _ := db.GostClientHost.Where(db.GostClientHost.UserCode.Eq(userCode)).Count()
	proxy, _ := db.GostClientProxy.Where(db.GostClientProxy.UserCode.Eq(userCode)).Count()
	p2p, _ := db.GostClientP2P.Where(db.GostClientP2P.UserCode.Eq(userCode)).Count()
	return tunnel + forward + host + proxy + p2p
}

// MonthUsedBytes 用户近30天滚动已用流量。
// MONTH_DATEONLY_LIST 已含今日，与仪表盘/用户管理页口径一致，不再额外叠加今日。
func MonthUsedBytes(userCode string) int64 {
	obs := cache.GetUserObsDateRange(cache.MONTH_DATEONLY_LIST, userCode)
	return obs.InputBytes + obs.OutputBytes
}

// MonthExhausted 判断用户月流量配额是否用尽(limitGB 单位 GB，0=无限制)。
func MonthExhausted(userCode string, limitGB int) bool {
	if limitGB <= 0 {
		return false
	}
	return MonthUsedBytes(userCode) >= int64(limitGB)*1024*1024*1024
}

// VerifyCreate 创建隧道前校验用户组限制：节点/套餐白名单(自建节点豁免)+隧道数量+月流量。
func VerifyCreate(tx *query.Query, userCode, nodeCode, configCode string) error {
	group := Load(tx, userCode)
	if group == nil {
		return nil
	}
	// 用户自建节点豁免节点/套餐白名单，但仍计隧道数量与月流量。
	selfBound, _ := tx.GostNodeBind.Where(
		tx.GostNodeBind.NodeCode.Eq(nodeCode),
		tx.GostNodeBind.UserCode.Eq(userCode),
	).Count()
	if selfBound == 0 {
		if !AllowNode(group, nodeCode) {
			return errors.New("当前用户组不允许使用该节点")
		}
		if !AllowConfig(group, configCode) {
			return errors.New("当前用户组不允许使用该套餐")
		}
	}
	if group.TunnelLimit > 0 && TunnelCount(tx, userCode) >= int64(group.TunnelLimit) {
		return errors.New("隧道数量已达用户组上限")
	}
	if group.MonthTraffic > 0 && MonthExhausted(userCode, group.MonthTraffic) {
		return errors.New("本月流量已用完")
	}
	return nil
}

// LimitsCached 供 FRP 连接钩子使用，60s 缓存避免每连接查库；管理员/无组返回零值(=不限制)。
func LimitsCached(userCode string) cache.UserGroupLimit {
	if limit, found := cache.GetUserGroupLimit(userCode); found {
		return limit
	}
	db, _, _ := repository.Get("")
	var limit cache.UserGroupLimit
	if group := Load(db, userCode); group != nil {
		limit.MonthTraffic = group.MonthTraffic
		limit.TunnelLimit = group.TunnelLimit
	}
	cache.SetUserGroupLimit(userCode, limit)
	return limit
}
