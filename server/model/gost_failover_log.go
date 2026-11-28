package model

// 故障转移/入口变更原因。
const (
	FAILOVER_REASON_FAILOVER      = "failover"       // 主节点离线转移
	FAILOVER_REASON_FAILBACK      = "failback"       // 主节点恢复回切
	FAILOVER_REASON_PORT_CONFLICT = "port_conflict"  // 备节点无法同端口，放弃转移
)

// GostFailoverLog 故障转移/入口变更历史（E3）。
type GostFailoverLog struct {
	Base
	TunnelCode string `gorm:"column:tunnel_code;size:100;index"`
	TunnelType int    `gorm:"column:tunnel_type"`
	UserCode   string `gorm:"column:user_code;size:100;index"`
	FromNode   string `gorm:"column:from_node;size:100"`
	ToNode     string `gorm:"column:to_node;size:100"`
	OldEntry   string `gorm:"column:old_entry;size:300"`
	NewEntry   string `gorm:"column:new_entry;size:300"`
	Reason     string `gorm:"column:reason;size:30;index"`
}
