//go:build !windows

package updater

import (
	"os"
	"os/exec"
	"syscall"
)

// AtomicReplace 用新二进制替换 exe（Unix 允许覆盖运行中的可执行文件）。
// 先备份 .old 以便回滚，再 rename 覆盖并赋可执行权限。
func AtomicReplace(newBinary, exePath string) error {
	old := exePath + ".old"
	_ = os.Remove(old)
	_ = copyFile(exePath, old) // best effort 备份
	if err := moveFile(newBinary, exePath); err != nil {
		_ = copyFile(old, exePath) // 回滚
		return err
	}
	return os.Chmod(exePath, 0755)
}

// Rollback 恢复 .old 备份。
func Rollback(exePath string) error {
	old := exePath + ".old"
	if _, err := os.Stat(old); err != nil {
		return err
	}
	if err := copyFile(old, exePath); err != nil {
		return err
	}
	return os.Chmod(exePath, 0755)
}

// CleanupOld 启动时清理 .old（best effort）。
func CleanupOld(exePath string) {
	_ = os.Remove(exePath + ".old")
}

func detachCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
