package engine

import (
	"errors"
	"fmt"
	"server/global"
	"server/model"
	"server/pkg/utils"
	"server/repository/cache"
	"server/repository/query"
	"server/service/common/warn_msg"
	"time"

	v1 "github.com/SianHH/frp-package/pkg/config/v1"
	"github.com/lesismal/arpc"
	"go.uber.org/zap"
)

type ARpcClientEngine struct {
	code   string
	client *arpc.Client
}

func (e *ARpcClientEngine) IsRunning() bool {
	return e.client.CheckState() == nil
}

func NewARpcClientEngine(code string, client *arpc.Client) *ARpcClientEngine {
	return &ARpcClientEngine{code: code, client: client}
}

func (e *ARpcClientEngine) Stop(msg string) {
	_ = e.client.CallAsync("stop", msg, func(c *arpc.Context, err error) {

	}, time.Second*30)
}

func (e *ARpcClientEngine) Logger() ([]ClientLogEntry, error) {
	var logs []ClientLogEntry
	if err := e.client.Call("logger", nil, &logs, time.Second*10); err != nil {
		return nil, err
	}
	return logs, nil
}

func (e *ARpcClientEngine) PortCheck(tx *query.Query, ip, port string) error {
	var relay string
	if err := e.client.Call("port_check", port, &relay, time.Second*5); err != nil {
		return err
	}
	if relay != "success" {
		return errors.New(relay)
	}
	return nil
}

// buildHealthCheck 由表单字段构造 frp 健康检查配置；类型为空或间隔≤0 时不启用。
func buildHealthCheck(hcType, path string, interval, timeout, maxFailed int) v1.HealthCheckConfig {
	if hcType == "" || interval <= 0 {
		return v1.HealthCheckConfig{}
	}
	return v1.HealthCheckConfig{
		Type:            hcType,
		Path:            path,
		IntervalSeconds: interval,
		TimeoutSeconds:  timeout,
		MaxFailed:       maxFailed,
	}
}

// forwardBandwidth 计算端口转发的带宽限制：用户显式设置优先，否则回退套餐限速。
func forwardBandwidth(forward model.GostClientForward) (limit, mode string) {
	limit = forward.BandwidthLimit
	if limit == "" {
		limit = fmt.Sprintf("%dKB", forward.Limiter*128)
	}
	mode = forward.BandwidthMode
	if mode == "" {
		mode = "client"
	}
	return limit, mode
}

func (e *ARpcClientEngine) HostConfig(tx *query.Query, hostCode string) error {
	host, err := tx.GostClientHost.Preload(tx.GostClientHost.Node).Where(tx.GostClientHost.Code.Eq(hostCode)).First()
	if err != nil {
		return err
	}
	warnMsg := warn_msg.GetHostWarnMsg(*host)
	if warnMsg != "" {
		_ = e.RemoveHost(tx, *host, host.Node)
		return errors.New(warnMsg)
	}
	auth, err := tx.GostAuth.Where(tx.GostAuth.TunnelCode.Eq(hostCode)).First()
	if err != nil {
		return err
	}
	var data = HostConfig{
		Key:       host.Code,
		UpdateTag: host.UpdatedAt.Format(time.DateTime) + "#" + host.Node.UpdatedAt.Format(time.DateTime),
		BaseCfg:   e.generateServerCommonCfg(host.PoolCount, host.Node, *auth),
		IsHttps:   host.TargetHttps == 1,
		Http: HTTPProxyConfig{
			HTTPProxyConfig: v1.HTTPProxyConfig{
				ProxyBaseConfig: v1.ProxyBaseConfig{
					Name: host.Code + "_http",
					Type: "http",
					Metadatas: map[string]string{
						"user":     auth.User,
						"password": auth.Password,
					},
					ProxyBackend: v1.ProxyBackend{
						LocalIP:   host.TargetIp,
						LocalPort: utils.StrMustInt(host.TargetPort),
					},
					LoadBalancer: v1.LoadBalancerConfig{
						Group:    host.LoadBalancerGroup,
						GroupKey: host.LoadBalancerGroupKey,
					},
					HealthCheck: buildHealthCheck(host.HealthCheckType, host.HealthCheckPath, host.HealthCheckInterval, host.HealthCheckTimeout, host.HealthCheckMaxFailed),
				},
				DomainConfig: v1.DomainConfig{
					CustomDomains: []string{
						host.Node.GetDomainHost(host.DomainPrefix, host.GetCustomDomain(), cache.GetNodeCustomDomain(host.NodeCode)),
					},
				},
				Locations:         host.GetLocations(),
				HostHeaderRewrite: host.HostHeaderRewrite,
			},
			Transport: ProxyTransport{
				UseEncryption:        host.UseEncryption == 1,
				UseCompression:       host.UseCompression == 1,
				BandwidthLimit:       fmt.Sprintf("%dKB", host.Limiter*128),
				BandwidthLimitMode:   "client",
				ProxyProtocolVersion: "",
			},
		},
	}

	go utils.Retry(func() error {
		var relay string
		if err := e.client.Call("host_config", data, &relay, time.Second*30); err != nil {
			global.Logger.Error("send message fail", zap.Error(err))
			return err
		}
		if relay != "success" {
			global.Logger.Error("send message fail", zap.String("relay", relay))
		}
		return nil
	}, 5, time.Second*5)
	return nil
}

