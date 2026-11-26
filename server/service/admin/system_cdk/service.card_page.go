package service

import (
	"gorm.io/gen"
	"server/pkg/bean"
	"server/repository"
	"time"
)

type CardPageReq struct {
	bean.PageParam
	BatchCode string `json:"batchCode"`
	Status    int    `json:"status"`
	Secret    string `json:"secret"`
	Account   string `json:"account"`
}

type CardItem struct {
	Code      string `json:"code"`
	BatchCode string `json:"batchCode"`
	BatchName string `json:"batchName"`
	Secret    string `json:"secret"`
	Status    int    `json:"status"`
	Account   string `json:"account"`
	UsedAt    string `json:"usedAt"`
	CreatedAt string `json:"createdAt"`
}

func (service *service) CardPage(req CardPageReq) (list []CardItem, total int64) {
	db, _, _ := repository.Get("")
	var where []gen.Condition
	if req.BatchCode != "" {
		where = append(where, db.SystemCdk.BatchCode.Eq(req.BatchCode))
	}
	if req.Status > 0 {
		where = append(where, db.SystemCdk.Status.Eq(req.Status))
	}
	if req.Secret != "" {
		where = append(where, db.SystemCdk.Secret.Like("%"+req.Secret+"%"))
	}
	if req.Account != "" {
		where = append(where, db.SystemCdk.Account.Like("%"+req.Account+"%"))
	}
	cards, total, _ := db.SystemCdk.Where(where...).Order(db.SystemCdk.Id.Desc()).FindByPage(req.GetOffset(), req.GetLimit())

	batchNames := make(map[string]string)
	for _, card := range cards {
		if _, ok := batchNames[card.BatchCode]; !ok {
			batchNames[card.BatchCode] = ""
			batch, _ := db.SystemCdkBatch.Where(db.SystemCdkBatch.Code.Eq(card.BatchCode)).First()
			if batch != nil {
				batchNames[card.BatchCode] = batch.Name
			}
		}
		var usedAt string
		if card.UsedAt > 0 {
			usedAt = time.Unix(card.UsedAt, 0).Format(time.DateTime)
		}
		list = append(list, CardItem{
			Code:      card.Code,
			BatchCode: card.BatchCode,
			BatchName: batchNames[card.BatchCode],
			Secret:    card.Secret,
			Status:    card.Status,
			Account:   card.Account,
			UsedAt:    usedAt,
			CreatedAt: card.CreatedAt.Format(time.DateTime),
		})
	}
	return list, total
}
