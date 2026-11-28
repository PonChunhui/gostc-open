package task

import (
	"fmt"
	"server/global"
	"server/model"
	"server/repository"
	cache2 "server/repository/cache"
	"server/repository/query"
	"time"

	"go.uber.org/zap"
)

// metricCollect 周期采集速率/连接数指标：计算滑动速率、刷新实时 gauge 与曲线、落历史时序。
func metricCollect() {
	cfg := global.Config.Metric
	if !cfg.Enable {
		return
	}
	db, _, log := repository.Get("")
	today := time.Now().Format(time.DateOnly)
	now := time.Now().Unix()
	window := "1m"
	if cfg.AggregateSec > 0 {
		window = fmt.Sprintf("%ds", cfg.AggregateSec)
	}

	nodes, _ := db.GostNode.Find()
	for _, n := range nodes {
		obs := cache2.GetNodeObs(today, n.Code)
		load := cache2.GetNodeLoad(n.Code)
		// 连接数 = 本采集周期内新建用户连接数（frps newUserConn 插件累加）。
		// 客户端上报的瞬时活跃连接数实测恒 0（frp workConn 生命周期语义），不采用。
		recordMetric(db, model.GOST_OBS_ORIGIN_KIND_NODE, n.Code, obs, float64(cache2.TakeNodeConn(n.Code)), now, window)
		// 节点在线客户端数/注册隧道数：来自节点 frps 实时上报
		recordCount(db, model.GOST_OBS_ORIGIN_KIND_NODE, n.Code, model.METRIC_CLIENT_COUNT, float64(load.ClientCount), now, window)
		recordCount(db, model.GOST_OBS_ORIGIN_KIND_NODE, n.Code, model.METRIC_TUNNEL_COUNT, float64(load.TunnelCount), now, window)
	}
	clients, _ := db.GostClient.Find()
	for _, c := range clients {
		recordMetric(db, model.GOST_OBS_ORIGIN_KIND_CLIENT, c.Code, cache2.GetClientObs(today, c.Code), -1, now, window)
		tc, _ := db.GostClientHost.Where(db.GostClientHost.ClientCode.Eq(c.Code), db.GostClientHost.Enable.Eq(1)).Count()
		fc, _ := db.GostClientForward.Where(db.GostClientForward.ClientCode.Eq(c.Code), db.GostClientForward.Enable.Eq(1)).Count()
		pc, _ := db.GostClientProxy.Where(db.GostClientProxy.ClientCode.Eq(c.Code), db.GostClientProxy.Enable.Eq(1)).Count()
		tc2, _ := db.GostClientTunnel.Where(db.GostClientTunnel.ClientCode.Eq(c.Code), db.GostClientTunnel.Enable.Eq(1)).Count()
		p2, _ := db.GostClientP2P.Where(db.GostClientP2P.ClientCode.Eq(c.Code), db.GostClientP2P.Enable.Eq(1)).Count()
		recordCount(db, model.GOST_OBS_ORIGIN_KIND_CLIENT, c.Code, model.METRIC_TUNNEL_COUNT, float64(tc+fc+pc+tc2+p2), now, window)
	}
	users, _ := db.SystemUser.Find()
	for _, u := range users {
		recordMetric(db, model.GOST_OBS_ORIGIN_KIND_USER, u.Code, cache2.GetUserObs(today, u.Code), -1, now, window)
		cc, _ := db.GostClient.Where(db.GostClient.UserCode.Eq(u.Code)).Count()
		recordCount(db, model.GOST_OBS_ORIGIN_KIND_USER, u.Code, model.METRIC_CLIENT_COUNT, float64(cc), now, window)
		ht, _ := db.GostClientHost.Where(db.GostClientHost.UserCode.Eq(u.Code), db.GostClientHost.Enable.Eq(1)).Count()
		ft, _ := db.GostClientForward.Where(db.GostClientForward.UserCode.Eq(u.Code), db.GostClientForward.Enable.Eq(1)).Count()
		pt, _ := db.GostClientProxy.Where(db.GostClientProxy.UserCode.Eq(u.Code), db.GostClientProxy.Enable.Eq(1)).Count()
		tt, _ := db.GostClientTunnel.Where(db.GostClientTunnel.UserCode.Eq(u.Code), db.GostClientTunnel.Enable.Eq(1)).Count()
		pp, _ := db.GostClientP2P.Where(db.GostClientP2P.UserCode.Eq(u.Code), db.GostClientP2P.Enable.Eq(1)).Count()
		recordCount(db, model.GOST_OBS_ORIGIN_KIND_USER, u.Code, model.METRIC_TUNNEL_COUNT, float64(ht+ft+pt+tt+pp), now, window)
	}
	for _, code := range gatherTunnelCodes(db) {
		recordMetric(db, model.GOST_OBS_ORIGIN_KIND_TUNNEL, code, cache2.GetTunnelObs(today, code), -1, now, window)
	}
	_ = log
}

