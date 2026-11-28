package certpick

import (
	"os"
	"strings"

	"server/global"
	"server/model"
	"server/pkg/crypto"
	"server/repository/query"
)

// covers 判断证书域名是否覆盖目标域名（精确或通配）。
func covers(certDomain, domain string) bool {
	if certDomain == domain {
		return true
	}
	if strings.HasPrefix(certDomain, "*.") {
		// 通配证书不覆盖裸域（TLS 语义）。
		return strings.HasSuffix(domain, "."+certDomain[2:])
	}
	return false
}

func readCert(c *model.SystemCert) (string, string) {
	cb, err := os.ReadFile(c.CertPath)
	if err != nil {
		return "", ""
	}
	kb, err := os.ReadFile(c.KeyPath)
	if err != nil {
		return "", ""
	}
	return string(cb), string(kb)
}

// ForHost 为域名解析隧道挑选证书：①引用 CertCode ②同域名/通配 SystemCert ③存量手动快照 ④无。
// 返回绑定域名与证书/私钥 PEM（可能为空）。
func ForHost(db *query.Query, host *model.GostClientHost) (domain, certPEM, keyPEM string) {
	domain = host.GetCustomDomain()
	if domain == "" {
		return "", "", ""
	}
	cp, kp := ForDomain(db, host, domain)
	return domain, cp, kp
}

// ForDomain 为指定域名挑选证书（故障转移换后缀时复用）。
func ForDomain(db *query.Query, host *model.GostClientHost, domain string) (certPEM, keyPEM string) {
	if domain == "" {
		return "", ""
	}
	// ① 显式引用
	if host.CertCode != "" {
		if c, err := db.SystemCert.Where(db.SystemCert.Code.Eq(host.CertCode), db.SystemCert.Status.Eq(model.CERT_STATUS_VALID)).First(); err == nil && c != nil {
			if cp, kp := readCert(c); cp != "" {
				return cp, kp
			}
		}
	}
	// ② 同域名/通配覆盖（先属主后平台）
	for _, owner := range []string{host.UserCode, "1"} {
		if certs, err := db.SystemCert.Where(db.SystemCert.UserCode.Eq(owner), db.SystemCert.Status.Eq(model.CERT_STATUS_VALID)).Find(); err == nil {
			for _, c := range certs {
				if covers(c.Domain, domain) {
					if cp, kp := readCert(c); cp != "" {
						return cp, kp
					}
				}
			}
		}
	}
	// ③ 存量手动快照
	if host.CustomCert != "" {
		return host.CustomCert, crypto.DecryptSafe(host.CustomKey, global.Config.Encrypt.Key)
	}
	return "", ""
}
