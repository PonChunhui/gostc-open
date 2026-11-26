package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// File 校验文件名并返回安装包绝对路径，防止路径穿越
func (service *service) File(name string) (string, error) {
	if name == "" {
		return "", errors.New("文件名不能为空")
	}
	if strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
		return "", errors.New("非法文件名")
	}
	clean := filepath.Base(name)
	if clean != name || clean == "." || clean == ".." {
		return "", errors.New("非法文件名")
	}
	path := filepath.Join(Dir(), clean)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "", errors.New("文件不存在")
	}
	return path, nil
}
