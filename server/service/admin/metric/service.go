package service

import (
	"server/model"
	"server/repository"
	cache2 "server/repository/cache"
	metricservice "server/service/normal/metric"
	"sort"
)

type service struct {
}

var Service *service

// Realtime 复用用户级实时指标读取（指标按对象编码全局存储，管理员可查任意对象）。
func (s *service) Realtime(req metricservice.RealtimeReq) []metricservice.Series {
	return metricservice.Service.Realtime(req)
}

// History 复用用户级历史指标读取。
func (s *service) History(req metricservice.HistoryReq) []metricservice.HistoryPoint {
	return metricservice.Service.History(req)
}

// TopReq 全站某类对象按即时指标排序取 TopN。
type TopReq struct {
	OriginKind int    `json:"originKind"` // 1节点 2客户端 3隧道 4用户
	Metric     string `json:"metric"`
	Limit      int    `json:"limit"`
}

type TopItem struct {
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

func (s *service) Top(req TopReq) []TopItem {
	if req.Metric == "" {
		req.Metric = model.METRIC_INPUT_RATE
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 10
	}
	db, _, _ := repository.Get("")
	names := map[string]string{}
	switch req.OriginKind {
	case 1:
		nodes, _ := db.GostNode.Find()
		for _, n := range nodes {
			names[n.Code] = n.Name
		}
	case 2:
		clients, _ := db.GostClient.Find()
		for _, c := range clients {
			names[c.Code] = c.Name
		}
	case 3:
		healths, _ := db.GostTunnelHealth.Find()
		for _, h := range healths {
			names[h.TunnelCode] = h.TunnelCode
		}
	case 4:
		users, _ := db.SystemUser.Find()
		for _, u := range users {
			names[u.Code] = u.Account
		}
	}
	items := make([]TopItem, 0, len(names))
	for code, name := range names {
		items = append(items, TopItem{
			Code: code, Name: name,
			Value: cache2.GetMetricGauge(req.OriginKind, code, req.Metric),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Value > items[j].Value })
	if len(items) > req.Limit {
		items = items[:req.Limit]
	}
	return items
}
