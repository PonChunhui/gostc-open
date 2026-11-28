package service

import (
	"server/model"
	"server/pkg/bean"
	"server/pkg/jwt"
	"server/repository"
	cache2 "server/repository/cache"
	"server/repository/query"
	"time"
)

type PageReq struct {
	bean.PageParam
	TunnelType   int    `json:"tunnelType"`
	HealthStatus int    `json:"healthStatus"`
	ClientCode   string `json:"clientCode"`
}

type Item struct {
	TunnelCode       string `json:"tunnelCode"`
	TunnelType       int    `json:"tunnelType"`
	Name             string `json:"name"`
	ClientCode       string `json:"clientCode"`
	NodeCode         string `json:"nodeCode"`
	ClientOnline     int    `json:"clientOnline"`
	TunnelRegistered int    `json:"tunnelRegistered"`
	BackendReachable int    `json:"backendReachable"`
	HealthStatus     int    `json:"healthStatus"`
	LastError        string `json:"lastError"`
	ProbeDisable     int    `json:"probeDisable"`
	LastCheckAt      string `json:"lastCheckAt"`
}

func (s *service) Page(claims jwt.Claims, req PageReq) (list []Item, total int64) {
	db, _, _ := repository.Get("")
	q := db.GostTunnelHealth.Where(db.GostTunnelHealth.UserCode.Eq(claims.Code)).Order(db.GostTunnelHealth.Id.Desc())
	if req.TunnelType > 0 {
		q = q.Where(db.GostTunnelHealth.TunnelType.Eq(req.TunnelType))
	}
	if req.HealthStatus > 0 {
		q = q.Where(db.GostTunnelHealth.HealthStatus.Eq(req.HealthStatus))
	}
	if req.ClientCode != "" {
		q = q.Where(db.GostTunnelHealth.ClientCode.Eq(req.ClientCode))
	}
	rows, total, _ := q.FindByPage(req.GetOffset(), req.GetLimit())
	for _, r := range rows {
		live := cache2.GetTunnelHealth(r.TunnelCode)
		item := Item{
			TunnelCode: r.TunnelCode, TunnelType: r.TunnelType, Name: tunnelName(db, r.TunnelType, r.TunnelCode),
			ClientCode: r.ClientCode, NodeCode: r.NodeCode,
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

type Overview struct {
	Total          int64 `json:"total"`
	Healthy        int64 `json:"healthy"`
	Abnormal       int64 `json:"abnormal"`
	Unknown        int64 `json:"unknown"`
	ClientOffline  int64 `json:"clientOffline"`
	Unregistered   int64 `json:"unregistered"`
	BackendBad     int64 `json:"backendBad"`
}

func (s *service) Overview(claims jwt.Claims) Overview {
	db, _, _ := repository.Get("")
	var ov Overview
	rows, _ := db.GostTunnelHealth.Where(db.GostTunnelHealth.UserCode.Eq(claims.Code)).Find()
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
