package service

import (
	"os"
	"path/filepath"
	"server/global"
	"sort"
	"time"
)

type ListItem struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	UpdatedAt string `json:"updatedAt"`
}

// Dir 客户端安装包存放目录（程序所在目录下的 downloads）
func Dir() string {
	return filepath.Join(global.BASE_PATH, "downloads")
}

func (service *service) List() []ListItem {
	list := make([]ListItem, 0)
	entries, err := os.ReadDir(Dir())
	if err != nil {
		return list
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		list = append(list, ListItem{
			Name:      entry.Name(),
			Size:      info.Size(),
			UpdatedAt: info.ModTime().Format(time.DateTime),
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})
	return list
}
