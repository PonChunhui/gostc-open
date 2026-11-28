package configs

// Encrypt 敏感凭据加密配置。Key 为空时关闭加密（兼容旧明文部署）；
// 配置后新写入的凭据加密存储，并对存量明文做幂等迁移。
type Encrypt struct {
	Key string `yaml:"key"`
}
