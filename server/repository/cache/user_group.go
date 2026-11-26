package cache

import (
	"fmt"
	"server/global"
	"time"
)

const user_group_limit_key = "user:group:limit:%s"

// UserGroupLimit 用户所属组的关键限额，供 FRP 连接钩子高频读取。
type UserGroupLimit struct {
	MonthTraffic int // 每月流量上限GB，0=无限制
	TunnelLimit  int // 隧道数量上限，0=无限制
}

func SetUserGroupLimit(userCode string, limit UserGroupLimit) {
	global.Cache.SetStruct(fmt.Sprintf(user_group_limit_key, userCode), limit, time.Second*60)
}

// GetUserGroupLimit 返回缓存的组限额；found=false 表示未命中(需回源加载)。
func GetUserGroupLimit(userCode string) (result UserGroupLimit, found bool) {
	err := global.Cache.GetStruct(fmt.Sprintf(user_group_limit_key, userCode), &result)
	return result, err == nil
}

func DelUserGroupLimit(userCode string) {
	global.Cache.Del(fmt.Sprintf(user_group_limit_key, userCode))
}
