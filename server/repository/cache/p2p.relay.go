package cache

import (
	"fmt"
	"sync"
	"time"

	"server/global"
)

const gost_obs_p2p_relay_key = "gost_plugins:obs:p2p:relay:%s"

var p2pRelayLock = &sync.Mutex{}

type P2PRelayObs struct {
	Bytes int64
}

// IncreaseP2PRelay 累计 P2P STCP 中继流量（待定时落库到 GostClientP2P.RelayBytes）。
func IncreaseP2PRelay(code string, bytes int64) {
	if code == "" || bytes <= 0 {
		return
	}
	p2pRelayLock.Lock()
	defer p2pRelayLock.Unlock()
	var data P2PRelayObs
	key := fmt.Sprintf(gost_obs_p2p_relay_key, code)
	_ = global.Cache.GetStruct(key, &data)
	data.Bytes += bytes
	global.Cache.SetStruct(key, data, time.Hour*24*40)
}

// TakeP2PRelay 取出并清零某隧道的中继流量增量（落库用）。
func TakeP2PRelay(code string) int64 {
	p2pRelayLock.Lock()
	defer p2pRelayLock.Unlock()
	var data P2PRelayObs
	key := fmt.Sprintf(gost_obs_p2p_relay_key, code)
	_ = global.Cache.GetStruct(key, &data)
	if data.Bytes > 0 {
		global.Cache.Del(key)
	}
	return data.Bytes
}
