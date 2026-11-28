package service

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"server/model"
	"server/repository"
	"server/repository/query"
)

// SettleOrder 结算订单（回调/查单共用）：验金额、幂等置已支付、发放积分或套餐、记账单。
// 通过「待支付→已支付」条件更新保证重复回调与并发只结算一次。
func SettleOrder(outTradeNo, tradeNo string, amount decimal.Decimal) error {
	db, _, log := repository.Get("")
	return db.Transaction(func(tx *query.Query) error {
		order, err := tx.SystemOrder.Where(tx.SystemOrder.Code.Eq(outTradeNo)).First()
		if err != nil {
			return errors.New("订单不存在")
		}
		if order.Status == model.ORDER_STATUS_PAID {
			return nil // 幂等：已结算
		}
		if !order.Amount.Equal(amount) {
			return errors.New("回调金额与订单不一致")
		}
		info, err := tx.SystemOrder.Where(
			tx.SystemOrder.Code.Eq(outTradeNo),
			tx.SystemOrder.Status.Eq(model.ORDER_STATUS_PENDING),
		).UpdateSimple(
			tx.SystemOrder.Status.Value(model.ORDER_STATUS_PAID),
			tx.SystemOrder.TradeNo.Value(tradeNo),
			tx.SystemOrder.PayAt.Value(time.Now().Unix()),
		)
		if err != nil {
			log.Error("结算订单失败", zap.Error(err))
			return errors.New("结算失败")
		}
		if info.RowsAffected == 0 {
			return nil // 并发已被结算
		}

		user, err := tx.SystemUser.Where(tx.SystemUser.Code.Eq(order.UserCode)).First()
		if err != nil {
			return errors.New("用户不存在")
		}
		switch order.Kind {
		case model.ORDER_KIND_GROUP:
			if user.GroupCode != "" {
				user.PrevGroupCode = user.GroupCode
			}
			user.GroupCode = order.TargetCode
			user.GroupExpAt = 0 // 框架版套餐永久有效，真实渠道接入后按套餐周期
		default: // recharge 充值积分
			user.Amount = user.Amount.Add(order.Amount)
			bill := &model.SystemBill{
				UserCode:     user.Code,
				OrderCode:    order.Code,
				Kind:         model.BILL_KIND_RECHARGE,
				Amount:       order.Amount,
				BalanceAfter: user.Amount,
				Remark:       order.Subject,
			}
			if err := tx.SystemBill.Create(bill); err != nil {
				log.Error("写入账单失败", zap.Error(err))
				return errors.New("记账失败")
			}
		}
		if err := tx.SystemUser.Save(user); err != nil {
			log.Error("更新用户余额失败", zap.Error(err))
			return errors.New("入账失败")
		}
		return nil
	})
}
