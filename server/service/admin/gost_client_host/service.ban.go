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

// Ban 封禁/解封域名解析。封禁后引擎下发被 warn_msg 拦截并移除节点配置，连接钩子同步拒绝；解封重新下发。
func (service *service) Ban(req BanReq) error {
	db, _, log := repository.Get("")
	if req.Status != model.STATUS_NORMAL && req.Status != model.STATUS_BANNED {
		return errors.New("状态错误")
	}
	return db.Transaction(func(tx *query.Query) error {
		host, _ := tx.GostClientHost.Where(tx.GostClientHost.Code.Eq(req.Code)).First()
		if host == nil {
			return errors.New("域名解析不存在")
		}
		host.Status = req.Status
		if req.Status == model.STATUS_BANNED {
			host.BanReason = req.Reason
		} else {
			host.BanReason = ""
		}
		if err := tx.GostClientHost.Save(host); err != nil {
			log.Error("封禁域名解析失败", zap.Error(err))
			return errors.New("操作失败")
		}
		cache.SetTunnelBan(host.Code, req.Status)
		engine.ClientHostConfig(tx, host.Code)
		return nil
	})
}