func (e *ARpcClientEngine) RemoveHost(tx *query.Query, host model.GostClientHost, node model.GostNode) error {
	go utils.Retry(func() error {
		var relay string
		if err := e.client.Call("remove_config", host.Code, &relay, time.Second*30); err != nil {
			global.Logger.Error("send message fail", zap.Error(err))
			return err
		}
		if relay != "success" {
			global.Logger.Error("send message fail", zap.String("relay", relay))
		}
		return nil
	}, 5, time.Second*5)
	return nil
}

func (e *ARpcClientEngine) ForwardConfig(tx *query.Query, forwardCode string) error {
	forward, err := tx.GostClientForward.Preload(tx.GostClientForward.Node).Where(tx.GostClientForward.Code.Eq(forwardCode)).First()
	if err != nil {
		return err
	}
	warnMsg := warn_msg.GetForwardWarnMsg(*forward)
	if warnMsg != "" {
		_ = e.RemoveForward(tx, *forward)
		return errors.New(warnMsg)
	}

	auth, err := tx.GostAuth.Where(tx.GostAuth.TunnelCode.Eq(forwardCode)).First()
	if err != nil {
		return err
	}

	bwLimit, bwMode := forwardBandwidth(*forward)
	var data = ForwardConfig{
		Key:       forward.Code,
		UpdateTag: forward.UpdatedAt.Format(time.DateTime) + "#" + forward.Node.UpdatedAt.Format(time.DateTime),
		BaseCfg:   e.generateServerCommonCfg(forward.PoolCount, forward.Node, *auth),
		PortRange: forward.PortRange,
		TCP: TCPProxyConfig{
			TCPProxyConfig: v1.TCPProxyConfig{
				ProxyBaseConfig: v1.ProxyBaseConfig{
					Name: forward.Code + "_tcp",
					Type: "tcp",
					Metadatas: map[string]string{
						"user":     auth.User,
						"password": auth.Password,
					},
					ProxyBackend: v1.ProxyBackend{
						LocalIP:   forward.TargetIp,
						LocalPort: utils.StrMustInt(forward.TargetPort),
					},
					LoadBalancer: v1.LoadBalancerConfig{
						Group:    forward.LoadBalancerGroup,
						GroupKey: forward.LoadBalancerGroupKey,
					},
					HealthCheck: buildHealthCheck(forward.HealthCheckType, "", forward.HealthCheckInterval, forward.HealthCheckTimeout, forward.HealthCheckMaxFailed),
				},
				RemotePort: utils.StrMustInt(forward.Port),
			},
			Transport: ProxyTransport{
				UseEncryption:      forward.UseEncryption == 1,
				UseCompression:     forward.UseCompression == 1,
				BandwidthLimit:     bwLimit,
				BandwidthLimitMode: bwMode,
				ProxyProtocolVersion: func() string {
					switch forward.ProxyProtocol {
					case 1:
						return "v1"
					case 2:
						return "v2"
					default:
						return ""
					}
				}(),
			},
		},
		UDP: UDPProxyConfig{
			UDPProxyConfig: v1.UDPProxyConfig{
				ProxyBaseConfig: v1.ProxyBaseConfig{
					Name: forward.Code + "_udp",
					Type: "udp",
					Metadatas: map[string]string{
						"user":     auth.User,
						"password": auth.Password,
					},
					ProxyBackend: v1.ProxyBackend{
						LocalIP:   forward.TargetIp,
						LocalPort: utils.StrMustInt(forward.TargetPort),
					},
					LoadBalancer: v1.LoadBalancerConfig{
						Group:    forward.LoadBalancerGroup,
						GroupKey: forward.LoadBalancerGroupKey,
					},
					HealthCheck: buildHealthCheck(forward.HealthCheckType, "", forward.HealthCheckInterval, forward.HealthCheckTimeout, forward.HealthCheckMaxFailed),
				},
				RemotePort: utils.StrMustInt(forward.Port),
			},
			Transport: ProxyTransport{
				UseEncryption:      forward.UseEncryption == 1,
				UseCompression:     forward.UseCompression == 1,
				BandwidthLimit:     bwLimit,
				BandwidthLimitMode: bwMode,
				ProxyProtocolVersion: func() string {
					switch forward.ProxyProtocol {
					case 1:
						return "v1"
					case 2:
						return "v2"
					default:
						return ""
					}
				}(),
			},
		},
		TCPMux: func() TCPMuxProxyConfig {
			if forward.Multiplexer == "" {
				return TCPMuxProxyConfig{}
			}
			return TCPMuxProxyConfig{
				TCPMuxProxyConfig: v1.TCPMuxProxyConfig{
					ProxyBaseConfig: v1.ProxyBaseConfig{
						Name: forward.Code + "_tcpmux",
						Type: "tcpmux",
						Metadatas: map[string]string{
							"user":     auth.User,
							"password": auth.Password,
						},
						ProxyBackend: v1.ProxyBackend{
							LocalIP:   forward.TargetIp,
							LocalPort: utils.StrMustInt(forward.TargetPort),
						},
						LoadBalancer: v1.LoadBalancerConfig{
							Group:    forward.LoadBalancerGroup,
							GroupKey: forward.LoadBalancerGroupKey,
						},
						HealthCheck: buildHealthCheck(forward.HealthCheckType, "", forward.HealthCheckInterval, forward.HealthCheckTimeout, forward.HealthCheckMaxFailed),
					},
					Multiplexer:     forward.Multiplexer,
					RouteByHTTPUser: forward.RouteByHTTPUser,
				},
				Transport: ProxyTransport{
					UseEncryption:      forward.UseEncryption == 1,
					UseCompression:     forward.UseCompression == 1,
					BandwidthLimit:     bwLimit,
					BandwidthLimitMode: bwMode,
				},
			}
		}(),
	}

	go utils.Retry(func() error {
		var relay string
		if err := e.client.Call("forward_config", data, &relay, time.Second*30); err != nil {
			global.Logger.Error("send message fail", zap.Error(err))
			return err
		}
		if relay != "success" {
			global.Logger.Error("send message fail", zap.String("relay", relay))
		}
		return nil
	}, 5, time.Second*5)
	return nil
}

