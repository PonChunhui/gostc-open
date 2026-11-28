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

// AdvancedReq 端口转发的 frp 高级能力（复用器/HTTP用户路由/带宽/范围端口/负载均衡/健康检查）。
type AdvancedReq struct {
	Code                 string `binding:"required" json:"code"`
	Multiplexer          string `json:"multiplexer"`
	RouteByHTTPUser      string `json:"routeByHTTPUser"`
	BandwidthLimit       string `json:"bandwidthLimit"`
	BandwidthMode        string `json:"bandwidthMode"`
	PortRange            string `json:"portRange"`
	LoadBalancerGroup    string `json:"loadBalancerGroup"`
	LoadBalancerGroupKey string `json:"loadBalancerGroupKey"`
	HealthCheckType      string `json:"healthCheckType"`
	HealthCheckInterval  int    `json:"healthCheckInterval"`
	HealthCheckTimeout   int    `json:"healthCheckTimeout"`
	HealthCheckMaxFailed int    `json:"healthCheckMaxFailed"`
	BackupNodeCode       string `json:"backupNodeCode"`
	FailoverEnable       int    `json:"failoverEnable"`
}

func (service *service) Advanced(claims jwt.Claims, req AdvancedReq) error {
	db, _, log := repository.Get("")
	if err := tunnel_adv.ValidateMultiplexer(req.Multiplexer); err != nil {
		return err
	}
	if err := tunnel_adv.ValidateBandwidth(req.BandwidthLimit, req.BandwidthMode); err != nil {
		return err
	}
	if err := tunnel_adv.ValidatePortRange(req.PortRange); err != nil {
		return err
	}
	if err := tunnel_adv.ValidateHealthCheck(req.HealthCheckType, "", req.HealthCheckInterval, req.HealthCheckTimeout, req.HealthCheckMaxFailed); err != nil {
		return err
	}
	return db.Transaction(func(tx *query.Query) error {
		forward, _ := tx.GostClientForward.Where(
			tx.GostClientForward.UserCode.Eq(claims.Code),
			tx.GostClientForward.Code.Eq(req.Code),
		).First()
		if forward == nil {
			return errors.New("操作失败")
		}
		forward.Multiplexer = req.Multiplexer
		forward.RouteByHTTPUser = req.RouteByHTTPUser
		forward.BandwidthLimit = req.BandwidthLimit
		forward.BandwidthMode = req.BandwidthMode
		forward.PortRange = req.PortRange
		forward.LoadBalancerGroup = req.LoadBalancerGroup
		forward.LoadBalancerGroupKey = req.LoadBalancerGroupKey
		forward.HealthCheckType = req.HealthCheckType
		forward.HealthCheckInterval = req.HealthCheckInterval
		forward.HealthCheckTimeout = req.HealthCheckTimeout
		forward.HealthCheckMaxFailed = req.HealthCheckMaxFailed
		forward.BackupNodeCode = req.BackupNodeCode
		if req.FailoverEnable == 1 {
			forward.FailoverEnable = 1
		} else {
			forward.FailoverEnable = 2
		}
		if err := tx.GostClientForward.Save(forward); err != nil {
			log.Error("修改端口转发高级配置失败", zap.Error(err))
			return errors.New("操作失败")
		}
		engine.ClientForwardConfig(tx, forward.Code)
		return nil
	})
}
