package alert

import (
	"server/model"
	"testing"
)

func TestMatchTarget(t *testing.T) {
	rule := func(user string, kind int, code string) *model.SystemAlertRule {
		return &model.SystemAlertRule{UserCode: user, TargetKind: kind, TargetCode: code}
	}
	cases := []struct {
		name       string
		rule       *model.SystemAlertRule
		objectKind int
		objectCode string
		userCode   string
		want       bool
	}{
		{"属主不一致拒绝", rule("u1", model.ALERT_TARGET_ALL, ""), model.ALERT_TARGET_TUNNEL, "t1", "u2", false},
		{"全部范围匹配", rule("u1", model.ALERT_TARGET_ALL, ""), model.ALERT_TARGET_TUNNEL, "t1", "u1", true},
		{"节点范围命中", rule("u1", model.ALERT_TARGET_NODE, "n1"), model.ALERT_TARGET_NODE, "n1", "u1", true},
		{"节点范围编码不符", rule("u1", model.ALERT_TARGET_NODE, "n1"), model.ALERT_TARGET_NODE, "n2", "u1", false},
		{"节点范围类型不符", rule("u1", model.ALERT_TARGET_NODE, "n1"), model.ALERT_TARGET_TUNNEL, "n1", "u1", false},
		{"隧道范围命中", rule("u1", model.ALERT_TARGET_TUNNEL, "t1"), model.ALERT_TARGET_TUNNEL, "t1", "u1", true},
		{"客户端范围命中", rule("u1", model.ALERT_TARGET_CLIENT, "c1"), model.ALERT_TARGET_CLIENT, "c1", "u1", true},
		{"未知范围拒绝", rule("u1", 99, ""), model.ALERT_TARGET_TUNNEL, "t1", "u1", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := matchTarget(c.rule, c.objectKind, c.objectCode, c.userCode); got != c.want {
				t.Fatalf("matchTarget=%v want %v", got, c.want)
			}
		})
	}
}
