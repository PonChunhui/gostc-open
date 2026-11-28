package service

import (
	"errors"
	"go.uber.org/zap"
	"server/model"
	"server/repository"
)

type CreateReq struct {
	Name     string `binding:"required" json:"name" label:"名称"`
	Strategy string `json:"strategy" label:"调度策略"`
	Remark   string `json:"remark" label:"备注"`
}

func validStrategy(s string) bool {
	switch s {
	case "", model.GOST_NODE_GROUP_STRATEGY_PRIORITY, model.GOST_NODE_GROUP_STRATEGY_WEIGHT, model.GOST_NODE_GROUP_STRATEGY_RANDOM:
		return true
	}
	return false
}

func (service *service) Create(req CreateReq) error {
	if !validStrategy(req.Strategy) {
		return errors.New("调度策略须为 priority/weight/random")
	}
	db, _, log := repository.Get("")
	group, _ := db.GostNodeGroup.Where(db.GostNodeGroup.Name.Eq(req.Name)).First()
	if group != nil {
		return errors.New("该分组名称已被使用")
	}
	strategy := req.Strategy
	if strategy == "" {
		strategy = model.GOST_NODE_GROUP_STRATEGY_PRIORITY
	}
	if err := db.GostNodeGroup.Create(&model.GostNodeGroup{
		Name:     req.Name,
		Strategy: strategy,
		Remark:   req.Remark,
	}); err != nil {
		log.Error("新增节点分组失败", zap.Error(err))
		return errors.New("操作失败")
	}
	return nil
}
