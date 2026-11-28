package configs

// DBResolver 数据库读写分离配置（仅 mysql 生效）。
// 写流量走主库（默认连接），读流量随机分发到 Replicas 从库。
type DBResolver struct {
	Enable   bool     `yaml:"enable"`
	Replicas []string `yaml:"replicas"` // 从库完整 DSN 列表
}
