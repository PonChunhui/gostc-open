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

// Ban 封禁/解封端口转发。封禁后引擎下发被 warn_msg 拦截并移除节点配置，连接钩子同步拒绝；解封重新下发。
func (service *service) Ban(req BanReq) error {
	db, _, log := repository.Get("")
	if req.Status != model.STATUS_NORMAL && req.Status != model.STATUS_BANNED {
		return errors.New("状态错误")
	}
	return db.Transaction(func(tx *query.Query) error {
		forward, _ := tx.GostClientForward.Where(tx.GostClientForward.Code.Eq(req.Code)).First()
		if forward == nil {
			return errors.New("端口转发不存在")
		}
		forward.Status = req.Status
		if req.Status == model.STATUS_BANNED {
			forward.BanReason = req.Reason
		} else {
			forward.BanReason = ""
		}
		if err := tx.GostClientForward.Save(forward); err != nil {
			log.Error("封禁端口转发失败", zap.Error(err))
			return errors.New("操作失败")
		}
		cache.SetTunnelBan(forward.Code, req.Status)
		engine.ClientForwardConfig(tx, forward.Code)
		return nil
	})
}