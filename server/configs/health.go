package configs

// Health 隧道健康探测配置。
type Health struct {
	Enable          bool `yaml:"enable"`           // 是否启用主动探测
	IntervalSec     int  `yaml:"interval-sec"`     // 探测周期（秒），默认 30
	Concurrency     int  `yaml:"concurrency"`      // 并发探测数，默认 16
	ProbeTimeoutSec int  `yaml:"probe-timeout-sec"`// 单次探测超时（秒），默认 5
}
