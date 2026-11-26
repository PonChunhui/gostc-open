package service

import (
	"errors"
	"go.uber.org/zap"
	"server/repository"
	"server/repository/cache"
)

type UpdateReq struct {
	Code         string `binding:"required" json:"code" label:"编号"`
	IndexValue   int    `json:"indexValue" label:"排序"`
	Name         string `binding:"required" json:"name" label:"名称"`
	Remark       string `json:"remark" label:"备注"`
	NodeCodes    string `json:"nodeCodes" label:"允许节点"`
	ConfigCodes  string `json:"configCodes" label:"允许套餐"`
	TunnelLimit  int    `json:"tunnelLimit" label:"隧道数量上限"`
	MonthTraffic int    `json:"monthTraffic" label:"每月流量上限GB"`
}

func (service *service) Update(req UpdateReq) error {
	db, _, log := repository.Get("")
	group, _ := db.SystemUserGroup.Where(db.SystemUserGroup.Code.Eq(req.Code)).First()
	if group == nil {
		return errors.New("用户组不存在")
	}
	other, _ := db.SystemUserGroup.Where(db.SystemUserGroup.Name.Eq(req.Name), db.SystemUserGroup.Code.Neq(req.Code)).First()
	if other != nil {
		return errors.New("该用户组名称已被使用")
	}
	group.IndexValue = req.IndexValue
	group.Name = req.Name
	group.Remark = req.Remark
	group.NodeCodes = req.NodeCodes
	group.ConfigCodes = req.ConfigCodes
	group.TunnelLimit = req.TunnelLimit
	group.MonthTraffic = req.MonthTraffic
	if err := db.SystemUserGroup.Save(group); err != nil {
		log.Error("修改用户组失败", zap.Error(err))
		return errors.New("操作失败")
	}
	// 立即失效组内用户的限额缓存，使变更即时生效。
	var userCodes []string
	_ = db.SystemUser.Where(db.SystemUser.GroupCode.Eq(group.Code)).Pluck(db.SystemUser.Code, &userCodes)
	for _, code := range userCodes {
		cache.DelUserGroupLimit(code)
	}
	return nil
}
