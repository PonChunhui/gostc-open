package cache

import (
	"github.com/patrickmn/go-cache"
	"server/global"
	"strconv"
	"time"
)

const (
	tunnel_health_key     = "tunnel:health:"
	tunnel_registered_key = "tunnel:registered:"
)

// TunnelHealth 隧道健康三态缓存（探测任务写入，API 读取）。
type TunnelHealth struct {
	TunnelCode       string
	TunnelType       int
	UserCode         string
	ClientCode       string
	NodeCode         string
	ClientOnline     int
	TunnelRegistered int
	BackendReachable int
	HealthStatus     int
	LastCheckAt      int64
	LastError        string
	ProbeDisable     int
}

func SetTunnelHealth(code string, h TunnelHealth) {
	global.Cache.SetStruct(tunnel_health_key+code, h, cache.NoExpiration)
}

func GetTunnelHealth(code string) (result TunnelHealth) {
	_ = global.Cache.GetStruct(tunnel_health_key+code, &result)
	return result
}

func DelTunnelHealth(code string) {
	global.Cache.Del(tunnel_health_key + code)
}

// SetTunnelRegistered 由 frps newProxy/closeProxy 插件维护的注册态。
func SetTunnelRegistered(code string, registered bool) {
	v := "2"
	if registered {
		v = "1"
	}
	global.Cache.SetString(tunnel_registered_key+code, v, cache.NoExpiration)
}

func GetTunnelRegistered(code string) bool {
	return global.Cache.GetString(tunnel_registered_key+code) == "1"
}

func DelTunnelRegistered(code string) {
	global.Cache.Del(tunnel_registered_key + code)
}

// AlertSilence 告警静默期标记。
func SetAlertSilence(key string, duration time.Duration) {
	global.Cache.SetString("alert:silence:"+key, "1", duration)
}

func InAlertSilence(key string) bool {
	return global.Cache.GetString("alert:silence:"+key) == "1"
}

// FailoverCooldown 故障转移冷却标记，防止主备节点抖动导致反复切换。
func SetFailoverCooldown(code string, duration time.Duration) {
	global.Cache.SetString("failover:cooldown:"+code, "1", duration)
}

func InFailoverCooldown(code string) bool {
	return global.Cache.GetString("failover:cooldown:"+code) == "1"
}

// FailoverOrigin 记录转移前的主节点，用于 failback 回切。
func SetFailoverOrigin(code, primary string) {
	global.Cache.SetString("failover:origin:"+code, primary, cache.NoExpiration)
}

func GetFailoverOrigin(code string) string {
	return global.Cache.GetString("failover:origin:" + code)
}

func DelFailoverOrigin(code string) {
	global.Cache.Del("failover:origin:" + code)
	global.Cache.Del("failover:recovered:" + code)
	global.Cache.Del("failover:domain:" + code)
}

// FailoverDomain 记录转移期间保持不变的公网域名（平台前缀域名 DNS 跟随时使用），
// 保证多次转移/回切过程中入口域名不漂移。
func SetFailoverDomain(code, domain string) {
	global.Cache.SetString("failover:domain:"+code, domain, cache.NoExpiration)
}

func GetFailoverDomain(code string) string {
	return global.Cache.GetString("failover:domain:" + code)
}

// FailoverRecovered 记录主节点恢复在线的起始时间（Unix 秒），用于回切稳定期。
func SetFailoverRecovered(code string, ts int64) {
	global.Cache.SetString("failover:recovered:"+code, strconv.FormatInt(ts, 10), cache.NoExpiration)
}

func GetFailoverRecovered(code string) int64 {
	v, _ := strconv.ParseInt(global.Cache.GetString("failover:recovered:"+code), 10, 64)
	return v
}
