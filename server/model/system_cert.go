package model

// SystemCert ACME 证书（M4）。DNS-01 签发与自动续期。
type SystemCert struct {
	Base
	UserCode      string `gorm:"index"`
	Domain        string // 域名，支持泛域名 *.example.com
	Provider      string // dnspod / cloudflare / alidns
	CredentialEnc string // DNS 凭据（JSON，AES 加密存储）
	CertPath      string // 证书文件路径
	KeyPath       string // 私钥文件路径
	ExpAt         int64  // 到期时间（unix 秒）
	AutoRenew     int    // 1 自动续期 2 关闭
	Status        int    // 0 待签发 1 有效 2 失败 3 已过期
	LastError     string
	PushSummary   string // 最近一次推送网关的结果摘要
}

const (
	CERT_PROVIDER_DNSPOD     = "dnspod"
	CERT_PROVIDER_CLOUDFLARE = "cloudflare"
	CERT_PROVIDER_ALIDNS     = "alidns"
)

const (
	CERT_STATUS_PENDING = 0
	CERT_STATUS_VALID   = 1
	CERT_STATUS_FAILED  = 2
	CERT_STATUS_EXPIRED = 3
)
