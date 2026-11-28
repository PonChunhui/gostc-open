package configs

// ACME 证书自动签发配置（M4）。
type ACME struct {
	Email    string `yaml:"email"`     // ACME 账户邮箱
	CA       string `yaml:"ca"`        // CA 目录 URL，空=Let's Encrypt 生产；可填 staging
	Provider string `yaml:"provider"`  // 默认 DNS provider
	CertDir  string `yaml:"cert-dir"`  // 证书输出目录，空=<base>/data/certs
	ProxyURL string `yaml:"proxy-url"` // 网关 proxy 基址（如 http://127.0.0.1:8080），签发后推送；空=不推送
}
