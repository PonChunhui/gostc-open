package task

import (
	"fmt"
	"math/rand"
	"server/global"
	"server/model"
	"server/pkg/crypto"
	"server/repository"
	cache2 "server/repository/cache"
	"server/repository/query"
	"server/service/common/alert"
	"server/service/common/dnsfollow"
	"server/service/common/entryfollow"
	"server/service/common/node_select"
	"server/service/engine"
	"time"

	"go.uber.org/zap"
)

// failoverItem 一条启用故障转移的隧道归一化描述。
type failoverItem struct {
	Code           string
	Type           int
	UserCode       string
	NodeCode       string
	BackupNodeCode string
}

const failoverCooldown = 5 * time.Minute

func failbackDelay() int {
	d := global.Config.Failover.FailbackDelaySec
	if d <= 0 {
		d = 300
	}
	return d
}

// failoverScan 周期性检查启用故障转移的隧道：主节点离线时切换到备节点（同端口硬约束）并重推配置、跟随入口；
// 主节点恢复且开启 failback 时稳定后回切。
func failoverScan() {
	if !global.Config.Health.Enable {
		return
	}
	db, _, log := repository.Get("")
	items := gatherFailoverItems(db)
	if len(items) == 0 {
		return
	}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, it := range items {
		// ---- failback：曾转移且主节点恢复 ----
		if origin := cache2.GetFailoverOrigin(it.Code); origin != "" && origin != it.NodeCode {
			if cache2.GetNodeOnline(origin) {
				if global.Config.Failover.Failback {
					rec := cache2.GetFailoverRecovered(it.Code)
					if rec == 0 {
						cache2.SetFailoverRecovered(it.Code, time.Now().Unix())
					} else if time.Now().Unix()-rec >= int64(failbackDelay()) {
						doSwitch(db, it, origin, model.FAILOVER_REASON_FAILBACK, log)
						cache2.DelFailoverOrigin(it.Code)
					}
				}
				continue // 主节点在线，不做 failover
			}
			if cache2.GetFailoverRecovered(it.Code) != 0 {
				cache2.SetFailoverRecovered(it.Code, 0)
			}
		}
		if cache2.GetNodeOnline(it.NodeCode) {
			continue // 主节点在线，无需转移
		}
		if cache2.InFailoverCooldown(it.Code) {
			continue // 冷却期内，防抖
		}
		backupCode := pickBackup(db, it, rnd)
		if backupCode == "" {
			alert.Dispatch(model.ALERT_EVENT_TUNNEL_DOWN, model.ALERT_TARGET_TUNNEL, it.Code, it.UserCode,
				"隧道主节点离线且无可用备节点",
				fmt.Sprintf("隧道 %s 主节点 %s 离线，未配置可用备节点，请尽快处理", it.Code, it.NodeCode))
			cache2.SetFailoverCooldown(it.Code, failoverCooldown)
			continue
		}
		// ---- 同端口硬约束：备节点给不了同端口则放弃转移 ----
		primary, _ := db.GostNode.Where(db.GostNode.Code.Eq(it.NodeCode)).First()
		backup, _ := db.GostNode.Where(db.GostNode.Code.Eq(backupCode)).First()
		if !entryfollow.CanServeSamePort(db, primary, backup, it.Type, it.Code) {
			writeFailoverLog(db, it, it.NodeCode, backupCode, "", "", model.FAILOVER_REASON_PORT_CONFLICT)
			alert.DispatchPayload(model.ALERT_EVENT_FAILOVER_PORT_CONFLICT, model.ALERT_TARGET_TUNNEL, it.Code, it.UserCode,
				"备节点无法提供同端口，放弃故障转移",
				fmt.Sprintf("隧道 %s 主节点 %s 离线，备节点 %s 无法提供相同公网端口，已放弃转移（端口不变约束）", it.Code, it.NodeCode, backupCode),
				map[string]any{"tunnel": it.Code, "type": it.Type, "from": it.NodeCode, "to": backupCode, "reason": "port_conflict"})
			cache2.SetFailoverCooldown(it.Code, failoverCooldown)
			continue
		}
		cache2.SetFailoverOrigin(it.Code, it.NodeCode)
		doSwitch(db, it, backupCode, model.FAILOVER_REASON_FAILOVER, log)
		cache2.SetFailoverCooldown(it.Code, failoverCooldown)
	}
}

