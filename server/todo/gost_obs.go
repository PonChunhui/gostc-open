package todo

import (
	"server/repository"
	"server/repository/cache"
	"server/service/common/tunnel_quota"
)

func gostObs() {
	db, _, _ := repository.Get("")
	nodes, _ := db.GostNode.Select(
		db.GostNode.Code,
		db.GostNode.LimitResetIndex,
	).Where().Find()
	for _, node := range nodes {
		cache.RefreshNodeObsLimit(node.Code, node.LimitResetIndex)
	}

	{
		rows, _ := db.GostClientHost.Where(db.GostClientHost.LimitTotal.Gt(0)).Find()
		for _, r := range rows {
			tunnel_quota.Refresh(db, r.Code, r.ChargingType, r.Cycle, r.ExpAt, r.CreatedAt)
		}
	}
	{
		rows, _ := db.GostClientForward.Where(db.GostClientForward.LimitTotal.Gt(0)).Find()
		for _, r := range rows {
			tunnel_quota.Refresh(db, r.Code, r.ChargingType, r.Cycle, r.ExpAt, r.CreatedAt)
		}
	}
	{
		rows, _ := db.GostClientTunnel.Where(db.GostClientTunnel.LimitTotal.Gt(0)).Find()
		for _, r := range rows {
			tunnel_quota.Refresh(db, r.Code, r.ChargingType, r.Cycle, r.ExpAt, r.CreatedAt)
		}
	}
	{
		rows, _ := db.GostClientProxy.Where(db.GostClientProxy.LimitTotal.Gt(0)).Find()
		for _, r := range rows {
			tunnel_quota.Refresh(db, r.Code, r.ChargingType, r.Cycle, r.ExpAt, r.CreatedAt)
		}
	}
	{
		rows, _ := db.GostClientP2P.Where(db.GostClientP2P.LimitTotal.Gt(0)).Find()
		for _, r := range rows {
			tunnel_quota.Refresh(db, r.Code, r.ChargingType, r.Cycle, r.ExpAt, r.CreatedAt)
		}
	}
}
