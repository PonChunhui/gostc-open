package service

import (
	"errors"
	"server/model"
	"server/pkg/bean"
	"server/pkg/jwt"
	"server/repository"
	"time"
)

type RuleReq struct {
	Code       string   `json:"code"`
	Name       string   `json:"name" binding:"required"`
	EventType  string   `json:"eventType" binding:"required"`
	TargetKind int      `json:"targetKind"`
	TargetCode string   `json:"targetCode"`
	Channels   []string `json:"channels"`
	Receivers  []string `json:"receivers"`
	SilenceSec int      `json:"silenceSec"`
	Enable     int      `json:"enable"`
}

func validEvent(e string) bool {
	switch e {
	case model.ALERT_EVENT_TUNNEL_DOWN, model.ALERT_EVENT_TUNNEL_UP, model.ALERT_EVENT_NODE_OFFLINE,
		model.ALERT_EVENT_TRAFFIC_EXHAUST, model.ALERT_EVENT_CERT_EXPIRE, model.ALERT_EVENT_NODE_FAILOVER,
		model.ALERT_EVENT_FAILOVER_PORT_CONFLICT:
		return true
	}
	return false
}

func (s *service) CreateRule(claims jwt.Claims, req RuleReq) error {
	if !validEvent(req.EventType) {
		return errors.New("不支持的告警事件类型")
	}
	if req.TargetKind == 0 {
		req.TargetKind = model.ALERT_TARGET_ALL
	}
	if len(req.Channels) == 0 {
		return errors.New("请至少选择一个告警渠道")
	}
	db, _, _ := repository.Get("")
	return db.SystemAlertRule.Create(&model.SystemAlertRule{
		UserCode: claims.Code, Name: req.Name, EventType: req.EventType,
		TargetKind: req.TargetKind, TargetCode: req.TargetCode,
		Channels: model.MarshalStringList(req.Channels), Receivers: model.MarshalStringList(req.Receivers),
		SilenceSec: req.SilenceSec, Enable: req.Enable,
	})
}

func (s *service) UpdateRule(claims jwt.Claims, req RuleReq) error {
	if req.Code == "" {
		return errors.New("缺少规则编码")
	}
	if !validEvent(req.EventType) {
		return errors.New("不支持的告警事件类型")
	}
	db, _, _ := repository.Get("")
	rule, err := db.SystemAlertRule.Where(db.SystemAlertRule.Code.Eq(req.Code), db.SystemAlertRule.UserCode.Eq(claims.Code)).First()
	if err != nil {
		return errors.New("规则不存在")
	}
	rule.Name = req.Name
	rule.EventType = req.EventType
	rule.TargetKind = req.TargetKind
	rule.TargetCode = req.TargetCode
	rule.Channels = model.MarshalStringList(req.Channels)
	rule.Receivers = model.MarshalStringList(req.Receivers)
	rule.SilenceSec = req.SilenceSec
	rule.Enable = req.Enable
	return db.SystemAlertRule.Save(rule)
}

func (s *service) DeleteRule(claims jwt.Claims, code string) error {
	db, _, _ := repository.Get("")
	_, err := db.SystemAlertRule.Where(db.SystemAlertRule.Code.Eq(code), db.SystemAlertRule.UserCode.Eq(claims.Code)).Delete()
	return err
}

type RulePageReq struct {
	bean.PageParam
	EventType string `json:"eventType"`
	Enable    int    `json:"enable"`
}

type RuleItem struct {
	Code       string   `json:"code"`
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

func (s *service) PageRule(claims jwt.Claims, req RulePageReq) (list []RuleItem, total int64) {
	db, _, _ := repository.Get("")
	q := db.SystemAlertRule.Where(db.SystemAlertRule.UserCode.Eq(claims.Code)).Order(db.SystemAlertRule.Id.Desc())
	if req.EventType != "" {
		q = q.Where(db.SystemAlertRule.EventType.Eq(req.EventType))
	}
	if req.Enable > 0 {
		q = q.Where(db.SystemAlertRule.Enable.Eq(req.Enable))
	}
	rules, total, _ := q.FindByPage(req.GetOffset(), req.GetLimit())
	for _, r := range rules {
		list = append(list, RuleItem{
			Code: r.Code, Name: r.Name, EventType: r.EventType, TargetKind: r.TargetKind,
			TargetCode: r.TargetCode, Channels: r.GetChannels(), Receivers: r.GetReceivers(),
			SilenceSec: r.SilenceSec, Enable: r.Enable, CreatedAt: r.CreatedAt.Format(time.DateTime),
		})
	}
	return list, total
}
