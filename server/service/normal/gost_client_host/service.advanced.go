package service

import (
	"errors"
	"go.uber.org/zap"
	"server/pkg/jwt"
	"server/repository"
	"server/repository/query"
	"server/service/common/tunnel_adv"
	"server/service/engine"
)

// AdvancedReq 域名解析的 frp 高级能力（HTTP 路由/回源 Host/负载均衡/健康检查）。
type AdvancedReq struct {
	Code                 string   `binding:"required" json:"code"`
	Locations            []string `json:"locations"`
	HostHeaderRewrite    string   `json:"hostHeaderRewrite"`
	LoadBalancerGroup    string   `json:"loadBalancerGroup"`
	LoadBalancerGroupKey string   `json:"loadBalancerGroupKey"`
	HealthCheckType      string   `json:"healthCheckType"`
	HealthCheckPath      string   `json:"healthCheckPath"`
	HealthCheckInterval  int      `json:"healthCheckInterval"`
	HealthCheckTimeout   int      `json:"healthCheckTimeout"`
	HealthCheckMaxFailed int      `json:"healthCheckMaxFailed"`
	BackupNodeCode       string   `json:"backupNodeCode"`
	FailoverEnable       int      `json:"failoverEnable"`
}

func (service *service) Advanced(claims jwt.Claims, req AdvancedReq) error {
	db, _, log := repository.Get("")
	if err := tunnel_adv.ValidateHealthCheck(req.HealthCheckType, req.HealthCheckPath, req.HealthCheckInterval, req.HealthCheckTimeout, req.HealthCheckMaxFailed); err != nil {
		return err
	}
	return db.Transaction(func(tx *query.Query) error {
		host, _ := tx.GostClientHost.Where(
			tx.GostClientHost.UserCode.Eq(claims.Code),
			tx.GostClientHost.Code.Eq(req.Code),
		).First()
		if host == nil {
			return errors.New("操作失败")
		}
		host.SetLocations(req.Locations)
		host.HostHeaderRewrite = req.HostHeaderRewrite
		host.LoadBalancerGroup = req.LoadBalancerGroup
		host.LoadBalancerGroupKey = req.LoadBalancerGroupKey
		host.HealthCheckType = req.HealthCheckType
		host.HealthCheckPath = req.HealthCheckPath
		host.HealthCheckInterval = req.HealthCheckInterval
		host.HealthCheckTimeout = req.HealthCheckTimeout
		host.HealthCheckMaxFailed = req.HealthCheckMaxFailed
		host.BackupNodeCode = req.BackupNodeCode
		if req.FailoverEnable == 1 {
			host.FailoverEnable = 1
		} else {
			host.FailoverEnable = 2
		}
		if err := tx.GostClientHost.Save(host); err != nil {
			log.Error("修改域名解析高级配置失败", zap.Error(err))
			return errors.New("操作失败")
		}
		engine.ClientHostConfig(tx, host.Code)
		return nil
	})
}
