// Package entryfollow 故障转移「入口跟随」E2：同端口硬约束 + 各类隧道公网入口计算。
// 公网端口必须保持不变；备节点无法提供同端口时拒绝转移。
package entryfollow

import (
	"strconv"

	"server/model"
	cache2 "server/repository/cache"
	"server/repository/query"
)

// CanServeSamePort 校验备节点能否提供与主节点相同的公网端口。
func CanServeSamePort(db *query.Query, primary, backup *model.GostNode, tunnelType int, tunnelCode string) bool {
	if primary == nil || backup == nil {
		return false
	}
	switch tunnelType {
	case model.TUNNEL_HEALTH_TYPE_HOST:
		// 网关端口/协议由 UrlTpl 决定，要求一致以保证访问端口不变。
		return backup.UrlTpl == primary.UrlTpl
	case model.TUNNEL_HEALTH_TYPE_FORWARD:
		f, err := db.GostClientForward.Where(db.GostClientForward.Code.Eq(tunnelCode)).First()
		if err != nil || f == nil {
			return false
		}
		return portFree(db, backup, f.Port, tunnelCode)
	case model.TUNNEL_HEALTH_TYPE_PROXY:
		p, err := db.GostClientProxy.Where(db.GostClientProxy.Code.Eq(tunnelCode)).First()
		if err != nil || p == nil {
			return false
		}
		return portFree(db, backup, p.Port, tunnelCode)
	case model.TUNNEL_HEALTH_TYPE_TUNNEL, model.TUNNEL_HEALTH_TYPE_P2P:
		_, pp := primary.GetOriginAddress()
		_, bp := backup.GetOriginAddress()
		return pp == bp
	}
	return false
}

// portFree 端口在备节点端口池内且未被其它 forward/proxy 占用。
func portFree(db *query.Query, backup *model.GostNode, port, selfCode string) bool {
	if port == "" {
		return false
	}
	if !backup.GetPorts(nil)[port] {
		return false
	}
	if n, _ := db.GostClientForward.Where(
		db.GostClientForward.NodeCode.Eq(backup.Code),
		db.GostClientForward.Port.Eq(port),
		db.GostClientForward.Enable.Eq(1),
		db.GostClientForward.Code.Neq(selfCode),
	).Count(); n > 0 {
		return false
	}
	if n, _ := db.GostClientProxy.Where(
		db.GostClientProxy.NodeCode.Eq(backup.Code),
		db.GostClientProxy.Port.Eq(port),
		db.GostClientProxy.Enable.Eq(1),
		db.GostClientProxy.Code.Neq(selfCode),
	).Count(); n > 0 {
		return false
	}
	return true
}

// Entry 计算隧道在指定节点上的公网入口字符串。
func Entry(db *query.Query, node *model.GostNode, tunnelType int, tunnelCode string) string {
	if node == nil {
		return ""
	}
	switch tunnelType {
	case model.TUNNEL_HEALTH_TYPE_HOST:
		h, err := db.GostClientHost.Where(db.GostClientHost.Code.Eq(tunnelCode)).First()
		if err != nil || h == nil {
			return ""
		}
		return node.GetDomainFull(h.DomainPrefix, h.GetCustomDomain(), cache2.GetNodeCustomDomain(node.Code))
	case model.TUNNEL_HEALTH_TYPE_FORWARD:
		f, err := db.GostClientForward.Where(db.GostClientForward.Code.Eq(tunnelCode)).First()
		if err != nil || f == nil {
			return ""
		}
		host, _ := node.GetOriginAddress()
		return host + ":" + f.Port
	case model.TUNNEL_HEALTH_TYPE_PROXY:
		p, err := db.GostClientProxy.Where(db.GostClientProxy.Code.Eq(tunnelCode)).First()
		if err != nil || p == nil {
			return ""
		}
		host, _ := node.GetOriginAddress()
		return host + ":" + p.Port
	case model.TUNNEL_HEALTH_TYPE_TUNNEL, model.TUNNEL_HEALTH_TYPE_P2P:
		host, port := node.GetOriginAddress()
		return host + ":" + strconv.Itoa(port)
	}
	return ""
}
