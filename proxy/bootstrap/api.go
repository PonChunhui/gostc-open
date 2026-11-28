package bootstrap

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"proxy/configs"
	"proxy/global"
	"proxy/pkg/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func verifyCertificateAndKey(cert, key string) error {
	_, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return fmt.Errorf("证书对验证失败: %v", err)
	}
	return nil
}

func InitApi() {
	if global.Config.ApiAddr == "" {
		return
	}

	if global.MODE == "dev" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.Default()
	engine.Use(middleware.Logger(global.Logger, true, func(c *gin.Context) bool {
		return true
	}))

	engine.POST("/domain", Domain)
	engine.POST("/domain/delete", DomainDelete)

	svc := &http.Server{
		Addr:    global.Config.ApiAddr,
		Handler: engine,
	}

	var err error
	go func() {
		err = svc.ListenAndServe()
	}()
	time.Sleep(time.Second)
	if err != nil {
		global.Logger.Warn("api server listen on address: "+global.Config.ApiAddr, zap.Error(err))
		Release()
		os.Exit(1)
	}
	global.Logger.Info("api server listen on address: " + global.Config.ApiAddr)
}

type DomainReq struct {
	Domain     string `json:"domain"`
	Target     string `json:"target"`
	Cert       string `json:"cert"`
	Key        string `json:"key"`
	ForceHttps int    `json:"forceHttps"`
}

func Domain(c *gin.Context) {
	var req DomainReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(500, err.Error())
		return
	}
	if req.Domain == "" {
		return
	}

	var certFile = fmt.Sprintf("%s/data/certs/%s.pem", global.BASE_PATH, req.Domain)
	var keyFile = fmt.Sprintf("%s/data/certs/%s.key", global.BASE_PATH, req.Domain)
	if req.Cert != "" && req.Key != "" {
		if err := verifyCertificateAndKey(req.Cert, req.Key); err != nil {
			global.Logger.Warn("cert valid fail", zap.Error(err))
			return
		}
		_ = os.WriteFile(certFile, []byte(req.Cert), 0644)
		_ = os.WriteFile(keyFile, []byte(req.Key), 0644)
	} else {
		certFile = ""
		keyFile = ""
	}

	if global.Config.Domains == nil {
		global.Config.Domains = make(map[string]configs.DomainConfig)
	}
	// 证书续期推送可能仅带 cert/key 而不带 target：保留已有路由配置，避免清空转发目标。
	existing, hasExisting := global.Config.Domains[req.Domain]
	var ex *configs.DomainConfig
	if hasExisting {
		ex = &existing
	}
	global.Config.Domains[req.Domain] = mergeDomainConfig(ex, req, certFile, keyFile)
	if err := saveConfig(); err != nil {
		global.Logger.Error("save config fail", zap.Error(err))
	}
}

// mergeDomainConfig 合并路由配置：续期/仅证书推送保留已有 target 与 force-https；
// ForceHttps<0 为哨兵，表示显式保留已有 force-https（即使带了 target）。
func mergeDomainConfig(existing *configs.DomainConfig, req DomainReq, certFile, keyFile string) configs.DomainConfig {
	target := req.Target
	forceHttps := req.ForceHttps == 1
	if existing != nil {
		if target == "" {
			target = existing.Target
			forceHttps = existing.ForceHttps
		} else if req.ForceHttps < 0 {
			forceHttps = existing.ForceHttps
		}
		if certFile == "" {
			certFile = existing.Cert
			keyFile = existing.Key
		}
	}
	return configs.DomainConfig{
		Target:       target,
		Cert:         certFile,
		Key:          keyFile,
		ForceHttps:   forceHttps,
		DisableHttps: false,
	}
}

type DomainDeleteReq struct {
	Domain    string `json:"domain"`
	PurgeCert bool   `json:"purgeCert"`
}

// DomainDelete 移除域名路由（故障转移/解绑/删证书时清理源网关），可选删除证书文件。
func DomainDelete(c *gin.Context) {
	var req DomainDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(500, err.Error())
		return
	}
	if req.Domain == "" {
		return
	}
	if global.Config.Domains != nil {
		delete(global.Config.Domains, req.Domain)
	}
	if req.PurgeCert {
		_ = os.Remove(fmt.Sprintf("%s/data/certs/%s.pem", global.BASE_PATH, req.Domain))
		_ = os.Remove(fmt.Sprintf("%s/data/certs/%s.key", global.BASE_PATH, req.Domain))
	}
	if err := saveConfig(); err != nil {
		global.Logger.Error("save config fail", zap.Error(err))
	}
	c.String(200, "success")
}

func saveConfig() error {
	marshal, err := yaml.Marshal(global.Config)
	if err != nil {
		return err
	}
	return os.WriteFile(global.BASE_PATH+"/data/config.yaml", marshal, 0644)
}
