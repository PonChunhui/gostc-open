// Package dnsfollow 故障转移「入口跟随」：把域名 A 记录更新到备节点 IP（E1）。
// 复用 ACME 的 DNS 凭据（SystemCert）或全局 configs.DNS；支持 DNSPod/Cloudflare/阿里云。
package dnsfollow

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"server/global"
	"server/model"
	"server/pkg/crypto"
	"server/repository"
	"server/service/common/acme"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/cloudflare/cloudflare-go"
	"github.com/nrdcg/dnspod-go"
)

// ZoneOf 推导注册域（zone）：配置优先，否则取域名后两段。
func ZoneOf(domain, configured string) string {
	if configured != "" {
		return configured
	}
	parts := strings.Split(strings.Trim(domain, "."), ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return domain
}

// SubOf 返回相对 zone 的主机名（sub_domain / RR）。
func SubOf(domain, zone string) string {
	domain = strings.Trim(domain, ".")
	zone = strings.Trim(zone, ".")
	if domain == zone {
		return "@"
	}
	return strings.TrimSuffix(domain, "."+zone)
}

// zoneFor 推导域名所属注册域：顶层配置 zone 仅在该域名确属其下时生效，否则取后两段。
// 避免多域名（gostc.shop / linux66.cn）共用一个配置 zone 导致写错 zone。
func zoneFor(domain string) string {
	cfgZone := global.Config.DNS.Zone
	if cfgZone != "" && strings.HasSuffix(strings.Trim(domain, "."), cfgZone) {
		return cfgZone
	}
	return ZoneOf(domain, "")
}

// Resolve 解析某域名的 DNS 凭据：优先该域名的 SystemCert（ACME），
// 其次按注册域匹配 configs.DNS.Extra（多服务商），最后回退顶层单服务商配置。
func Resolve(domain string) (provider string, cred acme.Credential, ttl int, zone string, ok bool) {
	db, _, _ := repository.Get("")
	if cert, _ := db.SystemCert.Where(db.SystemCert.Domain.Eq(domain)).First(); cert != nil && cert.Provider != "" {
		raw := crypto.DecryptSafe(cert.CredentialEnc, global.Config.Encrypt.Key)
		var c acme.Credential
		if json.Unmarshal([]byte(raw), &c) == nil {
			return cert.Provider, c, global.Config.DNS.TTL, zoneFor(domain), true
		}
	}
	cfg := global.Config.DNS
	z := zoneFor(domain)
	for _, e := range cfg.Extra {
		if e.Zone == z && e.Provider != "" {
			var c acme.Credential
			_ = json.Unmarshal([]byte(e.Credential), &c)
			t := e.TTL
			if t <= 0 {
				t = cfg.TTL
			}
			return e.Provider, c, t, e.Zone, true
		}
	}
	if cfg.Provider == "" {
		return "", acme.Credential{}, 0, "", false
	}
	var c acme.Credential
	_ = json.Unmarshal([]byte(cfg.Credential), &c)
	return cfg.Provider, c, cfg.TTL, z, true
}

// UpdateA 把 domain 的 A 记录更新为 ip（不存在则创建）。
func UpdateA(provider string, cred acme.Credential, domain, ip string, ttl int, zone string) error {
	if domain == "" || ip == "" {
		return errors.New("domain/ip 为空")
	}
	if ttl <= 0 {
		ttl = 60
	}
	zone = ZoneOf(domain, zone)
	switch provider {
	case model.CERT_PROVIDER_CLOUDFLARE:
		return updateCloudflare(cred, domain, ip, ttl, zone)
	case model.CERT_PROVIDER_DNSPOD:
		return updateDNSPod(cred, domain, ip, ttl, zone)
	case model.CERT_PROVIDER_ALIDNS:
		return updateAlidns(cred, domain, ip, ttl, zone)
	}
	return errors.New("不支持的 DNS provider：" + provider)
}

func updateCloudflare(cred acme.Credential, domain, ip string, ttl int, zone string) error {
	var api *cloudflare.API
	var err error
	if cred.AuthToken != "" {
		api, err = cloudflare.NewWithAPIToken(cred.AuthToken)
	} else if cred.AuthEmail != "" && cred.AuthKey != "" {
		api, err = cloudflare.New(cred.AuthKey, cred.AuthEmail)
	} else {
		return errors.New("Cloudflare 凭据缺失")
	}
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	zoneID, err := api.ZoneIDByName(zone)
	if err != nil {
		return err
	}
	rc := cloudflare.ZoneIdentifier(zoneID)
	records, _, err := api.ListDNSRecords(ctx, rc, cloudflare.ListDNSRecordsParams{Name: domain, Type: "A"})
	if err != nil {
		return err
	}
	if len(records) > 0 {
		_, err = api.UpdateDNSRecord(ctx, rc, cloudflare.UpdateDNSRecordParams{
			ID: records[0].ID, Type: "A", Name: domain, Content: ip, TTL: ttl,
		})
		return err
	}
	_, err = api.CreateDNSRecord(ctx, rc, cloudflare.CreateDNSRecordParams{
		Type: "A", Name: domain, Content: ip, TTL: ttl,
	})
	return err
}

func updateDNSPod(cred acme.Credential, domain, ip string, ttl int, zone string) error {
	if cred.LoginToken == "" {
		return errors.New("DNSPod 凭据缺失（loginToken）")
	}
	client := dnspod.NewClient(dnspod.CommonParams{LoginToken: cred.LoginToken, Format: "json"})
	domains, _, err := client.Domains.List()
	if err != nil {
		return err
	}
	var domainID string
	for _, d := range domains {
		if d.Name == zone {
			domainID = d.ID.String()
			break
		}
	}
	if domainID == "" {
		return errors.New("DNSPod 未找到 zone：" + zone)
	}
	sub := SubOf(domain, zone)
	records, _, err := client.Records.List(domainID, sub)
	if err != nil {
		return err
	}
	attr := dnspod.Record{Name: sub, Type: "A", Value: ip, TTL: strconv.Itoa(ttl), Line: "默认"}
	for _, r := range records {
		if r.Type == "A" {
			_, _, err = client.Records.Update(domainID, r.ID, attr)
			return err
		}
	}
	_, _, err = client.Records.Create(domainID, attr)
	return err
}

func updateAlidns(cred acme.Credential, domain, ip string, ttl int, zone string) error {
	if cred.APIKey == "" || cred.SecretKey == "" {
		return errors.New("阿里云 DNS 凭据缺失（apiKey+secretKey）")
	}
	region := cred.RegionID
	if region == "" {
		region = "cn-hangzhou"
	}
	client, err := alidns.NewClientWithAccessKey(region, cred.APIKey, cred.SecretKey)
	if err != nil {
		return err
	}
	sub := SubOf(domain, zone)
	desc := alidns.CreateDescribeDomainRecordsRequest()
	desc.DomainName = zone
	desc.RRKeyWord = sub
	desc.TypeKeyWord = "A"
	resp, err := client.DescribeDomainRecords(desc)
	if err != nil {
		return err
	}
	if len(resp.DomainRecords.Record) > 0 {
		upd := alidns.CreateUpdateDomainRecordRequest()
		upd.RecordId = resp.DomainRecords.Record[0].RecordId
		upd.RR = sub
		upd.Type = "A"
		upd.Value = ip
		upd.TTL = requests.NewInteger(ttl)
		_, err = client.UpdateDomainRecord(upd)
		return err
	}
	add := alidns.CreateAddDomainRecordRequest()
	add.DomainName = zone
	add.RR = sub
	add.Type = "A"
	add.Value = ip
	add.TTL = requests.NewInteger(ttl)
	_, err = client.AddDomainRecord(add)
	return err
}
