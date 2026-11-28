package service

import (
	"server/model"
	"server/pkg/bean"
	"server/repository"
	cache2 "server/repository/cache"
	"server/repository/query"
	"time"
)

// PageReq 管理员全站健康列表筛选（不限属主）。
type PageReq struct {
	bean.PageParam
	UserCode     string `json:"userCode"`
	NodeCode     string `json:"nodeCode"`
	TunnelType   int    `json:"tunnelType"`
	HealthStatus int    `json:"healthStatus"`
	ClientCode   string `json:"clientCode"`
	Keyword      string `json:"keyword"`
}

type Item struct {
	TunnelCode       string `json:"tunnelCode"`
	TunnelType       int    `json:"tunnelType"`
	Name             string `json:"name"`
	UserCode         string `json:"userCode"`
	Account          string `json:"account"`
	ClientCode       string `json:"clientCode"`
	NodeCode         string `json:"nodeCode"`
	NodeName         string `json:"nodeName"`
	ClientOnline     int    `json:"clientOnline"`
	TunnelRegistered int    `json:"tunnelRegistered"`
	BackendReachable int    `json:"backendReachable"`
	HealthStatus     int    `json:"healthStatus"`
	LastError        string `json:"lastError"`
	ProbeDisable     int    `json:"probeDisable"`
	LastCheckAt      string `json:"lastCheckAt"`
}

func (s *service) Page(req PageReq) (list []Item, total int64) {
	db, _, _ := repository.Get("")
	q := db.GostTunnelHealth.Where().Order(db.GostTunnelHealth.Id.Desc())
	if req.UserCode != "" {
		q = q.Where(db.GostTunnelHealth.UserCode.Eq(req.UserCode))
	}
	if req.NodeCode != "" {
		q = q.Where(db.GostTunnelHealth.NodeCode.Eq(req.NodeCode))
	}
	if req.TunnelType > 0 {
		q = q.Where(db.GostTunnelHealth.TunnelType.Eq(req.TunnelType))
	}
	if req.HealthStatus > 0 {
		q = q.Where(db.GostTunnelHealth.HealthStatus.Eq(req.HealthStatus))
	}
	if req.ClientCode != "" {
		q = q.Where(db.GostTunnelHealth.ClientCode.Eq(req.ClientCode))
	}
	if req.Keyword != "" {
		q = q.Where(db.GostTunnelHealth.TunnelCode.Like("%" + req.Keyword + "%"))
	}
	rows, total, _ := q.FindByPage(req.GetOffset(), req.GetLimit())
	accounts := accountMap(db, rows)
	nodes := nodeMap(db, rows)
	for _, r := range rows {
		live := cache2.GetTunnelHealth(r.TunnelCode)
		item := Item{
			TunnelCode: r.TunnelCode, TunnelType: r.TunnelType, Name: tunnelName(db, r.TunnelType, r.TunnelCode),
			UserCode: r.UserCode, Account: accounts[r.UserCode],
			ClientCode: r.ClientCode, NodeCode: r.NodeCode, NodeName: nodes[r.NodeCode],
			ClientOnline: r.ClientOnline, TunnelRegistered: r.TunnelRegistered, BackendReachable: r.BackendReachable,
			HealthStatus: r.HealthStatus, LastError: r.LastError, ProbeDisable: r.ProbeDisable,
		}
		if live.LastCheckAt > 0 {
			item.ClientOnline = live.ClientOnline
			item.TunnelRegistered = live.TunnelRegistered
			item.BackendReachable = live.BackendReachable
			item.HealthStatus = live.HealthStatus
			item.LastError = live.LastError
			item.LastCheckAt = time.Unix(live.LastCheckAt, 0).Format(time.DateTime)
		} else if r.LastCheckAt > 0 {
			item.LastCheckAt = time.Unix(r.LastCheckAt, 0).Format(time.DateTime)
		}
		list = append(list, item)
	}
	return list, total
}

type OverviewReq struct {
	UserCode   string `json:"userCode"`
	NodeCode   string `json:"nodeCode"`
	TunnelType int    `json:"tunnelType"`
}

