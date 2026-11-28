package configs

// DNS 入口跟随（E1）的全局 DNS 凭据兜底配置；优先复用域名的 SystemCert(ACME) 凭据。
// 多域名/多服务商：用 Extra 按注册域(zone)分别配置；未命中 Extra 时回退顶层单服务商配置。
type DNS struct {
	Provider   string    `yaml:"provider"`   // dnspod / cloudflare / alidns
	Credential string    `yaml:"credential"` // JSON(acme.Credential)
	TTL        int       `yaml:"ttl"`        // A 记录 TTL(秒)，默认 60
	Zone       string    `yaml:"zone"`       // 注册域覆盖（默认取域名后两段）
	Extra      []DNSZone `yaml:"extra"`      // 按 zone 的额外服务商凭据（如 linux66.cn 与 gostc.shop 不同商）
}

// DNSZone 单个注册域的 DNS 服务商凭据。
type DNSZone struct {
	Zone       string `yaml:"zone"`       // 注册域，如 linux66.cn
	Provider   string `yaml:"provider"`   // dnspod / cloudflare / alidns
	Credential string `yaml:"credential"` // JSON(acme.Credential)
	TTL        int    `yaml:"ttl"`        // 0 表示沿用顶层 TTL
}
