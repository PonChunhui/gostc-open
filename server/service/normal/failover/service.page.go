package service

import (
	"server/pkg/bean"
	"server/pkg/jwt"
	"server/repository"
	"time"
)

type PageReq struct {
	bean.PageParam
	TunnelCode string `json:"tunnelCode"`
	Reason     string `json:"reason"`
}

type Item struct {
	Code       string `json:"code"`
	TunnelCode string `json:"tunnelCode"`
	TunnelType int    `json:"tunnelType"`
	FromNode   string `json:"fromNode"`
	ToNode     string `json:"toNode"`
	OldEntry   string `json:"oldEntry"`
	NewEntry   string `json:"newEntry"`
	Reason     string `json:"reason"`
	CreatedAt  string `json:"createdAt"`
}

func (s *service) Page(claims jwt.Claims, req PageReq) (list []Item, total int64) {
	db, _, _ := repository.Get("")
	q := db.GostFailoverLog.Where(db.GostFailoverLog.UserCode.Eq(claims.Code)).Order(db.GostFailoverLog.Id.Desc())
	if req.TunnelCode != "" {
		q = q.Where(db.GostFailoverLog.TunnelCode.Eq(req.TunnelCode))
	}
	if req.Reason != "" {
		q = q.Where(db.GostFailoverLog.Reason.Eq(req.Reason))
	}
	logs, total, _ := q.FindByPage(req.GetOffset(), req.GetLimit())
	for _, l := range logs {
		list = append(list, Item{
			Code: l.Code, TunnelCode: l.TunnelCode, TunnelType: l.TunnelType,
			FromNode: l.FromNode, ToNode: l.ToNode, OldEntry: l.OldEntry, NewEntry: l.NewEntry,
			Reason: l.Reason, CreatedAt: l.CreatedAt.Format(time.DateTime),
		})
	}
	return list, total
}
