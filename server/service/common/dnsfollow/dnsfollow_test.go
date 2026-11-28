package dnsfollow

import (
	"path/filepath"
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	_ "github.com/ncruces/go-sqlite3/embed"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"server/configs"
	"server/global"
	"server/model"
	"server/repository/query"
)

func TestZoneOf(t *testing.T) {
	cases := []struct{ domain, configured, want string }{
		{"a.b.example.com", "", "example.com"},
		{"example.com", "", "example.com"},
		{"a.example.com", "custom.org", "custom.org"},
		{"x.cn", "", "x.cn"},
	}
	for _, c := range cases {
		if got := ZoneOf(c.domain, c.configured); got != c.want {
			t.Fatalf("ZoneOf(%q,%q)=%q want %q", c.domain, c.configured, got, c.want)
		}
	}
}

func TestSubOf(t *testing.T) {
	cases := []struct{ domain, zone, want string }{
		{"a.example.com", "example.com", "a"},
		{"a.b.example.com", "example.com", "a.b"},
		{"example.com", "example.com", "@"},
	}
	for _, c := range cases {
		if got := SubOf(c.domain, c.zone); got != c.want {
			t.Fatalf("SubOf(%q,%q)=%q want %q", c.domain, c.zone, got, c.want)
		}
	}
}

// TestResolveMultiZone 多域名/多服务商：gostc.shop 走顶层 cloudflare，linux66.cn 走 Extra dnspod；
// 配置 zone 不跨域名误用（other.example.com 回退顶层但 zone 取自身后两段）。
func TestResolveMultiZone(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(filepath.Join(t.TempDir(), "dns.db")), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	db.Exec("PRAGMA foreign_keys=OFF")
	if sqlDB, e := db.DB(); e == nil {
		defer sqlDB.Close()
	}
	if err := db.AutoMigrate(&model.SystemCert{}); err != nil {
		t.Fatal(err)
	}
	query.SetDefault(db)

	global.Config.DNS = configs.DNS{
		Provider: model.CERT_PROVIDER_CLOUDFLARE, Credential: `{"authToken":"cf"}`, TTL: 60, Zone: "gostc.shop",
		Extra: []configs.DNSZone{
			{Zone: "linux66.cn", Provider: model.CERT_PROVIDER_DNSPOD, Credential: `{"loginToken":"1,tk"}`},
		},
	}
	cases := []struct{ domain, provider, zone string }{
		{"abc.x01.gostc.shop", model.CERT_PROVIDER_CLOUDFLARE, "gostc.shop"},
		{"yng.linux66.cn", model.CERT_PROVIDER_DNSPOD, "linux66.cn"},
		{"fsg.linux66.cn", model.CERT_PROVIDER_DNSPOD, "linux66.cn"},
		{"a.other.example.com", model.CERT_PROVIDER_CLOUDFLARE, "example.com"},
	}
	for _, c := range cases {
		p, _, _, z, ok := Resolve(c.domain)
		if !ok || p != c.provider || z != c.zone {
			t.Fatalf("Resolve(%q)=(%q,%q,ok=%v) want (%q,%q)", c.domain, p, z, ok, c.provider, c.zone)
		}
	}
}