func (e *ARpcClientEngine) RemoveForward(tx *query.Query, forward model.GostClientForward) error {
	go utils.Retry(func() error {
		var relay string
		if err := e.client.Call("remove_config", forward.Code, &relay, time.Second*30); err != nil {
			global.Logger.Error("send message fail", zap.Error(err))
			return err
		}
		if relay != "success" {
			global.Logger.Error("send message fail", zap.String("relay", relay))
		}
		return nil
	}, 5, time.Second*5)
	return nil
}

func (e *ARpcClientEngine) TunnelConfig(tx *query.Query, tunnelCode string) error {
	tunnel, err := tx.GostClientTunnel.Preload(tx.GostClientTunnel.Node).Where(tx.GostClientTunnel.Code.Eq(tunnelCode)).First()
	if err != nil {
		return err
	}
	warnMsg := warn_msg.GetTunnelWarnMsg(*tunnel)
	if warnMsg != "" {
		_ = e.RemoveTunnel(tx, *tunnel, tunnel.Node)
		return errors.New(warnMsg)
	}
	auth, err := tx.GostAuth.Where(tx.GostAuth.TunnelCode.Eq(tunnelCode)).First()
	if err != nil {
		return err
	}
	var data = TunnelConfig{
		Key:       tunnel.Code,
		UpdateTag: tunnel.UpdatedAt.Format(time.DateTime) + "#" + tunnel.Node.UpdatedAt.Format(time.DateTime),
		BaseCfg:   e.generateServerCommonCfg(tunnel.PoolCount, tunnel.Node, *auth),
		STCP: STCPProxyConfig{
			STCPProxyConfig: v1.STCPProxyConfig{
				ProxyBaseConfig: v1.ProxyBaseConfig{
					Name: tunnel.Code + "_stcp",
					Type: "stcp",
					Metadatas: map[string]string{
						"user":     auth.User,
						"password": auth.Password,
					},
					ProxyBackend: v1.ProxyBackend{
						LocalIP:   tunnel.TargetIp,
						LocalPort: utils.StrMustInt(tunnel.TargetPort),
					},
				},
				Secretkey: tunnel.VKey + "_stcp",
			},
			Transport: ProxyTransport{
				UseEncryption:        tunnel.UseEncryption == 1,
				UseCompression:       tunnel.UseCompression == 1,
				BandwidthLimit:       fmt.Sprintf("%dKB", tunnel.Limiter*128),
				BandwidthLimitMode:   "client",
				ProxyProtocolVersion: "",
			},
		},
		SUDP: SUDPProxyConfig{
			SUDPProxyConfig: v1.SUDPProxyConfig{
				ProxyBaseConfig: v1.ProxyBaseConfig{
					Name: tunnel.Code + "_sudp",
					Type: "sudp",
					Metadatas: map[string]string{
						"user":     auth.User,
						"password": auth.Password,
					},
					ProxyBackend: v1.ProxyBackend{
						LocalIP:   tunnel.TargetIp,
						LocalPort: utils.StrMustInt(tunnel.TargetPort),
					},
				},
				Secretkey: tunnel.VKey + "_sudp",
			},
			Transport: ProxyTransport{
				UseEncryption:        tunnel.UseEncryption == 1,
				UseCompression:       tunnel.UseCompression == 1,
				BandwidthLimit:       fmt.Sprintf("%dKB", tunnel.Limiter*128),
				BandwidthLimitMode:   "client",
				ProxyProtocolVersion: "",
			},
		},
	}

	go utils.Retry(func() error {
		var relay string
		if err := e.client.Call("tunnel_config", data, &relay, time.Second*30); err != nil {
			global.Logger.Error("send message fail", zap.Error(err))
			return err
		}
		if relay != "success" {
			global.Logger.Error("send message fail", zap.String("relay", relay))
		}
		return nil
	}, 5, time.Second*5)
	return nil
}

