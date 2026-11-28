// Package acme 封装 lego ACME DNS-01 签发（M4）。
package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"server/model"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/dnspod"
	"github.com/go-acme/lego/v4/registration"
)

// Credential 各 DNS provider 的凭据（JSON 序列化后加密存储）。
type Credential struct {
	LoginToken string `json:"loginToken"` // dnspod
	AuthToken  string `json:"authToken"`  // cloudflare
	ZoneToken  string `json:"zoneToken"`  // cloudflare（可选）
	AuthEmail  string `json:"authEmail"`  // cloudflare（key 模式）
	AuthKey    string `json:"authKey"`    // cloudflare（key 模式）
	APIKey     string `json:"apiKey"`     // alidns
	SecretKey  string `json:"secretKey"`  // alidns
	RegionID   string `json:"regionId"`   // alidns（可选）
}

type acmeUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *acmeUser) GetEmail() string                        { return u.Email }
func (u *acmeUser) GetRegistration() *registration.Resource { return u.Registration }
func (u *acmeUser) GetPrivateKey() crypto.PrivateKey        { return u.key }

// buildProvider 按 provider 构造 DNS-01 challenge provider。
func buildProvider(provider string, cred Credential) (challenge.Provider, error) {
	switch provider {
	case model.CERT_PROVIDER_DNSPOD:
		if cred.LoginToken == "" {
			return nil, errors.New("DNSPod 凭据缺失（loginToken）")
		}
		cfg := dnspod.NewDefaultConfig()
		cfg.LoginToken = cred.LoginToken
		return dnspod.NewDNSProviderConfig(cfg)
	case model.CERT_PROVIDER_CLOUDFLARE:
		cfg := cloudflare.NewDefaultConfig()
		cfg.AuthToken = cred.AuthToken
		cfg.ZoneToken = cred.ZoneToken
		cfg.AuthEmail = cred.AuthEmail
		cfg.AuthKey = cred.AuthKey
		if cfg.AuthToken == "" && (cfg.AuthEmail == "" || cfg.AuthKey == "") {
			return nil, errors.New("Cloudflare 凭据缺失（authToken 或 authEmail+authKey）")
		}
		return cloudflare.NewDNSProviderConfig(cfg)
	case model.CERT_PROVIDER_ALIDNS:
		if cred.APIKey == "" || cred.SecretKey == "" {
			return nil, errors.New("阿里云 DNS 凭据缺失（apiKey+secretKey）")
		}
		cfg := alidns.NewDefaultConfig()
		cfg.APIKey = cred.APIKey
		cfg.SecretKey = cred.SecretKey
		if cred.RegionID != "" {
			cfg.RegionID = cred.RegionID
		}
		return alidns.NewDNSProviderConfig(cfg)
	default:
		return nil, errors.New("不支持的 DNS provider：" + provider)
	}
}

// Issue 通过 DNS-01 签发证书，返回证书 PEM、私钥 PEM 与到期时间。
func Issue(domain, provider string, cred Credential, email, ca string) (certPEM, keyPEM []byte, exp time.Time, err error) {
	if domain == "" {
		return nil, nil, time.Time{}, errors.New("域名为空")
	}
	if email == "" {
		return nil, nil, time.Time{}, errors.New("ACME 邮箱未配置（acme.email）")
	}
	p, err := buildProvider(provider, cred)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	user := &acmeUser{Email: email, key: privateKey}
	config := lego.NewConfig(user)
	if ca != "" {
		config.CADirURL = ca
	}
	config.Certificate.KeyType = certcrypto.EC256
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	if err = client.Challenge.SetDNS01Provider(p); err != nil {
		return nil, nil, time.Time{}, err
	}
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	user.Registration = reg
	res, err := client.Certificate.Obtain(certificate.ObtainRequest{Domains: []string{domain}, Bundle: true})
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	exp, err = ParseExpiry(res.Certificate)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	return res.Certificate, res.PrivateKey, exp, nil
}

// ParseExpiry 解析证书 PEM 的到期时间。
func ParseExpiry(certPEM []byte) (time.Time, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return time.Time{}, errors.New("证书 PEM 解析失败")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}
	return cert.NotAfter, nil
}
