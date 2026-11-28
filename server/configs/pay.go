package configs

// Pay 支付配置。本期仅沙箱可用；真实渠道配置位预留，未配置资质则拒绝下单。
type Pay struct {
	SandboxKey string     `yaml:"sandbox-key"` // 沙箱回调签名密钥（空则用默认值）
	Alipay     AlipayPay  `yaml:"alipay"`
	Wechat     WechatPay  `yaml:"wechat"`
	Stripe     StripePay  `yaml:"stripe"`
}

type AlipayPay struct {
	AppID      string `yaml:"app-id"`
	PrivateKey string `yaml:"private-key"`
	PublicKey  string `yaml:"public-key"`
}

type WechatPay struct {
	MchID  string `yaml:"mch-id"`
	AppID  string `yaml:"app-id"`
	APIKey string `yaml:"api-key"`
}

type StripePay struct {
	SecretKey     string `yaml:"secret-key"`
	WebhookSecret string `yaml:"webhook-secret"`
}