func gatherTunnelCodes(db *query.Query) []string {
	var codes []string
	hosts, _ := db.GostClientHost.Where(db.GostClientHost.Enable.Eq(1)).Find()
	for _, h := range hosts {
		codes = append(codes, h.Code)
	}
	forwards, _ := db.GostClientForward.Where(db.GostClientForward.Enable.Eq(1)).Find()
	for _, f := range forwards {
		codes = append(codes, f.Code)
	}
	proxies, _ := db.GostClientProxy.Where(db.GostClientProxy.Enable.Eq(1)).Find()
	for _, p := range proxies {
		codes = append(codes, p.Code)
	}
	tunnels, _ := db.GostClientTunnel.Where(db.GostClientTunnel.Enable.Eq(1)).Find()
	for _, t := range tunnels {
		codes = append(codes, t.Code)
	}
	p2ps, _ := db.GostClientP2P.Where(db.GostClientP2P.Enable.Eq(1)).Find()
	for _, p := range p2ps {
		codes = append(codes, p.Code)
	}
	return codes
}

// recordMetric 由累计字节快照差值计算速率，刷新 gauge/曲线并写历史。connections<0 表示不采集连接数。
func recordMetric(db *query.Query, kind int, code string, obs cache2.TunnelObs, connections float64, now int64, window string) {
	prev := cache2.GetObsSnapshot(kind, code)
	var inputRate, outputRate float64
	if prev.Ts > 0 && now > prev.Ts {
		elapsed := float64(now - prev.Ts)
		if obs.InputBytes >= prev.InputBytes {
			inputRate = float64(obs.InputBytes-prev.InputBytes) / elapsed
		}
		if obs.OutputBytes >= prev.OutputBytes {
			outputRate = float64(obs.OutputBytes-prev.OutputBytes) / elapsed
		}
	}
	cache2.SetObsSnapshot(kind, code, cache2.ObsSnapshot{InputBytes: obs.InputBytes, OutputBytes: obs.OutputBytes, Ts: now})

	cache2.SetMetricGauge(kind, code, model.METRIC_INPUT_RATE, inputRate)
	cache2.SetMetricGauge(kind, code, model.METRIC_OUTPUT_RATE, outputRate)
	cache2.PushMetricPoint(kind, code, model.METRIC_INPUT_RATE, now, inputRate)
	cache2.PushMetricPoint(kind, code, model.METRIC_OUTPUT_RATE, now, outputRate)
	writeHistory(db, kind, code, model.METRIC_INPUT_RATE, inputRate, window, now)
	writeHistory(db, kind, code, model.METRIC_OUTPUT_RATE, outputRate, window, now)

	if connections >= 0 {
		cache2.SetMetricGauge(kind, code, model.METRIC_CONNECTIONS, connections)
		cache2.PushMetricPoint(kind, code, model.METRIC_CONNECTIONS, now, connections)
		writeHistory(db, kind, code, model.METRIC_CONNECTIONS, connections, window, now)
	}
}

// recordCount 刷新计数类指标（隧道数/客户端数）的 gauge、曲线与历史。
func recordCount(db *query.Query, kind int, code, metric string, value float64, now int64, window string) {
	cache2.SetMetricGauge(kind, code, metric, value)
	cache2.PushMetricPoint(kind, code, metric, now, value)
	writeHistory(db, kind, code, metric, value, window, now)
}

func writeHistory(db *query.Query, kind int, code, metric string, value float64, window string, ts int64) {
	_ = db.GostMetric.Create(&model.GostMetric{
		OriginKind: kind, OriginCode: code, Metric: metric, Value: value, Window: window, Ts: ts,
	})
}

// metricClean 每日清理超过保留期的历史指标。
func metricClean() {
	cfg := global.Config.Metric
	days := cfg.RetentionDays
	if days <= 0 {
		days = 7
	}
	db, _, log := repository.Get("")
	cutoff := time.Now().AddDate(0, 0, -days).Unix()
	if _, err := db.GostMetric.Where(db.GostMetric.Ts.Lt(cutoff)).Delete(); err != nil {
		log.Warn("清理历史指标失败", zap.Error(err))
	}
}
