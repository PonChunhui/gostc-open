package rpc

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lesismal/arpc"
	arpcLog "github.com/lesismal/arpc/log"
	cache2 "github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"net"
	"server/global"
	"server/model"
	"server/pkg/crypto"
	"server/pkg/rpc_protocol/websocket"
	"server/repository"
	cache3 "server/repository/cache"
	"server/service/common/ban"
	"server/service/engine"
	"strings"
	"time"
)

func checkRegister(client *arpc.Client) {
	time.Sleep(time.Second * 120)
	_, ok := client.Get("registered")
	if !ok {
		_ = client.Notify("stop", "客户端连接注册超时", time.Second*5)
		time.Sleep(time.Second * 2)
		client.Stop()
	}
}

func InitFrp(ginEngine *gin.Engine, ln net.Listener) {
	if global.MODE == "dev" {
		arpcLog.SetLevel(arpcLog.LevelDebug)
	} else {
		arpcLog.SetLevel(arpcLog.LevelNone)
	}
	rpcServer := arpc.NewServer()
	rpcServer.Handler.SetReadTimeout(time.Second * 50)
	rpcServer.Handler.HandleConnected(func(client *arpc.Client) {
		go checkRegister(client)
	})
	rpcServer.Handler.HandleDisconnected(func(client *arpc.Client) {
		fmt.Println("HandleDisconnected")
		code, ok := client.Get("code")
		if ok {
			cache3.SetClientOnline(code.(string), false, cache2.NoExpiration)
			cache3.SetNodeOnline(code.(string), false, cache2.NoExpiration)
			cache3.DelNodeLoad(code.(string))
		}
	})
	rpcServer.Handler.HandleMessageDone(func(c *arpc.Client, m *arpc.Message) {
		//fmt.Println(string(m.Data()))
	})

	// 路由
	rpcServer.Handler.Handle("rpc/client/ping", func(c *arpc.Context) {
		code, ok := c.Client.Get("code")
		if ok {
			cache3.SetClientOnline(code.(string), true, cache2.NoExpiration)
		}
		_ = c.Write("success")
	})
	rpcServer.Handler.Handle("rpc/node/ping", func(c *arpc.Context) {
		code, ok := c.Client.Get("code")
		if ok {
			cache3.SetNodeOnline(code.(string), true, cache2.NoExpiration)
		}
		_ = c.Write("success")
	})
	rpcServer.Handler.Handle("rpc/node/load", func(c *arpc.Context) {
		var req NodeLoadReq
		if err := c.Bind(&req); err != nil {
			_ = c.Write(err.Error())
			return
		}
		code, ok := c.Client.Get("code")
		if ok {
			cache3.SetNodeOnline(code.(string), true, cache2.NoExpiration)
			cache3.SetNodeLoad(code.(string), cache3.NodeLoad{
				ClientCount:     req.ClientCount,
				TunnelCount:     req.TunnelCount,
				ConnectionCount: req.ConnectionCount,
			}, time.Minute)
		}
		_ = c.Write("success")
	})
	rpcServer.Handler.Handle("rpc/client/reg", func(c *arpc.Context) {
		var reply = make(map[string]string)
		if err := c.Bind(&reply); err != nil {
			_ = c.Write(err.Error())
			return
		}
		key := reply["key"]
		version := reply["version"]
		if key == "" {
			_ = c.Write("no key")
			return
		}

		client := c.Client
		client.Set("key", key)
		client.Set("version", version)
		client.Set("registered", 1)

		db, _, _ := repository.Get("")
		gostClient, _ := db.GostClient.Where(db.GostClient.Key.Eq(key)).First()
		if gostClient == nil {
			gostClient = &model.GostClient{}
		}
		if gostClient.Code != "" && ban.UserBanned(gostClient.UserCode) {
			_ = c.Write("账号已被封禁")
			return
		}

		value, ok := engine.EngineRegistry.Get(gostClient.Code)
		if ok {
			value.GetClient().Stop("客户端已在别处连接")
			time.Sleep(time.Second * 2)
		}

		e := engine.NewARpcClientEngine(gostClient.Code, client)
		clientEngine := engine.NewClientEngine(gostClient.Code, e)
		if gostClient.Code == "" {
			_ = c.Write("客户端不存在")
			return
		} else {
			engine.EngineRegistry.Set(clientEngine)
		}
		_ = c.Write("success")
		cache3.SetClientLastTime(gostClient.Code)
		cache3.SetClientOnline(gostClient.Code, true, cache2.NoExpiration)
		cache3.SetClientVersion(gostClient.Code, version)
		client.Set("code", gostClient.Code)
		client.Set("userCode", gostClient.UserCode)
		engine.ClientAllConfigUpdateByClientCode(db, gostClient.Code)
	})
	rpcServer.Handler.Handle("rpc/node/reg", func(c *arpc.Context) {
		var reply = make(map[string]string)
		if err := c.Bind(&reply); err != nil {
			_ = c.Write(err.Error())
			return
		}
		key := reply["key"]
		version := reply["version"]
		domain := reply["domain"]
		domainCache := reply["domain_cache"]
		if key == "" {
			_ = c.Write("no key")
			return
		}

		client := c.Client
		client.Set("key", key)
		client.Set("version", version)
		client.Set("registered", 1)

		db, _, _ := repository.Get("")

		gostNode, _ := db.GostNode.Where(db.GostNode.Key.Eq(key)).First()
		if gostNode == nil {
			_ = c.Write("节点不存在")
			return
		}

		value, ok := engine.EngineRegistry.Get(gostNode.Code)
		if ok {
			value.GetNode().Stop("节点已在别处连接")
			time.Sleep(time.Second)
		}

		e := engine.NewARpcNodeEngine(gostNode.Code, "", client)
		clientEngine := engine.NewNodeEngine(gostNode.Code, e)
		if gostNode.Code == "" {
			_ = c.Write("节点不存在")
			return
		} else {
			engine.EngineRegistry.Set(clientEngine)
		}
		_ = c.Write("success")
		cache3.SetNodeOnline(gostNode.Code, true, cache2.NoExpiration)
		cache3.SetNodeVersion(gostNode.Code, version)
		cache3.SetNodeCustomDomain(gostNode.Code, domain == "1")
		cache3.SetNodeCache(gostNode.Code, domainCache == "1")
		cache3.DelNodeLoad(gostNode.Code)
		client.Set("code", gostNode.Code)
		engine.NodeConfig(db, gostNode.Code)
		// 节点重连（含面板重启后）立即触发一次健康探测，由探测ground truth重建隧道注册态，
		// 避免重启后注册缓存丢失、客户端配置未变又不重新注册导致的"未注册"假象。
		if global.HealthProbeTrigger != nil {
			global.HealthProbeTrigger()
		}
	})

	rpcServer.Handler.Handle("rpc/metrics/input", func(c *arpc.Context) {
		var req TrafficOutput
		if err := c.Bind(&req); err != nil {
			_ = c.Write(err.Error())
			return
		}
		code := strings.Split(req.Name, "_")[0]
		tunnelInfo := cache3.GetTunnelInfo(code)
		go cache3.IncreaseObs(time.Now().Format(time.DateOnly), tunnelInfo.Code, tunnelInfo.ClientCode, tunnelInfo.NodeCode, tunnelInfo.UserCode, cache3.TunnelObs{
			InputBytes:  req.Total,
			OutputBytes: 0,
		})
		if strings.HasSuffix(req.Name, "_p2pstcp") {
			go cache3.IncreaseP2PRelay(code, req.Total)
		}
		_ = c.Write("success")
	})
	rpcServer.Handler.Handle("rpc/metrics/output", func(c *arpc.Context) {
		var req TrafficOutput
		if err := c.Bind(&req); err != nil {
			_ = c.Write(err.Error())
			return
		}
		code := strings.Split(req.Name, "_")[0]
		tunnelInfo := cache3.GetTunnelInfo(code)
		go cache3.IncreaseObs(time.Now().Format(time.DateOnly), tunnelInfo.Code, tunnelInfo.ClientCode, tunnelInfo.NodeCode, tunnelInfo.UserCode, cache3.TunnelObs{
			InputBytes:  0,
			OutputBytes: req.Total,
		})
		if strings.HasSuffix(req.Name, "_p2pstcp") {
			go cache3.IncreaseP2PRelay(code, req.Total)
		}
		_ = c.Write("success")
	})

	rpcServer.Handler.Handle("rpc/client/host_set_domain_certs", func(c *arpc.Context) {
		var req HostSetDomainCerts
		if err := c.Bind(&req); err != nil {
			_ = c.Write(err.Error())
			return
		}
		userCodeTemp, ok := c.Client.Get("userCode")
		if !ok {
			_ = c.Write("unknown error")
			return
		}
		userCode := userCodeTemp.(string)
		db, _, _ := repository.Get("")
		host, err := db.GostClientHost.Preload(
			db.GostClientHost.Node,
		).Where(
			db.GostClientHost.CustomDomain.Eq(req.Domain),
			db.GostClientHost.UserCode.Eq(userCode),
		).First()
		if err != nil {
			_ = c.Write("no tunnel matching the domain name was queried.")
			return
		}
		plainKey := crypto.DecryptSafe(host.CustomKey, global.Config.Encrypt.Key)
		if host.CustomCert == req.Cert && plainKey == req.Key {
			_ = c.Write("success")
			return
		}
		// 兼容旧客户端上报证书：存为存量快照（绑定 UI 已改为引用证书管理）。
		encKey, err := crypto.EncryptIfEnabled(req.Key, global.Config.Encrypt.Key)
		if err != nil {
			_ = c.Write(err.Error())
			return
		}
		host.CustomCert = req.Cert
		host.CustomKey = encKey
		if err := db.GostClientHost.Save(host); err != nil {
			_ = c.Write(err.Error())
			return
		}
		engine.NodeAddDomain(db, host.NodeCode, host.GetCustomDomain(), req.Cert, req.Key, host.CustomForceHttps)
		_ = c.Write("success")
	})

	// 入口
	ginEngine.GET("rpc/ws", func(c *gin.Context) {
		ln.(*websocket.Listener).Handler(c.Writer, c.Request)
	})

	go func() {
		if err := rpcServer.Serve(ln); err != nil {
			global.Logger.Fatal("rpc serve fail", zap.Error(err))
			panic(err)
		}
	}()
	time.Sleep(time.Second)
	global.Logger.Info("rpc server on address: " + global.Config.Address)
}

type TrafficOutput struct {
	Name      string
	ProxyType string
	Total     int64
}

type NodeLoadReq struct {
	ClientCount     int64 `json:"clientCount"`
	TunnelCount     int64 `json:"tunnelCount"`
	ConnectionCount int64 `json:"connectionCount"`
}

type HostSetDomainCerts struct {
	Domain string `json:"domain"`
	Cert   string `json:"cert"`
	Key    string `json:"key"`
}
