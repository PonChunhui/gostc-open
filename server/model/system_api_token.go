package model

// SystemApiToken 用户级长期 API Token（用于 OpenAPI/SDK 鉴权，可吊销）。
// 仅存储 Token 的 SHA-256 哈希与前缀，明文仅在创建时返回一次。
type SystemApiToken struct {
	Base
	UserCode   string `gorm:"column:user_code;index;comment:用户编号"`
	Name       string `gorm:"column:name;comment:名称"`
	TokenHash  string `gorm:"column:token_hash;index;comment:Token哈希"`
	TokenPrefix string `gorm:"column:token_prefix;comment:Token前缀(展示)"`
	ExpAt      int64  `gorm:"column:exp_at;default:0;comment:过期时间戳(0永不过期)"`
	LastUsedAt int64  `gorm:"column:last_used_at;default:0;comment:最近使用时间"`
}
