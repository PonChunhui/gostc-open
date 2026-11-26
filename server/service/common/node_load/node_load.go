package node_load

import (
	"fmt"
	"server/model"
	"server/repository/query"
)

// CountTunnels 统计节点上已分配的隧道总数（私有隧道/端口转发/域名解析/代理隧道/P2P）
func CountTunnels(tx *query.Query, nodeCode string) (total int64) {
	tunnelCount, _ := tx.GostClientTunnel.Where(tx.GostClientTunnel.NodeCode.Eq(nodeCode)).Count()
	forwardCount, _ := tx.GostClientForward.Where(tx.GostClientForward.NodeCode.Eq(nodeCode)).Count()
	hostCount, _ := tx.GostClientHost.Where(tx.GostClientHost.NodeCode.Eq(nodeCode)).Count()
	proxyCount, _ := tx.GostClientProxy.Where(tx.GostClientProxy.NodeCode.Eq(nodeCode)).Count()
	p2pCount, _ := tx.GostClientP2P.Where(tx.GostClientP2P.NodeCode.Eq(nodeCode)).Count()
	return tunnelCount + forwardCount + hostCount + proxyCount + p2pCount
}

type nodeTunnelCount struct {
	NodeCode string `gorm:"column:node_code"`
	Total    int64  `gorm:"column:total"`
}

// CountTunnelsByNode 按节点分组统计已分配的隧道总数，用于列表展示
func CountTunnelsByNode(tx *query.Query) map[string]int64 {
	var counts = make(map[string]int64)
	var collect = func(rows []nodeTunnelCount) {
		for _, row := range rows {
			counts[row.NodeCode] += row.Total
		}
	}

	var tunnelRows []nodeTunnelCount
	_ = tx.GostClientTunnel.
		Select(tx.GostClientTunnel.NodeCode, tx.GostClientTunnel.Id.Count().As("total")).
		Group(tx.GostClientTunnel.NodeCode).Scan(&tunnelRows)
	collect(tunnelRows)

	var forwardRows []nodeTunnelCount
	_ = tx.GostClientForward.
		Select(tx.GostClientForward.NodeCode, tx.GostClientForward.Id.Count().As("total")).
		Group(tx.GostClientForward.NodeCode).Scan(&forwardRows)
	collect(forwardRows)

	var hostRows []nodeTunnelCount
	_ = tx.GostClientHost.
		Select(tx.GostClientHost.NodeCode, tx.GostClientHost.Id.Count().As("total")).
		Group(tx.GostClientHost.NodeCode).Scan(&hostRows)
	collect(hostRows)

	var proxyRows []nodeTunnelCount
	_ = tx.GostClientProxy.
		Select(tx.GostClientProxy.NodeCode, tx.GostClientProxy.Id.Count().As("total")).
		Group(tx.GostClientProxy.NodeCode).Scan(&proxyRows)
	collect(proxyRows)

	var p2pRows []nodeTunnelCount
	_ = tx.GostClientP2P.
		Select(tx.GostClientP2P.NodeCode, tx.GostClientP2P.Id.Count().As("total")).
		Group(tx.GostClientP2P.NodeCode).Scan(&p2pRows)
	collect(p2pRows)

	return counts
}

// VerifyTunnelLimit 校验节点隧道容量，避免大量用户挤占同一节点
func VerifyTunnelLimit(tx *query.Query, node *model.GostNode) error {
	if node.TunnelLimit <= 0 {
		return nil
	}
	total := CountTunnels(tx, node.Code)
	if total >= int64(node.TunnelLimit) {
		return fmt.Errorf("该节点隧道数量已达上限(%d/%d)，请选择其他节点", total, node.TunnelLimit)
	}
	return nil
}
