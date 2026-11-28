package configs

// Cache 缓存后端配置。Type=local（默认，本地文件缓存）或 redis（多实例共享）。
type Cache struct {
	Type  string `yaml:"type"`
	Redis Redis  `yaml:"redis"`
}

// Redis go-redis 连接配置。
type Redis struct {
	Addr   string `yaml:"addr"`
	Pwd    string `yaml:"pwd"`
	DB     int    `yaml:"db"`
	Prefix string `yaml:"prefix"`
}
