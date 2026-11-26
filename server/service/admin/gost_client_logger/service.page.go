package service

import (
	"errors"
	"server/pkg/bean"
	"server/repository"
	"server/service/engine"
	"strings"
	"time"
)

type PageReq struct {
	bean.PageParam
	ClientCode string `json:"clientCode"`
	Type       string `json:"type"`
	Keyword    string `json:"keyword"`
}

type Item struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Message   string `json:"message"`
}

func (service *service) Page(req PageReq) (list []Item, total int64, err error) {
	db, _, _ := repository.Get("")
	client, _ := db.GostClient.Where(db.GostClient.Code.Eq(req.ClientCode)).First()
	if client == nil {
		return nil, 0, errors.New("客户端不存在")
	}
	logs, err := engine.ClientLogger(req.ClientCode)
	if err != nil {
		return nil, 0, err
	}
	// 最新日志在前，按类型/关键词过滤
	var filtered []engine.ClientLogEntry
	for i := len(logs) - 1; i >= 0; i-- {
		entry := logs[i]
		if req.Type != "" && !strings.EqualFold(entry.Type, req.Type) {
			continue
		}
		if req.Keyword != "" && !strings.Contains(entry.Message, req.Keyword) {
			continue
		}
		filtered = append(filtered, entry)
	}
	total = int64(len(filtered))
	offset, limit := req.GetOffset(), req.GetLimit()
	if offset < 0 {
		offset = 0
	}
	for i := offset; i < len(filtered) && i < offset+limit; i++ {
		entry := filtered[i]
		list = append(list, Item{
			Timestamp: entry.Timestamp.Format(time.DateTime),
			Type:      entry.Type,
			Message:   strings.TrimRight(entry.Message, "\r\n"),
		})
	}
	return list, total, nil
}
