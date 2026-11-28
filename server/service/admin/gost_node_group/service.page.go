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
	Code      string `json:"code"`
	Name      string `json:"name"`
	Strategy  string `json:"strategy"`
	Remark    string `json:"remark"`
	NodeCount int64  `json:"nodeCount"`
	CreatedAt string `json:"createdAt"`
}

func (service *service) Page(req PageReq) (list []Item, total int64) {
	db, _, _ := repository.Get("")
	var where []gen.Condition
	if req.Name != "" {
		where = append(where, db.GostNodeGroup.Name.Like("%"+req.Name+"%"))
	}
	groups, total, _ := db.GostNodeGroup.Where(where...).Order(db.GostNodeGroup.Id.Desc()).FindByPage(req.GetOffset(), req.GetLimit())
	for _, group := range groups {
		nodeCount, _ := db.GostNode.Where(db.GostNode.GroupCode.Eq(group.Code)).Count()
		list = append(list, Item{
			Code:      group.Code,
			Name:      group.Name,
			Strategy:  group.Strategy,
			Remark:    group.Remark,
			NodeCount: nodeCount,
			CreatedAt: group.CreatedAt.Format(time.DateTime),
		})
	}
	return list, total
}
