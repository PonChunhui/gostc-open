//go:build !windows && !darwin

package registry

import "errors"

// 其他平台暂未实现开机自启。

func Registered() bool { return false }

func Register() error { return errors.New("当前平台暂不支持开机自启") }

func UnRegister() {}
