package task

import (
	"server/model"
	"testing"
)

func TestDecideHealth(t *testing.T) {
	cases := []struct {
		name       string
		online     bool
		registered bool
		reachable  int
		want       int
	}{
		{"全正常", true, true, 1, model.TUNNEL_HEALTH_HEALTHY},
		{"未探测后端但在线已注册", true, true, 0, model.TUNNEL_HEALTH_HEALTHY},
		{"后端不可达", true, true, 2, model.TUNNEL_HEALTH_ABNORMAL},
		{"客户端离线", false, true, 1, model.TUNNEL_HEALTH_ABNORMAL},
		{"未注册", true, false, 1, model.TUNNEL_HEALTH_ABNORMAL},
		{"全异常", false, false, 2, model.TUNNEL_HEALTH_ABNORMAL},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := decideHealth(c.online, c.registered, c.reachable); got != c.want {
				t.Fatalf("decideHealth=%v want %v", got, c.want)
			}
		})
	}
}
