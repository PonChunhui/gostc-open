package proxy

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestPickCertificate(t *testing.T) {
	wild := &tls.Certificate{}
	bare := &tls.Certificate{}
	other := &tls.Certificate{}
	certMap := map[string]*tls.Certificate{
		"*.ningbo-2.node.linux66.cn": wild,
		"ningbo-2.node.linux66.cn":   bare,
		"yng.linux66.cn":             other,
	}
	cases := []struct {
		sni  string
		want *tls.Certificate
	}{
		{"probe.ningbo-2.node.linux66.cn", wild},   // 子域→通配符优先于父域裸证书
		{"a.b.ningbo-2.node.linux66.cn", wild},     // 多级子域回退到通配符
		{"ningbo-2.node.linux66.cn", bare},         // 裸域精确匹配
		{"yng.linux66.cn", other},                  // 其它精确匹配
		{"probe.yng.linux66.cn", nil},              // 无通配符且父域证书不覆盖子域
		{"", nil},
		{"unknown.example.com", nil},
	}
	for _, c := range cases {
		if got := pickCertificate(certMap, c.sni); got != c.want {
			t.Errorf("sni=%q got=%p want=%p", c.sni, got, c.want)
		}
	}
}

func testRouter(withDefault bool) *HostRouter {
	r := &HostRouter{
		level1Routes: map[string]*HostRoute{
			"x01.gostc.shop": {Host: "x01.gostc.shop"},
			"114.66.40.58":   {Host: "114.66.40.58"},
		},
		level2Routes: map[string]*HostRoute{
			"x01.gostc.shop": {Host: "*.x01.gostc.shop"},
		},
	}
	if withDefault {
		r.defaultRoute = &HostRoute{Host: "default"}
	}
	return r
}

func TestMatch(t *testing.T) {
	cases := []struct {
		name        string
		host        string
		withDefault bool
		want        string
		found       bool
	}{
		{"裸域名精确匹配", "x01.gostc.shop", false, "x01.gostc.shop", true},
		{"通配符子域名", "a.x01.gostc.shop", false, "*.x01.gostc.shop", true},
		{"通配符带非标端口", "a.x01.gostc.shop:1443", false, "*.x01.gostc.shop", true},
		{"多级子域带端口", "a.b.x01.gostc.shop:8001", false, "*.x01.gostc.shop", true},
		{"裸域名带端口", "x01.gostc.shop:1443", false, "x01.gostc.shop", true},
		{"IP路由", "114.66.40.58", false, "114.66.40.58", true},
		{"IP带端口", "114.66.40.58:8001", false, "114.66.40.58", true},
		{"未知域名无兜底", "unknown.example.com", false, "", false},
		{"未知域名走兜底", "unknown.example.com", true, "default", true},
		{"无Host走兜底", "", true, "default", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := testRouter(c.withDefault)
			req := &http.Request{Host: c.host}
			route, ok := r.Match(req)
			if ok != c.found {
				t.Fatalf("found=%v want %v", ok, c.found)
			}
			if ok && route.Host != c.want {
				t.Fatalf("route=%q want %q", route.Host, c.want)
			}
		})
	}
}