func (e *ARpcClientEngine) RemoveTunnel(tx *query.Query, tunnel model.GostClientTunnel, node model.GostNode) error {
	go utils.Retry(func() error {
		var relay string
		if err := e.client.Call("remove_config", tunnel.Code, &relay, time.Second*30); err != nil {
			global.Logger.Error("send message fail", zap.Error(err))
			return err
		}
		if relay != "success" {
			global.Logger.Error("send message fail", zap.String("relay", relay))
		}
		return nil
	}, 5, time.Second*5)
	return nil
}

func (e *ARpcClientEngine) P2PConfig(tx *query.Query, p2pCode string) error {
	p2p, err := tx.GostClientP2P.Preload(tx.GostClientP2P.Node).Where(tx.GostClientP2P.Code.Eq(p2pCode)).First()
	if err != nil {
		return err
	}
	warnMsg := warn_msg.GetP2PWarnMsg(*p2p)
	if warnMsg != "" {
		_ = e.RemoveP2P(tx, *p2p)
		return errors.New(warnMsg)
	}

	auth, err := tx.GostAuth.Where(tx.GostAuth.TunnelCode.Eq(p2p.Code)).First()
	if auth == nil {
		return err
	}

	var data = P2PConfig{
		Key:       p2p.Code,
		UpdateTag: p2p.UpdatedAt.Format(time.DateTime) + "#" + p2p.Node.UpdatedAt.Format(time.DateTime),
		BaseCfg:   e.generateServerCommonCfg(p2p.PoolCount, p2p.Node, *auth),
		XTCP: XTCPProxyConfig{
			XTCPProxyConfig: v1.XTCPProxyConfig{
				ProxyBaseConfig: v1.ProxyBaseConfig{
					Name: p2p.Code + "_p2pxtcp",
					Type: "xtcp",
					Metadatas: map[string]string{
						"user":     auth.User,
						"password": auth.Password,
					},
					ProxyBackend: v1.ProxyBackend{
						LocalIP:   p2p.TargetIp,
						LocalPort: utils.StrMustInt(p2p.TargetPort),
					},
				},
				Secretkey: p2p.VKey,
			},
			Transport: ProxyTransport{
				UseEncryption:  p2p.UseEncryption == 1,
				UseCompression: p2p.UseCompression == 1,
				//BandwidthLimit:       fmt.Sprintf("%dKB", p2p.Limiter*128),
				ProxyProtocolVersion: "",
			},
		},
		STCP: STCPProxyConfig{
			STCPProxyConfig: v1.STCPProxyConfig{
				ProxyBaseConfig: v1.ProxyBaseConfig{
					Name: p2p.Code + "_p2pstcp",
					Type: "stcp",
					Metadatas: map[string]string{
						"user":     auth.User,
						"password": auth.Password,
					},
					ProxyBackend: v1.ProxyBackend{
						LocalIP:   p2p.TargetIp,
						LocalPort: utils.StrMustInt(p2p.TargetPort),
					},
				},
				Secretkey: p2p.VKey,
			},
			Transport: ProxyTransport{
				UseEncryption:        p2p.UseEncryption == 1,
				UseCompression:       p2p.UseCompression == 1,
				BandwidthLimit:       fmt.Sprintf("%dKB", p2p.Limiter*128),
				BandwidthLimitMode:   "client",
				ProxyProtocolVersion: "",
			},
		},
	}
	// 判断是否需要中继
	if p2p.Forward != 1 || p2p.Node.P2PDisableForward == 1 {
		data.STCP = STCPProxyConfig{}
	}

	go utils.Retry(func() error {
		var relay string
		if err := e.client.Call("p2p_config", data, &relay, time.Second*30); err != nil {
			global.Logger.Error("send message fail", zap.Error(err))
			return err
		}
		if relay != "success" {
			global.Logger.Error("send message fail", zap.String("relay", relay))
		}
		return nil
	}, 5, time.Second*5)
	return nil
}

