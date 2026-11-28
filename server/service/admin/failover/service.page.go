package service

import (
	"server/pkg/bean"
	"server/repository"
	"time"
)

type PageReq struct {
	bean.PageParam
	UserCode   string `json:"userCode"`
	TunnelCode string `json:"tunnelCode"`
	Reason     string `json:"reason"`
}

type Item struct {
	Code       string `json:"code"`
	TunnelCode string `json:"tunnelCode"`
	TunnelType int    `json:"tunnelType"`
	UserCode   string `json:"userCode"`
	Account    string `json:"account"`
	FromNode   string `json:"fromNode"`
	ToNode     string `json:"toNode"`
	FromName   string `json:"fromName"`
	ToName     string `json:"toName"`
	OldEntry   string `json:"oldEntry"`
	NewEntry   string `json:"newEntry"`
	Reason     string `json:"reason"`
	CreatedAt  string `json:"createdAt"`
}

func (s *service) Page(req PageReq) (list []Item, total int64) {
	db, _, _ := repository.Get("")
	q := db.GostFailoverLog.Where().Order(db.GostFailoverLog.Id.Desc())
	if req.UserCode != "" {
		q = q.Where(db.GostFailoverLog.UserCode.Eq(req.UserCode))
	}
	if req.TunnelCode != "" {
		q = q.Where(db.GostFailoverLog.TunnelCode.Eq(req.TunnelCode))
	}
	if req.Reason != "" {
		q = q.Where(db.GostFailoverLog.Reason.Eq(req.Reason))
	}
	logs, total, _ := q.FindByPage(req.GetOffset(), req.GetLimit())

	accountMap := map[string]string{}
	nodeMap := map[string]string{}
	for _, l := range logs {
		accountMap[l.UserCode] = ""
		nodeMap[l.FromNode] = ""
		nodeMap[l.ToNode] = ""
	}
	if len(accountMap) > 0 {
		codes := make([]string, 0, len(accountMap))
		for c := range accountMap {
			codes = append(codes, c)
		}
		users, _ := db.SystemUser.Where(db.SystemUser.Code.In(codes...)).Find()
		for _, u := range users {
			accountMap[u.Code] = u.Account
		}
	}
	if len(nodeMap) > 0 {
		codes := make([]string, 0, len(nodeMap))
		for c := range nodeMap {
			codes = append(codes, c)
		}
		nodes, _ := db.GostNode.Where(db.GostNode.Code.In(codes...)).Find()
		for _, n := range nodes {
			nodeMap[n.Code] = n.Name
		}
	}
	for _, l := range logs {
		list = append(list, Item{
			Code: l.Code, TunnelCode: l.TunnelCode, TunnelType: l.TunnelType,
			UserCode: l.UserCode, Account: accountMap[l.UserCode],
			FromNode: l.FromNode, ToNode: l.ToNode, FromName: nodeMap[l.FromNode], ToName: nodeMap[l.ToNode],
			OldEntry: l.OldEntry, NewEntry: l.NewEntry, Reason: l.Reason,
			CreatedAt: l.CreatedAt.Format(time.DateTime),
		})
	}
	return list, total
}
