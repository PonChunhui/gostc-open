package model

// SystemTunnelTemplate 隧道模板：保存一类隧道的创建参数（JSON），便于批量/快速创建。
type SystemTunnelTemplate struct {
	Base
	UserCode string `gorm:"column:user_code;index;comment:用户编号"`
	Name     string `gorm:"column:name;comment:名称"`
	Kind     string `gorm:"column:kind;comment:隧道类型host/forward/proxy/tunnel/p2p"`
	Content  string `gorm:"column:content;comment:创建参数JSON"`
}

const (
	TUNNEL_TEMPLATE_KIND_HOST    = "host"
	TUNNEL_TEMPLATE_KIND_FORWARD = "forward"
	TUNNEL_TEMPLATE_KIND_PROXY   = "proxy"
	TUNNEL_TEMPLATE_KIND_TUNNEL  = "tunnel"
	TUNNEL_TEMPLATE_KIND_P2P     = "p2p"
)
