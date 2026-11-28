package task

import (
	"server/model"
	"server/repository"
	"server/service/common/alert"
	certservice "server/service/normal/cert"
	"time"

	"go.uber.org/zap"
)

// certRenew 每日自动续期：到期前 30 天内且开启自动续期的证书重新签发。
// 失败保留旧证书继续服务，并触发告警（CA 限流/凭据错误等）。
func certRenew() {
	db, _, log := repository.Get("")
	threshold := time.Now().AddDate(0, 0, 30).Unix()
	certs, err := db.SystemCert.Where(
		db.SystemCert.AutoRenew.Eq(1),
		db.SystemCert.ExpAt.Gt(0),
		db.SystemCert.ExpAt.Lt(threshold),
	).Find()
	if err != nil {
		log.Warn("证书续期扫描失败", zap.Error(err))
		return
	}
	for _, c := range certs {
		if err := certservice.IssueAndStore(c); err != nil {
			log.Warn("证书自动续期失败", zap.String("domain", c.Domain), zap.Error(err))
			alert.Dispatch(model.ALERT_EVENT_CERT_EXPIRE, model.ALERT_TARGET_ALL, c.Code, c.UserCode,
				"证书自动续期失败", "域名 "+c.Domain+" 续期失败："+err.Error()+"（旧证书继续服务）")
		}
	}
}
