package bootstrap

import (
	"proxy/configs"
	"testing"
)

func TestMergeDomainConfig(t *testing.T) {
	existing := &configs.DomainConfig{Target: "http://127.0.0.1:58431", Cert: "/old.pem", Key: "/old.key", ForceHttps: true}

	// 续期：仅 cert/key，无 target → 保留 target 与 force-https
	got := mergeDomainConfig(existing, DomainReq{Domain: "a.com", Cert: "c", Key: "k"}, "/new.pem", "/new.key")
	if got.Target != existing.Target || !got.ForceHttps || got.Cert != "/new.pem" {
		t.Fatalf("renew should preserve target/forceHttps and use new cert: %+v", got)
	}

	// 续期哨兵：带 target 但 ForceHttps=-1 → 保留 force-https
	got = mergeDomainConfig(existing, DomainReq{Domain: "a.com", Target: "http://127.0.0.1:58432", ForceHttps: -1}, "/new.pem", "/new.key")
	if got.Target != "http://127.0.0.1:58432" || !got.ForceHttps {
		t.Fatalf("sentinel -1 should preserve forceHttps: %+v", got)
	}

	// 显式 forceHttps=0 带 target → 覆盖为 false
	got = mergeDomainConfig(existing, DomainReq{Domain: "a.com", Target: "http://x", ForceHttps: 0}, "/new.pem", "/new.key")
	if got.ForceHttps {
		t.Fatalf("explicit 0 should override forceHttps: %+v", got)
	}

	// 仅证书无 certFile → 保留旧证书路径
	got = mergeDomainConfig(existing, DomainReq{Domain: "a.com"}, "", "")
	if got.Cert != "/old.pem" || got.Key != "/old.key" {
		t.Fatalf("empty cert should keep old paths: %+v", got)
	}

	// 首推（无 existing）→ 用 req 值
	got = mergeDomainConfig(nil, DomainReq{Domain: "b.com", Target: "http://y", ForceHttps: 1}, "/n.pem", "/n.key")
	if got.Target != "http://y" || !got.ForceHttps || got.Cert != "/n.pem" {
		t.Fatalf("first push should use req: %+v", got)
	}

	// 首推无 target → 空 target（创建占位路由）
	got = mergeDomainConfig(nil, DomainReq{Domain: "b.com"}, "/n.pem", "/n.key")
	if got.Target != "" {
		t.Fatalf("first push without target should be empty: %+v", got)
	}
}
