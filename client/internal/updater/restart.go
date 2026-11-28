package updater

import (
	"os"
	"os/exec"
)

// Relaunch 以分离方式重启当前可执行文件（沿用原参数），调用方随后退出进程。
func Relaunch(args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	detachCmd(cmd)
	return cmd.Start()
}
