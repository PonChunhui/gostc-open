package service

import (
	"errors"
	"go.uber.org/zap"
	"server/repository"
)

type UpdateReq struct {
	Code     string `binding:"required" json:"code" label:"编号"`
	Name     string `binding:"required" json:"name" label:"名称"`
	Strategy string `json:"strategy" label:"调度策略"`
	Remark   string `json:"remark" label:"备注"`
}

func (service *service) Update(req UpdateReq) error {
	if !validStrategy(req.Strategy) {
		return errors.New("调度策略须为 priority/weight/random")
	}
	db, _, log := repository.Get("")
	group, _ := db.GostNodeGroup.Where(db.GostNodeGroup.Code.Eq(req.Code)).First()
	if group == nil {
		return errors.New("分组不存在")
	}
	group.Name = req.Name
	group.Remark = req.Remark
	if req.Strategy != "" {
		group.Strategy = req.Strategy
	}
	if err := db.GostNodeGroup.Save(group); err != nil {
		log.Error("更新节点分组失败", zap.Error(err))
		return errors.New("操作失败")
	}
	return nil
}
