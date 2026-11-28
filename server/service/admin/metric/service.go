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

// ObjectItem 可供指标查询选择的对象（编码+名称）。
type ObjectItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// ObjectsReq 按对象类型列出可选对象。
type ObjectsReq struct {
	OriginKind int `json:"originKind"` // 1节点 2客户端 3隧道 4用户
}

// Objects 返回某类全部对象，供前端实时/历史曲线选择，避免手填编码。
func (s *service) Objects(req ObjectsReq) []ObjectItem {
	db, _, _ := repository.Get("")
	items := make([]ObjectItem, 0)
	switch req.OriginKind {
	case 1:
		nodes, _ := db.GostNode.Find()
		for _, n := range nodes {
			items = append(items, ObjectItem{Code: n.Code, Name: n.Name})
		}
	case 2:
		clients, _ := db.GostClient.Find()
		for _, c := range clients {
			items = append(items, ObjectItem{Code: c.Code, Name: c.Name})
		}
	case 3:
		hosts, _ := db.GostClientHost.Find()
		for _, h := range hosts {
			items = append(items, ObjectItem{Code: h.Code, Name: h.Name})
		}
		forwards, _ := db.GostClientForward.Find()
		for _, f := range forwards {
			items = append(items, ObjectItem{Code: f.Code, Name: f.Name})
		}
		proxies, _ := db.GostClientProxy.Find()
		for _, p := range proxies {
			items = append(items, ObjectItem{Code: p.Code, Name: p.Name})
		}
		tunnels, _ := db.GostClientTunnel.Find()
		for _, t := range tunnels {
			items = append(items, ObjectItem{Code: t.Code, Name: t.Name})
		}
		p2ps, _ := db.GostClientP2P.Find()
		for _, p := range p2ps {
			items = append(items, ObjectItem{Code: p.Code, Name: p.Name})
		}
	case 4:
		users, _ := db.SystemUser.Find()
		for _, u := range users {
			items = append(items, ObjectItem{Code: u.Code, Name: u.Account})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items
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
