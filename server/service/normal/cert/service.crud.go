package service

import (
	"encoding/json"
	"errors"
	"server/global"
	"server/model"
	"server/pkg/bean"
	"server/pkg/crypto"
	"server/pkg/jwt"
	"server/repository"
	"server/service/common/acme"
	"server/service/engine"
	"time"
)

type CertReq struct {
	Code       string          `json:"code"`
	Domain     string          `json:"domain" binding:"required"`
	Provider   string          `json:"provider" binding:"required"`
	Credential acme.Credential `json:"credential"`
	AutoRenew  int             `json:"autoRenew"`
}

func validProvider(p string) bool {
	switch p {
	case model.CERT_PROVIDER_DNSPOD, model.CERT_PROVIDER_CLOUDFLARE, model.CERT_PROVIDER_ALIDNS:
		return true
	}
	return false
}

func encryptCredential(cred acme.Credential) (string, error) {
	raw, err := json.Marshal(cred)
	if err != nil {
		return "", err
	}
	return crypto.EncryptIfEnabled(string(raw), global.Config.Encrypt.Key)
}

func (s *service) Create(claims jwt.Claims, req CertReq) error {
	if !validProvider(req.Provider) {
		return errors.New("不支持的 DNS provider")
	}
	enc, err := encryptCredential(req.Credential)
	if err != nil {
		return err
	}
	autoRenew := req.AutoRenew
	if autoRenew == 0 {
		autoRenew = 1
	}
	db, _, _ := repository.Get("")
	return db.SystemCert.Create(&model.SystemCert{
		UserCode: claims.Code, Domain: req.Domain, Provider: req.Provider,
		CredentialEnc: enc, AutoRenew: autoRenew, Status: model.CERT_STATUS_PENDING,
	})
}

func (s *service) Update(claims jwt.Claims, req CertReq) error {
	if req.Code == "" {
		return errors.New("缺少证书编码")
	}
	if !validProvider(req.Provider) {
		return errors.New("不支持的 DNS provider")
	}
	db, _, _ := repository.Get("")
	cert, err := db.SystemCert.Where(db.SystemCert.Code.Eq(req.Code), db.SystemCert.UserCode.Eq(claims.Code)).First()
	if err != nil {
		return errors.New("证书不存在")
	}
	// 已签发证书不允许改域名（网关路由/引用会悬空），需删除重建。
	if cert.Status == model.CERT_STATUS_VALID && req.Domain != cert.Domain {
		return errors.New("已签发的证书不允许修改域名，请删除后重新创建")
	}
	enc, err := encryptCredential(req.Credential)
	if err != nil {
		return err
	}
	cert.Domain = req.Domain
	cert.Provider = req.Provider
	cert.CredentialEnc = enc
	cert.AutoRenew = req.AutoRenew
	return db.SystemCert.Save(cert)
}

func (s *service) Delete(claims jwt.Claims, code string) error {
	db, _, _ := repository.Get("")
	cert, err := db.SystemCert.Where(db.SystemCert.Code.Eq(code), db.SystemCert.UserCode.Eq(claims.Code)).First()
	if err != nil {
		return errors.New("证书不存在")
	}
	// 被域名解析引用时拒删，避免悬空 CertCode。
	if n, _ := db.GostClientHost.Where(db.GostClientHost.CertCode.Eq(code)).Count(); n > 0 {
		return errors.New("证书正被域名解析引用，请先在绑定域名中解除引用")
	}
	// 清理各网关上的该域名路由与证书文件。
	for _, nodeCode := range resolveCertNodeCodes(db, cert, claims.Code == "1") {
		engine.NodeRemoveDomain(db, nodeCode, cert.Domain, true)
	}
	_, err = db.SystemCert.Where(db.SystemCert.Code.Eq(code), db.SystemCert.UserCode.Eq(claims.Code)).Delete()
	return err
}

type CertPageReq struct {
	bean.PageParam
	Domain   string `json:"domain"`
	Provider string `json:"provider"`
	Status   int    `json:"status"`
}

type CertItem struct {
	Code      string `json:"code"`
	Domain    string `json:"domain"`
	Provider  string `json:"provider"`
	CertPath  string `json:"certPath"`
	KeyPath   string `json:"keyPath"`
	ExpAt     string `json:"expAt"`
	AutoRenew int    `json:"autoRenew"`
	Status      int    `json:"status"`
	LastError   string `json:"lastError"`
	PushSummary string `json:"pushSummary"`
	CreatedAt   string `json:"createdAt"`
}

func (s *service) Page(claims jwt.Claims, req CertPageReq) (list []CertItem, total int64) {
	db, _, _ := repository.Get("")
	q := db.SystemCert.Where(db.SystemCert.UserCode.Eq(claims.Code)).Order(db.SystemCert.Id.Desc())
	if req.Domain != "" {
		q = q.Where(db.SystemCert.Domain.Like("%" + req.Domain + "%"))
	}
	if req.Provider != "" {
		q = q.Where(db.SystemCert.Provider.Eq(req.Provider))
	}
	if req.Status > 0 {
		q = q.Where(db.SystemCert.Status.Eq(req.Status))
	}
	certs, total, _ := q.FindByPage(req.GetOffset(), req.GetLimit())
	for _, c := range certs {
		exp := ""
		if c.ExpAt > 0 {
			exp = time.Unix(c.ExpAt, 0).Format(time.DateTime)
		}
		list = append(list, CertItem{
			Code: c.Code, Domain: c.Domain, Provider: c.Provider, CertPath: c.CertPath, KeyPath: c.KeyPath,
			ExpAt: exp, AutoRenew: c.AutoRenew, Status: c.Status, LastError: c.LastError,
			PushSummary: c.PushSummary, CreatedAt: c.CreatedAt.Format(time.DateTime),
		})
	}
	return list, total
}
