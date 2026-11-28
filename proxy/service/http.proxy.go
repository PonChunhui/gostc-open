package service

import (
	"proxy/configs"
	"proxy/global"
	"proxy/pkg/proxy"
	"sync"
)

var httpProxyServiceLock = &sync.Mutex{}

func UpdateHttpProxy(c configs.Config) {
	httpProxyServiceLock.Lock()
	defer httpProxyServiceLock.Unlock()
	var proxyRouteList []proxy.LoadRouteConfig
	var certs []proxy.LoadCertConfig
	appendRoute := func(host string, r configs.DomainConfig) {
		proxyRouteList = append(proxyRouteList, proxy.LoadRouteConfig{
			Host:     host,
			Target:   r.Target,
			Rewrite:  true,
			Sni:      r.Sni,
			Origin:   r.Origin,
			AutoTLS:  r.ForceHttps && !r.DisableHttps,
			AutoCert: global.Config.AutoSetCert,
		})
		if !r.DisableHttps && r.Cert != "" && r.Key != "" {
			certs = append(certs, proxy.LoadCertConfig{
				Host:     host,
				CertFile: r.Cert,
				KeyFile:  r.Key,
			})
		}
	}
	for host, r := range c.Domains {
		appendRoute(host, r)
	}
	for ip, r := range c.IPRoutes {
		appendRoute(ip, r)
	}
	global.Proxy.LoadRoutes(proxyRouteList)
	global.Proxy.LoadCerts(certs)
	global.Proxy.SetDefaultRoute(proxy.LoadRouteConfig{
		Host:    "default",
		Target:  c.Default.Target,
		Rewrite: true,
		Sni:     c.Default.Sni,
		Origin:  c.Default.Origin,
		AutoTLS: c.Default.ForceHttps && !c.Default.DisableHttps,
	})
}
