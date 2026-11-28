package model

// 隧道类型（健康探测用，与各类隧道一一对应）。
const (
	TUNNEL_HEALTH_TYPE_HOST    = 1 // 域名解析
	TUNNEL_HEALTH_TYPE_FORWARD = 2 // 端口转发
	TUNNEL_HEALTH_TYPE_PROXY   = 3 // 私有代理
	TUNNEL_HEALTH_TYPE_TUNNEL  = 4 // 私有隧道
	TUNNEL_HEALTH_TYPE_P2P     = 5 // P2P
)

// 健康状态。
const (
	TUNNEL_HEALTH_UNKNOWN  = 0 // 未知
	TUNNEL_HEALTH_HEALTHY  = 1 // 健康
	TUNNEL_HEALTH_ABNORMAL = 2 // 异常
)

// GostTunnelHealth 集中记录每条隧道的健康三态，独立于隧道主表，避免探测写入与业务写入争用同一行。
type GostTunnelHealth struct {
	Base
	TunnelCode       string `gorm:"column:tunnel_code;size:100;uniqueIndex"`
	TunnelType       int    `gorm:"column:tunnel_type;index"`
	UserCode         string `gorm:"column:user_code;size:100;index"`
	ClientCode       string `gorm:"column:client_code;size:100;index"`
	NodeCode         string `gorm:"column:node_code;size:100;index"`
	ClientOnline     int    `gorm:"column:client_online"`     // 1 在线 2 离线
	TunnelRegistered int    `gorm:"column:tunnel_registered"` // 1 已注册 2 未注册
	BackendReachable int    `gorm:"column:backend_reachable"` // 1 可达 2 不可达 0 未探测
	HealthStatus     int    `gorm:"column:health_status;index"`
	LastCheckAt      int64  `gorm:"column:last_check_at"`
	LastError        string `gorm:"column:last_error;size:500"`
	ProbeDisable     int    `gorm:"column:probe_disable"` // 1 停用主动探测
}
