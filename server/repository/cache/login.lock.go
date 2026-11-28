package cache

import (
	"server/global"
	"strconv"
	"time"
)

const (
	login_fail_key = "login:fail:"
	login_lock_key = "login:lock:"

	// MaxLoginFails 连续失败达到该次数触发账号锁定。
	MaxLoginFails = 5
	// LoginLockDuration 锁定时长。
	LoginLockDuration = time.Minute * 15
)

// LoginLocked 账号是否处于锁定状态。
func LoginLocked(account string) bool {
	return global.Cache.GetString(login_lock_key+account) == "1"
}

// LoginFail 记录一次失败；达到阈值则锁定账号并清零计数。
func LoginFail(account string) {
	n, _ := strconv.Atoi(global.Cache.GetString(login_fail_key + account))
	n++
	if n >= MaxLoginFails {
		global.Cache.SetString(login_lock_key+account, "1", LoginLockDuration)
		global.Cache.Del(login_fail_key + account)
		return
	}
	global.Cache.SetString(login_fail_key+account, strconv.Itoa(n), LoginLockDuration)
}

// LoginSuccess 登录成功，清除失败计数与锁定。
func LoginSuccess(account string) {
	global.Cache.Del(login_fail_key + account)
	global.Cache.Del(login_lock_key + account)
}
