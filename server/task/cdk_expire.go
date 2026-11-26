package task

import (
	"go.uber.org/zap"
	"server/repository"
	"server/repository/cache"
	"time"
)

// cdkExpire 卡密发放的限时用户组到期后，回落到原用户组。
func cdkExpire() {
	now := time.Now().Unix()
	db, _, log := repository.Get("")
	users, _ := db.SystemUser.Where(
		db.SystemUser.GroupExpAt.Gt(0),
		db.SystemUser.GroupExpAt.Lte(now),
	).Find()
	for _, user := range users {
		user.GroupCode = user.PrevGroupCode
		user.GroupExpAt = 0
		user.PrevGroupCode = ""
		if err := db.SystemUser.Save(user); err != nil {
			log.Error("用户组到期回落失败", zap.String("account", user.Account), zap.Error(err))
			continue
		}
		cache.DelUserGroupLimit(user.Code)
	}
}
