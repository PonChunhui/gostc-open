package task

import (
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/ncruces/go-sqlite3/gormlite"
	_ "github.com/ncruces/go-sqlite3/embed"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"server/global"
	"server/model"
	"server/pkg/memory"
	cache2 "server/repository/cache"
	"server/repository/query"
)

// setupFailoverDB 建立内存 sqlite + 本地缓存，供故障转移决策测试使用。
func setupFailoverDB(t *testing.T) *query.Query {
	t.Helper()
	global.Cache = memory.NewMemory(filepath.Join(t.TempDir(), "cache.db"), time.Hour)
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.GostNode{}, &model.GostNodeGroup{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return query.Use(db)
}

// TestPickBackup 覆盖故障转移备节点选择的核心决策：
// 显式备节点优先（须在线）→ 主节点分组按策略回退 → 无可用备节点返回空。
func TestPickBackup(t *testing.T) {
	q := setupFailoverDB(t)
	if err := q.GostNodeGroup.Create(&model.GostNodeGroup{Base: model.Base{Code: "G"}, Name: "g", Strategy: model.GOST_NODE_GROUP_STRATEGY_PRIORITY}); err != nil {
		t.Fatal(err)
	}
	mk := func(code, group string, pri int, online bool) {
		if err := q.GostNode.Create(&model.GostNode{Base: model.Base{Code: code}, Key: code + "-key", GroupCode: group, Priority: pri, Weight: 1}); err != nil {
			t.Fatal(err)
		}
		cache2.SetNodeOnline(code, online, time.Hour)
	}
	mk("P", "G", 0, false) // 主节点，离线
	mk("B", "G", 5, true)  // 显式备节点，在线（优先级数值较大）
	mk("M", "G", 1, true)  // 分组成员，在线（优先级最优）
	mk("X", "", 0, false)  // 无分组、离线

	rnd := rand.New(rand.NewSource(1))
	cases := []struct {
		name string
		it   failoverItem
		want string
	}{
		{"显式备节点在线优先", failoverItem{NodeCode: "P", BackupNodeCode: "B"}, "B"},
		{"显式备节点离线回退分组", failoverItem{NodeCode: "P", BackupNodeCode: "X"}, "M"},
		{"无显式备节点走分组策略", failoverItem{NodeCode: "P", BackupNodeCode: ""}, "M"},
		{"主节点无分组且无备节点", failoverItem{NodeCode: "X", BackupNodeCode: ""}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := pickBackup(q, c.it, rnd); got != c.want {
				t.Fatalf("pickBackup = %q, want %q", got, c.want)
			}
		})
	}
}

// TestPickBackupGroupAllOffline 分组内除主节点外全部离线时无可用备节点。
func TestPickBackupGroupAllOffline(t *testing.T) {
	q := setupFailoverDB(t)
	_ = q.GostNodeGroup.Create(&model.GostNodeGroup{Base: model.Base{Code: "G"}, Name: "g", Strategy: model.GOST_NODE_GROUP_STRATEGY_PRIORITY})
	_ = q.GostNode.Create(&model.GostNode{Base: model.Base{Code: "P"}, Key: "pk", GroupCode: "G", Priority: 0, Weight: 1})
	_ = q.GostNode.Create(&model.GostNode{Base: model.Base{Code: "M"}, Key: "mk", GroupCode: "G", Priority: 1, Weight: 1})
	cache2.SetNodeOnline("P", false, time.Hour)
	cache2.SetNodeOnline("M", false, time.Hour)
	rnd := rand.New(rand.NewSource(1))
	if got := pickBackup(q, failoverItem{NodeCode: "P"}, rnd); got != "" {
		t.Fatalf("expected no backup, got %q", got)
	}
}

// TestFailoverCooldown 验证切换冷却（防抖）缓存行为。
func TestFailoverCooldown(t *testing.T) {
	global.Cache = memory.NewMemory(filepath.Join(t.TempDir(), "cache.db"), time.Hour)
	code := "tunnel-cooldown"
	if cache2.InFailoverCooldown(code) {
		t.Fatal("should not be in cooldown initially")
	}
	cache2.SetFailoverCooldown(code, time.Minute)
	if !cache2.InFailoverCooldown(code) {
		t.Fatal("should be in cooldown after set")
	}
}

