package service

import (
	"errors"
	"go.uber.org/zap"
	"server/model"
	"server/pkg/jwt"
	"server/repository"
	"server/repository/cache"
	"server/repository/query"
	"strings"
	"time"
)

type CdkRedeemReq struct {
	Secret string `binding:"required" json:"secret" label:"卡密"`
}

type CdkRedeemResp struct {
	Amount     string `json:"amount"`
	GroupName  string `json:"groupName"`
	GroupDays  int    `json:"groupDays"`
	GroupExpAt string `json:"groupExpAt"`
}

// applyGroupGrant 将批次发放的用户组应用到用户（纯函数，便于单测）：
//  1. 已是目标组且永久(GroupExpAt==0) → 保持不变；
//  2. 永久卡(days==0) → 升级为目标组，永久有效；
//  3. 已是目标组且未到期 → 在原到期时间上续期；
//  4. 其他 → 记录原组用于到期回落，升级为目标组并设置到期时间。
func applyGroupGrant(user *model.SystemUser, groupCode string, days int, now time.Time) {
	if user.GroupCode == groupCode && user.GroupExpAt == 0 {
		return
	}
	if days == 0 {
		user.GroupCode = groupCode
		user.GroupExpAt = 0
		user.PrevGroupCode = ""
		return
	}
	if user.GroupCode == groupCode && user.GroupExpAt > now.Unix() {
		user.GroupExpAt += int64(days) * 86400
		return
	}
	if user.GroupCode == groupCode {
		user.PrevGroupCode = ""
	} else {
		user.PrevGroupCode = user.GroupCode
	}
	user.GroupCode = groupCode
	user.GroupExpAt = now.Unix() + int64(days)*86400
}

func (service *service) CdkRedeem(claims jwt.Claims, req CdkRedeemReq) (result CdkRedeemResp, err error) {
	secret := strings.ToUpper(strings.TrimSpace(req.Secret))
	if secret == "" {
		return result, errors.New("请输入卡密")
	}

	db, _, log := repository.Get("")
	var groupName string
	err = db.Transaction(func(tx *query.Query) error {
		card, _ := tx.SystemCdk.Where(tx.SystemCdk.Secret.Eq(secret)).First()
		if card == nil {
			return errors.New("卡密不存在")
		}
		if card.Status != model.CDK_STATUS_UNUSED {
			return errors.New("卡密已被使用或已停用")
		}
		batch, _ := tx.SystemCdkBatch.Where(tx.SystemCdkBatch.Code.Eq(card.BatchCode)).First()
		if batch == nil {
			return errors.New("卡密批次异常")
		}
		user, err := tx.SystemUser.Where(tx.SystemUser.Code.Eq(claims.Code)).First()
		if err != nil {
			return errors.New("账号错误")
		}

		now := time.Now()
		if batch.Amount.IsPositive() {
			user.Amount = user.Amount.Add(batch.Amount)
		}
		if batch.GroupCode != "" {
			applyGroupGrant(user, batch.GroupCode, batch.GroupDays, now)
		}

		// 条件更新兜底并发：仅未使用状态可被置为已兑换
		info, err := tx.SystemCdk.Where(
			tx.SystemCdk.Code.Eq(card.Code),
			tx.SystemCdk.Status.Eq(model.CDK_STATUS_UNUSED),
		).UpdateSimple(
			tx.SystemCdk.Status.Value(model.CDK_STATUS_USED),
			tx.SystemCdk.UserCode.Value(user.Code),
			tx.SystemCdk.Account.Value(user.Account),
			tx.SystemCdk.UsedAt.Value(now.Unix()),
		)
		if err != nil {
			log.Error("兑换卡密失败", zap.Error(err))
			return errors.New("兑换失败")
		}
		if info.RowsAffected == 0 {
			return errors.New("卡密已被使用或已停用")
		}
		if err := tx.SystemUser.Save(user); err != nil {
			log.Error("兑换卡密失败", zap.Error(err))
			return errors.New("兑换失败")
		}

		result.Amount = batch.Amount.String()
		result.GroupDays = batch.GroupDays
		if user.GroupExpAt > 0 {
			result.GroupExpAt = time.Unix(user.GroupExpAt, 0).Format(time.DateTime)
		}
		if batch.GroupCode != "" {
			group, _ := tx.SystemUserGroup.Where(tx.SystemUserGroup.Code.Eq(batch.GroupCode)).First()
			if group != nil {
				groupName = group.Name
			}
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	result.GroupName = groupName
	cache.DelUserGroupLimit(claims.Code)
	return result, nil
}
