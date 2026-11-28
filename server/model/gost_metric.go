package model

// 指标名称。
const (
	METRIC_CONNECTIONS  = "connections"   // 当前连接数（gauge）
	METRIC_INPUT_RATE   = "input_rate"    // 上行速率 bytes/s
	METRIC_OUTPUT_RATE  = "output_rate"   // 下行速率 bytes/s
	METRIC_TUNNEL_COUNT = "tunnel_count"  // 隧道数
	METRIC_CLIENT_COUNT = "client_count"  // 在线客户端数
)

// GostMetric 聚合后的历史指标时序（按窗口落库，供历史曲线查询）。
type GostMetric struct {
	Base
	OriginKind int     `gorm:"column:origin_kind;index"`
	OriginCode string  `gorm:"column:origin_code;size:100;index"`
	Metric     string  `gorm:"column:metric;size:50;index"`
	Value      float64 `gorm:"column:value"`
	Window     string  `gorm:"column:window;size:20"` // 聚合窗口，如 1m/5m/1h
	Ts         int64   `gorm:"column:ts;index"`       // 窗口时间戳（秒）
}
