package service

import (
	"gorm.io/gen"
	"server/model"
	"server/pkg/bean"
	"server/repository"
	"time"
)

type BatchPageReq struct {
	bean.PageParam
	Name string `json:"name"`
}

type BatchItem struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Remark    string `json:"remark"`
	GroupCode string `json:"groupCode"`
	GroupName string `json:"groupName"`
	GroupDays int    `json:"groupDays"`
	Amount    string `json:"amount"`
	Total     int    `json:"total"`
	UsedCount int64  `json:"usedCount"`
	CreatedAt string `json:"createdAt"`
}

func (service *service) BatchPage(req BatchPageReq) (list []BatchItem, total int64) {
	db, _, _ := repository.Get("")
	var where []gen.Condition
	if req.Name != "" {
		where = append(where, db.SystemCdkBatch.Name.Like("%"+req.Name+"%"))
	}
	batches, total, _ := db.SystemCdkBatch.Where(where...).Order(db.SystemCdkBatch.Id.Desc()).FindByPage(req.GetOffset(), req.GetLimit())
	for _, batch := range batches {
		usedCount, _ := db.SystemCdk.Where(
			db.SystemCdk.BatchCode.Eq(batch.Code),
			db.SystemCdk.Status.Eq(model.CDK_STATUS_USED),
		).Count()
		var groupName string
		if batch.GroupCode != "" {
			group, _ := db.SystemUserGroup.Where(db.SystemUserGroup.Code.Eq(batch.GroupCode)).First()
			if group != nil {
				groupName = group.Name
			}
		}
		list = append(list, BatchItem{
			Code:      batch.Code,
			Name:      batch.Name,
			Remark:    batch.Remark,
			GroupCode: batch.GroupCode,
			GroupName: groupName,
			GroupDays: batch.GroupDays,
			Amount:    batch.Amount.String(),
			Total:     batch.Total,
			UsedCount: usedCount,
			CreatedAt: batch.CreatedAt.Format(time.DateTime),
		})
	}
	return list, total
}
