//go:build !windows && !darwin

package global

import (
	"os"
	"path/filepath"
)

func defaultBasePath() string {
	exe, _ := os.Executable()
	return filepath.Dir(exe)
}
