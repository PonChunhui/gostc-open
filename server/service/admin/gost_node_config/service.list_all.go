package service

import (
	"server/repository"
)

type ListAllReq struct {
}

type ListAllItem struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	NodeCode     string `json:"nodeCode"`
	NodeName     string `json:"nodeName"`
	ChargingType int    `json:"chargingType"`
	LimitTotal   int    `json:"limitTotal"`
}

// ListAll 返回全部节点套餐(含所属节点)，供用户组白名单选择。
func (service *service) ListAll(req ListAllReq) (list []ListAllItem) {
	db, _, _ := repository.Get("")
	cfgs, _ := db.GostNodeConfig.Preload(db.GostNodeConfig.Node).Order(
		db.GostNodeConfig.NodeCode.Asc(),
		db.GostNodeConfig.IndexValue.Asc(),
		db.GostNodeConfig.Id.Desc(),
	).Find()
	for _, cfg := range cfgs {
		list = append(list, ListAllItem{
			Code:         cfg.Code,
			Name:         cfg.Name,
			NodeCode:     cfg.NodeCode,
			NodeName:     cfg.Node.Name,
			ChargingType: cfg.ChargingType,
			LimitTotal:   cfg.LimitTotal,
		})
	}
	return list
}
