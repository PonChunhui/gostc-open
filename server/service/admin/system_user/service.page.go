package service

import (
	"github.com/shopspring/decimal"
	"gorm.io/gen"
	"server/pkg/bean"
	"server/repository"
	"server/repository/cache"
	"time"
)

type PageReq struct {
	bean.PageParam
	Account string `json:"account"`
	Admin   int    `json:"admin"`
	Email   string `json:"email"`
	Status  int    `json:"status"`
}

type Item struct {
	Code          string          `json:"code"`
	Account       string          `json:"account"`
	Admin         int             `json:"admin"`
	Amount        decimal.Decimal `json:"amount"`
	Remark        string          `json:"remark"`
	Email         string          `json:"email"`
	CreatedAt     string          `json:"createdAt"`
	InputBytes    int64           `json:"inputBytes"`
	OutputBytes   int64           `json:"outputBytes"`
	GroupCode     string          `json:"groupCode"`
	GroupName     string          `json:"groupName"`
	GroupExpAt    string          `json:"groupExpAt"`
	PrevGroupName string          `json:"prevGroupName"`
	Status        int             `json:"status"`
	BanReason     string          `json:"banReason"`
}

func (service *service) Page(req PageReq) (list []Item, total int64) {
	db, _, _ := repository.Get("")
	var where []gen.Condition
	if req.Account != "" {
		where = append(where, db.SystemUser.Account.Like("%"+req.Account+"%"))
	}
	if req.Admin > 0 {
		where = append(where, db.SystemUser.Admin.Eq(req.Admin))
	}
	if req.Status > 0 {
		where = append(where, db.SystemUser.Status.Eq(req.Status))
	}
	users, total, _ := db.SystemUser.Preload(db.SystemUser.BindEmail).Where(where...).Order(db.SystemUser.Id.Desc()).FindByPage(req.GetOffset(), req.GetLimit())
	groups, _ := db.SystemUserGroup.Find()
	var groupNames = make(map[string]string)
	for _, group := range groups {
		groupNames[group.Code] = group.Name
	}
	for _, user := range users {
		obsInfo := cache.GetUserObsDateRange(cache.MONTH_DATEONLY_LIST, user.Code)
		var groupExpAt string
		if user.GroupExpAt > 0 {
			groupExpAt = time.Unix(user.GroupExpAt, 0).Format(time.DateTime)
		}
		list = append(list, Item{
			Code:          user.Code,
			Account:       user.Account,
			Admin:         user.Admin,
			Amount:        user.Amount,
			Remark:        user.Remark,
			Email:         user.BindEmail.Email,
			CreatedAt:     user.CreatedAt.Format(time.DateTime),
			InputBytes:    obsInfo.InputBytes,
			OutputBytes:   obsInfo.OutputBytes,
			GroupCode:     user.GroupCode,
			GroupName:     groupNames[user.GroupCode],
			GroupExpAt:    groupExpAt,
			PrevGroupName: groupNames[user.PrevGroupCode],
			Status:        user.Status,
			BanReason:     user.BanReason,
		})
	}
	return list, total
}