// doSwitch 执行切换 + 入口跟随（DNS/网关路由）+ 历史 + 结构化告警。
func doSwitch(db *query.Query, it failoverItem, target, reason string, log *zap.Logger) {
	primary, _ := db.GostNode.Where(db.GostNode.Code.Eq(it.NodeCode)).First()
	oldEntry := entryfollow.Entry(db, primary, it.Type, it.Code)
	if err := switchTunnelNode(db, it, target); err != nil {
		log.Warn("故障转移失败", zap.String("code", it.Code), zap.Error(err))
		return
	}
	backup, _ := db.GostNode.Where(db.GostNode.Code.Eq(target)).First()
	newEntry := entryfollow.Entry(db, backup, it.Type, it.Code)
	writeFailoverLog(db, it, it.NodeCode, target, oldEntry, newEntry, reason)

	// host：网关路由跟随 + 自定义域名 DNS 跟随
	if it.Type == model.TUNNEL_HEALTH_TYPE_HOST && backup != nil {
		host, _ := db.GostClientHost.Where(db.GostClientHost.Code.Eq(it.Code)).First()
		if host != nil {
			domain := host.GetCustomDomain()
			if domain == "" {
				domain = host.DomainPrefix + "." + backup.Domain
			}
			cert := crypto.DecryptSafe(host.CustomCert, global.Config.Encrypt.Key)
			key := crypto.DecryptSafe(host.CustomKey, global.Config.Encrypt.Key)
			engine.NodeAddDomain(db, target, domain, cert, key, host.CustomForceHttps)
			if custom := host.GetCustomDomain(); custom != "" && global.Config.Failover.DNSFollow {
				if provider, cred, ttl, zone, ok := dnsfollow.Resolve(custom); ok {
					ip, _ := backup.GetOriginAddress()
					if err := dnsfollow.UpdateA(provider, cred, custom, ip, ttl, zone); err != nil {
						log.Warn("DNS 跟随失败", zap.String("domain", custom), zap.Error(err))
					}
				}
			}
		}
	}

	title := "隧道已故障转移到备节点"
	if reason == model.FAILOVER_REASON_FAILBACK {
		title = "主节点恢复，隧道已回切"
	}
	alert.DispatchPayload(model.ALERT_EVENT_NODE_FAILOVER, model.ALERT_TARGET_TUNNEL, it.Code, it.UserCode,
		title,
		fmt.Sprintf("隧道 %s 由节点 %s 切换到 %s，入口 %s -> %s", it.Code, it.NodeCode, target, oldEntry, newEntry),
		map[string]any{"tunnel": it.Code, "type": it.Type, "from": it.NodeCode, "to": target,
			"oldEntry": oldEntry, "newEntry": newEntry, "reason": reason})
}

func writeFailoverLog(db *query.Query, it failoverItem, from, to, oldEntry, newEntry, reason string) {
	_ = db.GostFailoverLog.Create(&model.GostFailoverLog{
		TunnelCode: it.Code, TunnelType: it.Type, UserCode: it.UserCode,
		FromNode: from, ToNode: to, OldEntry: oldEntry, NewEntry: newEntry, Reason: reason,
	})
}

// pickBackup 选择备节点：优先显式备节点（须在线），否则从主节点所在分组按策略挑在线节点。
func pickBackup(db *query.Query, it failoverItem, rnd *rand.Rand) string {
	if it.BackupNodeCode != "" && cache2.GetNodeOnline(it.BackupNodeCode) {
		return it.BackupNodeCode
	}
	primary, _ := db.GostNode.Where(db.GostNode.Code.Eq(it.NodeCode)).First()
	if primary != nil && primary.GroupCode != "" {
		if n, err := node_select.SelectNode(db, primary.GroupCode, it.NodeCode, rnd); err == nil && n != nil {
			return n.Code
		}
	}
	return ""
}

