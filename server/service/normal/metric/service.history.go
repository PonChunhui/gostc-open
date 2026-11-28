package service

import (
	"server/repository"
	"time"
)

type HistoryReq struct {
	OriginKind int    `json:"originKind"`
	OriginCode string `json:"originCode"`
	Metric     string `json:"metric"`
	StartTs    int64  `json:"startTs"`
	EndTs      int64  `json:"endTs"`
}

type HistoryPoint struct {
	Ts    int64   `json:"ts"`
	Value float64 `json:"value"`
}

func (s *service) History(req HistoryReq) (list []HistoryPoint) {
	if req.Metric == "" || req.OriginCode == "" {
		return []HistoryPoint{}
	}
	now := time.Now().Unix()
	if req.EndTs <= 0 || req.EndTs > now {
		req.EndTs = now
	}
	if req.StartTs <= 0 {
		req.StartTs = req.EndTs - 86400 // 默认最近 24 小时
	}
	db, _, _ := repository.Get("")
	rows, _ := db.GostMetric.Where(
		db.GostMetric.OriginKind.Eq(req.OriginKind),
		db.GostMetric.OriginCode.Eq(req.OriginCode),
		db.GostMetric.Metric.Eq(req.Metric),
		db.GostMetric.Ts.Gte(req.StartTs),
		db.GostMetric.Ts.Lte(req.EndTs),
	).Order(db.GostMetric.Ts.Asc()).Limit(2000).Find()
	for _, r := range rows {
		list = append(list, HistoryPoint{Ts: r.Ts, Value: r.Value})
	}
	if list == nil {
		list = []HistoryPoint{}
	}
	return list
}
