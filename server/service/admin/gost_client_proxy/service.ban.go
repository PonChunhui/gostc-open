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

// Ban 封禁/解封代理隧道。封禁后引擎下发被 warn_msg 拦截并移除节点配置，连接钩子同步拒绝；解封重新下发。
func (service *service) Ban(req BanReq) error {
	db, _, log := repository.Get("")
	if req.Status != model.STATUS_NORMAL && req.Status != model.STATUS_BANNED {
		return errors.New("状态错误")
	}
	return db.Transaction(func(tx *query.Query) error {
		proxy, _ := tx.GostClientProxy.Where(tx.GostClientProxy.Code.Eq(req.Code)).First()
		if proxy == nil {
			return errors.New("代理隧道不存在")
		}
		proxy.Status = req.Status
		if req.Status == model.STATUS_BANNED {
			proxy.BanReason = req.Reason
		} else {
			proxy.BanReason = ""
		}
		if err := tx.GostClientProxy.Save(proxy); err != nil {
			log.Error("封禁代理隧道失败", zap.Error(err))
			return errors.New("操作失败")
		}
		cache.SetTunnelBan(proxy.Code, req.Status)
		engine.ClientProxyConfig(tx, proxy.Code)
		return nil
	})
}