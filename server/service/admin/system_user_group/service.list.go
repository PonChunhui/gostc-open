package service

import (
	"gorm.io/gen"
	"server/repository"
)

type ListReq struct {
	Name string `json:"name"`
}

type ListItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (service *service) List(req ListReq) (list []ListItem) {
	db, _, _ := repository.Get("")
	var where []gen.Condition
	if req.Name != "" {
		where = append(where, db.SystemUserGroup.Name.Like("%"+req.Name+"%"))
	}
	groups, _ := db.SystemUserGroup.Where(where...).Select(
		db.SystemUserGroup.Code,
		db.SystemUserGroup.Name,
	).Order(db.SystemUserGroup.IndexValue.Asc(), db.SystemUserGroup.Id.Desc()).Find()
	for _, group := range groups {
		list = append(list, ListItem{
			Code: group.Code,
			Name: group.Name,
		})
	}
	return list
}
