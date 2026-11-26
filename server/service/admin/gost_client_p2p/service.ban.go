package service

import (
	"errors"
	"go.uber.org/zap"
	"server/model"
	"server/repository"
	"server/repository/cache"
	"server/repository/query"
	"server/service/engine"
)

type BanReq struct {
	Code   string `binding:"required" json:"code" label:"编号"`
	Status int    `binding:"required" json:"status" label:"状态"`
	Reason string `json:"reason" label:"封禁原因"`
}

// Ban 封禁/解封P2P隧道。封禁后引擎下发被 warn_msg 拦截并移除节点配置，连接钩子同步拒绝；解封重新下发。
func (service *service) Ban(req BanReq) error {
	db, _, log := repository.Get("")
	if req.Status != model.STATUS_NORMAL && req.Status != model.STATUS_BANNED {
		return errors.New("状态错误")
	}
	return db.Transaction(func(tx *query.Query) error {
		p2p, _ := tx.GostClientP2P.Where(tx.GostClientP2P.Code.Eq(req.Code)).First()
		if p2p == nil {
			return errors.New("P2P隧道不存在")
		}
		p2p.Status = req.Status
		if req.Status == model.STATUS_BANNED {
			p2p.BanReason = req.Reason
		} else {
			p2p.BanReason = ""
		}
		if err := tx.GostClientP2P.Save(p2p); err != nil {
			log.Error("封禁P2P隧道失败", zap.Error(err))
			return errors.New("操作失败")
		}
		cache.SetTunnelBan(p2p.Code, req.Status)
		engine.ClientP2PConfig(tx, p2p.Code)
		return nil
	})
}