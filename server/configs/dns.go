package configs

// DNS 入口跟随（E1）的全局 DNS 凭据兜底配置；优先复用域名的 SystemCert(ACME) 凭据。
type DNS struct {
	Provider   string `yaml:"provider"`   // dnspod / cloudflare / alidns
	Credential string `yaml:"credential"` // JSON(acme.Credential)
	TTL        int    `yaml:"ttl"`        // A 记录 TTL(秒)，默认 60
	Zone       string `yaml:"zone"`       // 注册域覆盖（默认取域名后两段）
}
