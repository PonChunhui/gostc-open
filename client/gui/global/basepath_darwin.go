//go:build darwin

package global

import (
	"os"
	"path/filepath"
)

func defaultBasePath() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, "Library", "Application Support", "GOSTC")
	}
	exe, _ := os.Executable()
	return filepath.Dir(exe)
}