func (e *ARpcClientEngine) RemoveP2P(tx *query.Query, p2p model.GostClientP2P) error {
	go utils.Retry(func() error {
		var relay string
		if err := e.client.Call("remove_config", p2p.Code, &relay, time.Second*30); err != nil {
			global.Logger.Error("send message fail", zap.Error(err))
			return err
		}
		if relay != "success" {
			global.Logger.Error("send message fail", zap.String("relay", relay))
		}
		return nil
	}, 5, time.Second*5)
	return nil
}

func (e *ARpcClientEngine) ProxyConfig(tx *query.Query, proxyCode string) error {
	proxy, err := tx.GostClientProxy.Preload(tx.GostClientProxy.Node).Where(tx.GostClientProxy.Code.Eq(proxyCode)).First()
	if err != nil {
		return err
	}
	warnMsg := warn_msg.GetProxyWarnMsg(*proxy)
	if warnMsg != "" {
		_ = e.RemoveProxy(tx, *proxy)
		return errors.New(warnMsg)
	}

	auth, err := tx.GostAuth.Where(tx.GostAuth.TunnelCode.Eq(proxyCode)).First()
	if err != nil {
		return err
	}

	var data = ProxyConfig{
		Key:      proxy.Code,
		BaseCfg:  e.generateServerCommonCfg(proxy.PoolCount, proxy.Node, *auth),
		Name:     proxy.Code + "_proxy",
		Port:     utils.StrMustInt(proxy.Port),
		AuthUser: proxy.AuthUser,
		AuthPwd:  proxy.AuthPwd,
		Metadata: map[string]string{
			"user":     auth.User,
			"password": auth.Password,
		},
		Limiter:        fmt.Sprintf("%dKB", proxy.Limiter*128),
		UseEncryption:  proxy.UseEncryption,
		UseCompression: proxy.UseCompression,
		UpdateTag:      proxy.UpdatedAt.Format(time.DateTime) + "#" + proxy.Node.UpdatedAt.Format(time.DateTime),
	}

	go utils.Retry(func() error {
		var relay string
		if err := e.client.Call("proxy_config", data, &relay, time.Second*30); err != nil {
			global.Logger.Error("send message fail", zap.Error(err))
			return err
		}
		if relay != "success" {
			global.Logger.Error("send message fail", zap.String("relay", relay))
		}
		return nil
	}, 5, time.Second*5)
	return nil
}

