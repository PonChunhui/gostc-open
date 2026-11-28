package service

import (
	"server/pkg/bean"
	"server/repository"
	"time"
)

type PageReq struct {
	bean.PageParam
	Account string `json:"account"`
	Path    string `json:"path"`
	Method  string `json:"method"`
}

type Item struct {
	Code      string `json:"code"`
	Account   string `json:"account"`
	UserCode  string `json:"userCode"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	IP        string `json:"ip"`
	Status    int    `json:"status"`
	Detail    string `json:"detail"`
	CreatedAt string `json:"createdAt"`
}

func (service *service) Page(req PageReq) (list []Item, total int64) {
	db, _, _ := repository.Get("")
	q := db.SystemAuditLog.Order(db.SystemAuditLog.Id.Desc())
	if req.Account != "" {
		q = q.Where(db.SystemAuditLog.Account.Like("%" + req.Account + "%"))
	}
	if req.Path != "" {
		q = q.Where(db.SystemAuditLog.Path.Like("%" + req.Path + "%"))
	}
	if req.Method != "" {
		q = q.Where(db.SystemAuditLog.Method.Eq(req.Method))
	}
	logs, total, _ := q.FindByPage(req.GetOffset(), req.GetLimit())
	for _, l := range logs {
		list = append(list, Item{
			Code:      l.Code,
			Account:   l.Account,
			UserCode:  l.UserCode,
			Method:    l.Method,
			Path:      l.Path,
			IP:        l.IP,
			Status:    l.Status,
			Detail:    l.Detail,
			CreatedAt: l.CreatedAt.Format(time.DateTime),
		})
	}
	return list, total
}
