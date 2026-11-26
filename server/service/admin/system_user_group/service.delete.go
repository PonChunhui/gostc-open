package service

import (
	"errors"
	"go.uber.org/zap"
	"server/repository"
)

type DeleteReq struct {
	Code string `binding:"required" json:"code" label:"编号"`
}

func (service *service) Delete(req DeleteReq) error {
	db, _, log := repository.Get("")
	group, _ := db.SystemUserGroup.Where(db.SystemUserGroup.Code.Eq(req.Code)).First()
	if group == nil {
		return nil
	}
	count, _ := db.SystemUser.Where(db.SystemUser.GroupCode.Eq(req.Code)).Count()
	if count > 0 {
		return errors.New("请先移除该用户组下的所有用户")
	}
	if _, err := db.SystemUserGroup.Where(db.SystemUserGroup.Code.Eq(req.Code)).Delete(); err != nil {
		log.Error("删除用户组失败", zap.Error(err))
		return errors.New("操作失败")
	}
	return nil
}
