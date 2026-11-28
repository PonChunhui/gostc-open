// Package alert 告警分发：按规则匹配事件，经邮件/Webhook 渠道发送，带静默期与发送记录。
package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"server/global"
	"server/model"
	"server/pkg/email"
	cache2 "server/repository/cache"
	"server/repository"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// Dispatch 触发一次告警事件。objectKind 用 model.ALERT_TARGET_*；userCode 为资源属主。
// 异步执行，不阻塞调用方（探测任务/插件）。
func Dispatch(eventType string, objectKind int, objectCode, userCode, title, message string) {
	DispatchPayload(eventType, objectKind, objectCode, userCode, title, message, nil)
}

// DispatchPayload 同 Dispatch，但附带结构化 payload（仅 webhook 渠道携带），用于入口跟随等场景。
func DispatchPayload(eventType string, objectKind int, objectCode, userCode, title, message string, payload map[string]any) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.Logger.Warn("告警分发异常", zap.Any("recover", r))
			}
		}()
		dispatch(eventType, objectKind, objectCode, userCode, title, message, payload)
	}()
}

func dispatch(eventType string, objectKind int, objectCode, userCode, title, message string, payload map[string]any) {
	db, _, log := repository.Get("")
	rules, err := db.SystemAlertRule.Where(
		db.SystemAlertRule.Enable.Eq(1),
		db.SystemAlertRule.EventType.Eq(eventType),
	).Find()
	if err != nil {
		log.Warn("查询告警规则失败", zap.Error(err))
		return
	}
	for _, rule := range rules {
		if !matchTarget(rule, objectKind, objectCode, userCode) {
			continue
		}
		silenceKey := fmt.Sprintf("%s:%s:%s", rule.Code, eventType, objectCode)
		if cache2.InAlertSilence(silenceKey) {
			continue
		}
		silence := time.Duration(rule.SilenceSec) * time.Second
		if silence <= 0 {
			silence = time.Minute * 5
		}
		cache2.SetAlertSilence(silenceKey, silence)
		for _, channel := range rule.GetChannels() {
			result, detail := send(channel, rule.GetReceivers(), title, message, payload)
			logEntry := &model.SystemAlertLog{
				RuleCode:   rule.Code,
				UserCode:   rule.UserCode,
				EventType:  eventType,
				ObjectKind: objectKind,
				ObjectCode: objectCode,
				Channel:    channel,
				Result:     result,
				Message:    detail,
			}
			if err := db.SystemAlertLog.Create(logEntry); err != nil {
				log.Warn("写入告警记录失败", zap.Error(err))
			}
		}
	}
}

// matchTarget 规则目标匹配：属主一致 + 范围匹配。
func matchTarget(rule *model.SystemAlertRule, objectKind int, objectCode, userCode string) bool {
	if rule.UserCode != userCode {
		return false
	}
	switch rule.TargetKind {
	case model.ALERT_TARGET_ALL:
		return true
	case model.ALERT_TARGET_NODE:
		return objectKind == model.ALERT_TARGET_NODE && rule.TargetCode == objectCode
	case model.ALERT_TARGET_TUNNEL:
		return objectKind == model.ALERT_TARGET_TUNNEL && rule.TargetCode == objectCode
	case model.ALERT_TARGET_CLIENT:
		return objectKind == model.ALERT_TARGET_CLIENT && rule.TargetCode == objectCode
	}
	return false
}

// send 经指定渠道发送，返回结果(1成功2失败)与详情。
func send(channel string, receivers []string, title, message string, payload map[string]any) (int, string) {
	if len(receivers) == 0 {
		return 2, "无接收者"
	}
	switch channel {
	case "email":
		return sendEmail(receivers, title, message)
	case "webhook":
		return sendWebhook(receivers, title, message, payload)
	}
	return 2, "未知渠道:" + channel
}

func sendEmail(receivers []string, title, message string) (int, string) {
	var cfg model.SystemConfigEmail
	cache2.GetSystemConfigEmail(&cfg)
	if cfg.Enable != "1" {
		return 2, "邮件服务未启用"
	}
	port, _ := strconv.Atoi(cfg.Port)
	err := email.Send(email.Config{Host: cfg.Host, Port: port, User: cfg.User, Pwd: cfg.Pwd}, email.Body{
		Mails:    receivers,
		NickName: cfg.NickName,
		Title:    title,
		Body:     message,
	})
	if err != nil {
		return 2, err.Error()
	}
	return 1, "ok"
}

func sendWebhook(urls []string, title, message string, extra map[string]any) (int, string) {
	body := map[string]any{
		"title":     title,
		"message":   message,
		"timestamp": time.Now().Format(time.DateTime),
	}
	for k, v := range extra {
		body[k] = v
	}
	payload, _ := json.Marshal(body)
	client := &http.Client{Timeout: time.Second * 10}
	var lastErr string
	okCount := 0
	for _, u := range urls {
		resp, err := client.Post(u, "application/json", bytes.NewReader(payload))
		if err != nil {
			lastErr = err.Error()
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			okCount++
		} else {
			lastErr = "status " + resp.Status
		}
	}
	if okCount == len(urls) {
		return 1, "ok"
	}
	return 2, fmt.Sprintf("成功%d/%d %s", okCount, len(urls), lastErr)
}
