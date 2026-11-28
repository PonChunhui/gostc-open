package task

import (
	"fmt"
	"server/bootstrap"
	"server/global"
)

func init() {
	// 注入即时健康探测触发器，供节点重连/面板重启后重建隧道注册态。
	global.HealthProbeTrigger = ScheduleHealthProbe

	bootstrap.TaskFunc = func() {
		_, _ = global.Cron.AddFunc("0 0 * * *", gostObs)
		_, _ = global.Cron.AddFunc("0 0 * * *", tunnelClean) // 每天，清理旧隧道
		_, _ = global.Cron.AddFunc("0 * * * *", cdkExpire)   // 每小时，限时用户组到期回落

		healthSec := global.Config.Health.IntervalSec
		if healthSec <= 0 {
			healthSec = 30
		}
		_, _ = global.Cron.AddFunc(fmt.Sprintf("@every %ds", healthSec), healthProbe) // 隧道健康探测
		_, _ = global.Cron.AddFunc(fmt.Sprintf("@every %ds", healthSec), failoverScan) // 节点故障转移

		metricSec := global.Config.Metric.AggregateSec
		if metricSec <= 0 {
			metricSec = 60
		}
		_, _ = global.Cron.AddFunc(fmt.Sprintf("@every %ds", metricSec), metricCollect) // 指标采集
		_, _ = global.Cron.AddFunc("0 3 * * *", metricClean)                            // 每天，清理历史指标
		_, _ = global.Cron.AddFunc("0 4 * * *", certExpireScan)                          // 每天，证书到期扫描
		_, _ = global.Cron.AddFunc("0 5 * * *", certRenew)                             // 每天，ACME 证书自动续期
		_, _ = global.Cron.AddFunc("@every 5m", p2pRelayFlush)                        // P2P 中继流量落库
	}
}
