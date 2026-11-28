package model

// SystemAuditLog 管理/用户侧写操作审计日志（不记录请求体，避免泄露密码等敏感信息）。
type SystemAuditLog struct {
	Base
	UserCode string `gorm:"column:user_code;size:100;index;comment:操作者code"`
	Account  string `gorm:"column:account;size:100;comment:操作者账号"`
	Method   string `gorm:"column:method;size:10;comment:请求方法"`
	Path     string `gorm:"column:path;size:255;comment:请求路径"`
	IP       string `gorm:"column:ip;size:64;comment:来源IP"`
	Status   int    `gorm:"column:status;comment:HTTP状态码"`
	Detail   string `gorm:"column:detail;size:500;comment:备注"`
}
