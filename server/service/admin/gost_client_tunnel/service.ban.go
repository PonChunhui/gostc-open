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

// Ban 封禁/解封私有隧道。封禁后引擎下发被 warn_msg 拦截并移除节点配置，连接钩子同步拒绝；解封重新下发。
func (service *service) Ban(req BanReq) error {
	db, _, log := repository.Get("")
	if req.Status != model.STATUS_NORMAL && req.Status != model.STATUS_BANNED {
		return errors.New("状态错误")
	}
	return db.Transaction(func(tx *query.Query) error {
		tunnel, _ := tx.GostClientTunnel.Where(tx.GostClientTunnel.Code.Eq(req.Code)).First()
		if tunnel == nil {
			return errors.New("隧道不存在")
		}
		tunnel.Status = req.Status
		if req.Status == model.STATUS_BANNED {
			tunnel.BanReason = req.Reason
		} else {
			tunnel.BanReason = ""
		}
		if err := tx.GostClientTunnel.Save(tunnel); err != nil {
			log.Error("封禁私有隧道失败", zap.Error(err))
			return errors.New("操作失败")
		}
		cache.SetTunnelBan(tunnel.Code, req.Status)
		engine.ClientTunnelConfig(tx, tunnel.Code)
		return nil
	})
}