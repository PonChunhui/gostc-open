//go:build !windows && !darwin

package window

import (
	"log"
	"os/exec"
)

// ShowMain 其他平台没有内嵌窗口实现，退化为用系统浏览器打开管理界面。
func ShowMain() {
	url := mainURL()
	if err := exec.Command("xdg-open", url).Start(); err != nil {
		log.Println("打开浏览器失败:", err)
	}
}

// RefreshMain 其他平台无内嵌窗口，无需刷新。
func RefreshMain() {}

// CloseMain 其他平台无内嵌窗口，无需关闭。
func CloseMain() {}
