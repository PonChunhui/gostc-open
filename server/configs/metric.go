package configs

// Metric 实时指标与历史聚合配置。
type Metric struct {
	Enable        bool `yaml:"enable"`         // 是否启用指标采集
	AggregateSec  int  `yaml:"aggregate-sec"`  // 历史聚合周期（秒），默认 60
	RetentionDays int  `yaml:"retention-days"` // 历史保留天数，默认 7
}
