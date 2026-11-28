package configs

// Failover 入口跟随（P7）行为开关。
type Failover struct {
	DNSFollow        bool `yaml:"dns-follow"`         // 转移后更新自定义域名 A 记录（需 dns 凭据）
	Failback         bool `yaml:"failback"`           // 主节点恢复后自动回切
	FailbackDelaySec int  `yaml:"failback-delay-sec"` // 主节点恢复后稳定多久才回切（默认 300）
}
