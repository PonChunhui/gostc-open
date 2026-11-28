package model

import "github.com/shopspring/decimal"

const (
	BILL_KIND_RECHARGE = "recharge" // 充值入账
	BILL_KIND_CONSUME  = "consume"  // 消费扣减
	BILL_KIND_REFUND   = "refund"   // 退款
)

// SystemBill 资金账单流水，记录每次余额变动及变动后余额，便于对账。
type SystemBill struct {
	Base
	UserCode     string          `gorm:"column:user_code;index;comment:用户编号"`
	OrderCode    string          `gorm:"column:order_code;index;comment:关联订单编号"`
	Kind         string          `gorm:"column:kind;comment:类型recharge/consume/refund"`
	Amount       decimal.Decimal `gorm:"column:amount;default:0;comment:变动金额"`
	BalanceAfter decimal.Decimal `gorm:"column:balance_after;default:0;comment:变动后余额"`
	Remark       string          `gorm:"column:remark;size:500;default:'';not null;comment:备注"`
}
