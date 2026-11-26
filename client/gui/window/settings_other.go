//go:build !windows && !darwin

package window

import "log"

// ShowSettings 其他平台暂未实现设置窗口。
func ShowSettings() {
	log.Println("当前平台暂不支持监听设置窗口，请手动修改 data/config.yaml 后重启程序")
}

// CloseSettings 其他平台无设置窗口。
func CloseSettings() {}
