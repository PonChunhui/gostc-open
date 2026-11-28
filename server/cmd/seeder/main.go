package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"server/bootstrap"
	"server/global"
	"server/model"
	"server/pkg/utils"
	"server/repository"
	cache2 "server/repository/cache"
	"server/repository/query"
)

// 演示数据注入器：初始化 DB/缓存后写入一批演示数据，再 Sync 缓存。
// 用法: seeder.exe （需先停止 server，避免 SQLite 写锁与 cache.db 覆盖）
func main() {
	global.BASE_PATH = `C:\code\gostc-open-main\server\data\config.yaml`
	global.FLAG_DEV = true
	global.LOGGER_LEVEL = "warn"
	global.Init()

	bootstrap.InitLogger()
	bootstrap.InitConfig()
	bootstrap.InitPersistence()
	query.SetDefault(global.DB.GetDB())
	bootstrap.InitMemory()

	seed()

	bootstrap.Release()
	fmt.Println("DEMO_SEED_DONE")
}

func seed() {
	db, _, _ := repository.Get("")
	now := time.Now().Unix()
	rnd := rand.New(rand.NewSource(now))

	// ---------- 用户 ----------
	adminCode := "1"
	if u, err := db.SystemUser.Where(db.SystemUser.Admin.Eq(1)).First(); err == nil && u != nil {
		adminCode = u.Code
	} else {
		apwd, _ := utils.HashPassword("admin")
		adm := &model.SystemUser{Account: "admin", Password: apwd, Admin: 1, Status: 1, Amount: decimal.NewFromInt(1000)}
		adm.Code = "1"
		if err := db.SystemUser.Create(adm); err == nil {
			adminCode = adm.Code
		}
	}
	pwd, _ := utils.HashPassword("123456")
	userNames := []string{"alice", "bob", "carol"}
	var userCodes []string
	for i, name := range userNames {
		u := &model.SystemUser{
			Account:  name,
			Password: pwd,
			Admin:    0,
			Status:   1,
			Amount:   decimal.NewFromInt(int64(100 + i*250)),
		}
		if err := db.SystemUser.Create(u); err != nil {
			fmt.Println("create user err:", err)
			continue
		}
		userCodes = append(userCodes, u.Code)
	}
	allUsers := append([]string{adminCode}, userCodes...)

	// ---------- 节点分组 ----------
	grpCore := &model.GostNodeGroup{Name: "核心节点组", Strategy: model.GOST_NODE_GROUP_STRATEGY_PRIORITY, Remark: "按优先级调度"}
	grpEdge := &model.GostNodeGroup{Name: "边缘节点组", Strategy: model.GOST_NODE_GROUP_STRATEGY_WEIGHT, Remark: "按权重调度"}
	_ = db.GostNodeGroup.Create(grpCore)
	_ = db.GostNodeGroup.Create(grpEdge)

	// ---------- 节点 ----------
	type nodeSpec struct {
		name, domain, addr, httpPort, group string
		priority, weight                    int
		online                              bool
	}
	nodeSpecs := []nodeSpec{
		{"华东-01", "hd01.demo.gostc.shop", "10.0.0.1:7000", "8080", grpCore.Code, 1, 5, true},
		{"华北-02", "hb02.demo.gostc.shop", "10.0.0.2:7000", "8080", grpCore.Code, 2, 3, true},
		{"华南-03", "hn03.demo.gostc.shop", "10.0.0.3:7000", "8080", grpEdge.Code, 1, 2, true},
		{"海外-04", "ov04.demo.gostc.shop", "10.0.0.4:7000", "8080", grpEdge.Code, 2, 1, false},
	}
	var nodeCodes []string
	for _, s := range nodeSpecs {
		n := &model.GostNode{
			Key:  uuid.NewString(),
			Name: s.name, Domain: s.domain, Address: s.addr, HttpPort: s.httpPort,
			Protocol: "tcp", GroupCode: s.group, Priority: s.priority, Weight: s.weight,
			Web: 1, Tunnel: 1, Forward: 1, Proxy: 1, P2P: 1, AllowDomainMatcher: 1,
			UrlTpl: "https://{{DOMAIN}}:1443", ForwardPorts: "10001-11000",
		}
		if err := db.GostNode.Create(n); err != nil {
			fmt.Println("create node err:", err)
			continue
		}
		nodeCodes = append(nodeCodes, n.Code)
		cache2.SetNodeOnline(n.Code, s.online, time.Hour*24)
	}

	// ---------- 客户端 ----------
	type clientSpec struct {
		name   string
		user   int
		online bool
	}
	clientSpecs := []clientSpec{
		{"我的工作电脑", 0, true},
		{"我的Mac", 0, true},
		{"家庭NAS", 1, true},
		{"公司服务器", 2, false},
	}
	var clientCodes []string
	for _, s := range clientSpecs {
		c := &model.GostClient{Name: s.name, UserCode: userCodes[s.user], Key: uuid.NewString()}
		if err := db.GostClient.Create(c); err != nil {
			fmt.Println("create client err:", err)
			continue
		}
		clientCodes = append(clientCodes, c.Code)
		cache2.SetClientOnline(c.Code, s.online, time.Hour*24)
	}

	// ---------- 隧道 + 健康 ----------
	// health: online, registered, reachable, status, lastErr
	type tSpec struct {
		name                       string
		typ                        int
		user, node, client         int
		online, reg, reach, status int
		lastErr                    string
	}
	specs := []tSpec{
		{"企业官网", model.TUNNEL_HEALTH_TYPE_HOST, 0, 0, 0, 1, 1, 1, model.TUNNEL_HEALTH_HEALTHY, ""},
		{"个人博客", model.TUNNEL_HEALTH_TYPE_HOST, 1, 1, 2, 1, 1, 2, model.TUNNEL_HEALTH_ABNORMAL, "backend connect refused"},
		{"API文档站", model.TUNNEL_HEALTH_TYPE_HOST, 2, 2, 3, 1, 1, 1, model.TUNNEL_HEALTH_HEALTHY, ""},
		{"SSH远程", model.TUNNEL_HEALTH_TYPE_FORWARD, 0, 0, 0, 1, 1, 1, model.TUNNEL_HEALTH_HEALTHY, ""},
		{"远程桌面", model.TUNNEL_HEALTH_TYPE_FORWARD, 1, 1, 2, 1, 2, 0, model.TUNNEL_HEALTH_ABNORMAL, "proxy not registered"},
		{"SOCKS5代理", model.TUNNEL_HEALTH_TYPE_PROXY, 2, 2, 3, 1, 1, 1, model.TUNNEL_HEALTH_HEALTHY, ""},
		{"私密隧道", model.TUNNEL_HEALTH_TYPE_TUNNEL, 0, 0, 1, 2, 2, 0, model.TUNNEL_HEALTH_ABNORMAL, "client offline"},
		{"NAS同步", model.TUNNEL_HEALTH_TYPE_P2P, 1, 1, 2, 1, 1, 0, model.TUNNEL_HEALTH_HEALTHY, ""},
	}
	var tunnelCodes []string
	for i, s := range specs {
		uc, nc, cc := userCodes[s.user], nodeCodes[s.node], clientCodes[s.client]
		var code string
		switch s.typ {
		case model.TUNNEL_HEALTH_TYPE_HOST:
			h := &model.GostClientHost{Name: s.name, TargetIp: "192.168.1.10", TargetPort: "8080", DomainPrefix: fmt.Sprintf("demo%d", i), NodeCode: nc, ClientCode: cc, UserCode: uc, Enable: 1, Status: 1}
			_ = db.GostClientHost.Create(h)
			code = h.Code
		case model.TUNNEL_HEALTH_TYPE_FORWARD:
			f := &model.GostClientForward{Name: s.name, TargetIp: "192.168.1.10", TargetPort: "22", Port: fmt.Sprintf("1000%d", i), NodeCode: nc, ClientCode: cc, UserCode: uc, Enable: 1, Status: 1}
			_ = db.GostClientForward.Create(f)
			code = f.Code
		case model.TUNNEL_HEALTH_TYPE_PROXY:
			p := &model.GostClientProxy{Name: s.name, Protocol: "socks5", Port: "1080", AuthUser: "demo", AuthPwd: "demo123", NodeCode: nc, ClientCode: cc, UserCode: uc, Enable: 1, Status: 1}
			_ = db.GostClientProxy.Create(p)
			code = p.Code
		case model.TUNNEL_HEALTH_TYPE_TUNNEL:
			t := &model.GostClientTunnel{Name: s.name, TargetIp: "192.168.1.10", TargetPort: "3389", VKey: "demo-vkey", NodeCode: nc, ClientCode: cc, UserCode: uc, Enable: 1, Status: 1}
			_ = db.GostClientTunnel.Create(t)
			code = t.Code
		case model.TUNNEL_HEALTH_TYPE_P2P:
			p := &model.GostClientP2P{Name: s.name, TargetIp: "192.168.1.10", TargetPort: "5000", VKey: "demo-p2p", Forward: 1, NatType: "EasyNAT", RelayBytes: 1024 * 1024 * 512, NodeCode: nc, ClientCode: cc, UserCode: uc, Enable: 1, Status: 1}
			_ = db.GostClientP2P.Create(p)
			code = p.Code
		}
		if code == "" {
			continue
		}
		tunnelCodes = append(tunnelCodes, code)
		h := &model.GostTunnelHealth{
			TunnelCode: code, TunnelType: s.typ, UserCode: uc, ClientCode: cc, NodeCode: nc,
			ClientOnline: s.online, TunnelRegistered: s.reg, BackendReachable: s.reach,
			HealthStatus: s.status, LastCheckAt: now, LastError: s.lastErr,
		}
		_ = db.GostTunnelHealth.Create(h)
		cache2.SetTunnelHealth(code, cache2.TunnelHealth{
			TunnelCode: code, TunnelType: s.typ, UserCode: uc, ClientCode: cc, NodeCode: nc,
			ClientOnline: s.online, TunnelRegistered: s.reg, BackendReachable: s.reach,
			HealthStatus: s.status, LastCheckAt: now, LastError: s.lastErr,
		})
		cache2.SetTunnelRegistered(code, s.reg == 1)
	}

	// ---------- 告警规则 + 记录 ----------
	rules := []*model.SystemAlertRule{
		{UserCode: adminCode, Name: "全站隧道离线告警", EventType: model.ALERT_EVENT_TUNNEL_DOWN, TargetKind: model.ALERT_TARGET_ALL, Channels: model.MarshalStringList([]string{"email", "webhook"}), Receivers: model.MarshalStringList([]string{"admin@demo.local"}), SilenceSec: 300, Enable: 1},
		{UserCode: userCodes[0], Name: "华东节点离线", EventType: model.ALERT_EVENT_NODE_OFFLINE, TargetKind: model.ALERT_TARGET_NODE, TargetCode: nodeCodes[0], Channels: model.MarshalStringList([]string{"email"}), Receivers: model.MarshalStringList([]string{"alice@demo.local"}), SilenceSec: 600, Enable: 1},
		{UserCode: userCodes[1], Name: "流量耗尽提醒", EventType: model.ALERT_EVENT_TRAFFIC_EXHAUST, TargetKind: model.ALERT_TARGET_ALL, Channels: model.MarshalStringList([]string{"webhook"}), Receivers: model.MarshalStringList([]string{"https://hooks.demo.local/bob"}), SilenceSec: 3600, Enable: 1},
	}
	for _, r := range rules {
		_ = db.SystemAlertRule.Create(r)
	}
	events := []string{model.ALERT_EVENT_TUNNEL_DOWN, model.ALERT_EVENT_TUNNEL_UP, model.ALERT_EVENT_NODE_OFFLINE, model.ALERT_EVENT_TRAFFIC_EXHAUST, model.ALERT_EVENT_CERT_EXPIRE}
	channels := []string{"email", "webhook"}
	for i := 0; i < 12; i++ {
		result := 1
		msg := "告警发送成功"
		if i%4 == 3 {
			result = 2
			msg = "webhook 请求超时"
		}
		l := &model.SystemAlertLog{
			RuleCode: rules[i%len(rules)].Code, UserCode: rules[i%len(rules)].UserCode,
			EventType: events[i%len(events)], ObjectKind: model.ALERT_TARGET_TUNNEL, ObjectCode: tunnelCodes[i%len(tunnelCodes)],
			Channel: channels[i%2], Result: result, Message: msg,
		}
		_ = db.SystemAlertLog.Create(l)
	}

	// ---------- 指标：历史落库 + 实时缓存 ----------
	type metricObj struct {
		kind int
		code string
	}
	var objs []metricObj
	for _, c := range nodeCodes {
		objs = append(objs, metricObj{model.GOST_OBS_ORIGIN_KIND_NODE, c})
	}
	for _, c := range clientCodes {
		objs = append(objs, metricObj{model.GOST_OBS_ORIGIN_KIND_CLIENT, c})
	}
	for _, c := range allUsers {
		objs = append(objs, metricObj{model.GOST_OBS_ORIGIN_KIND_USER, c})
	}
	for _, c := range tunnelCodes {
		objs = append(objs, metricObj{model.GOST_OBS_ORIGIN_KIND_TUNNEL, c})
	}
	metrics := []string{model.METRIC_INPUT_RATE, model.METRIC_OUTPUT_RATE, model.METRIC_CONNECTIONS}
	for _, o := range objs {
		base := float64(50000 + rnd.Intn(500000))
		for _, m := range metrics {
			var hist []*model.GostMetric
			// 历史：过去 12 小时，每 10 分钟一个点
			for k := 72; k >= 1; k-- {
				ts := now - int64(k)*600
				val := base * (0.6 + rnd.Float64()*0.8)
				if m == model.METRIC_CONNECTIONS {
					val = float64(5 + rnd.Intn(120))
				}
				hist = append(hist, &model.GostMetric{OriginKind: o.kind, OriginCode: o.code, Metric: m, Value: val, Window: "10m", Ts: ts})
			}
			_ = db.GostMetric.Create(hist...)
			// 实时 gauge + 环形曲线（最近 60 点，每 5 秒）
			cur := base * (0.6 + rnd.Float64()*0.8)
			if m == model.METRIC_CONNECTIONS {
				cur = float64(5 + rnd.Intn(120))
			}
			cache2.SetMetricGauge(o.kind, o.code, m, cur)
			for k := 60; k >= 1; k-- {
				ts := now - int64(k)*5
				v := cur * (0.7 + rnd.Float64()*0.6)
				cache2.PushMetricPoint(o.kind, o.code, m, ts, v)
			}
		}
	}

	// ---------- 每日流量 ----------
	for _, o := range objs {
		for d := 0; d < 7; d++ {
			day := time.Now().AddDate(0, 0, -d).Format("2006-01-02")
			obs := &model.GostObs{
				OriginKind: o.kind, OriginCode: o.code, EventDate: day,
				InputBytes:  int64(1024*1024*100 + rnd.Intn(1024*1024*900)),
				OutputBytes: int64(1024*1024*80 + rnd.Intn(1024*1024*700)),
			}
			_ = db.GostObs.Create(obs)
		}
	}

	// ---------- 证书 ----------
	certs := []*model.SystemCert{
		{UserCode: adminCode, Domain: "*.demo.gostc.shop", Provider: model.CERT_PROVIDER_CLOUDFLARE, CertPath: "data/certs/_wildcard.demo.gostc.shop.pem", KeyPath: "data/certs/_wildcard.demo.gostc.shop.key", ExpAt: now + 86400*60, AutoRenew: 1, Status: model.CERT_STATUS_VALID},
		{UserCode: userCodes[0], Domain: "blog.demo.gostc.shop", Provider: model.CERT_PROVIDER_DNSPOD, CertPath: "data/certs/blog.demo.gostc.shop.pem", KeyPath: "data/certs/blog.demo.gostc.shop.key", ExpAt: now + 86400*10, AutoRenew: 1, Status: model.CERT_STATUS_VALID},
	}
	for _, c := range certs {
		_ = db.SystemCert.Create(c)
	}

	// ---------- 订单 + 账单 ----------
	orders := []*model.SystemOrder{
		{UserCode: userCodes[0], Subject: "余额充值 100 元", Kind: model.ORDER_KIND_RECHARGE, Amount: decimal.NewFromInt(100), Provider: "sandbox", Status: model.ORDER_STATUS_PAID, TradeNo: "SBX100001", PayAt: now - 3600, ExpAt: now + 86400},
		{UserCode: userCodes[1], Subject: "余额充值 50 元", Kind: model.ORDER_KIND_RECHARGE, Amount: decimal.NewFromInt(50), Provider: "sandbox", Status: model.ORDER_STATUS_PENDING, ExpAt: now + 1800},
		{UserCode: userCodes[2], Subject: "升级专业版用户组", Kind: model.ORDER_KIND_GROUP, TargetCode: "pro", Amount: decimal.NewFromInt(99), Provider: "sandbox", Status: model.ORDER_STATUS_PAID, TradeNo: "SBX100003", PayAt: now - 7200, ExpAt: now + 86400},
	}
	for _, o := range orders {
		_ = db.SystemOrder.Create(o)
		if o.Status == model.ORDER_STATUS_PAID {
			_ = db.SystemBill.Create(&model.SystemBill{UserCode: o.UserCode, OrderCode: o.Code, Kind: "recharge", Amount: o.Amount, BalanceAfter: o.Amount, Remark: o.Subject})
		}
	}

	// ---------- 客户端版本 ----------
	versions := []*model.SystemClientVersion{
		{Version: "v2.1.0", Platform: "windows", Arch: "amd64", DownloadURL: "https://dl.demo.local/gostc-windows-amd64.zip", Sha256: "demo-sha256-win", Channel: model.CLIENT_VERSION_CHANNEL_STABLE, GrayPercent: 100, Enable: 1, ReleaseNote: "新增健康监控与告警"},
		{Version: "v2.1.0", Platform: "linux", Arch: "amd64", DownloadURL: "https://dl.demo.local/gostc-linux-amd64.tar.gz", Sha256: "demo-sha256-linux", Channel: model.CLIENT_VERSION_CHANNEL_BETA, GrayPercent: 50, Enable: 1, ReleaseNote: "P2P NAT 预检"},
	}
	for _, v := range versions {
		_ = db.SystemClientVersion.Create(v)
	}

	// ---------- 隧道模板 ----------
	templates := []*model.SystemTunnelTemplate{
		{UserCode: userCodes[0], Name: "Web站点模板", Kind: model.TUNNEL_TEMPLATE_KIND_HOST, Content: `{"targetIp":"192.168.1.10","targetPort":"8080","targetHttps":0}`},
		{UserCode: userCodes[1], Name: "SSH转发模板", Kind: model.TUNNEL_TEMPLATE_KIND_FORWARD, Content: `{"targetIp":"192.168.1.10","targetPort":"22"}`},
	}
	for _, t := range templates {
		_ = db.SystemTunnelTemplate.Create(t)
	}

	// ---------- API Token ----------
	_ = db.SystemApiToken.Create(&model.SystemApiToken{UserCode: userCodes[0], Name: "演示令牌", TokenHash: "demo-token-hash", TokenPrefix: "gostc_demo...", ExpAt: 0})

	fmt.Printf("seeded: users=%d nodes=%d clients=%d tunnels=%d metricObjs=%d\n", len(userCodes), len(nodeCodes), len(clientCodes), len(tunnelCodes), len(objs))
}