// switchTunnelNode 将隧道切换到备节点并重新下发配置。
func switchTunnelNode(db *query.Query, it failoverItem, backupCode string) error {
	switch it.Type {
	case model.TUNNEL_HEALTH_TYPE_HOST:
		host, err := db.GostClientHost.Where(db.GostClientHost.Code.Eq(it.Code)).First()
		if err != nil {
			return err
		}
		host.NodeCode = backupCode
		if err := db.GostClientHost.Save(host); err != nil {
			return err
		}
		engine.ClientHostConfig(db, it.Code)
	case model.TUNNEL_HEALTH_TYPE_FORWARD:
		forward, err := db.GostClientForward.Where(db.GostClientForward.Code.Eq(it.Code)).First()
		if err != nil {
			return err
		}
		forward.NodeCode = backupCode
		if err := db.GostClientForward.Save(forward); err != nil {
			return err
		}
		engine.ClientForwardConfig(db, it.Code)
	case model.TUNNEL_HEALTH_TYPE_PROXY:
		proxy, err := db.GostClientProxy.Where(db.GostClientProxy.Code.Eq(it.Code)).First()
		if err != nil {
			return err
		}
		proxy.NodeCode = backupCode
		if err := db.GostClientProxy.Save(proxy); err != nil {
			return err
		}
		engine.ClientProxyConfig(db, it.Code)
	case model.TUNNEL_HEALTH_TYPE_TUNNEL:
		tunnel, err := db.GostClientTunnel.Where(db.GostClientTunnel.Code.Eq(it.Code)).First()
		if err != nil {
			return err
		}
		tunnel.NodeCode = backupCode
		if err := db.GostClientTunnel.Save(tunnel); err != nil {
			return err
		}
		engine.ClientTunnelConfig(db, it.Code)
	case model.TUNNEL_HEALTH_TYPE_P2P:
		p2p, err := db.GostClientP2P.Where(db.GostClientP2P.Code.Eq(it.Code)).First()
		if err != nil {
			return err
		}
		p2p.NodeCode = backupCode
		if err := db.GostClientP2P.Save(p2p); err != nil {
			return err
		}
		engine.ClientP2PConfig(db, it.Code)
	default:
		return fmt.Errorf("未知隧道类型 %d", it.Type)
	}
	return nil
}

func gatherFailoverItems(db *query.Query) []failoverItem {
	var items []failoverItem
	hosts, _ := db.GostClientHost.Where(db.GostClientHost.FailoverEnable.Eq(1)).Find()
	for _, h := range hosts {
		items = append(items, failoverItem{h.Code, model.TUNNEL_HEALTH_TYPE_HOST, h.UserCode, h.NodeCode, h.BackupNodeCode})
	}
	forwards, _ := db.GostClientForward.Where(db.GostClientForward.FailoverEnable.Eq(1)).Find()
	for _, f := range forwards {
		items = append(items, failoverItem{f.Code, model.TUNNEL_HEALTH_TYPE_FORWARD, f.UserCode, f.NodeCode, f.BackupNodeCode})
	}
	proxies, _ := db.GostClientProxy.Where(db.GostClientProxy.FailoverEnable.Eq(1)).Find()
	for _, p := range proxies {
		items = append(items, failoverItem{p.Code, model.TUNNEL_HEALTH_TYPE_PROXY, p.UserCode, p.NodeCode, p.BackupNodeCode})
	}
	tunnels, _ := db.GostClientTunnel.Where(db.GostClientTunnel.FailoverEnable.Eq(1)).Find()
	for _, t := range tunnels {
		items = append(items, failoverItem{t.Code, model.TUNNEL_HEALTH_TYPE_TUNNEL, t.UserCode, t.NodeCode, t.BackupNodeCode})
	}
	p2ps, _ := db.GostClientP2P.Where(db.GostClientP2P.FailoverEnable.Eq(1)).Find()
	for _, p := range p2ps {
		items = append(items, failoverItem{p.Code, model.TUNNEL_HEALTH_TYPE_P2P, p.UserCode, p.NodeCode, p.BackupNodeCode})
	}
	return items
}
