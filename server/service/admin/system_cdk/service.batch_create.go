package service

import (
	"errors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"server/model"
	"server/pkg/utils"
	"server/repository"
	"server/repository/query"
)

const cdkSecretDict = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type BatchCreateReq struct {
	Name      string          `binding:"required" json:"name" label:"批次名称"`
	Remark    string          `json:"remark" label:"备注"`
	GroupCode string          `json:"groupCode" label:"发放用户组"`
	GroupDays int             `json:"groupDays" label:"用户组有效天数"`
	Amount    decimal.Decimal `json:"amount" label:"发放积分"`
	Count     int             `binding:"required" json:"count" label:"生成数量"`
}

func (service *service) BatchCreate(req BatchCreateReq) error {
	if req.Count < 1 || req.Count > 1000 {
		return errors.New("生成数量需在1-1000之间")
	}
	if req.GroupDays < 0 {
		return errors.New("有效天数不能为负数")
	}
	if req.Amount.IsNegative() {
		return errors.New("发放积分不能为负数")
	}
	if req.GroupCode == "" && req.Amount.IsZero() {
		return errors.New("请至少配置发放用户组或发放积分")
	}

	db, _, log := repository.Get("")
	if req.GroupCode != "" {
		group, _ := db.SystemUserGroup.Where(db.SystemUserGroup.Code.Eq(req.GroupCode)).First()
		if group == nil {
			return errors.New("发放用户组不存在")
		}
	}

	return db.Transaction(func(tx *query.Query) error {
		batch := &model.SystemCdkBatch{
			Name:      req.Name,
			Remark:    req.Remark,
			GroupCode: req.GroupCode,
			GroupDays: req.GroupDays,
			Amount:    req.Amount,
			Total:     req.Count,
		}
		if err := tx.SystemCdkBatch.Create(batch); err != nil {
			log.Error("创建卡密批次失败", zap.Error(err))
			return errors.New("操作失败")
		}

		secrets := make(map[string]bool)
		cards := make([]*model.SystemCdk, 0, req.Count)
		for i := 0; i < req.Count; i++ {
			var secret string
			for retry := 0; retry < 5; retry++ {
				secret = utils.RandStr(20, cdkSecretDict)
				if secrets[secret] {
					secret = ""
					continue
				}
				exist, _ := tx.SystemCdk.Where(tx.SystemCdk.Secret.Eq(secret)).First()
				if exist == nil {
					break
				}
				secret = ""
			}
			if secret == "" {
				return errors.New("卡密生成冲突，请重试")
			}
			secrets[secret] = true
			cards = append(cards, &model.SystemCdk{
				BatchCode: batch.Code,
				Secret:    secret,
				Status:    model.CDK_STATUS_UNUSED,
			})
		}
		if err := tx.SystemCdk.CreateInBatches(cards, 100); err != nil {
			log.Error("生成卡密失败", zap.Error(err))
			return errors.New("操作失败")
		}
		return nil
	})
}
