package service

import (
	"server/model"
	cache2 "server/repository/cache"
)

type RealtimeReq struct {
	OriginKind int      `json:"originKind"`
	OriginCode string   `json:"originCode"`
	Metrics    []string `json:"metrics"`
}

type Series struct {
	Metric string               `json:"metric"`
	Gauge  float64              `json:"gauge"`
	Points []cache2.MetricPoint `json:"points"`
}

func (s *service) Realtime(req RealtimeReq) (list []Series) {
	metrics := req.Metrics
	if len(metrics) == 0 {
		metrics = []string{model.METRIC_INPUT_RATE, model.METRIC_OUTPUT_RATE, model.METRIC_CONNECTIONS}
	}
	for _, m := range metrics {
		list = append(list, Series{
			Metric: m,
			Gauge:  cache2.GetMetricGauge(req.OriginKind, req.OriginCode, m),
			Points: cache2.GetMetricPoints(req.OriginKind, req.OriginCode, m),
		})
	}
	return list
}
