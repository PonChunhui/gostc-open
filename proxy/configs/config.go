package configs

type Config struct {
	HTTPAddr    string `yaml:"http-addr"`
	HTTPSAddr   string `yaml:"https-addr"`
	Certs       string `yaml:"certs"`
	ApiAddr     string `yaml:"api-addr"`
	AutoSetCert bool   `yaml:"auto-set-cert"` // 自动申请证书（仅 80 端口生效）
	HTTP3       bool   `yaml:"http3"`         // 启用 HTTPS over QUIC（需 UDP 放行，默认关）

	Default  DomainConfig            `yaml:"default"`   // 兜底路由：无匹配 Host 时转发到 target（target 空则关闭）
	Domains  map[string]DomainConfig `yaml:"domains"`   // 域名路由（含 *. 通配符）
	IPRoutes map[string]DomainConfig `yaml:"ip-routes"` // IP 路由：按 IP（无 Host/SNI）精确匹配
}

type DomainConfig struct {
	Target       string `yaml:"target"`
	Cert         string `yaml:"cert"`
	Key          string `yaml:"key"`
	Sni          string `yaml:"sni"`           // 回源 TLS ServerName（空则用 target 主机名）
	Origin       string `yaml:"origin"`        // 回源 Origin 头改写（空则透传客户端 Origin）
	ForceHttps   bool   `yaml:"force-https"`   // HTTP 请求 301 跳转到 https://host:<https-port>
	DisableHttps bool   `yaml:"disable-https"` // 仅 HTTP 服务：不加载证书、不强制跳转
}
