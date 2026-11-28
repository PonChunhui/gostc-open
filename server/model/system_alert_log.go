package model

// SystemAlertLog 告警发送记录。
type SystemAlertLog struct {
	Base
	RuleCode   string `gorm:"column:rule_code;size:100;index"`
	UserCode   string `gorm:"column:user_code;size:100;index"`
	EventType  string `gorm:"column:event_type;size:50;index"`
	ObjectKind int    `gorm:"column:object_kind"`
	ObjectCode string `gorm:"column:object_code;size:100;index"`
	Channel    string `gorm:"column:channel;size:50"`
	Result     int    `gorm:"column:result"` // 1 成功 2 失败
	Message    string `gorm:"column:message;size:1000"`
}
