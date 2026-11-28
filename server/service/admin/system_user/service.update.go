package service

import (
	"errors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"server/pkg/utils"
	"server/repository"
	"server/repository/cache"
)

type UpdateReq struct {
	Code       string          `binding:"required" json:"code" label:"编号"`
	Account    string          `binding:"required" json:"account" label:"账号"`
	Amount     decimal.Decimal `binding:"required" json:"amount" label:"积分"`
	Password   string          `json:"password" label:"密码"`
	Admin      int             `binding:"required" json:"admin"`
	Remark     string          `json:"remark" label:"备注"`
	GroupCode  string          `json:"groupCode" label:"用户组"`
	GroupExpAt int64           `json:"groupExpAt" label:"用户组到期时间"`
}

func (service *service) Update(req UpdateReq) error {
	db, _, log := repository.Get("")
	user, _ := db.SystemUser.Where(db.SystemUser.Code.Eq(req.Code)).First()
	if user == nil {
		return errors.New("用户不存在")
	}
	user.Admin = req.Admin
	user.Account = req.Account
	user.Amount = req.Amount
	user.Remark = req.Remark
	user.GroupCode = req.GroupCode
	// 到期时间以请求字段为准：0=永久（清除回落组）；>0 保留既有回落组，到期由定时任务回落
	user.GroupExpAt = req.GroupExpAt
	if user.GroupExpAt <= 0 {
		user.GroupExpAt = 0
		user.PrevGroupCode = ""
	}
	if req.Password != "" {
		user.Salt = utils.RandStr(8, utils.AllDict)
		newHash, err := utils.HashPassword(req.Password)
		if err != nil {
			log.Error("生成密码哈希失败", zap.Error(err))
			return errors.New("操作失败")
		}
		user.Password = newHash
	}

	if err := db.SystemUser.Save(user); err != nil {
		log.Error("修改用户失败", zap.Error(err))
		return errors.New("操作失败")
	}
	cache.DelUserGroupLimit(user.Code)
	return nil
}
