package service

import (
	"errors"
	"server/pkg/bean"
	"server/repository"
	"server/repository/query"
	"time"
)

// RulePageReq 管理员全站告警规则列表（不限属主）。
type RulePageReq struct {
	bean.PageParam
	UserCode  string `json:"userCode"`
	EventType string `json:"eventType"`
	Enable    int    `json:"enable"`
}

type RuleItem struct {
	Code       string   `json:"code"`
	UserCode   string   `json:"userCode"`
	Account    string   `json:"account"`
	Name       string   `json:"name"`
	EventType  string   `json:"eventType"`
	TargetKind int      `json:"targetKind"`
	TargetCode string   `json:"targetCode"`
	Channels   []string `json:"channels"`
	Receivers  []string `json:"receivers"`
	SilenceSec int      `json:"silenceSec"`
	Enable     int      `json:"enable"`
	CreatedAt  string   `json:"createdAt"`
}

func (s *service) PageRule(req RulePageReq) (list []RuleItem, total int64) {
	db, _, _ := repository.Get("")
	q := db.SystemAlertRule.Where().Order(db.SystemAlertRule.Id.Desc())
	if req.UserCode != "" {
		q = q.Where(db.SystemAlertRule.UserCode.Eq(req.UserCode))
	}
	if req.EventType != "" {
		q = q.Where(db.SystemAlertRule.EventType.Eq(req.EventType))
	}
	if req.Enable > 0 {
		q = q.Where(db.SystemAlertRule.Enable.Eq(req.Enable))
	}
	rules, total, _ := q.FindByPage(req.GetOffset(), req.GetLimit())
	codes := make([]string, 0, len(rules))
	for _, r := range rules {
		codes = append(codes, r.UserCode)
	}
	accounts := accountsByCodes(db, codes)
	for _, r := range rules {
		list = append(list, RuleItem{
			Code: r.Code, UserCode: r.UserCode, Account: accounts[r.UserCode],
			Name: r.Name, EventType: r.EventType, TargetKind: r.TargetKind,
			TargetCode: r.TargetCode, Channels: r.GetChannels(), Receivers: r.GetReceivers(),
			SilenceSec: r.SilenceSec, Enable: r.Enable, CreatedAt: r.CreatedAt.Format(time.DateTime),
		})
	}
	return list, total
}

// DeleteRule 管理员可删除任意用户的告警规则。
func (s *service) DeleteRule(code string) error {
	if code == "" {
		return errors.New("缺少规则编码")
	}
	db, _, _ := repository.Get("")
	_, err := db.SystemAlertRule.Where(db.SystemAlertRule.Code.Eq(code)).Delete()
	return err
}

// LogPageReq 管理员全站告警发送记录。
type LogPageReq struct {
	bean.PageParam
	UserCode  string `json:"userCode"`
	EventType string `json:"eventType"`
	Result    int    `json:"result"`
}

type LogItem struct {
	Code       string `json:"code"`
	UserCode   string `json:"userCode"`
	Account    string `json:"account"`
	RuleCode   string `json:"ruleCode"`
	EventType  string `json:"eventType"`
	ObjectKind int    `json:"objectKind"`
	ObjectCode string `json:"objectCode"`
	Channel    string `json:"channel"`
	Result     int    `json:"result"`
	Message    string `json:"message"`
	CreatedAt  string `json:"createdAt"`
}

func (s *service) PageLog(req LogPageReq) (list []LogItem, total int64) {
	db, _, _ := repository.Get("")
	q := db.SystemAlertLog.Where().Order(db.SystemAlertLog.Id.Desc())
	if req.UserCode != "" {
		q = q.Where(db.SystemAlertLog.UserCode.Eq(req.UserCode))
	}
	if req.EventType != "" {
		q = q.Where(db.SystemAlertLog.EventType.Eq(req.EventType))
	}
	if req.Result > 0 {
		q = q.Where(db.SystemAlertLog.Result.Eq(req.Result))
	}
	rows, total, _ := q.FindByPage(req.GetOffset(), req.GetLimit())
	codes := make([]string, 0, len(rows))
	for _, r := range rows {
		codes = append(codes, r.UserCode)
	}
	accounts := accountsByCodes(db, codes)
	for _, r := range rows {
		list = append(list, LogItem{
			Code: r.Code, UserCode: r.UserCode, Account: accounts[r.UserCode],
			RuleCode: r.RuleCode, EventType: r.EventType, ObjectKind: r.ObjectKind,
			ObjectCode: r.ObjectCode, Channel: r.Channel, Result: r.Result,
			Message: r.Message, CreatedAt: r.CreatedAt.Format(time.DateTime),
		})
	}
	return list, total
}

func accountsByCodes(db *query.Query, codes []string) map[string]string {
	m := map[string]string{}
	if len(codes) == 0 {
		return m
	}
	users, _ := db.SystemUser.Where(db.SystemUser.Code.In(codes...)).Find()
	for _, u := range users {
		m[u.Code] = u.Account
	}
	return m
}