func (e *ARpcClientEngine) RemoveProxy(tx *query.Query, proxy model.GostClientProxy) error {
	go utils.Retry(func() error {
		var relay string
		if err := e.client.Call("remove_config", proxy.Code, &relay, time.Second*30); err != nil {
			global.Logger.Error("send message fail", zap.Error(err))
			return err
		}
		if relay != "success" {
			global.Logger.Error("send message fail", zap.String("relay", relay))
		}
		return nil
	}, 5, time.Second*5)
	return nil
}

func (e *ARpcClientEngine) generateServerCommonCfg(poolCount int, node model.GostNode, auth model.GostAuth) v1.ClientCommonConfig {
	serverAddr, serverPort := node.GetAddress()
	return v1.ClientCommonConfig{
		Auth: v1.AuthClientConfig{
			Token: node.Code,
		},
		ServerAddr: serverAddr,
		ServerPort: serverPort,
		Transport: v1.ClientTransportConfig{
			Protocol:  node.Protocol,
			PoolCount: poolCount,
		},
		Metadatas: map[string]string{
			"user":     auth.User,
			"password": auth.Password,
		},
		LoginFailExit: new(bool),
	}
}

func (e *ARpcClientEngine) CustomCfgConfig(tx *query.Query, cfgCode string) error {
	cfg, err := tx.FrpClientCfg.Where(tx.FrpClientCfg.Code.Eq(cfgCode)).First()
	if err != nil {
		return err
	}
	warnMsg := warn_msg.GetCfgWarnMsg(*cfg)
	if warnMsg != "" {
		_ = e.RemoveCustomCfg(tx, cfgCode)
		return errors.New(warnMsg)
	}

	var data = CustomCfgConfig{
		Key:       cfg.Code,
		UpdateTag: cfg.UpdatedAt.Format(time.DateTime),
		Type:      cfg.ContentType,
		Content:   cfg.Content,
	}
	go utils.Retry(func() error {
		var relay string
		if err := e.client.Call("custom_cfg_config", data, &relay, time.Second*30); err != nil {
			global.Logger.Error("send message fail", zap.Error(err))
			return err
		}
		if relay != "success" {
			global.Logger.Error("send message fail", zap.String("relay", relay))
		}
		return nil
	}, 5, time.Second*5)
	return nil
}

func (e *ARpcClientEngine) RemoveCustomCfg(tx *query.Query, cfgCode string) error {
	go utils.Retry(func() error {
		var relay string
		if err := e.client.Call("remove_config", cfgCode, &relay, time.Second*30); err != nil {
			global.Logger.Error("send message fail", zap.Error(err))
			return err
		}
		if relay != "success" {
			global.Logger.Error("send message fail", zap.String("relay", relay))
		}
		return nil
	}, 5, time.Second*5)
	return nil
}
