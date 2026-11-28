package entryfollow

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ncruces/go-sqlite3/gormlite"
	_ "github.com/ncruces/go-sqlite3/embed"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"server/global"
	"server/model"
	"server/pkg/memory"
	"server/repository/query"
)

func setupDB(t *testing.T) *query.Query {
	t.Helper()
	global.Cache = memory.NewMemory(filepath.Join(t.TempDir(), "cache.db"), time.Hour)
	db, err := gorm.Open(gormlite.Open(filepath.Join(t.TempDir(), "e.db")), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db.Exec("PRAGMA foreign_keys=OFF")
	if err := db.AutoMigrate(&model.GostNode{}, &model.GostClientForward{}, &model.GostClientProxy{}, &model.GostClientHost{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { _ = sqlDB.Close() })
	return query.Use(db)
}

func node(code, addr, pool, tpl string) *model.GostNode {
	return &model.GostNode{Base: model.Base{Code: code}, Key: code + "-k", Address: addr, ForwardPorts: pool, UrlTpl: tpl, Domain: "x.example.com"}
}

// TestSamePortForward 端口转发同端口硬约束：池内且空闲→通过；被占用/池外→拒绝。
func TestSamePortForward(t *testing.T) {
	q := setupDB(t)
	primary := node("P", "1.1.1.1:7000", "10000-10010", "")
	backup := node("B", "2.2.2.2:7000", "10000-10010", "")
	_ = q.GostClientForward.Create(&model.GostClientForward{Base: model.Base{Code: "F1"}, NodeCode: "P", Port: "10005", Enable: 1})

	if !CanServeSamePort(q, primary, backup, model.TUNNEL_HEALTH_TYPE_FORWARD, "F1") {
		t.Fatal("池内空闲端口应通过")
	}
	// 备节点已有同端口占用
	_ = q.GostClientForward.Create(&model.GostClientForward{Base: model.Base{Code: "F2"}, NodeCode: "B", Port: "10005", Enable: 1})
	if CanServeSamePort(q, primary, backup, model.TUNNEL_HEALTH_TYPE_FORWARD, "F1") {
		t.Fatal("备节点端口被占用应拒绝")
	}
	// 池外端口
	_ = q.GostClientForward.Create(&model.GostClientForward{Base: model.Base{Code: "F3"}, NodeCode: "P", Port: "20000", Enable: 1})
	if CanServeSamePort(q, primary, backup, model.TUNNEL_HEALTH_TYPE_FORWARD, "F3") {
		t.Fatal("池外端口应拒绝")
	}
}

// TestSamePortHostAndTunnel host 要求 UrlTpl 一致；tunnel/p2p 要求 bind 端口一致。
func TestSamePortHostAndTunnel(t *testing.T) {
	q := setupDB(t)
	p := node("P", "1.1.1.1:7000", "", "https://{{DOMAIN}}:1443")
	bSame := node("B", "2.2.2.2:7000", "", "https://{{DOMAIN}}:1443")
	bDiff := node("C", "3.3.3.3:7001", "", "https://{{DOMAIN}}:443")
	if !CanServeSamePort(q, p, bSame, model.TUNNEL_HEALTH_TYPE_HOST, "H") {
		t.Fatal("UrlTpl 一致应通过")
	}
	if CanServeSamePort(q, p, bDiff, model.TUNNEL_HEALTH_TYPE_HOST, "H") {
		t.Fatal("UrlTpl 不同应拒绝")
	}
	if !CanServeSamePort(q, p, bSame, model.TUNNEL_HEALTH_TYPE_TUNNEL, "T") {
		t.Fatal("bind 端口一致应通过")
	}
	if CanServeSamePort(q, p, bDiff, model.TUNNEL_HEALTH_TYPE_P2P, "T") {
		t.Fatal("bind 端口不同应拒绝")
	}
}

// TestEntry 各类隧道入口字符串。
func TestEntry(t *testing.T) {
	q := setupDB(t)
	n := node("B", "2.2.2.2:7000", "10000-10010", "https://{{DOMAIN}}:1443")
	_ = q.GostClientForward.Create(&model.GostClientForward{Base: model.Base{Code: "F1"}, NodeCode: "B", Port: "10005", Enable: 1})
	_ = q.GostClientHost.Create(&model.GostClientHost{Base: model.Base{Code: "H1"}, NodeCode: "B", DomainPrefix: "abc"})
	if got := Entry(q, n, model.TUNNEL_HEALTH_TYPE_FORWARD, "F1"); got != "2.2.2.2:10005" {
		t.Fatalf("forward entry=%s", got)
	}
	if got := Entry(q, n, model.TUNNEL_HEALTH_TYPE_TUNNEL, "T"); got != "2.2.2.2:7000" {
		t.Fatalf("tunnel entry=%s", got)
	}
	if got := Entry(q, n, model.TUNNEL_HEALTH_TYPE_HOST, "H1"); got != "https://abc.x.example.com:1443" {
		t.Fatalf("host entry=%s", got)
	}
}
