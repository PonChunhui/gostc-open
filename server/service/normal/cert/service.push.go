package service

import (
	"fmt"
	"strings"

	"server/global"
	"server/model"
	"server/repository/query"
	"server/service/engine"
)

// resolveCertNodeCodes 计算证书应推送的网关节点（含权限过滤）。
// 普通用户证书仅推自己隧道所在节点；管理员证书额外覆盖域名后缀匹配的平台节点。
func resolveCertNodeCodes(db *query.Query, cert *model.SystemCert, isAdmin bool) []string {
	set := map[string]bool{}
	own := func(userCode string) bool { return isAdmin || userCode == cert.UserCode }

	if hosts, err := db.GostClientHost.Where(db.GostClientHost.CustomDomain.Eq(cert.Domain)).Find(); err == nil {
		for _, h := range hosts {
			if own(h.UserCode) && h.NodeCode != "" {
				set[h.NodeCode] = true
			}
		}
	}
	if hosts, err := db.GostClientHost.Where(db.GostClientHost.CertCode.Eq(cert.Code)).Find(); err == nil {
		for _, h := range hosts {
			if own(h.UserCode) && h.NodeCode != "" {
				set[h.NodeCode] = true
			}
		}
	}
	if isAdmin {
		if nodes, err := db.GostNode.Find(); err == nil {
			for _, n := range nodes {
				if n.Domain == "" {
					continue
				}
				if cert.Domain == n.Domain || cert.Domain == "*."+n.Domain || strings.HasSuffix(cert.Domain, "."+n.Domain) {
					set[n.Code] = true
				}
			}
		}
	}
	codes := make([]string, 0, len(set))
	for c := range set {
		codes = append(codes, c)
	}
	return codes
}

// pushCert 将证书推送到所有关联网关（续期语义：保留 target/force-https），并返回推送摘要。
func pushCert(db *query.Query, cert *model.SystemCert, certPEM, keyPEM string) (string, error) {
	codes := resolveCertNodeCodes(db, cert, cert.UserCode == "1")
	for _, code := range codes {
		engine.NodeRenewDomain(db, code, cert.Domain, certPEM, keyPEM)
	}
	summary := fmt.Sprintf("推送网关 %d 个", len(codes))
	if len(codes) > 0 {
		summary += ": " + strings.Join(codes, ",")
	}
	// 兼容旧配置 acme.proxy-url（单网关直推）
	if global.Config.ACME.ProxyURL != "" {
		if err := pushToProxy(cert.Domain, certPEM, keyPEM); err != nil {
			return summary + "; proxy-url 失败", err
		}
		summary += "; proxy-url"
	}
	return summary, nil
}
