package task

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"server/global"
	"server/model"
	"server/repository"
	cache2 "server/repository/cache"
	"server/repository/query"
	"server/service/common/alert"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// probeItem 一条待探测隧道的归一化描述。
type probeItem struct {
	Code       string
	Type       int
	UserCode   string
	ClientCode string
	NodeCode   string
	HostHeader string // host 类型探测用 Host 头
	RemotePort string // forward 类型探测用远端口
}

// healthProbe 周期性探测全部启用隧道的健康三态，状态变化时落库并触发告警。
func healthProbe() {
	cfg := global.Config.Health
	if !cfg.Enable {
		return
	}
	timeout := time.Duration(cfg.ProbeTimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = time.Second * 5
	}
	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 16
	}

	db, _, log := repository.Get("")
	nodes, err := db.GostNode.Find()
	if err != nil {
		log.Warn("健康探测加载节点失败", zap.Error(err))
		return
	}
	nodeMap := make(map[string]*model.GostNode, len(nodes))
	for _, n := range nodes {
		nodeMap[n.Code] = n
	}

	items := gatherProbeItems(db, nodeMap)
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for _, it := range items {
		wg.Add(1)
		sem <- struct{}{}
		go func(it probeItem) {
			defer wg.Done()
			defer func() { <-sem }()
			probeOne(db, it, nodeMap[it.NodeCode], timeout)
		}(it)
	}
	wg.Wait()
	monitorNodes(nodes)
}

// monitorNodes 监测节点在线状态，由在线转离线时触发 node_offline 告警（面向管理员）。
func monitorNodes(nodes []*model.GostNode) {
	for _, n := range nodes {
		online := cache2.GetNodeOnline(n.Code)
		cur := "2"
		if online {
			cur = "1"
		}
		key := "node:health:prev:" + n.Code
		prev := global.Cache.GetString(key)
		if prev == "1" && cur == "2" {
			alert.Dispatch(model.ALERT_EVENT_NODE_OFFLINE, model.ALERT_TARGET_NODE, n.Code, "1",
				"节点离线告警", "节点 "+n.Name+"("+n.Code+") 已离线")
		}
		global.Cache.SetString(key, cur, time.Hour*24)
	}
}

func gatherProbeItems(db *query.Query, nodeMap map[string]*model.GostNode) []probeItem {
	var items []probeItem
	hosts, _ := db.GostClientHost.Where(db.GostClientHost.Enable.Eq(1)).Find()
	for _, h := range hosts {
		n := nodeMap[h.NodeCode]
		header := ""
		if n != nil {
			header = n.GetDomainHost(h.DomainPrefix, h.GetCustomDomain(), cache2.GetNodeCustomDomain(h.NodeCode))
		}
		items = append(items, probeItem{Code: h.Code, Type: model.TUNNEL_HEALTH_TYPE_HOST, UserCode: h.UserCode, ClientCode: h.ClientCode, NodeCode: h.NodeCode, HostHeader: header})
	}
	forwards, _ := db.GostClientForward.Where(db.GostClientForward.Enable.Eq(1)).Find()
	for _, f := range forwards {
		items = append(items, probeItem{Code: f.Code, Type: model.TUNNEL_HEALTH_TYPE_FORWARD, UserCode: f.UserCode, ClientCode: f.ClientCode, NodeCode: f.NodeCode, RemotePort: f.Port})
	}
	proxies, _ := db.GostClientProxy.Where(db.GostClientProxy.Enable.Eq(1)).Find()
	for _, p := range proxies {
		items = append(items, probeItem{Code: p.Code, Type: model.TUNNEL_HEALTH_TYPE_PROXY, UserCode: p.UserCode, ClientCode: p.ClientCode, NodeCode: p.NodeCode})
	}
	tunnels, _ := db.GostClientTunnel.Where(db.GostClientTunnel.Enable.Eq(1)).Find()
	for _, t := range tunnels {
		items = append(items, probeItem{Code: t.Code, Type: model.TUNNEL_HEALTH_TYPE_TUNNEL, UserCode: t.UserCode, ClientCode: t.ClientCode, NodeCode: t.NodeCode})
	}
	p2ps, _ := db.GostClientP2P.Where(db.GostClientP2P.Enable.Eq(1)).Find()
	for _, p := range p2ps {
		items = append(items, probeItem{Code: p.Code, Type: model.TUNNEL_HEALTH_TYPE_P2P, UserCode: p.UserCode, ClientCode: p.ClientCode, NodeCode: p.NodeCode})
	}
	return items
}

