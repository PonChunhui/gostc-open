package service

import (
	"errors"
	"go.uber.org/zap"
	"server/model"
	"server/repository"
	"server/repository/cache"
	"server/service/engine"
)

type BanReq struct {
	Code   string `binding:"required" json:"code" label:"编号"`
	Status int    `binding:"required" json:"status" label:"状态"`
	Reason string `json:"reason" label:"封禁原因"`
}

// Ban 封禁/解封用户。封禁立即踢下线其全部客户端并拦截登录、注册与连接；解封恢复在线客户端配置下发。
func (service *service) Ban(req BanReq) error {
	db, _, log := repository.Get("")
	user, _ := db.SystemUser.Where(db.SystemUser.Code.Eq(req.Code)).First()
	if user == nil {
		return errors.New("用户不存在")
	}
	if user.Admin == model.SYSTEM_IS_ADMIN {
		return errors.New("不能封禁管理员账号")
	}
	if req.Status != model.STATUS_NORMAL && req.Status != model.STATUS_BANNED {
		return errors.New("状态错误")
	}
	user.Status = req.Status
	if req.Status == model.STATUS_BANNED {
		user.BanReason = req.Reason
	} else {
		user.BanReason = ""
	}
	if err := db.SystemUser.Save(user); err != nil {
		log.Error("封禁用户失败", zap.Error(err))
		return errors.New("操作失败")
	}
	cache.SetUserBan(user.Code, req.Status)

	clients, _ := db.GostClient.Where(db.GostClient.UserCode.Eq(user.Code)).Find()
	for _, client := range clients {
		if req.Status == model.STATUS_BANNED {
			engine.ClientStop(client.Code, "账号已被封禁")
		} else {
			engine.ClientAllConfigUpdateByClientCode(db, client.Code)
		}
	}
	return nil
}