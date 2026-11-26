package service

import (
	"errors"
	"go.uber.org/zap"
	"server/model"
	"server/repository"
)

type CreateReq struct {
	IndexValue   int    `json:"indexValue" label:"排序"`
	Name         string `binding:"required" json:"name" label:"名称"`
	Remark       string `json:"remark" label:"备注"`
	NodeCodes    string `json:"nodeCodes" label:"允许节点"`
	ConfigCodes  string `json:"configCodes" label:"允许套餐"`
	TunnelLimit  int    `json:"tunnelLimit" label:"隧道数量上限"`
	MonthTraffic int    `json:"monthTraffic" label:"每月流量上限GB"`
}

func (service *service) Create(req CreateReq) error {
	db, _, log := repository.Get("")
	group, _ := db.SystemUserGroup.Where(db.SystemUserGroup.Name.Eq(req.Name)).First()
	if group != nil {
		return errors.New("该用户组名称已被使用")
	}
	if err := db.SystemUserGroup.Create(&model.SystemUserGroup{
		IndexValue:   req.IndexValue,
		Name:         req.Name,
		Remark:       req.Remark,
		NodeCodes:    req.NodeCodes,
		ConfigCodes:  req.ConfigCodes,
		TunnelLimit:  req.TunnelLimit,
		MonthTraffic: req.MonthTraffic,
	}); err != nil {
		log.Error("新增用户组失败", zap.Error(err))
		return errors.New("操作失败")
	}
	return nil
}
