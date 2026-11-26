package service

import (
	"errors"
	"server/model"
	"server/repository"
)

type CardExportReq struct {
	BatchCode string `binding:"required" json:"batchCode" label:"批次编号"`
}

// CardExport 导出批次内全部未使用卡密（已兑换/已停用不导出），可重复导出。
func (service *service) CardExport(req CardExportReq) ([]string, error) {
	db, _, _ := repository.Get("")
	batch, _ := db.SystemCdkBatch.Where(db.SystemCdkBatch.Code.Eq(req.BatchCode)).First()
	if batch == nil {
		return nil, errors.New("批次不存在")
	}
	cards, _ := db.SystemCdk.Where(
		db.SystemCdk.BatchCode.Eq(req.BatchCode),
		db.SystemCdk.Status.Eq(model.CDK_STATUS_UNUSED),
	).Order(db.SystemCdk.Id.Asc()).Find()
	var secrets []string
	for _, card := range cards {
		secrets = append(secrets, card.Secret)
	}
	return secrets, nil
}
