package task

import (
	"go.uber.org/zap"
	"server/repository"
	cache3 "server/repository/cache"
)

// p2pRelayFlush 将缓存中的 P2P STCP 中继流量增量持久化到 GostClientP2P.RelayBytes。
func p2pRelayFlush() {
	db, _, log := repository.Get("")
	list, err := db.GostClientP2P.Find()
	if err != nil {
		log.Error("查询P2P隧道失败", zap.Error(err))
		return
	}
	for _, p2p := range list {
		delta := cache3.TakeP2PRelay(p2p.Code)
		if delta <= 0 {
			continue
		}
		p2p.RelayBytes += delta
		if err := db.GostClientP2P.Save(p2p); err != nil {
			log.Error("P2P中继流量落库失败", zap.String("code", p2p.Code), zap.Error(err))
			// 落库失败把增量放回缓存，避免丢失
			cache3.IncreaseP2PRelay(p2p.Code, delta)
		}
	}
}
