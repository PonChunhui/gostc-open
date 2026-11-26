package model

import "github.com/shopspring/decimal"

const (
	SYSTEM_IS_ADMIN = 1 // 是管理员
	SYSTEM_NO_ADMIN = 2 // 不是管理员
)

type SystemUser struct {
	Base
	Account   string          `gorm:"column:account;size:100;uniqueIndex;comment:账号"`
	Password  string          `gorm:"column:password;comment:密码"`
	Salt      string          `gorm:"column:salt;size:8;comment:盐"`
	OtpKey    string          `gorm:"column:otp_key;index;default:'';not null"`
	Admin     int             `gorm:"column:admin;size:1;default:2;comment:是否为管理员"`
	Status    int             `gorm:"column:status;size:1;default:1;comment:状态"`
	BanReason string          `gorm:"column:ban_reason;size:500;default:'';not null;comment:封禁原因"`
	Amount    decimal.Decimal `gorm:"column:amount;default:0;comment:余额"`
	Remark    string          `gorm:"column:remark;size:500;default:'';not null;comment:备注"`
	GroupCode string          `gorm:"column:group_code;index:idx_system_user_grp_code;default:'';not null;comment:用户组编号"`
	// GroupExpAt 用户组到期时间戳（秒），0=永久；PrevGroupCode 为到期后回落的用户组编号。
	GroupExpAt    int64           `gorm:"column:group_exp_at;default:0;comment:用户组到期时间戳,0=永久"`
	PrevGroupCode string          `gorm:"column:prev_group_code;size:100;default:'';not null;comment:到期后回落的用户组编号"`
	BindEmail     SystemUserEmail `gorm:"foreignKey:UserCode;references:Code"`
}
