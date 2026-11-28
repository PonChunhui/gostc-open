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
	group, _ := db.GostNodeGroup.Where(db.GostNodeGroup.Code.Eq(req.Code)).First()
	if group == nil {
		return nil
	}
	count, _ := db.GostNode.Where(db.GostNode.GroupCode.Eq(req.Code)).Count()
	if count > 0 {
		return errors.New("请先移除该分组下的所有节点")
	}
	if _, err := db.GostNodeGroup.Where(db.GostNodeGroup.Code.Eq(req.Code)).Delete(); err != nil {
		log.Error("删除节点分组失败", zap.Error(err))
		return errors.New("操作失败")
	}
	return nil
}
