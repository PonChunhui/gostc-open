package event

import (
	"fmt"
	service "github.com/SianHH/frp-package/package"
	"github.com/SianHH/frp-package/package/frpc"
	v1 "github.com/SianHH/frp-package/pkg/config/v1"
	"github.com/lesismal/arpc"
	"os"
	"strconv"
	"strings"
	"time"
)

func ForwardHandle(client *arpc.Client, callback func(key, updateTag string), checkUpdate func(key, updateTag string) bool) {
	client.Handler.Handle("forward_config", func(c *arpc.Context) {
		var req ForwardConfig
		if err := c.Bind(&req); err != nil {
			_ = c.Write(err.Error())
			return
		}

		if !checkUpdate(req.Key, req.UpdateTag) {
			_ = c.Write("success")
			return
		}

		var proxyCfgs []v1.ProxyConfigurer
		if req.TCP.Name != "" {
			proxyCfgs = append(proxyCfgs, req.TCP.To())
		}
		if req.UDP.Name != "" {
			proxyCfgs = append(proxyCfgs, req.UDP.To())
		}
		if req.TCPMux.Name != "" {
			proxyCfgs = append(proxyCfgs, req.TCPMux.To())
		}
		// 范围端口：展开为多个 TCP 代理（每个远端端口一个，名称唯一）
		if ports := parsePortRange(req.PortRange); len(ports) > 0 && req.TCP.Name != "" {
			for _, port := range ports {
				tcp := req.TCP
				tcp.TCPProxyConfig.RemotePort = port
				tcp.TCPProxyConfig.ProxyBaseConfig.Name = fmt.Sprintf("%s_range_%d", req.Key, port)
				proxyCfgs = append(proxyCfgs, tcp.To())
			}
		}
		service.Del(req.Key)
		time.Sleep(time.Second)
		req.BaseCfg.Transport.ProxyURL = os.Getenv("GOSTC_TRANSPORT_PROXY_URL")
		svc, err := frpc.NewService(req.BaseCfg, proxyCfgs, nil)
		if err != nil {
			_ = c.Write(err)
			return
		}
		if err := svc.Start(); err != nil {
			_ = c.Write(err.Error())
			return
		}
		if err := service.Set(req.Key, svc); err != nil {
			svc.Stop()
			_ = c.Write(err.Error())
			return
		}
		callback(req.Key, req.UpdateTag)
		_ = c.Write("success")
	})
}

// parsePortRange 解析 "8000-8010" 形式的范围端口；非法或超过 1000 个端口时返回 nil。
func parsePortRange(s string) []int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return nil
	}
	start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || start <= 0 || end < start || end-start > 1000 {
		return nil
	}
	ports := make([]int, 0, end-start+1)
	for p := start; p <= end; p++ {
		ports = append(ports, p)
	}
	return ports
}
