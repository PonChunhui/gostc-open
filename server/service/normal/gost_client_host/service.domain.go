package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"os"
	"server/model"
	"server/pkg/utils"
	"server/repository"
	"server/repository/cache"
	"server/repository/query"
	"server/service/engine"
	"strings"
)

type DomainReq struct {
	Code             string `binding:"required" json:"code"`
	CustomDomain     string `json:"customDomain"`
	CertCode         string `json:"certCode"` // 引用证书管理中的证书（不再手动上传）
	CustomForceHttps int    `json:"customForceHttps"`
	DomainMatcher    int    `json:"domainMatcher"`
}

// certCoversDomain 判断证书域名是否覆盖绑定域名（精确或通配）。
func certCoversDomain(certDomain, domain string) bool {
	if certDomain == domain {
		return true
	}
	if strings.HasPrefix(certDomain, "*.") {
		return strings.HasSuffix(domain, "."+certDomain[2:])
	}
	return false
}

func (service *service) Domain(userCode string, req DomainReq) error {
	db, _, log := repository.Get("")
	if req.CustomDomain != "" && !utils.ValidateDomain(req.CustomDomain) {
		return errors.New("域名格式错误")
	}

	return db.Transaction(func(tx *query.Query) error {
		user, _ := tx.SystemUser.Where(tx.SystemUser.Code.Eq(userCode)).First()
		if user == nil {
			return errors.New("用户错误")
		}

		host, _ := tx.GostClientHost.Preload(
			tx.GostClientHost.Node,
		).Where(
			tx.GostClientHost.UserCode.Eq(user.Code),
			tx.GostClientHost.Code.Eq(req.Code),
		).First()
		if host == nil {
			return errors.New("操作失败")
		}

		if !cache.GetNodeOnline(host.NodeCode) {
			return errors.New("节点已离线，无法操作")
		}

		if !cache.GetNodeCustomDomain(host.NodeCode) {
			return errors.New("该节点未启用自定义域名")
		}

		if host.CustomDomain != "" {
			if host.CustomDomain != req.CustomDomain {
				validDomain, _ := tx.GostClientHostDomain.Where(
					tx.GostClientHostDomain.Domain.Eq(req.CustomDomain),
				).First()
				if validDomain != nil {
					return errors.New("该域名已被使用")
				}
			}

			if _, err := tx.GostClientHostDomain.Where(tx.GostClientHostDomain.Domain.Eq(host.CustomDomain)).Delete(); err != nil {
				return errors.New("绑定失败")
			}
		}

		node, _ := tx.GostNode.Where(tx.GostNode.Code.Eq(host.NodeCode)).First()
		if node == nil {
			return errors.New("节点错误")
		}

		if node.AllowDomainMatcher != 1 && req.DomainMatcher == 1 {
			return errors.New("该隧道使用的节点不支持绑定泛域名")
		}

		if req.CustomDomain != "" {
			if isSubdomain(req.CustomDomain, node.Domain) {
				return errors.New("请使用你自己的域名")
			}

			if err := tx.GostClientHostDomain.Create(&model.GostClientHostDomain{
				Domain: req.CustomDomain,
			}); err != nil {
				log.Error("保存自定义域名失败", zap.Error(err))
				return errors.New("操作失败")
			}
		}

		oldDomain := host.GetCustomDomain()

		// 证书仅支持引用证书管理（CertCode），不再手动上传。
		certPEM, keyPEM := "", ""
		if req.CertCode != "" {
			cert, _ := tx.SystemCert.Where(tx.SystemCert.Code.Eq(req.CertCode), tx.SystemCert.UserCode.Eq(user.Code)).First()
			if cert == nil {
				return errors.New("证书不存在或无权引用")
			}
			if cert.Status != model.CERT_STATUS_VALID {
				return errors.New("引用的证书尚未签发或已失效")
			}
			cb, err := os.ReadFile(cert.CertPath)
			if err != nil {
				return errors.New("读取证书文件失败")
			}
			kb, err := os.ReadFile(cert.KeyPath)
			if err != nil {
				return errors.New("读取私钥文件失败")
			}
			certPEM, keyPEM = string(cb), string(kb)
			host.CertCode = req.CertCode
		} else {
			host.CertCode = ""
		}
		if req.CustomDomain != "" && certPEM != "" {
			cert, _ := tx.SystemCert.Where(tx.SystemCert.Code.Eq(req.CertCode)).First()
			if cert != nil && !certCoversDomain(cert.Domain, host.GetCustomDomain()) {
				return errors.New("证书域名与绑定域名不匹配")
			}
		}

		host.CustomDomain = req.CustomDomain
		host.CustomCert = ""
		host.CustomKey = ""
		host.CustomForceHttps = req.CustomForceHttps
		host.CustomDomainMatcher = req.DomainMatcher
		if err := tx.GostClientHost.Save(host); err != nil {
			log.Error("保存用户域名解析失败", zap.Error(err))
			return errors.New("操作失败")
		}
		newDomain := host.GetCustomDomain()
		if newDomain != "" {
			engine.NodeAddDomain(tx, host.NodeCode, newDomain, certPEM, keyPEM, host.CustomForceHttps)
		}
		// 解绑或换域名时清理源网关旧路由。
		if oldDomain != "" && oldDomain != newDomain {
			engine.NodeRemoveDomain(tx, host.NodeCode, oldDomain, false)
		}
		engine.NodeIngress(tx, host.NodeCode)
		engine.ClientHostConfig(tx, host.Code)
		return nil
	})
}

func verifyCertificateAndKey(cert, key string) error {
	_, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return fmt.Errorf("证书对验证失败: %v", err)
	}
	return nil
}

// 判断是否为子域名
func isSubdomain(target, base string) bool {
	target = strings.ToLower(strings.TrimSuffix(target, "."))
	base = strings.ToLower(strings.TrimSuffix(base, "."))
	if target == base {
		return true
	}
	return strings.HasSuffix(target, "."+base)
}
