package certpick

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"server/global"
	"server/model"
	"server/pkg/memory"
	"server/repository/query"
	"time"
)

func setupDB(t *testing.T) *query.Query {
	t.Helper()
	dir := t.TempDir()
	db, err := gorm.Open(sqlite.Open(filepath.Join(dir, "t.db")), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.SystemCert{}, &model.GostClientHost{}, &model.GostNode{}); err != nil {
		t.Fatal(err)
	}
	q := query.Use(db)
	query.SetDefault(db)
	global.Cache = memory.NewMemory(filepath.Join(dir, "cache.db"), time.Hour)
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return q
}

func writePEM(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCovers(t *testing.T) {
	if !covers("*.x.com", "a.x.com") || !covers("*.x.com", "a.b.x.com") {
		t.Fatal("wildcard should cover subdomains")
	}
	if covers("*.x.com", "x.com") {
		t.Fatal("wildcard should not cover apex")
	}
	if !covers("x.com", "x.com") || covers("x.com", "a.x.com") {
		t.Fatal("exact match semantics")
	}
}

func TestForDomainPriority(t *testing.T) {
	q := setupDB(t)
	dir := t.TempDir()
	refCert := writePEM(t, dir, "ref.pem", "REFCERT")
	refKey := writePEM(t, dir, "ref.key", "REFKEY")
	wildCert := writePEM(t, dir, "wild.pem", "WILDCERT")
	wildKey := writePEM(t, dir, "wild.key", "WILDKEY")

	_ = q.SystemCert.Create(&model.SystemCert{Base: model.Base{Code: "c-ref"}, UserCode: "u1", Domain: "a.example.com", Status: model.CERT_STATUS_VALID, CertPath: refCert, KeyPath: refKey})
	_ = q.SystemCert.Create(&model.SystemCert{Base: model.Base{Code: "c-wild"}, UserCode: "1", Domain: "*.example.com", Status: model.CERT_STATUS_VALID, CertPath: wildCert, KeyPath: wildKey})

	host := &model.GostClientHost{Base: model.Base{Code: "h1"}, UserCode: "u1", CertCode: "c-ref", CustomCert: "SNAP", CustomKey: "SNAPKEY"}
	// ① 引用优先
	cp, kp := ForDomain(q, host, "a.example.com")
	if cp != "REFCERT" || kp != "REFKEY" {
		t.Fatalf("certCode ref should win: %s", cp)
	}
	// ② 无引用时通配（平台）覆盖
	host.CertCode = ""
	host.CustomCert = ""
	cp, _ = ForDomain(q, host, "b.example.com")
	if cp != "WILDCERT" {
		t.Fatalf("admin wildcard should cover: %s", cp)
	}
	// ③ 存量快照回退
	host.CustomCert = "SNAP"
	host.CustomKey = "SNAPKEY"
	cp, kp = ForDomain(q, host, "other.net")
	if cp != "SNAP" {
		t.Fatalf("snapshot fallback expected: %s", cp)
	}
	_ = kp
	// ④ 无
	host.CustomCert = ""
	cp, _ = ForDomain(q, host, "other.net")
	if cp != "" {
		t.Fatalf("no cert expected: %s", cp)
	}
}
