package service

import (
	"gorm.io/gen"
	"server/pkg/bean"
	"server/repository"
	"time"
)

type PageReq struct {
	bean.PageParam
	Name string `json:"name"`
}

type Item struct {
	Code         string `json:"code"`
	IndexValue   int    `json:"indexValue"`
	Name         string `json:"name"`
	Remark       string `json:"remark"`
	NodeCodes    string `json:"nodeCodes"`
	ConfigCodes  string `json:"configCodes"`
	TunnelLimit  int    `json:"tunnelLimit"`
	MonthTraffic int    `json:"monthTraffic"`
	UserCount    int64  `json:"userCount"`
	CreatedAt    string `json:"createdAt"`
}

func (service *service) Page(req PageReq) (list []Item, total int64) {
	db, _, _ := repository.Get("")
	var where []gen.Condition
	if req.Name != "" {
		where = append(where, db.SystemUserGroup.Name.Like("%"+req.Name+"%"))
	}
	groups, total, _ := db.SystemUserGroup.Where(where...).Order(db.SystemUserGroup.IndexValue.Asc(), db.SystemUserGroup.Id.Desc()).FindByPage(req.GetOffset(), req.GetLimit())
	for _, group := range groups {
		userCount, _ := db.SystemUser.Where(db.SystemUser.GroupCode.Eq(group.Code)).Count()
		list = append(list, Item{
			Code:         group.Code,
			IndexValue:   group.IndexValue,
			Name:         group.Name,
			Remark:       group.Remark,
			NodeCodes:    group.NodeCodes,
			ConfigCodes:  group.ConfigCodes,
			TunnelLimit:  group.TunnelLimit,
			MonthTraffic: group.MonthTraffic,
			UserCount:    userCount,
			CreatedAt:    group.CreatedAt.Format(time.DateTime),
		})
	}
	return list, total
}
