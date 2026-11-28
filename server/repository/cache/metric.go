package cache

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"server/global"
	"sync"
	"time"
)

const (
	metric_gauge_key = "metric:gauge:%d:%s:%s" // kind:code:metric
	metric_ring_key  = "metric:ring:%d:%s:%s"
	metric_ring_cap  = 120 // 实时曲线保留点数
)

var metricLock = &sync.RWMutex{}

// MetricPoint 实时曲线采样点。
type MetricPoint struct {
	Ts    int64   `json:"ts"`
	Value float64 `json:"value"`
}

type metricRing struct {
	Points []MetricPoint
}

// SetMetricGauge 写入即时指标值（连接数、速率等）。
func SetMetricGauge(kind int, code, metric string, value float64) {
	global.Cache.SetString(fmt.Sprintf(metric_gauge_key, kind, code, metric), fmt.Sprintf("%f", value), time.Minute*10)
}

func GetMetricGauge(kind int, code, metric string) float64 {
	s := global.Cache.GetString(fmt.Sprintf(metric_gauge_key, kind, code, metric))
	var v float64
	_, _ = fmt.Sscanf(s, "%f", &v)
	return v
}

// PushMetricPoint 追加实时曲线采样点（环形缓冲，保留最近 metric_ring_cap 个）。
func PushMetricPoint(kind int, code, metric string, ts int64, value float64) {
	metricLock.Lock()
	defer metricLock.Unlock()
	key := fmt.Sprintf(metric_ring_key, kind, code, metric)
	var ring metricRing
	_ = global.Cache.GetStruct(key, &ring)
	ring.Points = append(ring.Points, MetricPoint{Ts: ts, Value: value})
	if len(ring.Points) > metric_ring_cap {
		ring.Points = ring.Points[len(ring.Points)-metric_ring_cap:]
	}
	global.Cache.SetStruct(key, ring, time.Minute*30)
}

func GetMetricPoints(kind int, code, metric string) []MetricPoint {
	metricLock.RLock()
	defer metricLock.RUnlock()
	var ring metricRing
	_ = global.Cache.GetStruct(fmt.Sprintf(metric_ring_key, kind, code, metric), &ring)
	if ring.Points == nil {
		return []MetricPoint{}
	}
	return ring.Points
}

// 上一周期累计字节快照，用于计算速率。
const metric_obs_snapshot_key = "metric:obs:snapshot:%d:%s"

type ObsSnapshot struct {
	InputBytes  int64
	OutputBytes int64
	Ts          int64
}

func GetObsSnapshot(kind int, code string) (s ObsSnapshot) {
	_ = global.Cache.GetStruct(fmt.Sprintf(metric_obs_snapshot_key, kind, code), &s)
	return s
}

func SetObsSnapshot(kind int, code string, s ObsSnapshot) {
	global.Cache.SetStruct(fmt.Sprintf(metric_obs_snapshot_key, kind, code), s, cache.NoExpiration)
}

// 节点周期内新建用户连接计数（frps newUserConn 插件累加，采集时取走清零）。
const metric_node_conn_key = "metric:node:conn:%s"

type NodeConnObs struct {
	Count int64
}

func IncreaseNodeConn(code string) {
	if code == "" {
		return
	}
	key := fmt.Sprintf(metric_node_conn_key, code)
	var obs NodeConnObs
	_ = global.Cache.GetStruct(key, &obs)
	obs.Count++
	global.Cache.SetStruct(key, obs, cache.NoExpiration)
}

func TakeNodeConn(code string) int64 {
	key := fmt.Sprintf(metric_node_conn_key, code)
	var obs NodeConnObs
	_ = global.Cache.GetStruct(key, &obs)
	global.Cache.Del(key)
	return obs.Count
}
