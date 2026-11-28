package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"server/global"
	"server/model"
	"server/pkg/crypto"
	"server/pkg/jwt"
	"server/repository"
	"server/service/common/acme"
)

type UploadReq struct {
	Code string `json:"code" binding:"required"`
	Cert string `json:"cert" binding:"required"`
	Key  string `json:"key" binding:"required"`
}

// Issue 手动触发签发。
func (s *service) Issue(claims jwt.Claims, code string) error {
	db, _, _ := repository.Get("")
	cert, err := db.SystemCert.Where(db.SystemCert.Code.Eq(code), db.SystemCert.UserCode.Eq(claims.Code)).First()
	if err != nil {
		return errors.New("证书不存在")
	}
	return IssueAndStore(cert)
}

// Renew 手动续期（与签发同流程）。
func (s *service) Renew(claims jwt.Claims, code string) error {
	return s.Issue(claims, code)
}

// Upload 上传自有证书。
func (s *service) Upload(claims jwt.Claims, req UploadReq) error {
	db, _, _ := repository.Get("")
	cert, err := db.SystemCert.Where(db.SystemCert.Code.Eq(req.Code), db.SystemCert.UserCode.Eq(claims.Code)).First()
	if err != nil {
		return errors.New("证书不存在")
	}
	exp, err := acme.ParseExpiry([]byte(req.Cert))
	if err != nil {
		return errors.New("证书解析失败：" + err.Error())
	}
	certPath, keyPath, err := writeCertFiles(cert.Domain, []byte(req.Cert), []byte(req.Key))
	if err != nil {
		return err
	}
	cert.CertPath, cert.KeyPath, cert.ExpAt = certPath, keyPath, exp.Unix()
	cert.Status, cert.LastError = model.CERT_STATUS_VALID, ""
	summary, perr := pushCert(db, cert, req.Cert, req.Key)
	cert.PushSummary = summary
	if perr != nil {
		cert.LastError = "证书已保存，但推送网关失败：" + perr.Error()
	}
	return db.SystemCert.Save(cert)
}

// IssueAndStore 签发并落盘、更新记录、推送网关。失败保留旧证书。供 service 与续期任务复用。
func IssueAndStore(cert *model.SystemCert) error {
	db, _, _ := repository.Get("")
	fail := func(err error) error {
		cert.Status = model.CERT_STATUS_FAILED
		cert.LastError = err.Error()
		_ = db.SystemCert.Save(cert)
		return err
	}
	plain := crypto.DecryptSafe(cert.CredentialEnc, global.Config.Encrypt.Key)
	var cred acme.Credential
	if err := json.Unmarshal([]byte(plain), &cred); err != nil {
		return fail(errors.New("DNS 凭据解析失败：" + err.Error()))
	}
	certPEM, keyPEM, exp, err := acme.Issue(cert.Domain, cert.Provider, cred, global.Config.ACME.Email, global.Config.ACME.CA)
	if err != nil {
		return fail(err)
	}
	certPath, keyPath, err := writeCertFiles(cert.Domain, certPEM, keyPEM)
	if err != nil {
		return fail(err)
	}
	cert.CertPath, cert.KeyPath, cert.ExpAt = certPath, keyPath, exp.Unix()
	cert.Status, cert.LastError = model.CERT_STATUS_VALID, ""
	if err := db.SystemCert.Save(cert); err != nil {
		return fail(err)
	}
	summary, perr := pushCert(db, cert, string(certPEM), string(keyPEM))
	cert.PushSummary = summary
	if perr != nil {
		// 推送失败不改变证书有效状态，仅记录
		cert.LastError = "证书已签发，但推送网关失败：" + perr.Error()
	}
	_ = db.SystemCert.Save(cert)
	return nil
}

func writeCertFiles(domain string, certPEM, keyPEM []byte) (certPath, keyPath string, err error) {
	dir := global.Config.ACME.CertDir
	if dir == "" {
		dir = global.BASE_PATH + "/data/certs"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", err
	}
	name := strings.ReplaceAll(domain, "*", "_wildcard")
	certPath = filepath.Join(dir, name+".pem")
	keyPath = filepath.Join(dir, name+".key")
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return "", "", err
	}
	return certPath, keyPath, nil
}

func pushToProxy(domain, cert, key string) error {
	base := global.Config.ACME.ProxyURL
	if base == "" {
		return nil
	}
	body, _ := json.Marshal(map[string]string{"domain": domain, "cert": cert, "key": key})
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(strings.TrimRight(base, "/")+"/domain", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("网关返回状态 " + resp.Status)
	}
	return nil
}
