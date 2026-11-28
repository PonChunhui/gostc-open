package model

import "github.com/shopspring/decimal"

const (
	ORDER_STATUS_PENDING = 0 // 待支付
	ORDER_STATUS_PAID    = 1 // 已支付
	ORDER_STATUS_CLOSED  = 2 // 已关闭
	ORDER_STATUS_FAILED  = 3 // 失败

	ORDER_KIND_RECHARGE = "recharge" // 充值积分
	ORDER_KIND_GROUP    = "group"    // 购买用户组套餐
)

// SystemOrder 支付订单。OutTradeNo 即 Code（商户订单号），TradeNo 为网关交易号。
type SystemOrder struct {
	Base
	UserCode   string          `gorm:"column:user_code;index;comment:用户编号"`
	Subject    string          `gorm:"column:subject;comment:订单标题"`
	Kind       string          `gorm:"column:kind;default:'recharge';comment:类型recharge/group"`
	TargetCode string          `gorm:"column:target_code;comment:套餐用户组编号"`
	Amount     decimal.Decimal `gorm:"column:amount;default:0;comment:金额"`
	Provider   string          `gorm:"column:provider;index;comment:支付渠道"`
	Status     int             `gorm:"column:status;default:0;index;comment:状态0待支付1已支付2关闭3失败"`
	TradeNo    string          `gorm:"column:trade_no;comment:网关交易号"`
	PayAt      int64           `gorm:"column:pay_at;default:0;comment:支付时间"`
	ExpAt      int64           `gorm:"column:exp_at;default:0;comment:订单过期时间"`
	NotifyRaw  string          `gorm:"column:notify_raw;type:text;comment:回调原始报文"`
}