func probeOne(db *query.Query, it probeItem, node *model.GostNode, timeout time.Duration) {
	prev := cache2.GetTunnelHealth(it.Code)
	online := cache2.GetClientOnline(it.ClientCode)
	registered := cache2.GetTunnelRegistered(it.Code)

	reachable := 0 // 0 未探测
	lastErr := ""
	if prev.ProbeDisable != 1 && node != nil {
		switch it.Type {
		case model.TUNNEL_HEALTH_TYPE_HOST:
			reachable, lastErr = probeHTTP(node, it.HostHeader, timeout)
		case model.TUNNEL_HEALTH_TYPE_FORWARD:
			reachable, lastErr = probeTCP(node, it.RemotePort, timeout)
		}
	}

	status := decideHealth(online, registered, reachable)
	if !online {
		lastErr = "客户端离线"
	} else if !registered {
		lastErr = "隧道未在节点注册"
	}

	h := cache2.TunnelHealth{
		TunnelCode: it.Code, TunnelType: it.Type, UserCode: it.UserCode,
		ClientCode: it.ClientCode, NodeCode: it.NodeCode,
		ClientOnline: b2i(online), TunnelRegistered: b2i(registered),
		BackendReachable: reachable, HealthStatus: status,
		LastCheckAt: time.Now().Unix(), LastError: lastErr, ProbeDisable: prev.ProbeDisable,
	}
	cache2.SetTunnelHealth(it.Code, h)

	// 状态变化（或首次）才落库，并按需告警。
	if prev.LastCheckAt == 0 || prev.HealthStatus != status {
		upsertHealth(db, h)
		if prev.LastCheckAt != 0 {
			if status == model.TUNNEL_HEALTH_ABNORMAL {
				alert.Dispatch(model.ALERT_EVENT_TUNNEL_DOWN, model.ALERT_TARGET_TUNNEL, it.Code, it.UserCode,
					"隧道异常告警", fmt.Sprintf("隧道 %s 转为异常：%s", it.Code, lastErr))
			} else if status == model.TUNNEL_HEALTH_HEALTHY && prev.HealthStatus == model.TUNNEL_HEALTH_ABNORMAL {
				alert.Dispatch(model.ALERT_EVENT_TUNNEL_UP, model.ALERT_TARGET_TUNNEL, it.Code, it.UserCode,
					"隧道恢复通知", fmt.Sprintf("隧道 %s 已恢复健康", it.Code))
			}
		}
	}
}

func upsertHealth(db *query.Query, h cache2.TunnelHealth) {
	row, err := db.GostTunnelHealth.Where(db.GostTunnelHealth.TunnelCode.Eq(h.TunnelCode)).First()
	rec := &model.GostTunnelHealth{
		TunnelCode: h.TunnelCode, TunnelType: h.TunnelType, UserCode: h.UserCode,
		ClientCode: h.ClientCode, NodeCode: h.NodeCode, ClientOnline: h.ClientOnline,
		TunnelRegistered: h.TunnelRegistered, BackendReachable: h.BackendReachable,
		HealthStatus: h.HealthStatus, LastCheckAt: h.LastCheckAt, LastError: h.LastError, ProbeDisable: h.ProbeDisable,
	}
	if err != nil || row == nil {
		_ = db.GostTunnelHealth.Create(rec)
		return
	}
	rec.Id = row.Id
	rec.Code = row.Code
	rec.CreatedAt = row.CreatedAt
	_ = db.GostTunnelHealth.Save(rec)
}

// probeHTTP 经节点 HTTP 入口探测 host 隧道（服务器本地，绕过外网拦截）。
func probeHTTP(node *model.GostNode, hostHeader string, timeout time.Duration) (int, string) {
	ip, _ := node.GetOriginAddress()
	if ip == "" || node.HttpPort == "" || hostHeader == "" {
		return 0, ""
	}
	url := fmt.Sprintf("http://%s:%s/", ip, node.HttpPort)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 2, err.Error()
	}
	req.Host = hostHeader
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return 2, "探测失败:" + err.Error()
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 500 {
		return 2, fmt.Sprintf("上游返回 %d", resp.StatusCode)
	}
	if resp.StatusCode == 404 && strings.Contains(strings.ToLower(string(body)), "frp") {
		return 2, "frp 未找到转发(404)"
	}
	return 1, ""
}

// probeTCP 探测 forward 隧道远端端口可达性。
func probeTCP(node *model.GostNode, port string, timeout time.Duration) (int, string) {
	ip, _ := node.GetOriginAddress()
	if ip == "" || port == "" {
		return 0, ""
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), timeout)
	if err != nil {
		return 2, "端口不可达:" + err.Error()
	}
	_ = conn.Close()
	return 1, ""
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 2
}

// decideHealth 由三态判定综合健康状态：在线且已注册且后端非不可达 → 健康，否则异常。
func decideHealth(online, registered bool, reachable int) int {
	if online && registered && reachable != 2 {
		return model.TUNNEL_HEALTH_HEALTHY
	}
	return model.TUNNEL_HEALTH_ABNORMAL
}
