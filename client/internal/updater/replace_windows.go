//go:build windows

package updater

import (
	"os"
	"os/exec"
	"syscall"
)

// AtomicReplace 用新二进制替换正在运行的 exe（Windows 允许重命名运行中的 exe）。
// 旧文件改名为 .old 保留以便回滚，新文件移入原路径。
func AtomicReplace(newBinary, exePath string) error {
	old := exePath + ".old"
	_ = os.Remove(old)
	if err := os.Rename(exePath, old); err != nil {
		return err
	}
	if err := moveFile(newBinary, exePath); err != nil {
		_ = os.Rename(old, exePath) // 回滚
		return err
	}
	return nil
}

// Rollback 恢复 .old 备份。
func Rollback(exePath string) error {
	old := exePath + ".old"
	if _, err := os.Stat(old); err != nil {
		return err
	}
	_ = os.Remove(exePath)
	return os.Rename(old, exePath)
}

// CleanupOld 启动时清理上次升级遗留的 .old（best effort）。
func CleanupOld(exePath string) {
	_ = os.Remove(exePath + ".old")
}

func detachCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x00000008, // DETACHED_PROCESS
	}
}
