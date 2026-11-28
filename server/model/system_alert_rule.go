package model

// 告警事件类型。
const (
	ALERT_EVENT_TUNNEL_DOWN   = "tunnel_down"   // 隧道转为异常
	ALERT_EVENT_TUNNEL_UP     = "tunnel_up"     // 隧道恢复
	ALERT_EVENT_NODE_OFFLINE  = "node_offline"  // 节点离线
	ALERT_EVENT_TRAFFIC_EXHAUST = "traffic_exhaust" // 流量耗尽
	ALERT_EVENT_CERT_EXPIRE   = "cert_expire"   // 证书即将到期
	ALERT_EVENT_NODE_FAILOVER = "node_failover" // 隧道故障转移到备节点
	ALERT_EVENT_FAILOVER_PORT_CONFLICT = "failover_port_conflict" // 备节点无法提供同端口，放弃转移
)

// 告警目标范围。
const (
	ALERT_TARGET_ALL    = 1 // 全部对象
	ALERT_TARGET_NODE   = 2 // 指定节点
	ALERT_TARGET_TUNNEL = 3 // 指定隧道
	ALERT_TARGET_CLIENT = 4 // 指定客户端
)

// SystemAlertRule 告警规则。Channels/Receivers 以 JSON 数组字符串存储。
type SystemAlertRule struct {
	Base
	UserCode   string `gorm:"column:user_code;size:100;index"`
	Name       string `gorm:"column:name;size:100"`
	EventType  string `gorm:"column:event_type;size:50;index"`
	TargetKind int    `gorm:"column:target_kind"`
	TargetCode string `gorm:"column:target_code;size:100"`
	Channels   string `gorm:"column:channels;size:200"`  // ["email","webhook"]
	Receivers  string `gorm:"column:receivers;size:1000"`// email 收件人 / webhook URL 列表
	SilenceSec int    `gorm:"column:silence_sec"`        // 静默期（秒），同一对象同事件在此期间不重复告警
	Enable     int    `gorm:"column:enable;index"`
}

func (r SystemAlertRule) GetChannels() []string {
	return parseStringList(r.Channels)
}

func (r SystemAlertRule) GetReceivers() []string {
	return parseStringList(r.Receivers)
}
