package task

import (
	"server/bootstrap"
	"server/global"
)

func init() {
	bootstrap.TaskFunc = func() {
		_, _ = global.Cron.AddFunc("0 0 * * *", gostObs)
		_, _ = global.Cron.AddFunc("0 0 * * *", tunnelClean) // 每天，清理旧隧道
		_, _ = global.Cron.AddFunc("0 * * * *", cdkExpire)   // 每小时，限时用户组到期回落
	}
}
