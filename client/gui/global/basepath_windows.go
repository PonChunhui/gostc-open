//go:build windows

package global

import (
	"os"
	"path/filepath"
)

func defaultBasePath() string {
	exe, _ := os.Executable()
	return filepath.Dir(exe)
}
