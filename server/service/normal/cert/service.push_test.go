package service

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"server/model"
	"server/repository/query"
)

func setupPushDB(t *testing.T) *query.Query {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "t.db")), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.SystemCert{}, &model.GostClientHost{}, &model.GostNode{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return query.Use(db)
}

func TestResolveCertNodeCodesPermission(t *testing.T) {
	db := setupPushDB(t)
	_ = db.GostNode.Create(&model.GostNode{Base: model.Base{Code: "n1"}, Key: "k1", Domain: "x.example.com"})
	_ = db.GostNode.Create(&model.GostNode{Base: model.Base{Code: "n2"}, Key: "k2", Domain: "y.example.com"})
	// 属主 u1 的自定义域名隧道在 n1；他人 u2 的同域名隧道在 n2（不应被 u1 推送）
	_ = db.GostClientHost.Create(&model.GostClientHost{Base: model.Base{Code: "h1"}, UserCode: "u1", NodeCode: "n1", CustomDomain: "a.example.com"})
	_ = db.GostClientHost.Create(&model.GostClientHost{Base: model.Base{Code: "h2"}, UserCode: "u2", NodeCode: "n2", CustomDomain: "a.example.com"})
	cert := &model.SystemCert{Base: model.Base{Code: "c1"}, UserCode: "u1", Domain: "a.example.com"}

	got := resolveCertNodeCodes(db, cert, false)
	if len(got) != 1 || got[0] != "n1" {
		t.Fatalf("normal user should only push own node: %v", got)
	}
	// 管理员：同域名所有节点 + 后缀匹配平台节点
	got = resolveCertNodeCodes(db, cert, true)
	set := map[string]bool{}
	for _, c := range got {
		set[c] = true
	}
	if !set["n1"] || !set["n2"] {
		t.Fatalf("admin should cover both owners' nodes: %v", got)
	}
	// 平台通配证书 → 后缀匹配节点
	wild := &model.SystemCert{Base: model.Base{Code: "c2"}, UserCode: "1", Domain: "*.x.example.com"}
	got = resolveCertNodeCodes(db, wild, true)
	if len(got) != 1 || got[0] != "n1" {
		t.Fatalf("wildcard should match node suffix: %v", got)
	}
	// 引用 CertCode 的隧道也纳入
	_ = db.GostClientHost.Create(&model.GostClientHost{Base: model.Base{Code: "h3"}, UserCode: "u1", NodeCode: "n2", CertCode: "c1"})
	got = resolveCertNodeCodes(db, cert, false)
	set = map[string]bool{}
	for _, c := range got {
		set[c] = true
	}
	if !set["n1"] || !set["n2"] {
		t.Fatalf("certCode-referencing own tunnel should be included: %v", got)
	}
}
