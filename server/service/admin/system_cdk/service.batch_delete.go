package service

import (
	"errors"
	"go.uber.org/zap"
	"server/model"
	"server/repository"
	"server/repository/query"
)

type BatchDeleteReq struct {
	Code string `binding:"required" json:"code" label:"编号"`
}

func (service *service) BatchDelete(req BatchDeleteReq) error {
	db, _, log := repository.Get("")
	batch, _ := db.SystemCdkBatch.Where(db.SystemCdkBatch.Code.Eq(req.Code)).First()
	if batch == nil {
		return nil
	}
	usedCount, _ := db.SystemCdk.Where(
		db.SystemCdk.BatchCode.Eq(req.Code),
		db.SystemCdk.Status.Eq(model.CDK_STATUS_USED),
	).Count()
	if usedCount > 0 {
		return errors.New("该批次已有卡密被兑换，无法删除")
	}
	return db.Transaction(func(tx *query.Query) error {
		if _, err := tx.SystemCdk.Where(tx.SystemCdk.BatchCode.Eq(req.Code)).Delete(); err != nil {
			log.Error("删除批次卡密失败", zap.Error(err))
			return errors.New("操作失败")
		}
		if _, err := tx.SystemCdkBatch.Where(tx.SystemCdkBatch.Code.Eq(req.Code)).Delete(); err != nil {
			log.Error("删除卡密批次失败", zap.Error(err))
			return errors.New("操作失败")
		}
		return nil
	})
}
