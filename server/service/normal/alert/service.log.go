package service

import (
	"server/pkg/bean"
	"server/pkg/jwt"
	"server/repository"
	"time"
)

type LogPageReq struct {
	bean.PageParam
	EventType string `json:"eventType"`
	Result    int    `json:"result"`
}

type LogItem struct {
	Code       string `json:"code"`
	RuleCode   string `json:"ruleCode"`
	EventType  string `json:"eventType"`
	ObjectKind int    `json:"objectKind"`
	ObjectCode string `json:"objectCode"`
	Channel    string `json:"channel"`
	Result     int    `json:"result"`
	Message    string `json:"message"`
	CreatedAt  string `json:"createdAt"`
}

func (s *service) PageLog(claims jwt.Claims, req LogPageReq) (list []LogItem, total int64) {
	db, _, _ := repository.Get("")
	q := db.SystemAlertLog.Where(db.SystemAlertLog.UserCode.Eq(claims.Code)).Order(db.SystemAlertLog.Id.Desc())
	if req.EventType != "" {
		q = q.Where(db.SystemAlertLog.EventType.Eq(req.EventType))
	}
	if req.Result > 0 {
		q = q.Where(db.SystemAlertLog.Result.Eq(req.Result))
	}
	logs, total, _ := q.FindByPage(req.GetOffset(), req.GetLimit())
	for _, l := range logs {
		list = append(list, LogItem{
			Code: l.Code, RuleCode: l.RuleCode, EventType: l.EventType, ObjectKind: l.ObjectKind,
			ObjectCode: l.ObjectCode, Channel: l.Channel, Result: l.Result, Message: l.Message,
			CreatedAt: l.CreatedAt.Format(time.DateTime),
		})
	}
	return list, total
}
