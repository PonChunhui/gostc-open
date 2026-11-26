package service

import (
	"errors"
	"go.uber.org/zap"
	"server/model"
	"server/repository"
)

type CardBanReq struct {
	Code   string `binding:"required" json:"code" label:"编号"`
	Status int    `binding:"required" json:"status" label:"状态"`
}

// CardBan 停用/启用卡密，仅允许 未使用↔已停用 互转；已兑换卡密不可操作。
func (service *service) CardBan(req CardBanReq) error {
	if req.Status != model.CDK_STATUS_UNUSED && req.Status != model.CDK_STATUS_BANNED {
		return errors.New("状态参数错误")
	}
	db, _, log := repository.Get("")
	card, _ := db.SystemCdk.Where(db.SystemCdk.Code.Eq(req.Code)).First()
	if card == nil {
		return errors.New("卡密不存在")
	}
	if card.Status == model.CDK_STATUS_USED {
		return errors.New("已兑换的卡密无法操作")
	}
	if card.Status == req.Status {
		return nil
	}
	if _, err := db.SystemCdk.Where(db.SystemCdk.Code.Eq(req.Code)).UpdateSimple(db.SystemCdk.Status.Value(req.Status)); err != nil {
		log.Error("修改卡密状态失败", zap.Error(err))
		return errors.New("操作失败")
	}
	return nil
}