type Overview struct {
	Total         int64 `json:"total"`
	Healthy       int64 `json:"healthy"`
	Abnormal      int64 `json:"abnormal"`
	Unknown       int64 `json:"unknown"`
	ClientOffline int64 `json:"clientOffline"`
	Unregistered  int64 `json:"unregistered"`
	BackendBad    int64 `json:"backendBad"`
}

func (s *service) Overview(req OverviewReq) Overview {
	db, _, _ := repository.Get("")
	q := db.GostTunnelHealth.Where()
	if req.UserCode != "" {
		q = q.Where(db.GostTunnelHealth.UserCode.Eq(req.UserCode))
	}
	if req.NodeCode != "" {
		q = q.Where(db.GostTunnelHealth.NodeCode.Eq(req.NodeCode))
	}
	if req.TunnelType > 0 {
		q = q.Where(db.GostTunnelHealth.TunnelType.Eq(req.TunnelType))
	}
	var ov Overview
	rows, _ := q.Find()
	for _, r := range rows {
		live := cache2.GetTunnelHealth(r.TunnelCode)
		status, online, reg, reach := r.HealthStatus, r.ClientOnline, r.TunnelRegistered, r.BackendReachable
		if live.LastCheckAt > 0 {
			status, online, reg, reach = live.HealthStatus, live.ClientOnline, live.TunnelRegistered, live.BackendReachable
		}
		ov.Total++
		switch status {
		case model.TUNNEL_HEALTH_HEALTHY:
			ov.Healthy++
		case model.TUNNEL_HEALTH_ABNORMAL:
			ov.Abnormal++
		default:
			ov.Unknown++
		}
		if online == 2 {
			ov.ClientOffline++
		}
		if reg == 2 {
			ov.Unregistered++
		}
		if reach == 2 {
			ov.BackendBad++
		}
	}
	return ov
}

func accountMap(db *query.Query, rows []*model.GostTunnelHealth) map[string]string {
	m := map[string]string{}
	codes := make([]string, 0)
	for _, r := range rows {
		if r.UserCode != "" {
			codes = append(codes, r.UserCode)
		}
	}
	if len(codes) == 0 {
		return m
	}
	users, _ := db.SystemUser.Where(db.SystemUser.Code.In(codes...)).Find()
	for _, u := range users {
		m[u.Code] = u.Account
	}
	return m
}

func nodeMap(db *query.Query, rows []*model.GostTunnelHealth) map[string]string {
	m := map[string]string{}
	codes := make([]string, 0)
	for _, r := range rows {
		if r.NodeCode != "" {
			codes = append(codes, r.NodeCode)
		}
	}
	if len(codes) == 0 {
		return m
	}
	nodes, _ := db.GostNode.Where(db.GostNode.Code.In(codes...)).Find()
	for _, n := range nodes {
		m[n.Code] = n.Name
	}
	return m
}

func tunnelName(db *query.Query, tunnelType int, code string) string {
	switch tunnelType {
	case model.TUNNEL_HEALTH_TYPE_HOST:
		if r, err := db.GostClientHost.Where(db.GostClientHost.Code.Eq(code)).First(); err == nil {
			return r.Name
		}
	case model.TUNNEL_HEALTH_TYPE_FORWARD:
		if r, err := db.GostClientForward.Where(db.GostClientForward.Code.Eq(code)).First(); err == nil {
			return r.Name
		}
	case model.TUNNEL_HEALTH_TYPE_PROXY:
		if r, err := db.GostClientProxy.Where(db.GostClientProxy.Code.Eq(code)).First(); err == nil {
			return r.Name
		}
	case model.TUNNEL_HEALTH_TYPE_TUNNEL:
		if r, err := db.GostClientTunnel.Where(db.GostClientTunnel.Code.Eq(code)).First(); err == nil {
			return r.Name
		}
	case model.TUNNEL_HEALTH_TYPE_P2P:
		if r, err := db.GostClientP2P.Where(db.GostClientP2P.Code.Eq(code)).First(); err == nil {
			return r.Name
		}
	}
	return ""
}
