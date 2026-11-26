package model

import "github.com/shopspring/decimal"

const (
	CDK_STATUS_UNUSED = 1 // 未使用
	CDK_STATUS_USED   = 2 // 已兑换
	CDK_STATUS_BANNED = 3 // 已停用
)

// SystemCdkBatch 卡密批次：定义一批卡密兑换后发放的内容（用户组/积分，可组合）。
// GroupCode 为空表示不发放用户组；GroupDays=0 表示永久有效；Amount=0 表示不发放积分。
type SystemCdkBatch struct {
	Base
	Name      string          `gorm:"column:name;size:100;comment:批次名称"`
	Remark    string          `gorm:"column:remark;size:500;default:'';not null;comment:备注"`
	GroupCode string          `gorm:"column:group_code;size:100;default:'';not null;comment:发放用户组编号,空=不发"`
	GroupDays int             `gorm:"column:group_days;default:0;comment:用户组有效天数,0=永久"`
	Amount    decimal.Decimal `gorm:"column:amount;default:0;comment:发放积分"`
	Total     int             `gorm:"column:total;default:0;comment:生成数量"`
}

// SystemCdk 卡密：一卡一密，兑换后记录使用人与时间。
type SystemCdk struct {
	Base
	BatchCode string `gorm:"column:batch_code;size:100;index:idx_system_cdk_batchcode;comment:批次编号"`
	Secret    string `gorm:"column:secret;size:64;uniqueIndex;comment:卡密"`
	Status    int    `gorm:"column:status;size:1;default:1;comment:状态"`
	UserCode  string `gorm:"column:user_code;size:100;index;default:'';not null;comment:兑换用户编号"`
	Account   string `gorm:"column:account;size:100;default:'';not null;comment:兑换账号"`
	UsedAt    int64  `gorm:"column:used_at;default:0;comment:兑换时间戳"`
}
