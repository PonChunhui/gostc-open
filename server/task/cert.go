package task

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"server/global"
	"server/model"
	"server/pkg/crypto"
	"server/repository"
	"server/service/common/alert"
	"time"

	"go.uber.org/zap"
)

// certExpireScan 每日扫描自定义证书，临期（默认 30 天内）触发 cert_expire 告警。
func certExpireScan() {
	db, _, log := repository.Get("")
	hosts, err := db.GostClientHost.Where(db.GostClientHost.CustomCert.Neq("")).Find()
	if err != nil {
		log.Warn("证书到期扫描失败", zap.Error(err))
		return
	}
	threshold := time.Now().AddDate(0, 0, 30)
	for _, h := range hosts {
		certPEM := crypto.DecryptSafe(h.CustomCert, global.Config.Encrypt.Key)
		notAfter, ok := parseCertNotAfter(certPEM)
		if !ok {
			continue
		}
		if notAfter.Before(threshold) {
			days := int(time.Until(notAfter).Hours() / 24)
			alert.Dispatch(model.ALERT_EVENT_CERT_EXPIRE, model.ALERT_TARGET_TUNNEL, h.Code, h.UserCode,
				"证书即将到期告警", fmt.Sprintf("隧道 %s 的证书将于 %s 到期（剩余约 %d 天）", h.Name, notAfter.Format(time.DateOnly), days))
		}
	}
	// 证书管理（SystemCert）临期/过期扫描。
	certs, err := db.SystemCert.Where(db.SystemCert.ExpAt.Gt(0)).Find()
	if err != nil {
		log.Warn("证书管理到期扫描失败", zap.Error(err))
		return
	}
	now := time.Now()
	for _, c := range certs {
		exp := time.Unix(c.ExpAt, 0)
		if exp.Before(now) && c.Status != model.CERT_STATUS_EXPIRED {
			c.Status = model.CERT_STATUS_EXPIRED
			_ = db.SystemCert.Save(c)
		}
		if exp.Before(threshold) {
			days := int(time.Until(exp).Hours() / 24)
			alert.Dispatch(model.ALERT_EVENT_CERT_EXPIRE, model.ALERT_TARGET_ALL, c.Code, c.UserCode,
				"证书即将到期告警", fmt.Sprintf("证书 %s 将于 %s 到期（剩余约 %d 天）", c.Domain, exp.Format(time.DateOnly), days))
		}
	}
}

func parseCertNotAfter(certPEM string) (time.Time, bool) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return time.Time{}, false
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, false
	}
	return cert.NotAfter, true
}