// TestFailoverE2E 端到端集成演示：真实 failoverScan 代码路径 + 真实 SQLite + 缓存 + 告警。
// 场景：域名隧道主节点「海外-04」离线、备节点「华南-03」在线，开启健康开关后触发巡检，
// 断言隧道 NodeCode 自动切换到备节点，并生成 node_failover 告警记录。
// （无真实客户端连接时 engine 重推优雅跳过，不影响改库与告警。）
func TestFailoverE2E(t *testing.T) {
	global.Cache = memory.NewMemory(filepath.Join(t.TempDir(), "cache.db"), time.Hour)
	global.Logger = zap.NewNop()
	global.Config.Health.Enable = true

	db, err := gorm.Open(gormlite.Open(filepath.Join(t.TempDir(), "e2e.db")), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	db.Exec("PRAGMA foreign_keys=OFF") // 测试库免外键约束（不建关联的客户端/用户行）
	if sqlDB, e := db.DB(); e == nil {
		defer sqlDB.Close() // 释放文件句柄，便于 TempDir 清理
	}
	if err := db.AutoMigrate(
		&model.GostNode{}, &model.GostNodeGroup{},
		&model.GostClientHost{}, &model.GostClientForward{}, &model.GostClientProxy{},
		&model.GostClientTunnel{}, &model.GostClientP2P{},
		&model.SystemAlertRule{}, &model.SystemAlertLog{},
	); err != nil {
		t.Fatal(err)
	}
	q := query.Use(db)
	query.SetDefault(db) // 供 alert.dispatch 内 repository.Get 使用

	// 节点：海外-04 离线，华南-03 在线
	_ = q.GostNode.Create(&model.GostNode{Base: model.Base{Code: "off"}, Key: "off-k", Name: "海外-04"})
	_ = q.GostNode.Create(&model.GostNode{Base: model.Base{Code: "on"}, Key: "on-k", Name: "华南-03"})
	cache2.SetNodeOnline("off", false, time.Hour)
	cache2.SetNodeOnline("on", true, time.Hour)

	// 域名隧道：主节点=海外-04，显式备节点=华南-03，启用故障转移
	host := &model.GostClientHost{
		Base:           model.Base{Code: "host-demo"},
		Name:           "演示域名隧道",
		UserCode:       "u1",
		ClientCode:     "c1",
		NodeCode:       "off",
		BackupNodeCode: "on",
		FailoverEnable: 1,
		Enable:         1,
		TargetIp:       "192.168.1.10",
		TargetPort:     "8080",
	}
	if err := q.GostClientHost.Create(host); err != nil {
		t.Fatal(err)
	}

	// 告警规则：属主 u1，匹配 node_failover，全范围，邮件渠道（未配置邮件服务→记录失败但留痕）
	_ = q.SystemAlertRule.Create(&model.SystemAlertRule{
		Base:       model.Base{Code: "rule-1"},
		UserCode:   "u1",
		Name:       "故障转移告警",
		EventType:  model.ALERT_EVENT_NODE_FAILOVER,
		TargetKind: model.ALERT_TARGET_ALL,
		Enable:     1,
		Channels:   model.MarshalStringList([]string{"email"}),
	})

	t.Logf("转移前: 隧道 %s 主节点=%s(海外-04,离线) 备节点=%s(华南-03,在线)", host.Code, host.NodeCode, host.BackupNodeCode)

	// 触发真实巡检
	failoverScan()

	// 断言：NodeCode 已切换到备节点
	got, _ := q.GostClientHost.Where(q.GostClientHost.Code.Eq("host-demo")).First()
	if got == nil || got.NodeCode != "on" {
		t.Fatalf("期望切换到备节点 on，实际 NodeCode=%v", got)
	}
	t.Logf("转移后: 隧道 %s 主节点已切换=%s(华南-03) ✅", got.Code, got.NodeCode)

	// 断言：生成 node_failover 告警记录（异步，轮询等待）
	var logs []*model.SystemAlertLog
	for i := 0; i < 50; i++ {
		logs, _ = q.SystemAlertLog.Where(q.SystemAlertLog.ObjectCode.Eq("host-demo")).Find()
		if len(logs) > 0 {
			break
		}
		time.Sleep(40 * time.Millisecond)
	}
	if len(logs) == 0 {
		t.Fatal("期望生成 node_failover 告警记录，实际为空")
	}
	for _, l := range logs {
		t.Logf("告警记录: 事件=%s 对象=%s 渠道=%s 结果=%d(1成2败) 详情=%s ✅", l.EventType, l.ObjectCode, l.Channel, l.Result, l.Message)
	}

	// 冷却生效：再次巡检不应重复切换/告警（防抖）
	failoverScan()
	logs2, _ := q.SystemAlertLog.Where(q.SystemAlertLog.ObjectCode.Eq("host-demo")).Find()
	if len(logs2) != len(logs) {
		t.Fatalf("冷却期内不应重复告警: 之前%d条 之后%d条", len(logs), len(logs2))
	}
	t.Logf("冷却防抖: 再次巡检未重复告警（仍%d条）✅", len(logs2))
}

// TestFailoverPrefixDomainStable 平台自动生成的前缀域名在 DNS 跟随下入口域名保持不变：
// 转移/回切都沿用原节点后缀（abc.x01.gostc.shop），仅 A 记录指向变化；回切清理后域名缓存释放。
func TestFailoverPrefixDomainStable(t *testing.T) {
	global.Cache = memory.NewMemory(filepath.Join(t.TempDir(), "cache.db"), time.Hour)
	global.Logger = zap.NewNop()
	global.Config.Failover.DNSFollow = true
	global.Config.DNS.Provider = "" // 测试不触达真实 DNS API（Resolve 返回 ok=false）

	db, err := gorm.Open(gormlite.Open(filepath.Join(t.TempDir(), "prefix.db")), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	db.Exec("PRAGMA foreign_keys=OFF")
	if sqlDB, e := db.DB(); e == nil {
		defer sqlDB.Close()
	}
	if err := db.AutoMigrate(&model.GostNode{}, &model.GostClientHost{}, &model.GostFailoverLog{},
		&model.SystemAlertRule{}, &model.SystemAlertLog{}); err != nil {
		t.Fatal(err)
	}
	q := query.Use(db)
	query.SetDefault(db)

	_ = q.GostNode.Create(&model.GostNode{Base: model.Base{Code: "P"}, Key: "pk", Name: "主",
		Domain: "x01.gostc.shop", UrlTpl: "https://{{DOMAIN}}:1443", Address: "1.1.1.1:7000"})
	_ = q.GostNode.Create(&model.GostNode{Base: model.Base{Code: "B"}, Key: "bk", Name: "备",
		Domain: "x02.gostc.shop", UrlTpl: "https://{{DOMAIN}}:1443", Address: "2.2.2.2:7000"})
	_ = q.GostClientHost.Create(&model.GostClientHost{Base: model.Base{Code: "host-p"}, Name: "前缀域名",
		UserCode: "u1", ClientCode: "c1", NodeCode: "P", DomainPrefix: "abc", Enable: 1,
		TargetIp: "192.168.1.10", TargetPort: "8080"})

	it := failoverItem{Code: "host-p", Type: model.TUNNEL_HEALTH_TYPE_HOST, UserCode: "u1", NodeCode: "P"}
	doSwitch(q, it, "B", model.FAILOVER_REASON_FAILOVER, zap.NewNop())
	if got := cache2.GetFailoverDomain("host-p"); got != "abc.x01.gostc.shop" {
		t.Fatalf("转移后应保持原前缀域名, got %q", got)
	}
	// 回切：仍沿用存储的原域名，不漂移为 abc.x02
	it.NodeCode = "B"
	doSwitch(q, it, "P", model.FAILOVER_REASON_FAILBACK, zap.NewNop())
	if got := cache2.GetFailoverDomain("host-p"); got != "abc.x01.gostc.shop" {
		t.Fatalf("回切应保持原前缀域名, got %q", got)
	}
	logs, _ := q.GostFailoverLog.Where(q.GostFailoverLog.TunnelCode.Eq("host-p")).Find()
	for _, l := range logs {
		if l.NewEntry != "https://abc.x01.gostc.shop:1443" {
			t.Fatalf("入口域名应不变, reason=%s newEntry=%q", l.Reason, l.NewEntry)
		}
	}
	cache2.DelFailoverOrigin("host-p")
	if got := cache2.GetFailoverDomain("host-p"); got != "" {
		t.Fatalf("回切清理后域名缓存应释放, got %q", got)
	}
	t.Logf("前缀域名 DNS 跟随: 转移/回切入口均保持 https://abc.x01.gostc.shop:1443 ✅")
}
